package redis

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

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

// CheckHealth checks if the job queue is alive.
func (rs *RedisClient) CheckHealth() bool {

	randomNum := rand.Intn(10000)
	key := fmt.Sprintf("health:%d", randomNum)

	prefixedKey := rs.GetRedisPrefixedKey(key)

	val, err := rs.Get(ctx, prefixedKey).Result()
	if err != redis.Nil || val != "" {
		return false
	}

	rs.Set(ctx, prefixedKey, "1", 0)
	val, err = rs.Get(ctx, prefixedKey).Result()
	if err != redis.Nil && val != "1" {
		return false
	}

	rs.Del(ctx, prefixedKey)
	val, err = rs.Get(ctx, prefixedKey).Result()
	if err != redis.Nil || val != "" {
		return false
	}

	return true
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
