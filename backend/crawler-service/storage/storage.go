package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Storage handles data persistence
type Storage struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewStorage creates a new storage instance
func NewStorage(db *gorm.DB, redisClient *redis.Client) *Storage {
	return &Storage{
		db:          db,
		redisClient: redisClient,
	}
}

// CrawlResult represents a stored crawl result
type CrawlResult struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	JobID       string    `gorm:"index" json:"job_id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Links       string    `json:"links"` // JSON array
	Images      string    `json:"images"` // JSON array
	Meta        string    `json:"meta"`   // JSON object
	Depth       int       `json:"depth"`
	StatusCode  int       `json:"status_code"`
	CrawledAt  time.Time `json:"crawled_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// SaveResult saves a crawl result
Storage) SaveResultfunc (s *(ctx context.Context, result CrawlResult) error {
	result.CreatedAt = time.Now()

	if s.db != nil {
		return s.db.Create(&result).Error
	}

	// Fall back to Redis
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("crawl:result:%s:%d", result.JobID, result.ID)
	return s.redisClient.Set(ctx, key, data, 0).Err()
}

// GetResultsByJobID retrieves results for a job
func (s *Storage) GetResultsByJobID(ctx context.Context, jobID string, limit, offset int) ([]CrawlResult, error) {
	if s.db != nil {
		var results []CrawlResult
		err := s.db.Where("job_id = ?", jobID).Limit(limit).Offset(offset).Find(&results).Error
		return results, err
	}

	// Fall back to Redis - get keys matching pattern
	pattern := fmt.Sprintf("crawl:result:%s:*", jobID)
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var results []CrawlResult
	for i, key := range keys {
		if i >= limit+offset {
			break
		}
		if i < offset {
			continue
		}

		data, err := s.redisClient.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var result CrawlResult
		if err := json.Unmarshal(data, &result); err == nil {
			results = append(results, result)
		}
	}

	return results, nil
}

// CountResults counts results for a job
func (s *Storage) CountResults(ctx context.Context, jobID string) (int64, error) {
	if s.db != nil {
		var count int64
		err := s.db.Model(&CrawlResult{}).Where("job_id = ?", jobID).Count(&count).Error
		return count, err
	}

	// Fall back to Redis
	pattern := fmt.Sprintf("crawl:result:%s:*", jobID)
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}

	return int64(len(keys)), nil
}

// DeleteResults deletes results for a job
func (s *Storage) DeleteResults(ctx context.Context, jobID string) error {
	if s.db != nil {
		return s.db.Where("job_id = ?", jobID).Delete(&CrawlResult{}).Error
	}

	// Fall back to Redis
	pattern := fmt.Sprintf("crawl:result:%s:*", jobID)
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.redisClient.Del(ctx, keys...).Err()
	}

	return nil
}

// GetStats returns storage statistics
func (s *Storage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	if s.db != nil {
		var totalResults int64
		s.db.Model(&CrawlResult{}).Count(&totalResults)
		stats["total_results"] = totalResults
		var totalJobs int64
		s.db.Model(&CrawlResult{}).Distinct("job_id").Count(&totalJobs)
		stats["total_jobs"] = totalJobs
	} else {
		// Redis stats
		keys, _ := s.redisClient.Keys(ctx, "crawl:result:*").Result()
		stats["total_results"] = len(keys)
	}

	return stats, nil
}

// AutoMigrate creates the tables
func (s *Storage) AutoMigrate() error {
	if s.db == nil {
		return fmt.Errorf("database not configured")
	}
	return s.db.AutoMigrate(&CrawlResult{})
}

// SaveToCache saves data to Redis cache
func (s *Storage) SaveToCache(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	serialized, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.redisClient.Set(ctx, key, serialized, ttl).Err()
}

// GetFromCache retrieves data from Redis cache
func (s *Storage) GetFromCache(ctx context.Context, key string, dest interface{}) error {
	data, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
