# sk-pkg/cache - 高性能缓存库

这是一个灵活的 Go 语言缓存系统，支持多种缓存驱动（内存和 Redis），提供统一的接口，使得在不同缓存后端之间切换变得无缝。

## 安装

```bash
go get "github.com/sk-pkg/cache"
```

## 特性

- 支持内存缓存和 Redis 缓存
- 统一的 API 接口
- 内存缓存采用分片设计，减少锁竞争，提高并发性能
- 支持键值过期时间
- 支持原子递增和递减操作
- 支持批量清除操作

## 快速开始

### 内存缓存

```go
package main

import (
    "github.com/sk-pkg/cache"
    "log"
    "fmt"
)

func main() {
    // 创建内存缓存（默认驱动）
    c, err := cache.New()
    if err != nil {
        log.Fatal(err)
    }
    
    // 存储数据（有效期1秒）
    err = c.Put("user:1", "张三", 1)
    if err != nil {
        log.Fatal(err)
    }
    
    // 获取数据
    value, err := c.Get("user:1")
    if err != nil {
        log.Fatal(err)
    }
    
    if value != nil {
        fmt.Println("用户:", value.(string))
    }
}
```

### Redis 缓存

```go
package main

import (
    "github.com/sk-pkg/cache"
    "github.com/sk-pkg/cache/redis"
    "log"
    "fmt"
)

func main() {
    // Redis 配置
    cfg := redis.Config{
        Address:  "localhost:6379",  // Redis 服务器地址
        Password: "password",        // Redis 密码（可选）
        Prefix:   "myapp",           // 全局键前缀（可选）
    }
    
    // 创建 Redis 缓存
    c, err := cache.New(
        cache.WithDefaultDriver("redis"), 
        cache.WithRedisConfig(cfg),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 存储数据（有效期60秒）
    err = c.Put("user:1", "张三", 60)
    if err != nil {
        log.Fatal(err)
    }
    
    // 获取数据
    value, err := c.Get("user:1")
    if err != nil {
        log.Fatal(err)
    }
    
    if value != nil {
        fmt.Println("用户:", value.(string))
    }
}
```

## 高级用法

### 缓存管理器配置选项

```go
// 使用自定义前缀
c, err := cache.New(cache.WithPrefix("myapp"))

// 使用现有的 Redis 管理器
redisManager := redis.New(...)
c, err := cache.New(
    cache.WithDefaultDriver("redis"),
    cache.WithRedis(redisManager),
)
```

### 在缓存管理器中切换驱动

除了在初始化时指定默认驱动外，您还可以在运行时灵活地切换缓存驱动。缓存管理器提供了直接访问不同驱动实例的能力。

```go
// 创建缓存管理器
cacheManager, err := cache.New(
    cache.WithDefaultDriver("redis"),
    cache.WithRedisConfig(redisConfig),
)

// 使用内存缓存
err := cacheManager.Mem.Put("key", "value", 60)

// 使用Redis缓存
err := cacheManager.Redis.Put("key", "value", 60)
```

#### 使用场景

这种方式在以下场景特别有用：

1. **混合缓存策略**：
   - 将频繁访问的数据存储在内存缓存中
   - 将需要持久化的数据存储在Redis中

```go
// 频繁访问的配置存储在内存中
cacheManager.Mem.Put("app:config", configData, 300)

// 用户数据存储在Redis中以便跨服务共享
cacheManager.Redis.Put("user:1", userData, 3600)
```

2. **缓存降级**：
   - 当Redis不可用时，自动降级到内存缓存

```go
func getData(key string, cache *cache.Manager) (any, error) {
    // 尝试从Redis获取
    if cache.Redis != nil {
        value, err := cache.Redis.Get(key)
        if err == nil && value != nil {
            return value, nil
        }
    }
    
    // Redis失败或不可用，从内存缓存获取
    return cache.Mem.Get(key)
}
```

3. **数据分层**：
   - 根据数据特性选择不同的缓存驱动

```go
// 临时会话数据存储在内存中
cacheManager.Mem.Put("session:"+sessionID, sessionData, 1800)

// 共享计数器存储在Redis中以支持原子操作
newCount, _ := cacheManager.Redis.Increment("visits", 1)
```

### 永久存储数据

```go
// 永久存储数据（直到手动删除）
err := c.Forever("app:config", configData)
```

### 条件存储

```go
// 仅当键不存在时存储数据
err := c.Add("user:1", userData, 3600)
```

### 获取并删除

```go
// 获取数据并从缓存中删除
value, err := c.Pull("user:1")
```

### 检查键是否存在

```go
// 检查键是否存在
if c.Has("user:1") {
    // 键存在
}
```

### 递增/递减计数器

```go
// 递增计数器
newValue, err := c.Increment("visits", 1)

// 递减计数器
newValue, err := c.Decrement("remaining", 1)
```

### 清空缓存

```go
// 清空所有缓存
err := c.Flush()
```

## API 参考

### 缓存接口

`cache.go` 文件中定义了所有缓存实现必须实现的接口：

```go
type Cache interface {
    // 存储数据，指定过期时间（秒）
    Put(key string, value any, seconds int) error
    
    // 仅当键不存在时存储数据
    Add(key string, value any, seconds int) error
    
    // 获取数据
    Get(key string) (any, error)
    
    // 获取数据并删除
    Pull(key string) (any, error)
    
    // 检查键是否存在
    Has(key string) bool
    
    // 永久存储数据
    Forever(key string, value any) error
    
    // 删除键
    Forget(key string) (bool, error)
    
    // 递增整数值
    Increment(key string, n int) (int, error)
    
    // 递减整数值
    Decrement(key string, n int) (int, error)
    
    // 清空所有缓存
    Flush() error
}
```

### 内存缓存实现

`mem/mem.go` 提供了基于内存的缓存实现，具有以下特点：

- 使用分片设计减少锁竞争
- 自动清理过期项目
- 高性能并发访问

### Redis 缓存实现

`redis/redis.go` 提供了基于 Redis 的缓存实现，具有以下特点：

- 支持 Redis 连接配置
- 支持键前缀
- 支持 JSON 编码/解码
- 支持原子操作

## 性能考虑

### 内存缓存

内存缓存使用分片技术（默认 32 个分片）来减少锁竞争，提高并发性能。每个分片有自己的锁，这意味着不同键的操作可以并行执行，只要它们映射到不同的分片。

```go
// 内存缓存的分片数量
const cacheGroupCount = 32
```

### Redis 缓存

Redis 缓存使用连接池来管理 Redis 连接，这有助于减少连接开销并提高性能。对于高并发应用，建议使用专用的 Redis 实例。

## 最佳实践

1. **选择合适的驱动**：
   - 对于单机应用或不需要持久化的场景，使用内存缓存
   - 对于分布式应用或需要持久化的场景，使用 Redis 缓存

2. **合理设置过期时间**：
   - 避免缓存过期时间过短，导致频繁重建缓存
   - 避免过期时间过长，导致数据不一致

3. **使用前缀隔离不同应用的缓存**：
   ```go
   cache.New(cache.WithPrefix("myapp"))
   ```

4. **处理缓存未命中**：
   ```go
   value, err := cache.Get("key")
   if value == nil {
       // 缓存未命中，从数据源获取数据
       value = getFromDataSource()
       // 更新缓存
       cache.Put("key", value, 3600)
   }
   ```

5. **使用 Forever 存储不经常变化的数据**：
   ```go
   cache.Forever("app:config", configData)
   ```

6. **利用混合缓存策略**：
   - 对于热点数据，使用内存缓存提高访问速度
   - 对于需要持久化或共享的数据，使用 Redis 缓存

7. **实现缓存降级机制**：
   - 当 Redis 不可用时，自动降级到内存缓存
   - 避免缓存故障导致系统不可用

## 示例

更多示例可以参考 `example` 目录。

## 许可证

本项目采用 MIT 许可证。