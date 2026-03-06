package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func InitDB() error {
	// Check if DATABASE_URL is provided (Render provides this)
	databaseURL := os.Getenv("DATABASE_URL")

	var dsn string

	if databaseURL != "" {
		// Use DATABASE_URL directly
		// Format: postgres://user:password@host:port/database
		dsn = convertDatabaseURL(databaseURL)
		log.Println("Using DATABASE_URL for connection")
	} else {
		// Use individual environment variables
		config := Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "sin_admin"),
			Password: getEnv("DB_PASSWORD", "SinEngine@123"),
			DBName:   getEnv("DB_NAME", "sin_engine"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		}

		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
		)
		log.Println("Using individual DB config for connection")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		log.Printf("⚠️ Database connection failed: %v", err)
		log.Println("Continuing without database...")
		return nil // Don't fail, just continue without DB
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("⚠️ Database pool error: %v", err)
		return nil
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Database connected successfully")
	return nil
}

// convertDatabaseURL converts Render's DATABASE_URL to GORM format
func convertDatabaseURL(url string) string {
	// Quick and dirty conversion
	// postgres://user:pass@host:port/dbname -> host=host port=port user=user password=pass dbname=dbname sslmode=disable
	result := "sslmode=disable "

	// Extract host and port
	// Format: postgres://user:pass@host:port/dbname
	// Remove prefix
	url = url[11:] // remove "postgres://"

	// Find @ to split user:pass from host:port
	atIdx := -1
	for i, c := range url {
		if c == '@' {
			atIdx = i
			break
		}
	}

	if atIdx == -1 {
		return result
	}

	// user:pass
	cred := url[:atIdx]
	hostPart := url[atIdx+1:]

	// Split user:pass
	colonIdx := -1
	for i, c := range cred {
		if c == ':' {
			colonIdx = i
			break
		}
	}

	if colonIdx != -1 {
		result += "user=" + cred[:colonIdx] + " password=" + cred[colonIdx+1:] + " "
	}

	// Split host:port/dbname
	slashIdx := -1
	for i, c := range hostPart {
		if c == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx != -1 {
		hostPort := hostPart[:slashIdx]
		dbname := hostPart[slashIdx+1:]

		// Split host:port
		hostColonIdx := -1
		for i, c := range hostPort {
			if c == ':' {
				hostColonIdx = i
				break
			}
		}

		if hostColonIdx != -1 {
			result += "host=" + hostPort[:hostColonIdx] + " port=" + hostPort[hostColonIdx+1:] + " "
		} else {
			result += "host=" + hostPort + " port=5432 "
		}

		result += "dbname=" + dbname
	}

	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func CloseDB() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
