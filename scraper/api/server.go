package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/danglnh07/newsaggr/scraper/db"
	_ "github.com/danglnh07/newsaggr/scraper/docs"
	"github.com/danglnh07/newsaggr/scraper/util"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server struct
type Server struct {
	mux     *gin.Engine
	queries *db.Queries
	config  *util.Config
	logger  *slog.Logger
}

// Constructor method for Server
func NewServer(queries *db.Queries, config *util.Config, logger *slog.Logger) *Server {
	return &Server{
		mux:     gin.Default(),
		queries: queries,
		config:  config,
		logger:  logger,
	}
}

// Method to register handler
func (server *Server) RegisterHandler() {
	api := server.mux.Group("/api")
	{
		// Article's routes
		articles := api.Group("/articles")
		{
			articles.GET("/:id", server.GetArticle)
			articles.GET("", server.ListArticles)
		}

		// Source's routes
		sources := api.Group("/sources")
		{
			sources.GET("/:id", server.GetSource)
			sources.GET("", server.ListSources)
			sources.POST("", server.CreateSource)
			sources.PUT("/:id", server.UpdateSource)
			sources.DELETE("/:id", server.DeleteSource)
		}

		// Swagger route
		api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

// Method to start the server
func (server *Server) Start() error {
	server.RegisterHandler()
	return server.mux.Run(":8080")
}

// General error response
type ErrorResponse struct {
	Message string `json:"error"`
}

// Helper method: extract query parameter for pagination
func (server *Server) GetPagingParams(ctx *gin.Context) (int, int) {
	const (
		defaultPageID   = 1
		defaultPageSize = 5
		maxPageSize     = 10
	)

	pageID, err := strconv.Atoi(ctx.Query("page_id"))
	if err != nil || pageID < 1 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid page_id parameter"})
		return 0, 0
	}

	pageSize, err := strconv.Atoi(ctx.Query("page_size"))
	if err != nil || pageSize < 1 || pageSize > maxPageSize {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid page_size parameter"})
		return 0, 0
	}

	return pageID, pageSize
}
