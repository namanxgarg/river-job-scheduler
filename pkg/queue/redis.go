package queue

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/namangarg/job-scheduler/pkg/job"
	"github.com/sirupsen/logrus"
)

// RedisQueue implements the Queue interface using Redis
type RedisQueue struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisQueue creates a new Redis queue instance
func NewRedisQueue(addr string, password string, db int) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisQueue{
		client: client,
		logger: logrus.New(),
	}, nil
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, j *job.Job) error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}

	// Use Redis Sorted Set for scheduled jobs
	if j.ScheduledFor.After(time.Now()) {
		return q.client.ZAdd(ctx, "scheduled_jobs", &redis.Z{
			Score:  float64(j.ScheduledFor.Unix()),
			Member: data,
		}).Err()
	}

	// Use Redis List for immediate jobs
	return q.client.RPush(ctx, q.queueKey(j.Queue), data).Err()
}

// Dequeue retrieves a job from the queue
func (q *RedisQueue) Dequeue(ctx context.Context, queueName string) (*job.Job, error) {
	// First, check for scheduled jobs that are due
	now := time.Now().Unix()
	scheduledJobs, err := q.client.ZRangeByScore(ctx, "scheduled_jobs", &redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatInt(now, 10),
		Offset: 0,
		Count:  1,
	}).Result()

	if err != nil {
		return nil, err
	}

	if len(scheduledJobs) > 0 {
		// Remove the job from scheduled set
		err = q.client.ZRem(ctx, "scheduled_jobs", scheduledJobs[0]).Err()
		if err != nil {
			return nil, err
		}

		var j job.Job
		if err := json.Unmarshal([]byte(scheduledJobs[0]), &j); err != nil {
			return nil, err
		}
		return &j, nil
	}

	// If no scheduled jobs, check the regular queue
	data, err := q.client.BLPop(ctx, 0, q.queueKey(queueName)).Result()
	if err != nil {
		return nil, err
	}

	if len(data) < 2 {
		return nil, nil
	}

	var j job.Job
	if err := json.Unmarshal([]byte(data[1]), &j); err != nil {
		return nil, err
	}

	return &j, nil
}

// Retry adds a job back to the queue for retry
func (q *RedisQueue) Retry(ctx context.Context, j *job.Job) error {
	j.Status = job.StatusRetrying
	j.RetryCount++
	j.ScheduledFor = time.Now().Add(j.NextRetryDelay())
	return q.Enqueue(ctx, j)
}

// Remove removes a job from the queue
func (q *RedisQueue) Remove(ctx context.Context, jobID string) error {
	// Implementation depends on how you want to handle job removal
	// This is a simplified version
	return q.client.Del(ctx, "job:"+jobID).Err()
}

// queueKey returns the Redis key for a queue
func (q *RedisQueue) queueKey(queueName string) string {
	return "queue:" + queueName
}

// Close closes the Redis connection
func (q *RedisQueue) Close() error {
	return q.client.Close()
}
