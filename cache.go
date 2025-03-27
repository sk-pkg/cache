// Package cache provides a flexible caching system with support for multiple drivers.
// It offers a unified interface for both memory and Redis caching mechanisms,
// allowing for seamless switching between different cache backends.
package cache

import (
	"github.com/sk-pkg/cache/mem"
	"github.com/sk-pkg/cache/redis"
	redisManager "github.com/sk-pkg/redis"
)

const (
	// MemCache represents the memory cache driver identifier
	MemCache = "mem"
	// RedisCache represents the Redis cache driver identifier
	RedisCache = "redis"
	// DefaultPrefix is the default key prefix used for all cache entries
	DefaultPrefix = "go_cache:"
)

// Cache defines the interface for all cache implementations.
// Any cache driver must implement these methods to be compatible with the cache manager.
type Cache interface {
	// Put stores data in the cache for a specified duration.
	//
	// Parameters:
	//   - key: The unique identifier for the cached item
	//   - value: The data to be stored in the cache
	//   - seconds: The time-to-live in seconds (0 means no expiration)
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   err := cache.Put("user:1", userData, 3600) // Cache for 1 hour
	Put(key string, value any, seconds int) error

	// Add stores data in the cache only if the key does not already exist.
	//
	// Parameters:
	//   - key: The unique identifier for the cached item
	//   - value: The data to be stored in the cache
	//   - seconds: The time-to-live in seconds (0 means no expiration)
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   err := cache.Add("user:1", userData, 3600) // Cache for 1 hour if not exists
	Add(key string, value any, seconds int) error

	// Get retrieves data from the cache.
	//
	// Parameters:
	//   - key: The unique identifier for the cached item
	//
	// Returns:
	//   - any: The cached data if found, nil otherwise
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   data, err := cache.Get("user:1")
	//   if err != nil {
	//     // Handle error
	//   }
	//   if data == nil {
	//     // Key not found
	//   }
	Get(key string) (any, error)

	// Pull retrieves data from the cache and then removes it.
	//
	// Parameters:
	//   - key: The unique identifier for the cached item
	//
	// Returns:
	//   - any: The cached data if found, nil otherwise
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   data, err := cache.Pull("user:1")
	//   // After this call, "user:1" is no longer in the cache
	Pull(key string) (any, error)

	// Has checks if an item exists in the cache.
	//
	// Parameters:
	//   - key: The unique identifier for the cached item
	//
	// Returns:
	//   - bool: true if the item exists, false otherwise
	//
	// Example:
	//   if cache.Has("user:1") {
	//     // Item exists in cache
	//   }
	Has(key string) bool

	// Forever stores data in the cache permanently (until manually removed).
	//
	// Parameters:
	//   - key: The unique identifier for the cached item
	//   - value: The data to be stored in the cache
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   err := cache.Forever("app:config", configData)
	Forever(key string, value any) error

	// Forget removes an item from the cache.
	//
	// Parameters:
	//   - key: The unique identifier for the cached item
	//
	// Returns:
	//   - bool: true if the item was removed, false otherwise
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   removed, err := cache.Forget("user:1")
	Forget(key string) (bool, error)

	// Increment increases the integer value of a key by the given amount.
	//
	// Parameters:
	//   - key: The unique identifier for the cached item
	//   - n: The amount to increment by
	//
	// Returns:
	//   - int: The new value after incrementing
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   newValue, err := cache.Increment("visits", 1)
	Increment(key string, n int) (int, error)

	// Decrement decreases the integer value of a key by the given amount.
	//
	// Parameters:
	//   - key: The unique identifier for the cached item
	//   - n: The amount to decrement by
	//
	// Returns:
	//   - int: The new value after decrementing
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   newValue, err := cache.Decrement("remaining", 1)
	Decrement(key string, n int) (int, error)

	// Flush removes all items from the cache.
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   err := cache.Flush()
	Flush() error
}

// Manager provides a unified interface to work with different cache implementations.
// It supports both memory and Redis cache backends.
type Manager struct {
	// Mem is the memory cache implementation
	Mem mem.Cache
	// Redis is the Redis cache implementation
	Redis *redis.Cache
	// defaultCache is the currently active cache implementation
	defaultCache Cache
}

// Option is a function type used for configuring the cache manager.
type Option func(*option)

// option holds the configuration options for the cache manager.
type option struct {
	defaultDriver string
	prefix        string
	redis         *redisManager.Manager
	redisConfig   redis.Config
}

// WithDefaultDriver sets the default cache driver to use.
//
// Parameters:
//   - driver: The cache driver identifier (e.g., "mem" or "redis")
//
// Returns:
//   - Option: A configuration option function
//
// Example:
//
//	cache.New(cache.WithDefaultDriver("redis"))
func WithDefaultDriver(driver string) Option {
	return func(o *option) {
		o.defaultDriver = driver
	}
}

// WithPrefix sets the key prefix for all cache entries.
//
// Parameters:
//   - prefix: The prefix to use for all cache keys
//
// Returns:
//   - Option: A configuration option function
//
// Example:
//
//	cache.New(cache.WithPrefix("myapp"))
func WithPrefix(prefix string) Option {
	return func(o *option) {
		o.prefix = prefix + ":"
	}
}

// WithRedis sets an existing Redis manager to use for Redis caching.
//
// Parameters:
//   - redis: An initialized Redis manager instance
//
// Returns:
//   - Option: A configuration option function
//
// Example:
//
//	redisManager := redis.New(...)
//	cache.New(cache.WithRedis(redisManager))
func WithRedis(redis *redisManager.Manager) Option {
	return func(o *option) {
		o.redis = redis
	}
}

// WithRedisConfig sets the Redis connection configuration.
//
// Parameters:
//   - redisConfig: The Redis connection configuration
//
// Returns:
//   - Option: A configuration option function
//
// Example:
//
//	config := redis.Config{
//	  Address: "localhost:6379",
//	  Password: "password",
//	  Prefix: "myapp",
//	}
//	cache.New(cache.WithRedisConfig(config))
func WithRedisConfig(redisConfig redis.Config) Option {
	return func(o *option) {
		o.redisConfig = redisConfig
	}
}

// Put stores data in the cache for a specified duration using the default cache driver.
//
// Parameters:
//   - key: The unique identifier for the cached item
//   - value: The data to be stored in the cache
//   - seconds: The time-to-live in seconds (0 means no expiration)
//
// Returns:
//   - error: Any error that occurred during the operation
func (m *Manager) Put(key string, value any, seconds int) error {
	return m.defaultCache.Put(key, value, seconds)
}

// Add stores data in the cache only if the key does not already exist using the default cache driver.
//
// Parameters:
//   - key: The unique identifier for the cached item
//   - value: The data to be stored in the cache
//   - seconds: The time-to-live in seconds (0 means no expiration)
//
// Returns:
//   - error: Any error that occurred during the operation
func (m *Manager) Add(key string, value any, seconds int) error {
	return m.defaultCache.Add(key, value, seconds)
}

// Get retrieves data from the cache using the default cache driver.
//
// Parameters:
//   - key: The unique identifier for the cached item
//
// Returns:
//   - any: The cached data if found, nil otherwise
//   - error: Any error that occurred during the operation
func (m *Manager) Get(key string) (any, error) {
	return m.defaultCache.Get(key)
}

// Pull retrieves data from the cache and then removes it using the default cache driver.
//
// Parameters:
//   - key: The unique identifier for the cached item
//
// Returns:
//   - any: The cached data if found, nil otherwise
//   - error: Any error that occurred during the operation
func (m *Manager) Pull(key string) (any, error) {
	return m.defaultCache.Pull(key)
}

// Has checks if an item exists in the cache using the default cache driver.
//
// Parameters:
//   - key: The unique identifier for the cached item
//
// Returns:
//   - bool: true if the item exists, false otherwise
func (m *Manager) Has(key string) bool {
	return m.defaultCache.Has(key)
}

// Forever stores data in the cache permanently using the default cache driver.
//
// Parameters:
//   - key: The unique identifier for the cached item
//   - value: The data to be stored in the cache
//
// Returns:
//   - error: Any error that occurred during the operation
func (m *Manager) Forever(key string, value any) error {
	return m.defaultCache.Forever(key, value)
}

// Forget removes an item from the cache using the default cache driver.
//
// Parameters:
//   - key: The unique identifier for the cached item
//
// Returns:
//   - bool: true if the item was removed, false otherwise
//   - error: Any error that occurred during the operation
func (m *Manager) Forget(key string) (bool, error) {
	return m.defaultCache.Forget(key)
}

// Increment increases the integer value of a key by the given amount using the default cache driver.
//
// Parameters:
//   - key: The unique identifier for the cached item
//   - n: The amount to increment by
//
// Returns:
//   - int: The new value after incrementing
//   - error: Any error that occurred during the operation
func (m *Manager) Increment(key string, n int) (int, error) {
	return m.defaultCache.Increment(key, n)
}

// Decrement decreases the integer value of a key by the given amount using the default cache driver.
//
// Parameters:
//   - key: The unique identifier for the cached item
//   - n: The amount to decrement by
//
// Returns:
//   - int: The new value after decrementing
//   - error: Any error that occurred during the operation
func (m *Manager) Decrement(key string, n int) (int, error) {
	return m.defaultCache.Decrement(key, n)
}

// Flush removes all items from the cache using the default cache driver.
//
// Returns:
//   - error: Any error that occurred during the operation
func (m *Manager) Flush() error {
	return m.defaultCache.Flush()
}

// New creates a new cache manager with the specified options.
//
// Parameters:
//   - opts: A variadic list of Option functions to configure the cache manager
//
// Returns:
//   - *Manager: The initialized cache manager
//   - error: Any error that occurred during initialization
//
// Example:
//
//	// Create a memory cache
//	memCache, err := cache.New()
//
//	// Create a Redis cache
//	redisCache, err := cache.New(
//	  cache.WithDefaultDriver("redis"),
//	  cache.WithRedisConfig(redis.Config{
//	    Address: "localhost:6379",
//	  }),
//	)
func New(opts ...Option) (*Manager, error) {
	// Initialize options with default prefix
	opt := &option{prefix: DefaultPrefix}

	// Apply all provided option functions
	for _, f := range opts {
		f(opt)
	}

	manager := &Manager{}

	// Initialize memory cache (always available)
	manager.Mem = mem.Init()

	// Initialize Redis cache if Redis configuration is provided
	if opt.redis != nil || opt.redisConfig != (redis.Config{}) {
		redisCache, err := redis.Init(
			redis.WithPrefix(opt.prefix),
			redis.WithRedisConfig(opt.redisConfig),
			redis.WithRedisManager(opt.redis),
		)
		if err != nil {
			return nil, err
		}
		manager.Redis = redisCache
	}

	// Set the default cache driver based on configuration
	switch opt.defaultDriver {
	case RedisCache:
		// Use Redis if available, otherwise fall back to memory cache
		if manager.Redis != nil {
			manager.defaultCache = manager.Redis
		} else {
			manager.defaultCache = manager.Mem
		}
	default:
		// Use memory cache as the default
		manager.defaultCache = manager.Mem
	}

	return manager, nil
}
