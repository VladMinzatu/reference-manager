package web

import (
	"github.com/gin-gonic/gin"
)

/*
Next steps:
- Add a form to add new categories
- Add a form to add new references
- Add a form to edit existing references
- Button to delete references with confirmation
- Interactivity to reorder categories and references
- Optional: implement lazy loading / infinite scroll for references using hx-trigger="revealed"
- Optional: Add a search bar

- DDD review and refactor: service -> aggregates and how aggregate roots integrate the repository

- Implement error handling for the server start.
- Add logging
- Configs for db settings
- Add graceful shutdown
*/
func StartServer(handler *Handler) error {
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./web/static")

	// Load templates
	r.LoadHTMLGlob("web/templates/*.html")

	// Routes
	r.GET("/", handler.Index)
	r.GET("/categories/:id/references", handler.CategoryReferences)
	r.GET("/add-category-form", handler.AddCategoryForm)
	r.POST("/categories", handler.CreateCategory)

	return r.Run(":8080")
}
