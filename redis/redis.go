package redis

import (
	"encoding/json"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/sk-pkg/redis"
)

type Option func(*option)

type option struct {
	prefix       string
	redisManager *redis.Manager
	redisConfig  Config
}

type Config struct {
	Address  string
	Password string
	Prefix   string
}

type Cache struct {
	redis  *redis.Manager
	prefix string
}

func WithPrefix(prefix string) Option {
	return func(o *option) {
		o.prefix = prefix
	}
}

func WithRedisManager(redis *redis.Manager) Option {
	return func(o *option) {
		o.redisManager = redis
	}
}

func WithRedisConfig(redisConfig Config) Option {
	return func(o *option) {
		o.redisConfig = redisConfig
	}
}

func Init(opts ...Option) (*Cache, error) {
	opt := &option{}
	for _, f := range opts {
		f(opt)
	}

	redisManager := opt.redisManager
	if opt.redisConfig.Address != "" {
		redisManager = redis.New(
			redis.WithPrefix(opt.redisConfig.Prefix),
			redis.WithAddress(opt.redisConfig.Address),
			redis.WithPassword(opt.redisConfig.Password),
		)
	}

	rdsCache := &Cache{
		redis:  redisManager,
		prefix: opt.prefix,
	}

	return rdsCache, nil
}

func (c Cache) Put(key string, value any, seconds int) error {
	return c.redis.Set(c.prefix+key, value, seconds)
}

func (c Cache) Add(key string, value any, seconds int) error {
	if !c.Has(key) {
		return c.redis.Set(c.prefix+key, value, seconds)
	}

	return nil
}

func (c Cache) Get(key string) (any, error) {
	bytes, err := c.redis.Get(c.prefix + key)
	if err != nil {
		return nil, err
	}

	var value any
	err = json.Unmarshal(bytes, &value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (c Cache) Pull(key string) (any, error) {
	value, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	ok, err := c.Forget(key)
	if err != nil && !ok {
		return nil, err
	}

	return value, nil
}

func (c Cache) Has(key string) bool {
	exists, _ := c.redis.Exists(c.prefix + key)
	return exists
}

func (c Cache) Forever(key string, value any) error {
	return c.redis.Set(c.prefix+key, value, 0)
}

func (c Cache) Forget(key string) (bool, error) {
	return c.redis.Del(c.prefix + key)
}

func (c Cache) Increment(key string, n int) (int, error) {
	conn := c.redis.ConnPool.Get()
	defer conn.Close()

	return redigo.Int(conn.Do("INCRBY", c.prefix+key, n))
}

func (c Cache) Decrement(key string, n int) (int, error) {
	conn := c.redis.ConnPool.Get()
	defer conn.Close()

	return redigo.Int(conn.Do("DECRBY", c.prefix+key, n))
}

func (c Cache) Flush() error {
	return c.redis.BatchDel(c.prefix)
}
