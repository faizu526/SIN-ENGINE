package ranker

import (
	"math"
	"strings"
	"time"
)

// Scorer calculates various relevance scores
type Scorer struct{}

// NewScorer creates a new scorer
func NewScorer() *Scorer {
	return &Scorer{}
}

// ScoreTextMatch calculates text match score
func (s *Scorer) ScoreTextMatch(query, text string) float64 {
	queryTerms := strings.Fields(strings.ToLower(query))
	textLower := strings.ToLower(text)

	if len(queryTerms) == 0 {
		return 0
	}

	matches := 0
	for _, term := range queryTerms {
		if strings.Contains(textLower, term) {
			matches++
		}
	}

	return float64(matches) / float64(len(queryTerms))
}

// ScoreTitleBoost applies boost for title matches
func (s *Scorer) ScoreTitleBoost(query, title string) float64 {
	queryTerms := strings.Fields(strings.ToLower(query))
	titleLower := strings.ToLower(title)

	matches := 0
	for _, term := range queryTerms {
		if strings.Contains(titleLower, term) {
			matches++
		}
	}

	// Higher weight for title matches
	if matches > 0 {
		return float64(matches) * 0.5
	}
	return 0
}

// ScorePopularity calculates popularity-based score
func (s *Scorer) ScorePopularity(views int64, likes int64, shares int64) float64 {
	// Weighted combination of engagement metrics
	score := float64(views)*0.1 + float64(likes)*0.3 + float64(shares)*0.5

	// Normalize to 0-1 range using log scale
	if score > 0 {
		score = math.Log(score+1) / 10
		if score > 1 {
			score = 1
		}
	}

	return score
}

// ScoreRecency calculates recency-based score
func (s *Scorer) ScoreRecency(createdAt time.Time) float64 {
	age := time.Since(createdAt)
	days := age.Hours() / 24

	// Exponential decay based on age
	// 0 days = 1.0, 30 days = 0.36, 90 days = 0.04
	score := math.Exp(-days / 30)

	return score
}

// ScoreCategoryMatch calculates category match score
func (s *Scorer) ScoreCategoryMatch(queryCategory, itemCategory string) float64 {
	if strings.EqualFold(queryCategory, itemCategory) {
		return 1.0
	}

	// Check for related categories
	related := map[string][]string{
		"web":     {"webapp", "website", "application"},
		"network": {"networking", "infrastructure", "network"},
		"mobile":  {"android", "ios", "app"},
	}

	if relatedCats, ok := related[strings.ToLower(queryCategory)]; ok {
		for _, cat := range relatedCats {
			if strings.EqualFold(cat, itemCategory) {
				return 0.7
			}
		}
	}

	return 0
}

// ScoreDifficultyMatch calculates difficulty match score
func (s *Scorer) ScoreDifficultyMatch(userLevel, contentLevel string) float64 {
	levels := map[string]int{
		"beginner":     1,
		"easy":         1,
		"basic":        1,
		"intermediate": 2,
		"medium":       2,
		"advanced":     3,
		"hard":         3,
		"expert":       4,
	}

	userLevelVal, userOk := levels[strings.ToLower(userLevel)]
	contentLevelVal, contentOk := levels[strings.ToLower(contentLevel)]

	if !userOk || !contentOk {
		return 0.5 // Default if unknown
	}

	// Perfect match
	if userLevelVal == contentLevelVal {
		return 1.0
	}

	// One level off
	if math.Abs(float64(userLevelVal-contentLevelVal)) == 1 {
		return 0.7
	}

	// More than one level off
	return 0.3
}

// ScoreTagOverlap calculates tag overlap score
func (s *Scorer) ScoreTagOverlap(queryTags, itemTags []string) float64 {
	if len(queryTags) == 0 || len(itemTags) == 0 {
		return 0
	}

	queryLower := toLowerSlice(queryTags)
	itemLower := toLowerSlice(itemTags)

	matches := 0
	for _, qt := range queryLower {
		for _, it := range itemLower {
			if qt == it {
				matches++
				break
			}
		}
	}

	return float64(matches) / float64(len(queryLower))
}

// ScoreClickThrough calculates CTR-based score
func (s *Scorer) ScoreClickThrough(clicks, impressions int64) float64 {
	if impressions == 0 {
		return 0.5 // Default
	}

	ctr := float64(clicks) / float64(impressions)

	// Scale CTR (typically 0-10%) to 0-1
	score := ctr * 10
	if score > 1 {
		score = 1
	}

	return score
}

// CalculateFinalScore combines all scoring components
func (s *Scorer) CalculateFinalScore(components ScoreComponents) float64 {
	weights := map[string]float64{
		"textMatch":  0.30,
		"titleBoost": 0.20,
		"popularity": 0.20,
		"recency":    0.15,
		"category":   0.05,
		"difficulty": 0.05,
		"tags":       0.05,
	}

	score := 0.0
	score += components.TextMatch * weights["textMatch"]
	score += components.TitleBoost * weights["titleBoost"]
	score += components.Popularity * weights["popularity"]
	score += components.Recency * weights["recency"]
	score += components.Category * weights["category"]
	score += components.Difficulty * weights["difficulty"]
	score += components.Tags * weights["tags"]

	return score
}

// ScoreComponents holds all scoring components
type ScoreComponents struct {
	TextMatch  float64
	TitleBoost float64
	Popularity float64
	Recency    float64
	Category   float64
	Difficulty float64
	Tags       float64
}

// toLowerSlice converts string slice to lowercase
func toLowerSlice(slice []string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = strings.ToLower(s)
	}
	return result
}
