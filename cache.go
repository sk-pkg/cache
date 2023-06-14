package cache

import (
	"github.com/sk-pkg/cache/mem"
	"github.com/sk-pkg/cache/redis"
	redisManager "github.com/sk-pkg/redis"
)

const (
	MemCache      = "mem"
	RedisCache    = "redis"
	DefaultPrefix = "go_cache:"
)

type Manager interface {
	// Put 缓存中存储数据
	Put(key string, value interface{}, seconds int) error

	// Add 方法将只存储缓存中不存在的数据。
	Add(key string, value interface{}, seconds int) error

	// Get 方法用于从缓存中获取数据。
	// 如果该数据在缓存中不存在，那么该方法将返回 nil, nil
	Get(key string) (interface{}, error)

	// Pull 方法用于从缓存中获取到数据之后再删除它。
	// 如果该数据在缓存中不存在，那么该方法将返回 nil, nil
	Pull(key string) (interface{}, error)

	// Has 方法可以用于判断缓存项是否存在。如果值为 nil，则该方法将会返回 false
	Has(key string) bool

	// Forever 方法可用于持久化将数据存储到缓存中。
	// 因为这些数据不会过期，所以必须通过 Forget 方法从缓存中手动删除它们
	Forever(key string, value interface{}) error

	// Forget 方法从缓存中删除这些数据
	Forget(key string) (bool, error)

	// Increment 方法增加指定key的值 int 值
	Increment(key string, n int) (int, error)

	// Decrement 方法减少指定key的值 int 值
	Decrement(key string, n int) (int, error)

	// Flush 方法清空所有的缓存
	Flush() error
}

type Option func(*option)

type option struct {
	driver      string
	prefix      string
	redis       *redisManager.Manager
	redisConfig redis.Config
}

func WithDriver(driver string) Option {
	return func(o *option) {
		o.driver = driver
	}
}

func WithPrefix(prefix string) Option {
	return func(o *option) {
		o.prefix = prefix + ":"
	}
}

func WithRedis(redis *redisManager.Manager) Option {
	return func(o *option) {
		o.redis = redis
	}
}

func WithRedisConfig(redisConfig redis.Config) Option {
	return func(o *option) {
		o.redisConfig = redisConfig
	}
}

func New(opts ...Option) (Manager, error) {
	opt := &option{prefix: DefaultPrefix}
	for _, f := range opts {
		f(opt)
	}

	switch opt.driver {
	case RedisCache:
		return redis.Init(
			redis.WithPrefix(opt.prefix),
			redis.WithRedisConfig(opt.redisConfig),
			redis.WithRedisManager(opt.redis),
		)
	default:
		return mem.Init(), nil
	}
}
