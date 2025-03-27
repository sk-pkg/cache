// Package redis provides a Redis implementation of the cache interface.
// It allows storing and retrieving data from Redis with various operations
// such as Put, Get, Increment, Decrement, etc.
package redis

import (
	"encoding/json"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/sk-pkg/redis"
)

// Option is a function type that configures the option struct.
type Option func(*option)

// option holds configuration parameters for the Redis cache.
type option struct {
	prefix       string
	redisManager *redis.Manager
	redisConfig  Config
}

// Config holds Redis connection configuration parameters.
type Config struct {
	Address  string // Redis server address in format "host:port"
	Password string // Redis server password, can be empty if no authentication required
	Prefix   string // Global prefix for all Redis keys
}

// Cache implements the cache interface using Redis as the storage backend.
type Cache struct {
	redis  *redis.Manager // Redis connection manager
	prefix string         // Key prefix for this cache instance
}

// WithPrefix returns an Option that sets the key prefix for the cache.
// The prefix will be prepended to all keys when storing or retrieving data.
//
// Example:
//
//	cache, _ := Init(WithPrefix("app:"))
//	// Keys will be stored as "app:keyname" in Redis
func WithPrefix(prefix string) Option {
	return func(o *option) {
		o.prefix = prefix
	}
}

// WithRedisManager returns an Option that sets a custom Redis manager.
// This allows reusing an existing Redis connection pool.
//
// Example:
//
//	redisManager := redis.New(...)
//	cache, _ := Init(WithRedisManager(redisManager))
func WithRedisManager(redis *redis.Manager) Option {
	return func(o *option) {
		o.redisManager = redis
	}
}

// WithRedisConfig returns an Option that configures the Redis connection.
// This is used when a new Redis connection needs to be established.
//
// Example:
//
//	config := Config{
//	    Address: "localhost:6379",
//	    Password: "secret",
//	    Prefix: "myapp:",
//	}
//	cache, _ := Init(WithRedisConfig(config))
func WithRedisConfig(redisConfig Config) Option {
	return func(o *option) {
		o.redisConfig = redisConfig
	}
}

// Init creates and initializes a new Redis cache with the provided options.
// It returns a pointer to the initialized Cache and any error encountered.
//
// If a RedisManager is provided via WithRedisManager, it will be used.
// Otherwise, if RedisConfig is provided via WithRedisConfig, a new Redis
// connection will be established.
//
// Example:
//
//	// Using configuration
//	config := redis.Config{
//	    Address: "localhost:6379",
//	    Password: "secret",
//	    Prefix: "myapp:",
//	}
//	cache, err := redis.Init(redis.WithRedisConfig(config))
//
//	// Using existing Redis manager
//	redisManager := redis.New(...)
//	cache, err := redis.Init(redis.WithRedisManager(redisManager))
func Init(opts ...Option) (*Cache, error) {
	opt := &option{}
	// Apply all provided options to the option struct
	for _, f := range opts {
		f(opt)
	}

	redisManager := opt.redisManager
	// If Redis address is provided, create a new Redis manager
	if opt.redisConfig.Address != "" {
		redisManager = redis.New(
			redis.WithPrefix(opt.redisConfig.Prefix),
			redis.WithAddress(opt.redisConfig.Address),
			redis.WithPassword(opt.redisConfig.Password),
		)
	}

	// Create and return the cache instance
	rdsCache := &Cache{
		redis:  redisManager,
		prefix: opt.prefix,
	}

	return rdsCache, nil
}

// Put stores a value in the cache for the specified duration in seconds.
// If seconds is 0, the value will be stored indefinitely.
//
// Parameters:
//   - key: The key under which to store the value
//   - value: The value to store (will be JSON encoded)
//   - seconds: The time-to-live in seconds (0 for indefinite)
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := cache.Put("user:123", userData, 3600) // Store for 1 hour
func (c Cache) Put(key string, value any, seconds int) error {
	return c.redis.Set(c.prefix+key, value, seconds)
}

// Add stores a value in the cache only if the key does not already exist.
// If the key exists, the operation is a no-op and returns nil.
//
// Parameters:
//   - key: The key under which to store the value
//   - value: The value to store (will be JSON encoded)
//   - seconds: The time-to-live in seconds (0 for indefinite)
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	// Only sets the value if "user:123" doesn't exist
//	err := cache.Add("user:123", userData, 3600)
func (c Cache) Add(key string, value any, seconds int) error {
	// Only set the value if the key doesn't exist
	if !c.Has(key) {
		return c.redis.Set(c.prefix+key, value, seconds)
	}

	return nil
}

// Get retrieves a value from the cache.
// If the key does not exist or has expired, it returns nil and an error.
//
// Parameters:
//   - key: The key to retrieve
//
// Returns:
//   - any: The retrieved value, or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	value, err := cache.Get("user:123")
//	if err != nil {
//	    // Handle error
//	}
//	userData := value.(map[string]interface{})
func (c Cache) Get(key string) (any, error) {
	// Get the raw bytes from Redis
	bytes, err := c.redis.Get(c.prefix + key)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data
	var value any
	err = json.Unmarshal(bytes, &value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// Pull retrieves a value from the cache and then removes it.
// This is equivalent to calling Get followed by Forget.
//
// Parameters:
//   - key: The key to retrieve and remove
//
// Returns:
//   - any: The retrieved value, or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	// Get the value and remove it in one operation
//	value, err := cache.Pull("user:123")
func (c Cache) Pull(key string) (any, error) {
	// Get the value first
	value, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	// Then delete the key
	_, err = c.Forget(key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// Has checks if a key exists in the cache.
//
// Parameters:
//   - key: The key to check
//
// Returns:
//   - bool: true if the key exists, false otherwise
//
// Example:
//
//	if cache.Has("user:123") {
//	    // Key exists
//	}
func (c Cache) Has(key string) bool {
	exists, _ := c.redis.Exists(c.prefix + key)
	return exists
}

// Forever stores a value in the cache indefinitely (without expiration).
// This is equivalent to calling Put with seconds=0.
//
// Parameters:
//   - key: The key under which to store the value
//   - value: The value to store (will be JSON encoded)
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := cache.Forever("app:config", configData)
func (c Cache) Forever(key string, value any) error {
	return c.redis.Set(c.prefix+key, value, 0)
}

// Forget removes a key from the cache.
//
// Parameters:
//   - key: The key to remove
//
// Returns:
//   - bool: true if the key was removed, false if it didn't exist
//   - error: Any error encountered during the operation
//
// Example:
//
//	removed, err := cache.Forget("user:123")
func (c Cache) Forget(key string) (bool, error) {
	return c.redis.Del(c.prefix + key)
}

// Increment atomically increments the integer value of a key by the given amount.
// If the key does not exist, it is set to the amount.
//
// Parameters:
//   - key: The key to increment
//   - n: The amount to increment by
//
// Returns:
//   - int: The new value after incrementing
//   - error: Any error encountered during the operation
//
// Example:
//
//	newValue, err := cache.Increment("visits", 1)
//	// newValue is the updated counter
func (c Cache) Increment(key string, n int) (int, error) {
	// Get a connection from the pool
	conn := c.redis.ConnPool.Get()
	defer conn.Close()

	// Execute INCRBY command and return the result
	return redigo.Int(conn.Do("INCRBY", c.prefix+key, n))
}

// Decrement atomically decrements the integer value of a key by the given amount.
// If the key does not exist, it is set to the negative of the amount.
//
// Parameters:
//   - key: The key to decrement
//   - n: The amount to decrement by
//
// Returns:
//   - int: The new value after decrementing
//   - error: Any error encountered during the operation
//
// Example:
//
//	newValue, err := cache.Decrement("remaining", 1)
//	// newValue is the updated counter
func (c Cache) Decrement(key string, n int) (int, error) {
	// Get a connection from the pool
	conn := c.redis.ConnPool.Get()
	defer conn.Close()

	// Execute DECRBY command and return the result
	return redigo.Int(conn.Do("DECRBY", c.prefix+key, n))
}

// Flush removes all keys with the cache prefix from Redis.
// This effectively clears the entire cache.
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := cache.Flush()
//	// All keys with the cache prefix are now removed
func (c Cache) Flush() error {
	return c.redis.BatchDel(c.prefix)
}
