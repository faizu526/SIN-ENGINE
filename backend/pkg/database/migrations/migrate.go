package migrations

import (
	"log"

	"github.com/sin-engine/pkg/database/models"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	// Auto Migrate - Like Django's makemigrations + migrate
	err := db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.APIKey{},
		&models.Scan{},
		&models.Vulnerability{},
		&models.Exploit{},
		&models.ExploitChain{},
		&models.Report{},
		&models.Dork{},
		&models.DorkResult{},
		&models.Breach{},
		&models.BreachMonitor{},
		&models.ReconTarget{},
		&models.ReconResult{},
		&models.Subdomain{},
	)

	if err != nil {
		return err
	}

	// Create indexes manually for better performance
	log.Println("📊 Creating indexes...")

	// User indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);")

	// Scan indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_scans_user_id ON scans(user_id);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_scans_status ON scans(status);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_scans_created_at ON scans(created_at);")

	// Vulnerability indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_scan_id ON vulnerabilities(scan_id);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_severity ON vulnerabilities(severity);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_type ON vulnerabilities(type);")

	// Recon indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_subdomains_domain ON subdomains(domain);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_recon_targets_user_id ON recon_targets(user_id);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_recon_targets_target ON recon_targets(target);")

	// Breach indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_breaches_email ON breaches(email);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_breaches_domain ON breaches(domain);")

	log.Println("✅ Migrations completed successfully")

	// Seed initial data if needed
	return seedData(db)
}

func seedData(db *gorm.DB) error {
	log.Println("🌱 Seeding initial data...")

	// Check if admin exists
	var count int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)

	if count == 0 {
		// Create default admin user
		admin := &models.User{
			Username:   "admin",
			Email:      "admin@sin.engine",
			FullName:   "System Administrator",
			Role:       "admin",
			IsActive:   true,
			IsVerified: true,
		}
		admin.SetPassword("Admin@123") // Change this in production

		if err := db.Create(admin).Error; err != nil {
			return err
		}
		log.Println("✅ Default admin user created")
	}

	// Seed default dorks
	seedDefaultDorks(db)

	return nil
}

func seedDefaultDorks(db *gorm.DB) {
	dorks := []models.Dork{
		{
			Name:        "GitHub API Keys",
			Query:       "filename:.env OR filename:.env.production OR extension:env AND (api_key OR apikey OR secret)",
			Engine:      models.EngineGitHub,
			Category:    "credentials",
			Description: "Find exposed API keys and secrets in GitHub",
			Tags:        []string{"credentials", "api-keys", "secrets"},
			IsPublic:    true,
		},
		{
			Name:        "Database Dumps",
			Query:       "extension:sql AND (INSERT INTO VALUES) AND (password OR pass OR user)",
			Engine:      models.EngineGitHub,
			Category:    "database",
			Description: "Find exposed database dumps",
			Tags:        []string{"database", "sql", "dump"},
			IsPublic:    true,
		},
		{
			Name:        "Shodan Open Ports",
			Query:       "port:22,23,3389,3306,5432,6379,27017",
			Engine:      models.EngineShodan,
			Category:    "network",
			Description: "Find exposed services on common ports",
			Tags:        []string{"network", "ports", "services"},
			IsPublic:    true,
		},
		{
			Name:        "WordPress Admin Panels",
			Query:       "http.title:'wp-admin' OR http.title:'WordPress Login' OR http.html:'/wp-admin'",
			Engine:      models.EngineShodan,
			Category:    "web",
			Description: "Find WordPress admin panels",
			Tags:        []string{"wordpress", "admin", "cms"},
			IsPublic:    true,
		},
	}

	for _, dork := range dorks {
		var count int64
		db.Model(&models.Dork{}).Where("name = ?", dork.Name).Count(&count)
		if count == 0 {
			db.Create(&dork)
		}
	}

	log.Println("✅ Default dorks seeded")
}
