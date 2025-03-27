// Package mem provides an in-memory cache implementation with expiration support.
// It uses a sharded approach to reduce lock contention and improve performance
// in concurrent environments.
package mem

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// cacheGroupCount defines the number of shards for the cache.
// Sharding helps reduce lock contention in concurrent environments.
const cacheGroupCount = 32

// Cache is a collection of cache shards that together form the complete cache.
// Operations on the cache are distributed across shards based on key hashing.
type Cache []*cache

// cache represents a single shard of the cache system.
// Each shard has its own lock to reduce contention.
type cache struct {
	items        map[string]item // Map of cached items
	janitor      *janitor        // Reference to the cleanup process
	sync.RWMutex                 // Lock for concurrent access
}

// item represents a single cached value with its expiration time.
type item struct {
	value      any   // The stored value
	Expiration int64 // Unix nano timestamp when the item expires (0 = no expiration)
}

// Init creates and initializes a new in-memory cache.
// It creates multiple cache shards and sets up janitors for each shard
// to clean up expired items.
//
// Returns:
//   - Cache: An initialized cache ready for use
//
// Example:
//
//	cache := mem.Init()
//	cache.Put("key", "value", 60) // Store for 60 seconds
func Init() Cache {
	const itemCount = 256 // Initial capacity for each shard's map

	// Create the cache with the specified number of shards
	c := make(Cache, cacheGroupCount)
	for i := 0; i < cacheGroupCount; i++ {
		// Initialize each shard with its own map
		c[i] = &cache{items: make(map[string]item, itemCount)}

		// Start a janitor for each shard to clean up expired items
		runJanitor(c[i], time.Second)
		// Set finalizer to ensure janitor is stopped when cache is garbage collected
		runtime.SetFinalizer(c[i], stopJanitor)
	}

	return c
}

// getGroup returns the appropriate cache shard for the given key.
// It uses a hash function to determine which shard should handle the key.
//
// Parameters:
//   - key: The cache key to determine the shard for
//
// Returns:
//   - *cache: The cache shard responsible for the key
func (c Cache) getGroup(key string) *cache {
	return c[uint(fnv32(key))%uint(cacheGroupCount)]
}

// Put stores a value in the cache with the specified expiration time.
// If the key already exists, its value will be overwritten.
//
// Parameters:
//   - key: The key under which to store the value
//   - value: The value to store
//   - seconds: The time-to-live in seconds (0 for no expiration)
//
// Returns:
//   - error: Always nil (for interface compatibility)
//
// Example:
//
//	cache.Put("user:123", userData, 3600) // Store for 1 hour
func (c Cache) Put(key string, value any, seconds int) error {
	var e int64
	// Calculate expiration time if seconds > 0
	if seconds > 0 {
		e = time.Now().Add(time.Duration(seconds) * time.Second).UnixNano()
	}

	// Create the cache item
	data := item{
		value:      value,
		Expiration: e,
	}

	// Get the appropriate shard and store the item
	group := c.getGroup(key)
	group.Lock()
	group.items[key] = data
	group.Unlock()

	return nil
}

// Add adds a value to the cache only if the key does not already exist.
// If the key exists, the operation is a no-op.
//
// Parameters:
//   - key: The key under which to store the value
//   - value: The value to store
//   - seconds: The time-to-live in seconds (0 for no expiration)
//
// Returns:
//   - error: Always nil (for interface compatibility)
//
// Example:
//
//	// Only sets the value if "user:123" doesn't exist
//	cache.Add("user:123", userData, 3600)
func (c Cache) Add(key string, value any, seconds int) error {
	group := c.getGroup(key)
	group.Lock()

	// Check if the key already exists
	_, ok := group.items[key]
	if !ok {
		// Key doesn't exist, add it with expiration if specified
		var e int64
		if seconds > 0 {
			e = time.Now().Add(time.Duration(seconds) * time.Second).UnixNano()
		}

		group.items[key] = item{
			value:      value,
			Expiration: e,
		}
	}
	group.Unlock()

	return nil
}

// Get retrieves a value from the cache.
// If the key does not exist, it returns nil without an error.
//
// Parameters:
//   - key: The key to retrieve
//
// Returns:
//   - any: The retrieved value, or nil if not found
//   - error: Always nil (for interface compatibility)
//
// Example:
//
//	value, _ := cache.Get("user:123")
//	if value != nil {
//	    userData := value.(UserData)
//	}
func (c Cache) Get(key string) (any, error) {
	group := c.getGroup(key)
	group.RLock()

	// Get the item from the cache (returns zero value if not found)
	i, _ := group.items[key]
	group.RUnlock()

	return i.value, nil
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
//	value, _ := cache.Pull("user:123")
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
	group := c.getGroup(key)
	group.RLock()

	// Check if the key exists in the map
	_, ok := group.items[key]
	group.RUnlock()

	return ok
}

// Forever stores a value in the cache indefinitely (without expiration).
// This is equivalent to calling Put with seconds=0.
//
// Parameters:
//   - key: The key under which to store the value
//   - value: The value to store
//
// Returns:
//   - error: Always nil (for interface compatibility)
//
// Example:
//
//	cache.Forever("app:config", configData)
func (c Cache) Forever(key string, value any) error {
	group := c.getGroup(key)
	group.Lock()
	// Store with no expiration (Expiration = 0)
	group.items[key] = item{value: value}
	group.Unlock()

	return nil
}

// Forget removes a key from the cache.
//
// Parameters:
//   - key: The key to remove
//
// Returns:
//   - bool: Always true (for interface compatibility)
//   - error: Always nil (for interface compatibility)
//
// Example:
//
//	removed, _ := cache.Forget("user:123")
func (c Cache) Forget(key string) (bool, error) {
	group := c.getGroup(key)

	group.Lock()
	// Remove the key from the map
	delete(group.items, key)
	group.Unlock()

	return true, nil
}

// Increment atomically increments the integer value of a key by the given amount.
// If the key does not exist, it is set to the amount.
// If the value is not an integer, an error is returned.
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
//	newValue, _ := cache.Increment("visits", 1)
//	// newValue is the updated counter
func (c Cache) Increment(key string, n int) (int, error) {
	group := c.getGroup(key)

	// Use defer to ensure the lock is released even if an error occurs
	defer group.Unlock()
	group.Lock()

	// Check if the key exists
	v, ok := group.items[key]
	if !ok {
		// Key doesn't exist, create it with the increment value
		group.items[key] = item{value: n}
		return n, nil
	}

	// Check if the value is an integer
	nv, ok := v.value.(int)
	if !ok {
		return 0, fmt.Errorf("Invalid type: expected int, got %T", v.value)
	}

	// Increment the value
	nv += n
	v.value = nv
	group.items[key] = v

	return nv, nil
}

// Decrement atomically decrements the integer value of a key by the given amount.
// If the key does not exist, an error is returned.
// If the value is not an integer, an error is returned.
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
//	newValue, _ := cache.Decrement("remaining", 1)
//	// newValue is the updated counter
func (c Cache) Decrement(key string, n int) (int, error) {
	group := c.getGroup(key)

	// Use defer to ensure the lock is released even if an error occurs
	defer group.Unlock()
	group.Lock()

	// Check if the key exists
	v, ok := group.items[key]
	if !ok {
		return n, errors.New("Undefined key: " + key)
	}

	// Check if the value is an integer
	nv, ok := v.value.(int)
	if !ok {
		return 0, errors.New("Invalid type ")
	}

	// Decrement the value
	nv -= n
	v.value = nv
	group.items[key] = v

	return nv, nil
}

// Flush removes all items from the cache.
// This operation clears all shards.
//
// Returns:
//   - error: Always nil (for interface compatibility)
//
// Example:
//
//	cache.Flush() // Clear the entire cache
func (c Cache) Flush() error {
	// Iterate through all shards
	for _, group := range c {
		group.Lock()

		// Clear the map
		clear(group.items)

		group.Unlock()
	}

	return nil
}

// fnv32 is a hash function for strings based on the FNV algorithm.
// It's used to determine which cache shard should handle a given key.
//
// Parameters:
//   - key: The string to hash
//
// Returns:
//   - uint32: The hash value
func fnv32(key string) uint32 {
	hash := uint32(2166136261)       // FNV offset basis
	const prime32 = uint32(16777619) // FNV prime
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
