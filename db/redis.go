package db

import (
	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedis() {
	opt, err := redis.ParseURL("redis://127.0.0.1:6379/0")
	if err != nil {
		panic(err)
	}

	Rdb = redis.NewClient(opt)
}
