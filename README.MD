# Cache

Cache缓存，支持内存、Redis作为缓存驱动

## 安装

```shell
go get "github.com/sk-pkg/cache"
```

## 快速开始
```go
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
```
## 可配置项
1.Driver 缓存驱动，默认为内存`mem`驱动，此项为可选。当驱动为`redis`时，需要传入redis配置（WithRedisConfig()）或redisManager（WithRedis()）
```go
cache.WithDriver("redis")
```
2.Prefix缓存前缀，默认为`go_cache`，此项可选
```go
cache.WithPrefix("go_cache")
```
3.Redis指向`*redis.Manager`[redisManager](https://github.com/sk-pkg/redis/blob/main/redis.go) ，此项可选
```go
cache.WithRedis(redis *redis.Manager)
```
4.RedisConfig redis连接配置，此项可选
```go
cache.WithRedisConfig(Config{
    Address: "192.168.10.10:6379",
    Password: "auth",
    Prefix:  "test",
})
```
## 可用方法
1.Put 缓存中存储数据
```go
Put(key string, value interface{}, seconds int) error
```
2.Add 方法将只存储缓存中不存在的数据。
```go
Add(key string, value interface{}, seconds int) error
```
3.Get 方法用于从缓存中获取数据，如果该数据在缓存中不存在，那么该方法将返回 nil, nil
```go
Get(key string) (interface{}, error)
```
4.Pull 方法用于从缓存中获取到数据之后再删除它，如果该数据在缓存中不存在，那么该方法将返回 nil, nil
```go
Pull(key string) (interface{}, error)
```
5.Has 方法可以用于判断缓存项是否存在。如果值为 nil，则该方法将会返回 false
```go
Has(key string) bool
```
6.Forever 方法可用于持久化将数据存储到缓存中，因为这些数据不会过期，所以必须通过 Forget 方法从缓存中手动删除它们
```go
Forever(key string, value interface{}) error
```
7.Forget 方法从缓存中删除这些数据
```go
Forget(key string) (bool, error)
```
8.Increment 方法增加指定key的值 int 值
```go
Increment(key string, n int) (int, error)
```
9.Decrement 方法减少指定key的值 int 值
```go
Decrement(key string, n int) (int, error)
```
10.Flush 方法清空所有的缓存
```go
Flush() error
```
## 注意
在使用`redis`作为驱动时，会出现一些报错`panic: interface conversion: interface {} is float64, not int` 

原因：使用 Golang 解析 JSON 格式数据时，若以 interface{} 接收数据，则会按照下列规则进行解析：
```go
bool, for JSON booleans

float64, for JSON numbers

string, for JSON strings

[]interface{}, for JSON arrays

map[string]interface{}, for JSON objects

nil for JSON null
```