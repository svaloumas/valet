package redis

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/svaloumas/valet/internal/core/domain"
	"github.com/svaloumas/valet/internal/core/port"
	"github.com/svaloumas/valet/pkg/apperrors"
	rs "github.com/svaloumas/valet/pkg/redis"
)

var _ port.Storage = &Redis{}
var ctx = context.Background()

// Redis represents a redis client.
type Redis struct {
	*rs.RedisClient
}

// New returns a redis client.
func New(url string, poolSize, minIdleConns int, keyPrefix string) *Redis {
	client := rs.New(url, poolSize, minIdleConns, keyPrefix)
	rs := &Redis{
		RedisClient: client,
	}
	return rs
}

// CreateJob adds a new job to the repository.
func (rs *Redis) CreateJob(j *domain.Job) error {
	key := rs.getRedisKeyForJob(j.ID)
	value, err := json.Marshal(j)
	if err != nil {
		return err
	}

	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetJob fetches a job from the repository.
func (rs *Redis) GetJob(id string) (*domain.Job, error) {
	key := rs.getRedisKeyForJob(id)
	val, err := rs.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "job"}
		}
		return nil, err
	}

	var j *domain.Job
	err = json.Unmarshal(val, &j)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// GetJobs fetches all jobs from the repository, optionally filters the jobs by status.
func (rs *Redis) GetJobs(status domain.JobStatus) ([]*domain.Job, error) {
	var keys []string
	key := rs.GetRedisPrefixedKey("job:*")
	iter := rs.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	jobs := []*domain.Job{}
	for _, key := range keys {
		value, err := rs.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		j := &domain.Job{}
		if err := json.Unmarshal(value, j); err != nil {
			return nil, err
		}
		if status == domain.Undefined || j.Status == status {
			jobs = append(jobs, j)
		}
	}

	// ORDER BY created_at ASC
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(*jobs[j].CreatedAt)
	})
	return jobs, nil
}

// GetJobsByPipelineID fetches the jobs of the specified pipeline.
func (rs *Redis) GetJobsByPipelineID(pipelineID string) ([]*domain.Job, error) {
	var keys []string
	key := rs.GetRedisPrefixedKey("job:*")
	iter := rs.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	jobs := []*domain.Job{}
	for _, key := range keys {
		value, err := rs.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		j := &domain.Job{}
		if err := json.Unmarshal(value, j); err != nil {
			return nil, err
		}
		if j.PipelineID == pipelineID {
			jobs = append(jobs, j)
		}
	}

	// ORDER BY created_at ASC
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(*jobs[j].CreatedAt)
	})
	return jobs, nil
}

// UpdateJob adds a new job to the repository.
func (rs *Redis) UpdateJob(id string, j *domain.Job) error {
	key := rs.getRedisKeyForJob(id)
	value, err := json.Marshal(j)
	if err != nil {
		return err
	}

	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// DeleteJob deletes a job from the repository.
func (rs *Redis) DeleteJob(id string) error {
	key := rs.getRedisKeyForJob(id)
	_, err := rs.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
func (rs *Redis) GetDueJobs() ([]*domain.Job, error) {
	var keys []string
	key := rs.GetRedisPrefixedKey("job:*")
	iter := rs.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	dueJobs := []*domain.Job{}
	for _, key := range keys {
		value, err := rs.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		j := &domain.Job{}
		if err := json.Unmarshal(value, j); err != nil {
			return nil, err
		}
		if j.IsScheduled() {
			if j.RunAt.Before(time.Now()) && j.Status == domain.Pending {
				dueJobs = append(dueJobs, j)
			}
		}
	}

	// ORDER BY run_at ASC
	sort.Slice(dueJobs, func(i, j int) bool {
		return dueJobs[i].RunAt.Before(*dueJobs[j].RunAt)
	})
	return dueJobs, nil
}

// CreateJobResult adds a new job result to the repository.
func (rs *Redis) CreateJobResult(result *domain.JobResult) error {
	key := rs.getRedisKeyForJobResult(result.JobID)
	value, err := json.Marshal(result)
	if err != nil {
		return err
	}

	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetJobResult fetches a job result from the repository.
func (rs *Redis) GetJobResult(jobID string) (*domain.JobResult, error) {
	key := rs.getRedisKeyForJobResult(jobID)
	val, err := rs.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, &apperrors.NotFoundErr{ID: jobID, ResourceName: "job result"}
		}
		return nil, err
	}

	var result *domain.JobResult
	err = json.Unmarshal(val, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateJobResult updates a job result to the repository.
func (rs *Redis) UpdateJobResult(jobID string, result *domain.JobResult) error {
	key := rs.getRedisKeyForJobResult(jobID)
	value, err := json.Marshal(result)
	if err != nil {
		return err
	}

	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// DeleteJobResult deletes a job result from the repository.
func (rs *Redis) DeleteJobResult(jobID string) error {
	key := rs.getRedisKeyForJobResult(jobID)
	_, err := rs.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

// CreatePipeline adds a new pipeline and of its jobs to the repository.
func (rs *Redis) CreatePipeline(p *domain.Pipeline) error {
	err := rs.Watch(ctx, func(tx *redis.Tx) error {

		for _, j := range p.Jobs {
			key := rs.getRedisKeyForJob(j.ID)
			value, err := json.Marshal(j)
			if err != nil {
				return err
			}

			err = rs.Set(ctx, key, value, 0).Err()
			if err != nil {
				return err
			}
		}

		key := rs.getRedisKeyForPipeline(p.ID)
		value, err := json.Marshal(p)
		if err != nil {
			return err
		}

		err = rs.Set(ctx, key, value, 0).Err()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// GetPipeline fetches a pipeline from the repository.
func (rs *Redis) GetPipeline(id string) (*domain.Pipeline, error) {
	key := rs.getRedisKeyForPipeline(id)
	val, err := rs.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, &apperrors.NotFoundErr{ID: id, ResourceName: "pipeline"}
		}
		return nil, err
	}

	var p *domain.Pipeline
	err = json.Unmarshal(val, &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetPipelines fetches all pipelines from the repository, optionally filters the pipelines by status.
func (rs *Redis) GetPipelines(status domain.JobStatus) ([]*domain.Pipeline, error) {
	var keys []string
	key := rs.GetRedisPrefixedKey("pipeline:*")
	iter := rs.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	pipelines := []*domain.Pipeline{}
	for _, key := range keys {
		value, err := rs.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		p := &domain.Pipeline{}
		if err := json.Unmarshal(value, p); err != nil {
			return nil, err
		}
		if status == domain.Undefined || p.Status == status {
			pipelines = append(pipelines, p)
		}
	}

	// ORDER BY created_at ASC
	sort.Slice(pipelines, func(i, j int) bool {
		return pipelines[i].CreatedAt.Before(*pipelines[j].CreatedAt)
	})
	return pipelines, nil

}

// UpdatePipeline updates a pipeline to the repository.
func (rs *Redis) UpdatePipeline(id string, p *domain.Pipeline) error {
	key := rs.getRedisKeyForPipeline(id)
	value, err := json.Marshal(p)
	if err != nil {
		return err
	}

	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// DeletePipeline deletes a pipeline and all its jobs from the repository.
func (rs *Redis) DeletePipeline(id string) error {
	err := rs.Watch(ctx, func(tx *redis.Tx) error {
		var keys []string
		key := rs.GetRedisPrefixedKey("job:*")
		iter := rs.Scan(ctx, 0, key, 0).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return err
		}

		jobs := []*domain.Job{}
		for _, key := range keys {
			value, err := rs.Get(ctx, key).Bytes()
			if err != nil {
				return err
			}
			j := &domain.Job{}
			if err := json.Unmarshal(value, j); err != nil {
				return err
			}
			if j.PipelineID == id {
				jobs = append(jobs, j)
			}
		}
		for _, j := range jobs {
			key = rs.getRedisKeyForJobResult(j.ID)
			_, err := rs.Del(ctx, key).Result()
			if err != nil {
				return err
			}
			key = rs.getRedisKeyForJob(j.ID)
			_, err = rs.Del(ctx, key).Result()
			if err != nil {
				return err
			}
		}
		key = rs.getRedisKeyForPipeline(id)
		_, err := rs.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (rs *Redis) getRedisKeyForPipeline(pipelineID string) string {
	return rs.GetRedisPrefixedKey("pipeline:" + pipelineID)
}

func (rs *Redis) getRedisKeyForJob(jobID string) string {
	return rs.GetRedisPrefixedKey("job:" + jobID)
}

func (rs *Redis) getRedisKeyForJobResult(jobID string) string {
	return rs.GetRedisPrefixedKey("jobresult:" + jobID)
}
