package web

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/service"
	"github.com/gin-gonic/gin"
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

func (h *Handler) Index(c *gin.Context) {
	categories, _ := h.svc.GetAllCategories()
	var data IndexData

	if len(categories) == 0 {
		c.Status(http.StatusOK)
		h.template.ExecuteTemplate(c.Writer, "index.html", data)
		return
	}
	activeCategoryId := categories[0].Id
	activeCategoryName := categories[0].Name

	references := h.renderReferences(activeCategoryId)

	data.Categories = categories
	data.ActiveCategoryId = activeCategoryId
	data.ReferenceData = ReferenceData{
		CategoryName: activeCategoryName,
		References:   references,
	}

	c.Status(http.StatusOK)
	h.template.ExecuteTemplate(c.Writer, "index.html", data)
}

func (h *Handler) CategoryReferences(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.String(http.StatusBadRequest, "Invalid path")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid category id")
		return
	}
	references := h.renderReferences(id)

	categoryName := c.Query("categoryName")
	data := ReferenceData{
		CategoryName: categoryName,
		References:   references,
	}

	c.Status(http.StatusOK)
	h.template.ExecuteTemplate(c.Writer, "references", data)
}

func (h *Handler) AddCategoryForm(c *gin.Context) {
	c.Status(http.StatusOK)
	h.template.ExecuteTemplate(c.Writer, "add-category-form.html", nil)
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
