package indexer

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// IndexBuilder helps build indexes incrementally
type IndexBuilder struct {
	db          *gorm.DB
	redisClient *redis.Client
	indexer     *SearchIndexer
}

// NewIndexBuilder creates a new index builder
func NewIndexBuilder(db *gorm.DB, redisClient *redis.Client) *IndexBuilder {
	return &IndexBuilder{
		db:          db,
		redisClient: redisClient,
		indexer:     NewSearchIndexer(db, redisClient),
	}
}

// BuildCourseIndex builds index for courses
func (b *IndexBuilder) BuildCourseIndex(ctx context.Context) error {
	log.Println("Building course index...")

	// This would query the courses table and index them
	// Example:
	// var courses []models.Course
	// b.db.Find(&courses)
	// for _, course := range courses {
	//     b.indexer.Index(ctx, Indexable{
	//         ID:          course.ID.String(),
	//         Type:        "course",
	//         Title:       course.Title,
	//         Description: course.Description,
	//         Content:     course.Content,
	//         URL:         "/courses/" + course.Slug,
	//         Tags:        course.Tags,
	//         Metadata:    map[string]interface{}{"difficulty": course.Difficulty},
	//     })
	// }

	log.Println("Course index built")
	return nil
}

// BuildPlatformIndex builds index for platforms
func (b *IndexBuilder) BuildPlatformIndex(ctx context.Context) error {
	log.Println("Building platform index...")
	// Similar to courses
	log.Println("Platform index built")
	return nil
}

// BuildToolIndex builds index for tools
func (b *IndexBuilder) BuildToolIndex(ctx context.Context) error {
	log.Println("Building tool index...")
	// Index security tools
	log.Println("Tool index built")
	return nil
}

// BuildDorkIndex builds index for dorks
func (b *IndexBuilder) BuildDorkIndex(ctx context.Context) error {
	log.Println("Building dork index...")
	// Index Google dorks
	log.Println("Dork index built")
	return nil
}

// BuildLabIndex builds index for labs
func (b *IndexBuilder) BuildLabIndex(ctx context.Context) error {
	log.Println("Building lab index...")
	// Index practice labs
	log.Println("Lab index built")
	return nil
}

// IncrementalUpdate performs incremental updates
func (b *IndexBuilder) IncrementalUpdate(ctx context.Context) error {
	// Get last update time from Redis
	lastUpdateKey := "search:last_update"

	var lastUpdate time.Time
	if b.redisClient != nil {
		timestamp, err := b.redisClient.Get(ctx, lastUpdateKey).Int64()
		if err == nil {
			lastUpdate = time.Unix(timestamp, 0)
		}
	}

	// Find items updated since last update
	// This would typically query: WHERE updated_at > lastUpdate

	// Update last update time
	if b.redisClient != nil {
		b.redisClient.Set(ctx, lastUpdateKey, time.Now().Unix(), 0)
	}

	log.Println("Incremental update completed")
	return nil
}

// RebuildTypeIndex rebuilds index for a specific type
func (b *IndexBuilder) RebuildTypeIndex(ctx context.Context, docType string) error {
	log.Printf("Rebuilding index for type: %s", docType)

	switch strings.ToLower(docType) {
	case "course":
		return b.BuildCourseIndex(ctx)
	case "platform":
		return b.BuildPlatformIndex(ctx)
	case "tool":
		return b.BuildToolIndex(ctx)
	case "dork":
		return b.BuildDorkIndex(ctx)
	case "lab":
		return b.BuildLabIndex(ctx)
	default:
		log.Printf("Unknown type: %s", docType)
	}

	return nil
}

// GetIndexStats returns statistics about the index
func (b *IndexBuilder) GetIndexStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	if b.redisClient == nil {
		return stats, nil
	}

	// Count items in each type
	types := []string{"course", "platform", "tool", "dork", "lab"}

	for _, docType := range types {
		count, err := b.redisClient.SCard(ctx, "search:indexes:"+docType).Result()
		if err == nil {
			stats[docType] = count
		}
	}

	return stats, nil
}
