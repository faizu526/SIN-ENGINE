package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Message represents a chat message
type Message struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

// ShortTermMemory manages short-term conversation memory (Redis-backed)
type ShortTermMemory struct {
	client *redis.Client
	limit  int
}

// NewShortTermMemory creates a new short-term memory
func NewShortTermMemory(client *redis.Client, limit int) *ShortTermMemory {
	return &ShortTermMemory{
		client: client,
		limit:  limit,
	}
}

// Add adds a message to short-term memory
func (m *ShortTermMemory) Add(ctx context.Context, userID string, msg Message) error {
	if m.client == nil {
		return nil // No-op if Redis not available
	}

	key := fmt.Sprintf("memory:short:%s", userID)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Add to list
	err = m.client.RPush(ctx, key, string(data)).Err()
	if err != nil {
		return err
	}

	// Trim to limit
	err = m.client.LTrim(ctx, key, -int64(m.limit), -1).Err()
	if err != nil {
		return err
	}

	// Set expiry (24 hours)
	return m.client.Expire(ctx, key, 24*time.Hour).Err()
}

// GetRecent gets recent messages
func (m *ShortTermMemory) GetRecent(ctx context.Context, userID string, count int) ([]Message, error) {
	if m.client == nil {
		return []Message{}, nil
	}

	key := fmt.Sprintf("memory:short:%s", userID)
	results, err := m.client.LRange(ctx, key, 0, int64(count-1)).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]Message, 0, len(results))
	for _, r := range results {
		var msg Message
		if err := json.Unmarshal([]byte(r), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// Clear clears short-term memory for a user
func (m *ShortTermMemory) Clear(ctx context.Context, userID string) error {
	if m.client == nil {
		return nil
	}

	key := fmt.Sprintf("memory:short:%s", userID)
	return m.client.Del(ctx, key).Err()
}

// GetAll gets all messages (for debugging)
func (m *ShortTermMemory) GetAll(ctx context.Context, userID string) ([]Message, error) {
	return m.GetRecent(ctx, userID, m.limit)
}
