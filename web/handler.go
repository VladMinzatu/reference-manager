package web

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/service"
)

type Handler struct {
	svc      *service.ReferenceService
	template *template.Template
}

func NewHandler(svc *service.ReferenceService) *Handler {
	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))
	return &Handler{svc: svc, template: tmpl}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	categories, _ := h.svc.GetAllCategories()
	type CategoryData struct {
		Category   model.Category
		References []template.HTML
	}
	var data []CategoryData

	for _, cat := range categories {
		renderer := NewHTMLReferenceRenderer(h.template)
		refs, _ := h.svc.GetReferences(cat.Id, false)
		for _, ref := range refs {
			ref.Render(renderer)
		}
		data = append(data, CategoryData{
			Category:   cat,
			References: renderer.Collect(),
		})
	}

	h.template.ExecuteTemplate(w, "index.html", data)
}

type HTMLReferenceRenderer struct {
	tmpl      *template.Template
	collected []template.HTML
}

func NewHTMLReferenceRenderer(tmpl *template.Template) *HTMLReferenceRenderer {
	return &HTMLReferenceRenderer{tmpl: tmpl, collected: make([]template.HTML, 0)}
}

func (r *HTMLReferenceRenderer) RenderBook(ref model.BookReference) {
	r.Render("_book", ref)
}
func (r *HTMLReferenceRenderer) RenderLink(ref model.LinkReference) {
	r.Render("_link", ref)
}
func (r *HTMLReferenceRenderer) RenderNote(ref model.NoteReference) {
	r.Render("_note", ref)
}

func (r *HTMLReferenceRenderer) Render(rendererName string, ref model.Reference) {
	var buf bytes.Buffer
	err := r.tmpl.ExecuteTemplate(&buf, rendererName, ref)
	if err != nil {
		// TODO: it's better to propagate and crash
		slog.Error("failed to render template", "template", rendererName, "error", err)
		return
	}
	r.collected = append(r.collected, template.HTML(buf.String()))
}

func (r *HTMLReferenceRenderer) Collect() []template.HTML {
	return r.collected
}
