package web

import "net/http"

/*
Next steps:
- Put categories on sidebar and interact with references for one category at a time
- Add a form to add new categories
- Add a form to add new references
- Add a form to edit existing references
- Button to delete references with confirmation
- Interactivity to reorder categories and references
- Optional: add pagination for references
- Optional: Add a search bar

- Implement error handling for the server start.
- Add logging
- Configs for db settings
- Add graceful shutdown
*/
func StartServer(handler *Handler) error {
	http.HandleFunc("/", handler.Index)

	return http.ListenAndServe(":8080", nil)
}
