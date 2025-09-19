package service

import (
	"log/slog"
	"os"
	"testing"

	"github.com/danglnh07/newsaggr/scraper/db"
	"github.com/danglnh07/newsaggr/scraper/util"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var (
	scraper *RssScraper
	logger  = slog.New(slog.NewTextHandler(os.Stdout, nil))
)

// Main entry point of this package test
func TestMain(m *testing.M) {
	// Load config
	config := util.LoadConfig("../.env")

	// Create queries, connect database and run auto migration
	queries := db.NewQueries()

	if err := queries.ConnectDB(config.DBConn); err != nil {
		logger.Error("Error connecting database", "error", err)
		os.Exit(1)
	}

	if err := queries.AutoMigration(); err != nil {
		logger.Error("Error running auto migration", "error", err)
		os.Exit(1)
	}

	// Create scraper
	scraper = NewRssScraper(queries)

	os.Exit(m.Run())
}

// Test RSS scraping
func TestScrape(t *testing.T) {
	// Insert some sources
	sources := []db.Source{
		{
			Model:    gorm.Model{},
			Link:     "https://cloudblog.withgoogle.com/rss/",
			Provider: "cloud.google.com/blog",
			Category: "engineering",
		},
		{
			Model:    gorm.Model{},
			Link:     "https://blog.google/rss/",
			Provider: "blog.google",
			Category: "engineering",
		},
		{
			Model:    gorm.Model{},
			Link:     "https://feeds.feedburner.com/GDBcode",
			Provider: "https://developers.googleblog.com",
			Category: "engineering",
		},
	}

	// Create source in database
	result := scraper.queries.DB.CreateInBatches(&sources, len(sources))
	require.NoError(t, result.Error)

	// Refresh inserted sources to ensure IDs are populated (safe guard)
	links := make([]string, 0, len(sources))
	for _, s := range sources {
		links = append(links, s.Link)
	}
	var inserted []db.Source
	result = scraper.queries.DB.Where("link IN ?", links).Find(&inserted)
	require.NoError(t, result.Error)
	require.Equal(t, len(sources), len(inserted))

	// Run scraping
	err := scraper.Run()
	require.NoError(t, err)

	// Check result and clean up
	for _, src := range inserted {
		require.NotZero(t, src.ID, "source ID should be set")

		var count int64
		res := scraper.queries.DB.Model(&db.Article{}).Where("source_id = ?", src.ID).Count(&count)
		require.NoError(t, res.Error)
		require.Greater(t, count, int64(0), "expected articles for source %s", src.Link)

		// Permanently remove articles for this source
		res = scraper.queries.DB.Where("source_id = ?", src.ID).Unscoped().Delete(&db.Article{})
		require.NoError(t, res.Error)
		require.Greater(t, res.RowsAffected, int64(0), "no articles deleted for source %d", src.ID)

		// Permanently remove the source
		res = scraper.queries.DB.Unscoped().Delete(&db.Source{}, src.ID)
		require.NoError(t, res.Error)
		require.Equal(t, int64(1), res.RowsAffected, "source not deleted")
	}
}
