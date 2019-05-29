/**
 * Copyright (c) 2014-2015, GoBelieve     
 * All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA  02111-1307  USA
 */

package main
import "time"
import "fmt"
import log "github.com/golang/glog"
import "github.com/gomodule/redigo/redis"

const ROOM_EVENT_ENTER = 1
const ROOM_EVENT_LEAVE = 2

type AppRoomMessage struct {
	appid int64
	timestamp int
	*RoomMessage
}

type RoomEvent struct {
	appid int64
	uid int64
	room_id int64
	session_id string
	event int
	timestamp int
}


type RoomMessageDeliver struct {
	wt chan *AppRoomMessage
	event_chan chan *RoomEvent
}

func NewRoomMessageDeliver() *RoomMessageDeliver {
	usd := &RoomMessageDeliver{}
	usd.wt = make(chan *AppRoomMessage, 10000)
	usd.event_chan = make(chan *RoomEvent, 10000)
	return usd
}

func (usd *RoomMessageDeliver) PublishEvent(appid int64, uid int64, room_id int64, session_id string, event int) bool {
	now := int(time.Now().Unix())
	e := &RoomEvent{appid, uid, room_id, session_id, event, now}
	
	begin := time.Now()
	select {
	case usd.event_chan <- e:
		end := time.Now()
		duration := end.Sub(begin)
		if duration > time.Millisecond*500 {
			log.Infof("publish room event slow:%d, %d, %d", appid, uid)
		}
		return true
	case <- time.After(60*time.Second):
		log.Infof("publish room event timed out:%d, %d", appid, uid)
		return false
	}
}

func (usd *RoomMessageDeliver) SaveRoomMessage(appid int64, msg *RoomMessage) bool {
	now := int(time.Now().Unix())	
	m := &AppRoomMessage{appid, now, msg}

	begin := time.Now()
	
	select {
	case usd.wt <- m:
		end := time.Now()
		duration := end.Sub(begin)
		if duration > time.Millisecond*500 {
			log.Infof("save room message to wt slow:%d, %d, %d", appid, msg.sender, msg.receiver)
		}
		return true
	case <- time.After(60*time.Second):
		log.Infof("save room message to wt timed out:%d, %d, %d", appid, msg.sender, msg.receiver)
		return false
	}
}


func (usd *RoomMessageDeliver) deliverEvents(events []*RoomEvent) {
	conn := redis_pool.Get()
	defer conn.Close()
	
	begin := time.Now()	
	conn.Send("MULTI")	
	for _, event := range(events) {
		content := fmt.Sprintf("%d\n%d\n%d\n%d\n%d\n%s",
			event.appid, event.uid, event.room_id, event.event, event.timestamp, event.session_id)
		queue_name := "room_events"
		conn.Send("LPUSH", queue_name, content)
	}
	_, err := redis.Values(conn.Do("EXEC"))
	
	end := time.Now()
	duration := end.Sub(begin)
	if err != nil {
		log.Info("multi lpush error:", err)
		return
	}
	log.Infof("event mmulti lpush:%d time:%s success", len(events), duration)
}


func (usd *RoomMessageDeliver) deliver(messages []*AppRoomMessage) {
	conn := redis_pool.Get()
	defer conn.Close()
	
	begin := time.Now()	
	conn.Send("MULTI")	
	for _, msg := range(messages) {
		content := fmt.Sprintf("%d\n%d\n%d\n%s", msg.sender, msg.receiver,
			msg.timestamp, msg.content)
		queue_name := fmt.Sprintf("rooms_%d_%d", msg.appid, msg.receiver)
		conn.Send("LPUSH", queue_name, content)
	}
	res, err := redis.Values(conn.Do("EXEC"))
	
	end := time.Now()
	duration := end.Sub(begin)
	if err != nil {
		log.Info("multi lpush error:", err)
		return
	}
	log.Infof("mmulti lpush:%d time:%s success", len(messages), duration)


	rooms := make(map[string]struct{})
	for index, v := range res {
		count, ok := v.(int64)

		if !ok {
			continue
		}

		//*2 for reduce call ltrim times
		if count <= int64(config.room_message_limit*2) {
			continue
		}
		
		if index >= len(messages) {
			log.Error("index out of bound")
			continue
		}
		msg := messages[index]
		queue_name := fmt.Sprintf("rooms_%d_%d", msg.appid, msg.receiver)
		rooms[queue_name] = struct{}{}
	}

	if len(rooms) == 0 {
		return
	}
	
	conn.Send("MULTI")
	for queue_name, _ := range rooms {
		conn.Send("LTRIM", queue_name, 0, config.room_message_limit - 1)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		log.Warning("ltrim room list err:", err)
	}
}

func (usd *RoomMessageDeliver) run() {
	messages := make([]*AppRoomMessage, 0, 10000)	
	for {
		messages = messages[:0]
		
		m := <- usd.wt
		messages = append(messages, m)

	Loop:
		for {
			select {
			case m = <- usd.wt:
				messages = append(messages, m)
			default:
				break Loop
			}
		}
		usd.deliver(messages)
	}
}


func (usd *RoomMessageDeliver) runEvent() {
	events := make([]*RoomEvent, 0, 100)
	for {
		events = events[:0]
		
		e := <- usd.event_chan
		events = append(events, e)

	Loop:
		for {
			select {
			case e = <- usd.event_chan:
				events = append(events, e)
			default:
				break Loop
			}
		}
		usd.deliverEvents(events)
	}
}


func (usd *RoomMessageDeliver) Start() {
	go usd.run()
	go usd.runEvent()
}
