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


# 问题排查
## ⚠️ Config file does not end with a newline character ##
im*.cfg 配置文件必须有一个空白行(且只有一行) ⚠️⚠️

