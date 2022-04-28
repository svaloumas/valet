package redis

import (
	"github.com/go-redis/redis/v8"
)

// RedisClient represents a redis client.
type RedisClient struct {
	*redis.Client
	KeyPrefix string
}

// New returns a redis client.
func New(url string, poolSize, minIdleConns int, keyPrefix string) *RedisClient {
	rs := new(RedisClient)

	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	rs.KeyPrefix = keyPrefix

	rs.Client = redis.NewClient(&redis.Options{
		Addr:         opt.Addr,
		Password:     opt.Password,
		DB:           opt.DB,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		TLSConfig:    opt.TLSConfig,
	})
	return rs
}

// Close terminates any storage connections gracefully.
func (rs *RedisClient) Close() error {
	return rs.Client.Close()
}

func (rs *RedisClient) GetRedisPrefixedKey(key string) string {
	if rs.KeyPrefix != "" {
		return rs.KeyPrefix + ":" + key
	}
	return key
}
