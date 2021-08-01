# 命令行设置go env -w
```sh    
    go env -w GOBIN=/Users/phpjungle/go/bin
    go env -w GOPROXY=https://goproxy.cn,direct
```

# 依赖 (VPN 支持)
```sh
    dep ensure

    go get github.com/valyala/gorpc
    go get github.com/richmonkey/cfg
    go get github.com/importcjj/sensitive
    go get github.com/gorilla/websocket
    go get github.com/googollee/go-engine.io
    go get github.com/gomodule/redigo/redis
    go get github.com/golang/glog
    go get github.com/go-sql-driver/mysql
    go get github.com/bitly/go-simplejson
```

# 编译安装
```
    mkdir bin
    make 
    make install
```

# SQL
```sql
mysql -e "source db.sql" gobelieve -uroot -hlocalhost

-- 创建账号
CREATE USER 'im'@'localhost' IDENTIFIED WITH mysql_native_password BY '123456';

GRANT ALL PRIVILEGES ON gobelieve.* TO'im'@'localhost';

FLUSH PRIVILEGES;

```

# 目录结构
```sh
/Users/phpjungle/im
.
├── data
│   └── pending
└── logs
    ├── im
    ├── imr
    └── ims
```
# access_token_$TOKEN: token 生成
```
    连接im服务器token存储在redis的hash对象中,脱离API服务器测试时，可以手工生成。
    $token就是客户端需要获得的, 用来连接im服务器的认证信息。
    key:access_token_$token
    field:
        user_id:用户id
        app_id:应用id
```

# token.php
```php
    <?php
    const APP_ID = 7;
    $data = file_get_contents("php://input");

    $data = json_decode($data, true);

    // {"data":{"token":"f64fdae9aa2e536e36becef55850b01d","cache_token":true}}
    if (isset($data['uid'])) {
        $chat_id = $data['uid'];
        $token = md5(sprintf("%s_%s", APP_ID, $chat_id));
        $resp = ['data' => ['token' => $token, 'cache_token' => set_access_token($chat_id)]];

        echo json_encode($resp);
    } else {
        $resp = ['data' => ['token' => '']];
        echo json_encode($resp);
    }

    function set_access_token($chat_id) {
        if ($chat_id) {
            $host = 'localhost';
            $redis = new PJRedis($host, 6379, null);
            $info = $redis->info();

            // var_dump($info);

            $redis->select(0);
            $key = sprintf("access_token_%s", md5(sprintf("%s_%s", APP_ID, $chat_id)));

            $stat = $redis->hash_sets($key, ['app_id' => APP_ID, 'user_id' => $chat_id]);

            return $stat;
        }
    }
```
# 问题排查
## ⚠️ Config file does not end with a newline character ##
im*.cfg 配置文件必须有一个空白行(且只有一行) ⚠️⚠️

