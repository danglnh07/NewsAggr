package api

import (
	"errors"
	"net/http"

	"github.com/danglnh07/newsaggr/scraper/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Article response struct for GET actions
type ArticleResponse struct {
	ID            uint    `json:"id"`
	Title         string  `json:"title"`
	Url           string  `json:"url"`
	Image         *string `json:"image"`
	PublishedDate string  `json:"published_date"`
	Category      string  `json:"category"`
}

// GetArticle godoc
// @Summary      Get an article by ID
// @Description  Retrieve a single article along with its source information
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Article ID"
// @Success      200  {object}  ArticleResponse
// @Failure      404  {object}  ErrorResponse  "Article not found"
// @Failure      500  {object}  ErrorResponse  "Failed to get article"
// @Router       /api/articles/{id} [get]
func (server *Server) GetArticle(ctx *gin.Context) {
	// Get ID from path parameter
	id := ctx.Param("id")

	// Fetch article from database
	var article db.Article
	result := server.queries.DB.Preload("Source").First(&article, id)
	if result.Error != nil {
		// If ID not match any record
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "Article not found"})
			return
		}

		// Other database error
		server.logger.Error("GET /api/articles/:id: Failed to get article", "error", result.Error)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to get article"})
		return
	}

	var image *string = nil
	if article.Image.Valid {
		image = &article.Image.String
	}

	// Return article back to client
	ctx.JSON(http.StatusOK, ArticleResponse{
		ID:            article.ID,
		Title:         article.Title,
		Url:           article.Url,
		Image:         image,
		PublishedDate: article.PublishedDate,
		Category:      article.Source.Category,
	})
}

// ListArticles godoc
// @Summary      List articles
// @Description  Retrieve a paginated list of articles with their details
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        page_id    query     int  true   "Page number"
// @Param        page_size  query     int  true   "Number of items per page"
// @Success      200  {array}   ArticleResponse
// @Failure      500  {object}  ErrorResponse  "Failed to list articles"
// @Router       /api/articles [get]
func (server *Server) ListArticles(ctx *gin.Context) {
	// Get pagination parameters
	pageID, pageSize := server.GetPagingParams(ctx)
	if pageID == 0 || pageSize == 0 {
		// Error already handled in GetPagingParams
		return
	}

	// Fetch articles from database with pagination
	var articles []db.Article
	result := server.queries.DB.Preload("Source").Limit(int(pageSize)).Offset(int((pageID - 1) * pageSize)).Find(&articles)
	if result.Error != nil {
		server.logger.Error("GET /api/articles: Failed to list articles", "error", result.Error)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to list articles"})
		return
	}

	resp := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		var image *string = nil
		if article.Image.Valid {
			image = &article.Image.String
		}

		resp[i] = ArticleResponse{
			ID:            article.ID,
			Title:         article.Title,
			Url:           article.Url,
			Image:         image,
			PublishedDate: article.PublishedDate,
		}
	}

	// Return the result back to client
	ctx.JSON(http.StatusOK, resp)
}
