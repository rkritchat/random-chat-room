package config

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Cfg struct {
	Rdb *redis.Client
}

func InitConfig() *Cfg {
	return &Cfg{
		Rdb: initRDB(),
	}
}

func initRDB() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})

	pong := rdb.Ping(context.Background())
	if pong.Val() != "PONG" {
		panic("cannot connect redis")
	}
	return rdb
}

func (c *Cfg) Free() {
	if c.Rdb != nil {
		_ = c.Rdb.Close()
	}
}
