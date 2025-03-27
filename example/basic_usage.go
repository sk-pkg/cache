package main

import (
	"fmt"
	"github.com/sk-pkg/cache"
	"github.com/sk-pkg/cache/redis"
	"log"
	"time"
)

func main() {
	// Demonstrate memory cache
	demoMemCache()

	// Demonstrate Redis cache
	// Uncomment the following line to test Redis cache
	// demoRedisCache()
}

func demoMemCache() {
	fmt.Println("=== Memory Cache Demonstration ===")

	// Create memory cache (default driver)
	c, err := cache.New(cache.WithPrefix("demo"))
	if err != nil {
		log.Fatal(err)
	}

	// Basic operations
	basicOperations(c)

	// Advanced operations
	advancedOperations(c)
}

func demoRedisCache() {
	fmt.Println("=== Redis Cache Demonstration ===")

	// Redis configuration
	cfg := redis.Config{
		Address:  "localhost:6379", // Please modify to your Redis server address
		Password: "",               // Set password here if needed
		Prefix:   "demo",
	}

	// Create Redis cache
	c, err := cache.New(
		cache.WithDefaultDriver("redis"),
		cache.WithRedisConfig(cfg),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Basic operations
	basicOperations(c)

	// Advanced operations
	advancedOperations(c)
}

func basicOperations(c *cache.Manager) {
	fmt.Println("\n--- Basic Operations ---")

	// Store data (expires in 5 seconds)
	err := c.Put("user:1", "John", 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Store data: user:1 = John (expires in 5 seconds)")

	// Get data
	value, err := c.Get("user:1")
	if err != nil {
		log.Fatal(err)
	}
	if value != nil {
		fmt.Printf("Get data: user:1 = %s\n", value.(string))
	}

	// Check if key exists
	exists := c.Has("user:1")
	fmt.Printf("Key exists: user:1 = %v\n", exists)

	// Wait for data to expire
	fmt.Println("Waiting 5 seconds for data to expire...")
	time.Sleep(6 * time.Second)

	// Get data again
	value, _ = c.Get("user:1")
	if value == nil {
		fmt.Println("Data expired: user:1 = nil")
	}
}

func advancedOperations(c *cache.Manager) {
	fmt.Println("\n--- Advanced Operations ---")

	// Store data permanently
	err := c.Forever("config:app", "Application configuration data")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Store data permanently: config:app = Application configuration data")

	// Conditional storage (only when key doesn't exist)
	err = c.Add("user:2", "David", 60)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Conditional storage: user:2 = David (only when key doesn't exist)")

	// Try conditional storage again (won't overwrite)
	err = c.Add("user:2", "Michael", 60)
	if err != nil {
		log.Fatal(err)
	}
	value, _ := c.Get("user:2")
	fmt.Printf("After conditional storage again: user:2 = %s (should still be David)\n", value.(string))

	// Increment counter
	err = c.Put("counter", 10, 60)
	if err != nil {
		log.Fatal(err)
	}
	newValue, err := c.Increment("counter", 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Increment counter: counter + 5 = %d\n", newValue)

	// Decrement counter
	newValue, err = c.Decrement("counter", 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Decrement counter: counter - 3 = %d\n", newValue)

	// Get and delete
	value, err = c.Pull("user:2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Get and delete: user:2 = %s\n", value.(string))

	// Verify key has been deleted
	exists := c.Has("user:2")
	fmt.Printf("Key exists: user:2 = %v (should be false)\n", exists)

	// Flush all cache
	fmt.Println("Flushing all cache...")
	err = c.Flush()
	if err != nil {
		log.Fatal(err)
	}

	// Verify cache has been flushed
	value, _ = c.Get("config:app")
	fmt.Printf("Get after flush: config:app = %v (should be nil)\n", value)
}
