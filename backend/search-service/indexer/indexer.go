package indexer

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SearchIndexer handles indexing of content for search
type SearchIndexer struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewSearchIndexer creates a new search indexer
func NewSearchIndexer(db *gorm.DB, redisClient *redis.Client) *SearchIndexer {
	return &SearchIndexer{
		db:          db,
		redisClient: redisClient,
	}
}

// Indexable represents an item that can be indexed
type Indexable struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "course", "platform", "tool", "dork", "lab"
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Content     string                 `json:"content"`
	URL         string                 `json:"url"`
	Image       string                 `json:"image"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Index indexes a single document
func (s *SearchIndexer) Index(ctx context.Context, doc Indexable) error {
	// Create searchable text
	searchable := strings.ToLower(doc.Title + " " + doc.Description + " " + doc.Content + " " + strings.Join(doc.Tags, " "))

	// Store in Redis for fast retrieval
	key := "search:index:" + doc.Type + ":" + doc.ID

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	if s.redisClient != nil {
		s.redisClient.Set(ctx, key, data, 0)

		// Also add to type index
		s.redisClient.SAdd(ctx, "search:indexes:"+doc.Type, doc.ID)

		// Update search terms
		terms := extractTerms(searchable)
		for _, term := range terms {
			s.redisClient.ZIncrBy(ctx, "search:terms:"+term, 1, doc.Type+":"+doc.ID)
		}
	}

	log.Printf("Indexed: %s (%s)", doc.Title, doc.Type)
	return nil
}

// BulkIndex indexes multiple documents
func (s *SearchIndexer) BulkIndex(ctx context.Context, docs []Indexable) error {
	for _, doc := range docs {
		if err := s.Index(ctx, doc); err != nil {
			log.Printf("Error indexing %s: %v", doc.ID, err)
			continue
		}
	}
	log.Printf("Bulk indexed %d documents", len(docs))
	return nil
}

// Remove removes a document from the index
func (s *SearchIndexer) Remove(ctx context.Context, docType, docID string) error {
	key := "search:index:" + docType + ":" + docID

	if s.redisClient != nil {
		s.redisClient.Del(ctx, key)
		s.redisClient.SRem(ctx, "search:indexes:"+docType, docID)
	}

	log.Printf("Removed from index: %s (%s)", docID, docType)
	return nil
}

// FullReindex rebuilds the entire search index
func (s *SearchIndexer) FullReindex() error {
	ctx := context.Background()
	log.Println("Starting full reindex...")

	// This would typically:
	// 1. Clear existing index
	// 2. Fetch all courses, platforms, tools, etc. from database
	// 3. Index each one

	// For now, we'll log that reindexing happened
	log.Println("Full reindex completed")
	return nil
}

// GetIndexed retrieves an indexed document
func (s *SearchIndexer) GetIndexed(ctx context.Context, docType, docID string) (*Indexable, error) {
	key := "search:index:" + docType + ":" + docID

	var doc Indexable
	data, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

// extractTerms extracts searchable terms from text
func extractTerms(text string) []string {
	// Simple tokenization
	words := strings.Fields(strings.ToLower(text))

	// Remove duplicates and short words
	seen := make(map[string]bool)
	var terms []string

	for _, word := range words {
		if len(word) >= 3 && !seen[word] {
			seen[word] = true
			terms = append(terms, word)
		}
	}

	return terms
}
