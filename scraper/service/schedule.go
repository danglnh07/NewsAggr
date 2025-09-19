package service

import (
	"log/slog"

	"github.com/robfig/cron/v3"
)

// Scheduler struct
type Scheduler struct {
	c          *cron.Cron
	RssScraper *RssScraper
	logger     *slog.Logger
}

// Constructor method of Scheduler
func NewScheduler(rss *RssScraper, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		c:          cron.New(cron.WithSeconds()),
		RssScraper: rss,
		logger:     logger,
	}
}

// Start cron job
func (scheduler *Scheduler) Start() {
	// Run scraping for every hour
	_, err := scheduler.c.AddFunc("0 0 * * * *", func() {
		err := scheduler.RssScraper.Run()
		if err != nil {
			scheduler.logger.Error("Failed to run RSS scraping", "error", err)
			return
		}
	})

	if err != nil {
		scheduler.logger.Error("Failed to set up cron job for RSS scraper", "error", err)
		return
	}

	scheduler.c.Start()
}
