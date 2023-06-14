package main

import (
	"github.com/sk-pkg/cache"
	"github.com/sk-pkg/cache/redis"
	"log"
)

func main() {
	cfg := redis.Config{
		Address: "192.168.10.10:6379",
		Prefix:  "test",
	}
	// 内存缓存
	// c, err := cache.New()

	// redis缓存
	c, err := cache.New(cache.WithDriver("redis"), cache.WithRedisConfig(cfg))
	if err != nil {
		log.Fatal(err)
	}

	err = c.Put("a", "a", 1)
	if err != nil {
		log.Fatal(err)
	}

	a, err := c.Get("a")
	if a == nil || a.(string) != "a" {
		log.Fatal("Got a failed value from mem cache:", a)
	}
}
