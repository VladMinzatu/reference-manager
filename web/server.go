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
	r.DELETE("/categories/:id", handler.DeleteCategory)
	r.GET("/add-reference-form", handler.AddReferenceForm)
	r.POST("/references", handler.CreateReference)
	r.DELETE("/references/:id", handler.DeleteReference)
	r.GET("/books/:id/edit", handler.EditBookForm)
	r.PUT("/books/:id", handler.UpdateBook)
	r.GET("/links/:id/edit", handler.EditLinkForm)
	r.PUT("/links/:id", handler.UpdateLink)
	r.GET("/notes/:id/edit", handler.EditNoteForm)
	r.PUT("/notes/:id", handler.UpdateNote)
	r.GET("/categories/:id/edit", handler.EditCategoryForm)
	r.POST("/categories/:id", handler.UpdateCategory)
	r.PUT("/categories/reorder", handler.ReorderCategories)
	r.PUT("/references/reorder", handler.ReorderReferences)

	return r.Run(":8080")
}
