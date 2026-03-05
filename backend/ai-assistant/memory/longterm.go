package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// KnowledgeItem represents a piece of learned knowledge
type KnowledgeItem struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index"`
	Topic     string    `json:"topic" gorm:"index"`
	Content   string    `json:"content"`
	Source    string    `json:"source"`
	Tags      string    `json:"tags"` // JSON array stored as string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name
func (KnowledgeItem) TableName() return "knowledge_items"

// LongTermMemory manages long-term learning memory (PostgreSQL + Redis)
type LongTermMemory struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewLongTermMemory creates a new long-term memory
func NewLongTermMemory(db *gorm.DB, redis *redis.Client) *LongTermMemory {
	m := &LongTermMemory{
		db:    db,
		redis: redis,
	}

	// Auto-migrate
	if db != nil {
		db.AutoMigrate(&KnowledgeItem{})
	}

	return m
}

// Store stores a knowledge item
func (m *LongTermMemory) Store(ctx context.Context, userID string, item KnowledgeItem) error {
	item.UserID = userID
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	if m.db != nil {
		if err := m.db.Create(&item).Error; err != nil {
			return err
		}
	}

	// Also cache in Redis for fast retrieval
	if m.redis != nil {
		cacheKey := fmt.Sprintf("knowledge:%s:%s", userID, item.Topic)
		data, _ := json.Marshal(item)
		m.redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}

	return nil
}

// Search searches for relevant knowledge
func (m *LongTermMemory) Search(ctx context.Context, query string, limit int) ([]KnowledgeItem, error) {
	// First try Redis cache
	if m.redis != nil {
		cacheKey := fmt.Sprintf("knowledge:search:%s", query)
		data, err := m.redis.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var items []KnowledgeItem
			if json.Unmarshal(data, &items) == nil {
				return items, nil
			}
		}
	}

	// Query database
	if m.db == nil {
		return []KnowledgeItem{}, nil
	}

	var items []KnowledgeItem
	err := m.db.Where("topic ILIKE ? OR content ILIKE ?", 
		"%"+query+"%", "%"+query+"%").
		Limit(limit).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	// Cache results
	if m.redis != nil && len(items) > 0 {
		cacheKey := fmt.Sprintf("knowledge:search:%s", query)
		data, _ := json.Marshal(items)
		m.redis.Set(ctx, cacheKey, data, 1*time.Hour)
	}

	return items, nil
}

// Get gets knowledge by topic
func (m *LongTermMemory) Get(ctx context.Context, userID, topic string) (*KnowledgeItem, error) {
	// Try cache first
	if m.redis != nil {
		cacheKey := fmt.Sprintf("knowledge:%s:%s", userID, topic)
		data, err := m.redis.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var item KnowledgeItem
			if json.Unmarshal(data, &item) == nil {
				return &item, nil
			}
		}
	}

	if m.db == nil {
		return nil, nil
	}

	var item KnowledgeItem
	err := m.db.Where("user_id = ? AND topic = ?", userID, topic).First(&item).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// GetAll gets all knowledge for a user
func (m *LongTermMemory) GetAll(ctx context.Context, userID string) ([]KnowledgeItem, error) {
	if m.db == nil {
		return []KnowledgeItem{}, nil
	}

	var items []KnowledgeItem
	err := m.db.Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

// Delete deletes knowledge by topic
func (m *LongTermMemory) Delete(ctx context.Context, userID, topic string) error {
	if m.db != nil {
		if err := m.db.Where("user_id = ? AND topic = ?", userID, topic).Delete(&KnowledgeItem{}).Error; err != nil {
			return err
		}
	}

	// Invalidate cache
	if m.redis != nil {
		cacheKey := fmt.Sprintf("knowledge:%s:%s", userID, topic)
		m.redis.Del(ctx, cacheKey)
	}

	return nil
}

// Clear clears all knowledge for a user
func (m *LongTermMemory) Clear(ctx context.Context, userID string) error {
	if m.db != nil {
		if err := m.db.Where("user_id = ?", userID).Delete(&KnowledgeItem{}).Error; err != nil {
			return err
		}
	}

	// Clear related cache keys
	if m.redis != nil {
		keys, _ := m.redis.Keys(ctx, fmt.Sprintf("knowledge:%s:*", userID)).Result()
		if len(keys) > 0 {
			m.redis.Del(ctx, keys...)
		}
	}

	return nil
}
