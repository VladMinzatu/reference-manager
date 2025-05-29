package web

import "net/http"

func StartServer(handler *Handler) error {
	http.HandleFunc("/", handler.Index)

	return http.ListenAndServe(":8080", nil)
}
