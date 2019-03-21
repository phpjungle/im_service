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

import log "github.com/golang/glog"
import "unsafe"
import "sync/atomic"
import "github.com/bitly/go-simplejson"
import "github.com/satori/go.uuid"

type RoomClient struct {
	*Connection
	room_id int64
	session_id string
}

func (client *RoomClient) Logout() {
	if client.room_id > 0 {
		channel := GetRoomChannel(client.room_id)
		channel.UnsubscribeRoom(client.appid, client.room_id)

		route := app_route.FindOrAddRoute(client.appid)
		route.RemoveRoomClient(client.room_id, client.Client())

		deliver := GetRoomMessageDeliver(client.room_id)
		deliver.PublishEvent(client.appid, client.uid, client.room_id,
			client.session_id, ROOM_EVENT_LEAVE)
	}
}

func (client *RoomClient) HandleMessage(msg *Message) {
	switch msg.cmd {
	case MSG_ENTER_ROOM:
		client.HandleEnterRoom(msg.body.(*Room))
	case MSG_LEAVE_ROOM:
		client.HandleLeaveRoom(msg.body.(*Room))
	case MSG_ROOM_IM:
		client.HandleRoomIM(msg)
	}
}

func (client *RoomClient) HandleEnterRoom(room *Room){
	if client.uid == 0 {
		log.Warning("client has't been authenticated")
		return
	}

	room_id := room.RoomID()
	log.Info("enter room id:", room_id)
	if room_id == 0 || client.room_id == room_id {
		return
	}
	route := app_route.FindOrAddRoute(client.appid)
	if client.room_id > 0 {
		channel := GetRoomChannel(client.room_id)
		channel.UnsubscribeRoom(client.appid, client.room_id)

		route.RemoveRoomClient(client.room_id, client.Client())
	}

	client.room_id = room_id
	route.AddRoomClient(client.room_id, client.Client())
	channel := GetRoomChannel(client.room_id)
	channel.SubscribeRoom(client.appid, client.room_id)

	u1, err := uuid.NewV4()
	if err != nil {
		client.session_id = ""
		log.Warning("generate uuid err:", err)
	} else {
		client.session_id = u1.String()
	}

	deliver := GetRoomMessageDeliver(client.room_id)
	deliver.PublishEvent(client.appid, client.uid, client.room_id,
		client.session_id, ROOM_EVENT_ENTER)
}

func (client *RoomClient) Client() *Client {
	p := unsafe.Pointer(client.Connection)
	return (*Client)(p)
}

func (client *RoomClient) HandleLeaveRoom(room *Room) {
	if client.uid == 0 {
		log.Warning("client has't been authenticated")
		return
	}

	room_id := room.RoomID()
	log.Info("leave room id:", room_id)
	if room_id == 0 {
		return
	}
	if client.room_id != room_id {
		return
	}

	route := app_route.FindOrAddRoute(client.appid)
	route.RemoveRoomClient(client.room_id, client.Client())
	channel := GetRoomChannel(client.room_id)
	channel.UnsubscribeRoom(client.appid, client.room_id)
	deliver := GetRoomMessageDeliver(client.room_id)
	deliver.PublishEvent(client.appid, client.uid, client.room_id, client.session_id, ROOM_EVENT_LEAVE)
	client.room_id = 0
	client.session_id = ""
}

func (client *RoomClient) HandleRoomIM(msg *Message) {
	room_im := msg.body.(*RoomMessage)
	seq := msg.seq
	if client.uid == 0 {
		log.Warning("client has't been authenticated")
		return
	}
	room_id := room_im.receiver
	if room_id != client.room_id {
		log.Warningf("room id:%d is't client's room id:%d\n", room_id, client.room_id)
		return
	}

	fb := atomic.LoadInt32(&client.forbidden) 
	if (fb == 1) {
		log.Infof("room id:%d client:%d, %d is forbidden", room_id, client.appid, client.uid)
		return
	}

	//todo 删除如下代码，临时补丁
	obj, err := simplejson.NewJson([]byte(room_im.content))
	if err == nil {
		elem := obj.Get("custom_elem")
		if elem != nil {
			action, err := elem.Get("user_action").Int64()
			if err == nil && action == 1003 {
				log.Warningf("invalid room message, room id:%d sender:%d content:%s", room_id, room_im.sender, room_im.content)
				return
			}
		}
		elem = obj.Get("user_action")
		if elem != nil {
			action, err := obj.Get("user_action").Int64()
			if err == nil && action == 1003 {
				log.Warningf("invalid room message, room id:%d sender:%d content:%s", room_id, room_im.sender, room_im.content)
				return
			}
		}
	}
	
	if (msg.flag & MESSAGE_FLAG_UNPERSISTENT) == 0 {
		deliver := GetRoomMessageDeliver(room_im.receiver)
		deliver.SaveRoomMessage(client.appid, room_im)
	}
	
	m := &Message{cmd:MSG_ROOM_IM, body:room_im}
	route := app_route.FindOrAddRoute(client.appid)
	clients := route.FindRoomClientSet(room_id)
	for c, _ := range(clients) {
		if c == client.Client() {
			continue
		}
		c.EnqueueNonBlockMessage(m)
	}

	amsg := &AppMessage{appid:client.appid, receiver:room_id, msg:m}
	channel := GetRoomChannel(client.room_id)
	channel.PublishRoom(amsg)

	client.wt <- &Message{cmd: MSG_ACK, body: &MessageACK{int32(seq)}}
}
