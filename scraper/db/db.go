package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Queries struct
type Queries struct {
	DB *gorm.DB
}

// Constructor method for Queries
func NewQueries() *Queries {
	return &Queries{}
}

// Connect to database
func (queries *Queries) ConnectDB(connStr string) error {
	conn, err := gorm.Open(postgres.Open(connStr))
	if err != nil {
		return err
	}

	queries.DB = conn
	return nil
}

// Run auto migration
func (queries *Queries) AutoMigration() error {
	return queries.DB.AutoMigrate(&Source{}, &Article{})
}

func (queries *Queries) Seed() error {
	// Insert some sources
	sources := []Source{
		{
			Link:     "https://cloudblog.withgoogle.com/rss/",
			Provider: "cloud.google.com/blog",
			Category: "engineering",
		},
		{
			Link:     "https://blog.google/rss/",
			Provider: "blog.google",
			Category: "engineering",
		},
		{
			Link:     "https://feeds.feedburner.com/GDBcode",
			Provider: "https://developers.googleblog.com",
			Category: "engineering",
		},
	}

	// Create source in database
	for _, source := range sources {
		var dest Source
		// use the unique link as the lookup condition and create if not exists
		if err := queries.DB.Where("link = ?", source.Link).FirstOrCreate(&dest, source).Error; err != nil {
			return err
		}
	}
	return nil
}
