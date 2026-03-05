package ranker

import (
	"context"
	"math"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RankerEngine handles result ranking
type RankerEngine struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewRankerEngine creates a new ranker engine
func NewRankerEngine(db *gorm.DB, redisClient *redis.Client) *RankerEngine {
	return &RankerEngine{
		db:          db,
		redisClient: redisClient,
	}
}

// Rank ranks search results
func (r *RankerEngine) Rank(results []Rankable) ([]RankedResult, error) {
	var ranked []RankedResult

	for _, result := range results {
		score := r.calculateScore(result)
		ranked = append(ranked, RankedResult{
			Result: result,
			Score:  score,
		})
	}

	// Sort by score descending
	r.sortByScore(ranked)

	return ranked, nil
}

// Rerank re-ranks search results based on query
func (r *RankerEngine) Rerank(results []SearchResultItem, query string) ([]SearchResultItem, error) {
	// Get user context if available
	ctx := context.Background()

	// Boost factors
	boosts := r.getBoosts(ctx, query)

	for i := range results {
		baseScore := results[i].Score

		// Apply boosts
		boost := 1.0

		// Popularity boost
		if pop, ok := boosts["popularity"]; ok {
			boost += pop * 0.1
		}

		// Recency boost
		if rec, ok := boosts["recency"]; ok {
			boost += rec * 0.05
		}

		results[i].Score = baseScore * boost
	}

	// Sort by new scores
	r.sortResults(results)

	return results, nil
}

// calculateScore calculates a relevance score for a result
func (r *RankerEngine) calculateScore(result Rankable) float64 {
	score := result.GetBaseScore()

	// Title match boost
	if result.GetTitleMatch() {
		score *= 1.5
	}

	// Tag match boost
	if result.GetTagMatch() {
		score *= 1.2
	}

	// Popularity factor
	score += result.GetPopularity() * 0.1

	// Recency factor
	score += result.GetRecencyScore()

	return score
}

// getBoosts gets boost factors for a query
func (r *RankerEngine) getBoosts(ctx context.Context, query string) map[string]float64 {
	boosts := make(map[string]float64)

	if r.redisClient == nil {
		return boosts
	}

	// Get popularity score
	popKey := "search:popularity:" + query
	popularity, err := r.redisClient.Get(ctx, popKey).Float64()
	if err == nil {
		boosts["popularity"] = popularity
	}

	// Get recency score
	recKey := "search:recency:" + query
	recency, err := r.redisClient.Get(ctx, recKey).Float64()
	if err == nil {
		boosts["recency"] = recency
	}

	return boosts
}

// sortByScore sorts results by score
func (r *RankerEngine) sortByScore(results []RankedResult) {
	// Simple bubble sort for now
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// sortResults sorts search results by score
func (r *RankerEngine) sortResults(results []SearchResultItem) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// Rankable interface for items that can be ranked
type Rankable interface {
	GetBaseScore() float64
	GetTitleMatch() bool
	GetTagMatch() bool
	GetPopularity() float64
	GetRecencyScore() float64
}

// RankedResult represents a ranked search result
type RankedResult struct {
	Result Rankable
	Score  float64
}

// SearchResultItem represents a search result for ranking
type SearchResultItem struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	URL         string                 `json:"url"`
	Image       string                 `json:"image"`
	Score       float64                `json:"score"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GetBaseScore returns the base score
func (s *SearchResultItem) GetBaseScore() float64 {
	return s.Score
}

// GetTitleMatch returns if query matched in title
func (s *SearchResultItem) GetTitleMatch() bool {
	return false // Would be calculated during search
}

// GetTagMatch returns if query matched in tags
func (s *SearchResultItem) GetTagMatch() bool {
	return false
}

// GetPopularity returns popularity score
func (s *SearchResultItem) GetPopularity() float64 {
	if pop, ok := s.Metadata["popularity"].(float64); ok {
		return pop
	}
	return 0
}

// GetRecencyScore returns recency score
func (s *SearchResultItem) GetRecencyScore() float64 {
	// Higher score for newer content
	// This would use actual timestamps
	return math.Random() * 0.5
}
