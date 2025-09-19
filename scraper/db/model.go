package db

import (
	"database/sql"

	"gorm.io/gorm"
)

// RSS source model
type Source struct {
	gorm.Model
	Link     string `json:"link" gorm:"unique"`
	Provider string `json:"provider"`
	Category string `json:"category"`
}

// Article model
type Article struct {
	gorm.Model
	SourceID      uint           `json:"source_id"`
	Source        Source         `json:"source" gorm:"foreignKey:SourceID"`
	Title         string         `json:"title"`
	Url           string         `json:"url" gorm:"unique"` // The article URL
	Image         sql.NullString `json:"image"`
	PublishedDate string         `json:"published_date"`
}
