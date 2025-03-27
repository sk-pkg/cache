package mem

import "time"

// janitor is responsible for periodically cleaning up expired cache items.
// It runs as a separate goroutine and can be stopped when no longer needed.
type janitor struct {
	Interval time.Duration // How frequently the janitor checks for expired items
	stop     chan struct{} // Channel used to signal the janitor to stop
}

// run starts the janitor's cleanup process for a given cache.
// It periodically scans the cache and removes expired items.
//
// Parameters:
//   - c: The cache shard to clean up
//
// The function runs indefinitely until signaled to stop via the stop channel.
func (j *janitor) run(c *cache) {
	// Create a ticker that triggers at the specified interval
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			{
				// Get current time for expiration comparison
				now := time.Now().UnixNano()
				c.Lock()
				// Iterate through all items in the cache
				for k, v := range c.items {
					// If the item has an expiration time and it's in the past, delete it
					if v.Expiration > 0 && now > v.Expiration {
						delete(c.items, k)
					}
				}
				c.Unlock()
			}
		case <-j.stop:
			// Stop the ticker and exit the function when signaled
			ticker.Stop()
			return
		}
	}
}

// runJanitor creates and starts a new janitor for a cache shard.
// It initializes the janitor with the specified cleanup interval and
// starts its cleanup process in a separate goroutine.
//
// Parameters:
//   - c: The cache shard to attach the janitor to
//   - ci: The cleanup interval duration
//
// Example:
//
//	cache := &cache{items: make(map[string]item)}
//	runJanitor(cache, time.Minute) // Run cleanup every minute
func runJanitor(c *cache, ci time.Duration) {
	// Create a new janitor with the specified interval
	j := &janitor{
		Interval: ci,
		stop:     make(chan struct{}),
	}

	// Attach the janitor to the cache
	c.janitor = j

	// Start the janitor's cleanup process in a separate goroutine
	go j.run(c)
}

// stopJanitor signals the janitor to stop its cleanup process.
// This function is typically called when the cache is being destroyed
// or when cleanup is no longer needed.
//
// Parameters:
//   - c: The cache shard whose janitor should be stopped
//
// This function is designed to be used with runtime.SetFinalizer to ensure
// the janitor is stopped when the cache is garbage collected.
func stopJanitor(c *cache) {
	// Send a signal to the janitor to stop
	c.janitor.stop <- struct{}{}
}
