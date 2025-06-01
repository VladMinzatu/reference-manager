package web

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/service"
)

type Handler struct {
	svc      *service.ReferenceService
	template *template.Template
}

type IndexData struct {
	Categories       []model.Category
	ActiveCategoryId int64
	ReferenceData    ReferenceData
}

type ReferenceData struct {
	CategoryName string
	References   []template.HTML
}

func NewHandler(svc *service.ReferenceService) *Handler {
	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))
	return &Handler{svc: svc, template: tmpl}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	categories, _ := h.svc.GetAllCategories()
	var data IndexData

	if len(categories) == 0 {
		h.template.ExecuteTemplate(w, "index.html", data)
	}
	activeCategoryId := categories[0].Id
	activeCategoryName := categories[0].Name

	references := h.renderReferences(activeCategoryId)

	data.Categories = categories
	data.ActiveCategoryId = activeCategoryId
	data.ReferenceData = ReferenceData{
		CategoryName: activeCategoryName,
		References:   references}

	h.template.ExecuteTemplate(w, "index.html", data)
}

func (h *Handler) CategoryReferences(w http.ResponseWriter, r *http.Request) {
	// TODO: Use gorilla/mux or similar to handle path parameters more elegantly
	path := r.URL.Path // e.g., "/category/{id}/references"
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "category" || parts[2] == "" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
	}
	id, _ := strconv.ParseInt(parts[2], 10, 64)
	references := h.renderReferences(id)

	categoryName := r.URL.Query().Get("categoryName")
	data := ReferenceData{
		CategoryName: categoryName,
		References:   references,
	}

	h.template.ExecuteTemplate(w, "references", data)
}

func (h *Handler) renderReferences(categoryId int64) []template.HTML {
	refs, _ := h.svc.GetReferences(categoryId, false)
	renderer := NewHTMLReferenceRenderer(h.template)
	for _, ref := range refs {
		ref.Render(renderer)
	}
	return renderer.collected
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
