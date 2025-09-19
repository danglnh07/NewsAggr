// @title RSS Scraper API
// @version 1.0
// @description This is the API for RSS Scraper service
// @host localhost:8080
// @BasePath /
package main

import (
	"log/slog"
	"os"

	"github.com/danglnh07/newsaggr/scraper/api"
	"github.com/danglnh07/newsaggr/scraper/db"
	"github.com/danglnh07/newsaggr/scraper/service"
	"github.com/danglnh07/newsaggr/scraper/util"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Load config
	config := util.LoadConfig(".env")

	// Create queries, connect database, run auto migration and seed data
	queries := db.NewQueries()

	if err := queries.ConnectDB(config.DBConn); err != nil {
		logger.Error("Error connecting database", "error", err)
		os.Exit(1)
	}

	if err := queries.AutoMigration(); err != nil {
		logger.Error("Error running auto migration", "error", err)
		os.Exit(1)
	}

	if err := queries.Seed(); err != nil {
		logger.Error("Error create seed data", "error", err)
		os.Exit(1)
	}

	// Run the cron job
	rss := service.NewRssScraper(queries)
	scheduler := service.NewScheduler(rss, logger)
	scheduler.Start()

	// Create and run server
	server := api.NewServer(queries, config, logger)
	if err := server.Start(); err != nil {
		logger.Error("Error staring server", "error", err)
		os.Exit(1)
	}
}
