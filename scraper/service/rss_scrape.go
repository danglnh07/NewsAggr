package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/danglnh07/newsaggr/scraper/db"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"
)

// Scraper struct
type RssScraper struct {
	queries *db.Queries
}

// Constructor method for Scraper
func NewRssScraper(queries *db.Queries) *RssScraper {
	return &RssScraper{
		queries: queries,
	}
}

// Scrape from an individual RSS source
func (scraper *RssScraper) Scrape(source db.Source) error {
	// Avoid nil slice
	articles := make([]db.Article, 0)

	// Parse the RSS feed
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(source.Link)
	if err != nil {
		return err
	}

	// Loop through each item and create articles
	for _, item := range feed.Items {
		image := sql.NullString{String: "", Valid: false} // Null string
		if item.Image != nil {
			image.String = item.Image.URL
			image.Valid = true
		}

		article := db.Article{
			Model:         gorm.Model{},
			SourceID:      source.ID,
			Title:         item.Title,
			Url:           item.Link,
			Image:         image,
			PublishedDate: item.Published,
		}
		articles = append(articles, article)
	}

	// Add all articles into database
	if len(articles) != 0 {
		result := scraper.queries.DB.CreateInBatches(&articles, len(articles))
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

// Run scraping for all RSS sources in database
func (scraper *RssScraper) Run() error {
	// Get all the sources
	var sources []db.Source
	result := scraper.queries.DB.Find(&sources)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no rss source found in database")
		}

		return result.Error
	}

	// Start scraping each source in a separate goroutine
	var (
		wg    sync.WaitGroup
		mutex sync.Mutex
		errs  = make([]string, 0)
	)
	for _, source := range sources {
		wg.Add(1)
		go func(src db.Source) {
			defer wg.Done()
			err := scraper.Scrape(src)
			if err != nil {
				mutex.Lock()
				errs = append(errs, fmt.Sprintf("error scraping source %s: %v", src.Link, err))
				mutex.Unlock()
			}
		}(source)
	}

	wg.Wait()

	// Check if there is any error
	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("error: \n%s", strings.Join(errs, "\n"))
}
