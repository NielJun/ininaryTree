package main

import (
	"fmt"
	"github.com/go-redis/redis"
)

var redisDb *redis.Client

func initRedis() (err error) {
	redisDb = redis.NewClient(&redis.Options{
		Addr:     "175.24.15.152:6379",
		Password: "meimima1234.",
		DB:       0,
	})

	_,err = redisDb.Ping().Result()
	return

}
func main() {

	err:= initRedis()
	if err != nil {
		fmt.Printf("链接redis服务器失败,%v",err)
		return
	}
	fmt.Printf("链接redis服务器成功")

}
