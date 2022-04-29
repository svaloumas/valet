package jobqueue

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/svaloumas/valet/internal/config"
	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/pkg/log"
	rs "github.com/svaloumas/valet/pkg/redis"
)

var _ port.JobQueue = &redisqueue{}
var ctx = context.Background()

type redisqueue struct {
	*rs.RedisClient
	logger *logrus.Logger
}

// NewRedisQueue returns a redis queue.
func NewRedisQueue(cfg config.Redis, loggingFormat string) *redisqueue {
	logger := log.NewLogger("redisqueue", loggingFormat)
	client := rs.New(cfg.URL, cfg.PoolSize, cfg.MinIdleConns, cfg.KeyPrefix, logger)
	rs := &redisqueue{
		RedisClient: client,
		logger:      logger,
	}
	return rs
}

func (q *redisqueue) Push(j *domain.Job) error {
	key := q.GetRedisPrefixedKey("job-queue")
	value, err := json.Marshal(j)
	if err != nil {
		q.logger.Errorf("could not marshal job: %s", err)
		return err
	}
	_, err = q.LPush(ctx, key, value).Result()
	if err != nil {
		q.logger.Errorf("error while LPUSH job message: %s", err)
		return err
	}
	return nil
}

func (q *redisqueue) Pop() *domain.Job {
	key := q.GetRedisPrefixedKey("job-queue")
	val, err := q.RPop(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			q.logger.Errorf("could not RPOP job message: %s", err)
		}
		return nil
	}
	var j *domain.Job
	err = json.Unmarshal(val, &j)
	if err != nil {
		q.logger.Errorf("could not unmarshal message body: %s", err)
		return nil
	}
	return j
}

func (q *redisqueue) Close() {
	q.RedisClient.Close()
}
