package main

import (
	"fmt"
	"github.com/sk-pkg/cache"
	"github.com/sk-pkg/cache/redis"
	"log"
)

func main() {
	// Create a cache manager that supports both memory and Redis
	cacheManager, err := createMixedCacheManager()
	if err != nil {
		log.Fatal(err)
	}

	// Demonstrate mixed cache strategy
	demoMixedCacheStrategy(cacheManager)

	// Demonstrate cache degradation
	demoCacheDegradation(cacheManager)
}

func createMixedCacheManager() (*cache.Manager, error) {
	// Redis configuration
	cfg := redis.Config{
		Address:  "localhost:6379", // Please modify to your Redis server address
		Password: "",               // Set password here if needed
		Prefix:   "mixed",
	}

	// Create a cache manager that supports both memory and Redis
	// Default to Redis, but memory cache is also available
	return cache.New(
		cache.WithDefaultDriver("redis"),
		cache.WithRedisConfig(cfg),
		cache.WithPrefix("mixed"),
	)
}

func demoMixedCacheStrategy(c *cache.Manager) {
	fmt.Println("=== Mixed Cache Strategy Demonstration ===")

	// Store frequently accessed configuration data in memory
	err := c.Mem.Put("config:app", "Application configuration data", 300)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Store configuration in memory: config:app")

	// Store user data that needs to be shared across services in Redis
	err = c.Redis.Put("user:1", "John's user data", 3600)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Store user data in Redis: user:1")

	// Read configuration from memory
	configData, err := c.Mem.Get("config:app")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Read configuration from memory: %s\n", configData.(string))

	// Read user data from Redis
	userData, err := c.Redis.Get("user:1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Read user data from Redis: %s\n", userData.(string))

	// Store data using default driver (Redis)
	err = c.Put("session:123", "Session data", 60)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Store session data using default driver: session:123")

	// Read data using default driver
	sessionData, err := c.Get("session:123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Read session data using default driver: %s\n", sessionData.(string))
}

func demoCacheDegradation(c *cache.Manager) {
	fmt.Println("\n=== Cache Degradation Demonstration ===")

	// Store data in Redis
	err := c.Redis.Put("important:data", "Important data", 3600)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Store important data in Redis")

	// Also store a copy in memory (as backup)
	err = c.Mem.Put("important:data", "Important data (memory backup)", 3600)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Store backup data in memory")

	// Function to simulate getting data from Redis or memory
	data := getDataWithFallback(c, "important:data")
	fmt.Printf("Retrieved data: %s\n", data)

	// Simulate Redis unavailability
	fmt.Println("\nSimulating Redis unavailability...")
	// In a real application, this would be a Redis connection failure
	// For demonstration, we directly get from memory
	data = getDataWithFallback(c, "important:data")
	fmt.Printf("Data retrieved when Redis is unavailable: %s\n", data)
}

// Function with degradation capability for data retrieval
func getDataWithFallback(c *cache.Manager, key string) string {
	// First try to get from Redis
	if c.Redis != nil {
		value, err := c.Redis.Get(key)
		if err == nil && value != nil {
			return fmt.Sprintf("%s (from Redis)", value.(string))
		}
		// In a real application, we would check if err is a connection error
		fmt.Println("Redis retrieval failed, degrading to memory cache")
	}

	// Redis failed or unavailable, get from memory cache
	value, _ := c.Mem.Get(key)
	if value != nil {
		return fmt.Sprintf("%s (from memory)", value.(string))
	}

	return "Data does not exist"
}
