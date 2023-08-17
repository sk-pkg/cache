package mem

import (
	"errors"
	"runtime"
	"sync"
	"time"
)

const cacheGroupCount = 32

type Cache []*cache

type cache struct {
	items   map[string]item
	janitor *janitor
	sync.RWMutex
}

type item struct {
	value      interface{}
	Expiration int64
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
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
		stop:     make(chan bool),
	}

	c.janitor = j

	go j.run(c)
}

func stopJanitor(c *cache) {
	c.janitor.stop <- true
}

func Init() Cache {
	c := make(Cache, cacheGroupCount)
	for i := 0; i < cacheGroupCount; i++ {
		c[i] = &cache{items: make(map[string]item, 256)}

		runJanitor(c[i], time.Second)
		runtime.SetFinalizer(c[i], stopJanitor)
	}

	return c
}

func (c Cache) getGroup(key string) *cache {
	return c[uint(fnv32(key))%uint(cacheGroupCount)]
}

func (c Cache) Put(key string, value interface{}, seconds int) error {
	var e int64
	if seconds > 0 {
		e = time.Now().Add(time.Duration(seconds) * time.Second).UnixNano()
	}

	data := item{
		value:      value,
		Expiration: e,
	}

	group := c.getGroup(key)
	group.Lock()

	group.items[key] = data
	group.Unlock()

	return nil
}

func (c Cache) Add(key string, value interface{}, seconds int) error {
	group := c.getGroup(key)
	group.Lock()

	_, ok := group.items[key]
	if !ok {
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

func (c Cache) Get(key string) (interface{}, error) {
	group := c.getGroup(key)
	group.RLock()

	i, _ := group.items[key]
	group.RUnlock()

	return i.value, nil
}

func (c Cache) Pull(key string) (interface{}, error) {
	value, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	_, err = c.Forget(key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (c Cache) Has(key string) bool {
	group := c.getGroup(key)
	group.RLock()

	_, ok := group.items[key]
	group.RUnlock()

	return ok
}

func (c Cache) Forever(key string, value interface{}) error {
	group := c.getGroup(key)
	group.Lock()
	group.items[key] = item{value: value}
	group.Unlock()

	return nil
}

func (c Cache) Forget(key string) (bool, error) {
	group := c.getGroup(key)

	group.Lock()
	delete(group.items, key)

	group.Unlock()

	return true, nil
}

func (c Cache) Increment(key string, n int) (int, error) {
	group := c.getGroup(key)
	group.RLock()

	v, ok := group.items[key]
	if !ok {
		group.items[key] = item{value: n}
		group.RUnlock()
		return n, nil
	}

	nv, ok := v.value.(int)
	if !ok {
		return 0, errors.New("Invalid type ")
	}

	nv += n
	v.value = nv
	group.items[key] = v
	group.RUnlock()

	return nv, nil
}

func (c Cache) Decrement(key string, n int) (int, error) {
	group := c.getGroup(key)
	group.RLock()

	v, ok := group.items[key]
	if !ok {
		group.RUnlock()
		return n, errors.New("Undefined key: " + key)
	}

	nv, ok := v.value.(int)
	if !ok {
		return 0, errors.New("Invalid type ")
	}

	nv -= n
	v.value = nv
	group.items[key] = v
	group.RUnlock()

	return nv, nil
}

func (c Cache) Flush() error {
	for _, group := range c {
		group.items = map[string]item{}
	}

	return nil
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
