package web

import (
	"net/http"
	"text/template"
)

type Handler struct {
	templates *template.Template
}

func NewHandler() *Handler {
	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))
	return &Handler{templates: tmpl}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "index.html", nil)
}
