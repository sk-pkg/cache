# sk-pkg/cache - High-Performance Cache Library

This is a flexible Go language caching system that supports multiple cache drivers (memory and Redis), providing a unified interface that makes switching between different cache backends seamless.

## Installation

```bash
go get "github.com/sk-pkg/cache"
```

## Features

- Support for memory cache and Redis cache
- Unified API interface
- Memory cache uses sharded design to reduce lock contention and improve concurrent performance
- Support for key-value expiration time
- Support for atomic increment and decrement operations
- Support for batch clearing operations

## Quick Start

### Memory Cache

```go
package main

import (
    "github.com/sk-pkg/cache"
    "log"
    "fmt"
)

func main() {
    // Create memory cache (default driver)
    c, err := cache.New()
    if err != nil {
        log.Fatal(err)
    }
    
    // Store data (expires in 1 second)
    err = c.Put("user:1", "John", 1)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve data
    value, err := c.Get("user:1")
    if err != nil {
        log.Fatal(err)
    }
    
    if value != nil {
        fmt.Println("User:", value.(string))
    }
}
```

### Redis Cache

```go
package main

import (
    "github.com/sk-pkg/cache"
    "github.com/sk-pkg/cache/redis"
    "log"
    "fmt"
)

func main() {
    // Redis configuration
    cfg := redis.Config{
        Address:  "localhost:6379",  // Redis server address
        Password: "password",        // Redis password (optional)
        Prefix:   "myapp",           // Global key prefix (optional)
    }
    
    // Create Redis cache
    c, err := cache.New(
        cache.WithDefaultDriver("redis"), 
        cache.WithRedisConfig(cfg),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Store data (expires in 60 seconds)
    err = c.Put("user:1", "John", 60)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve data
    value, err := c.Get("user:1")
    if err != nil {
        log.Fatal(err)
    }
    
    if value != nil {
        fmt.Println("User:", value.(string))
    }
}
```

## Advanced Usage

### Cache Manager Configuration Options

```go
// Use custom prefix
c, err := cache.New(cache.WithPrefix("myapp"))

// Use existing Redis manager
redisManager := redis.New(...)
c, err := cache.New(
    cache.WithDefaultDriver("redis"),
    cache.WithRedis(redisManager),
)
```

### Switching Drivers in Cache Manager

In addition to specifying the default driver at initialization, you can flexibly switch cache drivers at runtime. The cache manager provides the ability to directly access different driver instances.

```go
// Create cache manager
cacheManager, err := cache.New(
    cache.WithDefaultDriver("redis"),
    cache.WithRedisConfig(redisConfig),
)

// Use memory cache
err := cacheManager.Mem.Put("key", "value", 60)

// Use Redis cache
err := cacheManager.Redis.Put("key", "value", 60)
```

#### Use Cases

This approach is particularly useful in the following scenarios:

1. **Mixed Cache Strategy**:
   - Store frequently accessed data in memory cache
   - Store data that needs persistence in Redis

```go
// Frequently accessed configuration stored in memory
cacheManager.Mem.Put("app:config", configData, 300)

// User data stored in Redis for cross-service sharing
cacheManager.Redis.Put("user:1", userData, 3600)
```

2. **Cache Degradation**:
   - Automatically degrade to memory cache when Redis is unavailable

```go
func getData(key string, cache *cache.Manager) (any, error) {
    // Try to get from Redis
    if cache.Redis != nil {
        value, err := cache.Redis.Get(key)
        if err == nil && value != nil {
            return value, nil
        }
    }
    
    // Redis failed or unavailable, get from memory cache
    return cache.Mem.Get(key)
}
```

3. **Data Layering**:
   - Choose different cache drivers based on data characteristics

```go
// Temporary session data stored in memory
cacheManager.Mem.Put("session:"+sessionID, sessionData, 1800)

// Shared counters stored in Redis to support atomic operations
newCount, _ := cacheManager.Redis.Increment("visits", 1)
```

### Permanent Data Storage

```go
// Store data permanently (until manually deleted)
err := c.Forever("app:config", configData)
```

### Conditional Storage

```go
// Store data only if the key does not exist
err := c.Add("user:1", userData, 3600)
```

### Retrieve and Delete

```go
// Get data and remove it from the cache
value, err := c.Pull("user:1")
```

### Check if Key Exists

```go
// Check if key exists
if c.Has("user:1") {
    // Key exists
}
```

### Increment/Decrement Counters

```go
// Increment counter
newValue, err := c.Increment("visits", 1)

// Decrement counter
newValue, err := c.Decrement("remaining", 1)
```

### Clear Cache

```go
// Clear all cache
err := c.Flush()
```

## API Reference

### Cache Interface

The `cache.go` file defines the interface that all cache implementations must implement:

```go
type Cache interface {
    // Store data with specified expiration time (seconds)
    Put(key string, value any, seconds int) error
    
    // Store data only if the key does not exist
    Add(key string, value any, seconds int) error
    
    // Retrieve data
    Get(key string) (any, error)
    
    // Retrieve data and delete
    Pull(key string) (any, error)
    
    // Check if key exists
    Has(key string) bool
    
    // Store data permanently
    Forever(key string, value any) error
    
    // Delete key
    Forget(key string) (bool, error)
    
    // Increment integer value
    Increment(key string, n int) (int, error)
    
    // Decrement integer value
    Decrement(key string, n int) (int, error)
    
    // Clear all cache
    Flush() error
}
```

### Memory Cache Implementation

`mem/mem.go` provides a memory-based cache implementation with the following features:

- Uses sharded design to reduce lock contention
- Automatically cleans up expired items
- High-performance concurrent access

### Redis Cache Implementation

`redis/redis.go` provides a Redis-based cache implementation with the following features:

- Support for Redis connection configuration
- Support for key prefixes
- Support for JSON encoding/decoding
- Support for atomic operations

## Performance Considerations

### Memory Cache

Memory cache uses sharding technology (default 32 shards) to reduce lock contention and improve concurrent performance. Each shard has its own lock, which means operations on different keys can be executed in parallel as long as they map to different shards.

```go
// Number of shards for memory cache
const cacheGroupCount = 32
```

### Redis Cache

Redis cache uses connection pooling to manage Redis connections, which helps reduce connection overhead and improve performance. For high-concurrency applications, it is recommended to use a dedicated Redis instance.

## Best Practices

1. **Choose the Appropriate Driver**:
   - Use memory cache for single-machine applications or scenarios that don't require persistence
   - Use Redis cache for distributed applications or scenarios that require persistence

2. **Set Reasonable Expiration Times**:
   - Avoid setting cache expiration times too short, which can lead to frequent cache rebuilding
   - Avoid setting expiration times too long, which can lead to data inconsistency

3. **Use Prefixes to Isolate Cache for Different Applications**:
   ```go
   cache.New(cache.WithPrefix("myapp"))
   ```

4. **Handle Cache Misses**:
   ```go
   value, err := cache.Get("key")
   if value == nil {
       // Cache miss, get data from data source
       value = getFromDataSource()
       // Update cache
       cache.Put("key", value, 3600)
   }
   ```

5. **Use Forever for Infrequently Changing Data**:
   ```go
   cache.Forever("app:config", configData)
   ```

6. **Leverage Mixed Cache Strategy**:
   - Use memory cache for hot data to improve access speed
   - Use Redis cache for data that needs persistence or sharing

7. **Implement Cache Degradation Mechanism**:
   - Automatically degrade to memory cache when Redis is unavailable
   - Avoid system unavailability due to cache failures

## Examples

For more examples, refer to the `example` directory.

## License

This project is licensed under the MIT License.