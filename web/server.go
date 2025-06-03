package web

import "net/http"

/*
Next steps:
- Add a form to add new categories
- Add a form to add new references
- Add a form to edit existing references
- Button to delete references with confirmation
- Interactivity to reorder categories and references
- Optional: add pagination for references
- Optional: Add a search bar

- DDD review and refactor: service -> aggregates and how aggregate roots integrate the repository

- Implement error handling for the server start.
- Add logging
- Configs for db settings
- Add graceful shutdown
*/
func StartServer(handler *Handler) error {
	http.HandleFunc("/", handler.Index)
	http.HandleFunc("/category/", handler.CategoryReferences)
	http.HandleFunc("/add-category-modal-form", handler.AddCategoryModalForm)

	return http.ListenAndServe(":8080", nil)
}
