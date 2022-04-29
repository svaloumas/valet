package redis

import (
	"context"
	"io/ioutil"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

// RedisClient represents a redis client.
type RedisClient struct {
	*redis.Client
	KeyPrefix string
	logger    *logrus.Logger
}

// New returns a redis client.
func New(
	url string, poolSize, minIdleConns int,
	keyPrefix string, logger *logrus.Logger) *RedisClient {

	rs := new(RedisClient)

	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	rs.KeyPrefix = keyPrefix
	if logger != nil {
		rs.logger = logger
	} else {
		rs.logger = &logrus.Logger{Out: ioutil.Discard}
	}

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
	res, err := rs.Ping(ctx).Result()
	if err != nil {
		rs.logger.Errorf("ping returned error: %s", err)
		return false
	}
	return res == "PONG"
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
