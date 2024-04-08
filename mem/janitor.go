package mem

import "time"

type janitor struct {
	Interval time.Duration
	stop     chan struct{}
}

func (j *janitor) run(c *cache) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			{
				now := time.Now().UnixNano()
				c.Lock()
				for k, v := range c.items {
					if v.Expiration > 0 && now > v.Expiration {
						delete(c.items, k)
					}
				}
				c.Unlock()
			}
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func runJanitor(c *cache, ci time.Duration) {
	j := &janitor{
		Interval: ci,
		stop:     make(chan struct{}),
	}

	c.janitor = j

	go j.run(c)
}

func stopJanitor(c *cache) {
	c.janitor.stop <- struct{}{}
}
