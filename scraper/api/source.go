package api

import (
	"errors"
	"net/http"

	"github.com/danglnh07/newsaggr/scraper/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Response struct for resource
type SourceResponse struct {
	ID       uint   `json:"id"`
	Link     string `json:"link"`
	Provider string `json:"provider"`
	Category string `json:"category"`
}

// GetSource godoc
// @Summary      Get a news source by ID
// @Description  Retrieve a single news source from the database using its ID
// @Tags         sources
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Source ID"
// @Success      200  {object}  SourceResponse
// @Failure      404  {object}  ErrorResponse  "Source not found"
// @Failure      500  {object}  ErrorResponse  "Failed to get source"
// @Router       /api/sources/{id} [get]
func (server *Server) GetSource(ctx *gin.Context) {
	// Get ID from path parameter
	id := ctx.Param("id")

	// Fetch source from database
	var source db.Source
	result := server.queries.DB.First(&source, id)
	if result.Error != nil {
		// If ID not match any record
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "Source not found"})
			return
		}

		// Other database error
		server.logger.Error("GET /api/sources/:id: Failed to get source", "error", result.Error)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to get source"})
		return
	}

	// Return the result back to client
	ctx.JSON(http.StatusOK, SourceResponse{
		ID:       source.ID,
		Link:     source.Link,
		Provider: source.Provider,
		Category: source.Category,
	})
}

// ListSources godoc
// @Summary      List news sources
// @Description  Retrieve a paginated list of news sources
// @Tags         sources
// @Accept       json
// @Produce      json
// @Param        page_id    query     int  true   "Page number"
// @Param        page_size  query     int  true   "Number of items per page"
// @Success      200  {array}   SourceResponse
// @Failure      500  {object}  ErrorResponse  "Failed to list sources"
// @Router       /api/sources [get]
func (server *Server) ListSources(ctx *gin.Context) {
	// Get pagination parameters
	pageID, pageSize := server.GetPagingParams(ctx)
	if pageID == 0 || pageSize == 0 {
		// Error already handled in GetPagingParams
		return
	}

	// Fetch sources from database with pagination
	var sources []db.Source
	result := server.queries.DB.Limit(int(pageSize)).Offset(int((pageID - 1) * pageSize)).Find(&sources)
	if result.Error != nil {
		server.logger.Error("GET /api/sources: Failed to list sources", "error", result.Error)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to list sources"})
		return
	}

	resp := make([]SourceResponse, len(sources))
	for i, source := range sources {
		resp[i] = SourceResponse{
			ID:       source.ID,
			Link:     source.Link,
			Provider: source.Provider,
			Category: source.Category,
		}
	}

	// Return the result back to client
	ctx.JSON(http.StatusOK, resp)
}

// Request struct for create resource action
type CreateSourceRequest struct {
	Link     string `json:"link" binding:"required"`
	Provider string `json:"provider" binding:"required"`
	Category string `json:"category" binding:"required"`
}

// CreateSource godoc
// @Summary      Create a new news source
// @Description  Add a new news source to the database
// @Tags         sources
// @Accept       json
// @Produce      json
// @Param        source  body      CreateSourceRequest  true  "Source details"
// @Success      201  {object}  SourceResponse
// @Failure      400  {object}  ErrorResponse  "Invalid request body"
// @Failure      500  {object}  ErrorResponse  "Failed to create source"
// @Router       /api/sources [post]
func (server *Server) CreateSource(ctx *gin.Context) {
	var req CreateSourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.logger.Error("POST /api/sources: Invalid request body", "error", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request body"})
		return
	}

	var source = db.Source{
		Model:    gorm.Model{},
		Link:     req.Link,
		Provider: req.Provider,
		Category: req.Category,
	}
	result := server.queries.DB.Create(&source)
	if result.Error != nil {
		server.logger.Error("POST /api/sources: Failed to create source", "error", result.Error)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to create source"})
		return
	}

	ctx.JSON(http.StatusCreated, SourceResponse{
		ID:       source.ID,
		Link:     source.Link,
		Provider: source.Provider,
		Category: source.Category,
	})
}

// Request struct for update source action
type UpdateSourceRequest struct {
	Link     string `json:"link"`
	Provider string `json:"provider"`
	Category string `json:"category"`
}

// UpdateSource godoc
// @Summary      Update a news source
// @Description  Update details of an existing news source by ID
// @Tags         sources
// @Accept       json
// @Produce      json
// @Param        id      path      int                  true  "Source ID"
// @Param        source  body      UpdateSourceRequest  true  "Updated source details"
// @Success      200  {object}  SourceResponse
// @Failure      400  {object}  ErrorResponse  "Invalid request body"
// @Failure      404  {object}  ErrorResponse  "Source not found"
// @Failure      500  {object}  ErrorResponse  "Failed to update source"
// @Router       /api/sources/{id} [put]
func (server *Server) UpdateSource(ctx *gin.Context) {
	// Parse and validate request body
	var req UpdateSourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.logger.Error("PUT /api/sources/:id: Invalid request body", "error", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request body"})
		return
	}

	// Get ID from path parameter
	id := ctx.Param("id")
	var source db.Source
	result := server.queries.DB.First(&source, id)
	if result.Error != nil {
		// If ID not match any record
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "Source not found"})
			return
		}

		// Other database error
		server.logger.Error("PUT /api/sources/:id: Failed to get source", "error", result.Error)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to get source"})
		return
	}

	// Update new values if provided
	if req.Link != "" {
		source.Link = req.Link
	}

	if req.Provider != "" {
		source.Provider = req.Provider
	}

	if req.Category != "" {
		source.Category = req.Category
	}

	// Save changed to database
	result = server.queries.DB.Save(&source)
	if result.Error != nil {
		server.logger.Error("PUT /api/sources/:id: Failed to update source", "error", result.Error)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update source"})
		return
	}

	// Return result back to client
	ctx.JSON(http.StatusOK, SourceResponse{
		ID:       source.ID,
		Link:     source.Link,
		Provider: source.Provider,
		Category: source.Category,
	})
}

// DeleteSource godoc
// @Summary      Delete a news source
// @Description  Remove an existing news source by ID
// @Tags         sources
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Source ID"
// @Success      204  "No Content"
// @Failure      404  {object}  ErrorResponse  "Source not found"
// @Failure      500  {object}  ErrorResponse  "Failed to delete source"
// @Router       /api/sources/{id} [delete]
func (server *Server) DeleteSource(ctx *gin.Context) {
	// Get ID from path parameter
	id := ctx.Param("id")
	var source db.Source
	result := server.queries.DB.First(&source, id)
	if result.Error != nil {
		// If ID not match any record
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "Source not found"})
			return
		}

		// Other database error
		server.logger.Error("DELETE /api/sources/:id: Failed to get source", "error", result.Error)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to get source"})
		return
	}

	// Delete the source
	result = server.queries.DB.Delete(&source)
	if result.Error != nil {
		server.logger.Error("DELETE /api/sources/:id: Failed to delete source", "error", result.Error)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to delete source"})
		return
	}

	// Return no content status
	ctx.Status(http.StatusNoContent)
}
