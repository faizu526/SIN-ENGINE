package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// JobQueue manages crawl job queue using Redis
type JobQueue struct {
	redisClient *redis.Client
	key         string
}

// NewJobQueue creates a new job queue
func NewJobQueue(redisClient *redis.Client) *JobQueue {
	return &JobQueue{
		redisClient: redisClient,
		key:         "crawler:jobs",
	}
}

// Job represents a crawl job
type Job struct {
	ID          string                 `json:"id"`
	URL         string                 `json:"url"`
	Depth       int                    `json:"depth"`
	MaxDepth    int                    `json:"max_depth"`
	UserAgent   string                 `json:"user_agent"`
	Concurrent  int                    `json:"concurrent"`
	Timeout     int                    `json:"timeout"`
	Filters     []string               `json:"filters"`
	Priority    int                    `json:"priority"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Push adds a job to the queue
func (q *JobQueue) Push(ctx context.Context, job Job) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// Use sorted set with priority as score
	score := float64(job.Priority)
	if !job.ScheduledAt.IsZero() {
		score = float64(job.ScheduledAt.Unix())
	}

	err = q.redisClient.ZAdd(ctx, q.key, redis.Z{
		Score:  score,
		Member: jobData,
	}).Err()

	if err != nil {
		log.Printf("Error pushing job to queue: %v", err)
		return err
	}

	log.Printf("Job %s pushed to queue with priority %d", job.ID, job.Priority)
	return nil
}

// Pop removes and returns the highest priority job
func (q *JobQueue) Pop(ctx context.Context) ([]byte, error) {
	// Get highest priority job (highest score)
	result, err := q.redisClient.ZPopMax(ctx, q.key).Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	jobData, ok := result[0].Member.(string)
	if !ok {
		return nil, nil
	}

	return []byte(jobData), nil
}

// Peek returns the highest priority job without removing it
func (q *JobQueue) Peek(ctx context.Context) (*Job, error) {
	result, err := q.redisClient.ZRevRange(ctx, q.key, 0, 0).Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	var job Job
	if err := json.Unmarshal([]byte(result[0]), &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// Len returns the number of jobs in the queue
func (q *JobQueue) Len(ctx context.Context) (int64, error) {
	return q.redisClient.ZCard(ctx, q.key).Result()
}

// Clear removes all jobs from the queue
func (q *JobQueue) Clear(ctx context.Context) error {
	return q.redisClient.Del(ctx, q.key).Err()
}

// GetStats returns queue statistics
func (q *JobQueue) GetStats(ctx context.Context) (map[string]interface{}, error) {
	length, err := q.Len(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"length": length,
		"key":    q.key,
		"paused": false,
	}, nil
}
