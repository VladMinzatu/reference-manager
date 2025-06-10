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

type SidebarData struct {
	Categories       []model.Category
	ActiveCategoryId int64
}

type ReferencesData struct {
	CategoryName string
	References   []template.HTML
}

func NewHandler(svc *service.ReferenceService) *Handler {
	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))
	return &Handler{svc: svc, template: tmpl}
}

func (h *Handler) Index(c *gin.Context) {
	categories, _ := h.svc.GetAllCategories()
	var activeCategoryId int64
	var activeCategoryName string
	var references []template.HTML

	if len(categories) > 0 {
		activeCategoryId = categories[0].Id
		activeCategoryName = categories[0].Name
		references = h.renderReferences(activeCategoryId)
	}

	// Render the full page with both components
	c.HTML(http.StatusOK, "index.html", gin.H{
		"sidebar": SidebarData{
			Categories:       categories,
			ActiveCategoryId: activeCategoryId,
		},
		"references": ReferencesData{
			CategoryName: activeCategoryName,
			References:   references,
		},
	})
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
	data := ReferencesData{
		CategoryName: categoryName,
		References:   references,
	}

	c.HTML(http.StatusOK, "references", data)
}

func (h *Handler) AddCategoryForm(c *gin.Context) {
	c.Status(http.StatusOK)
	h.template.ExecuteTemplate(c.Writer, "add-category-form.html", nil)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "Category name required")
		return
	}
	category, err := h.svc.AddCategory(name)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create category")
		return
	}
	categories, _ := h.svc.GetAllCategories()

	data := SidebarData{
		Categories:       categories,
		ActiveCategoryId: category.Id,
	}

	c.HTML(http.StatusOK, "sidebar", data)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
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

	if err := h.svc.DeleteCategory(id); err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete category")
		return
	}

	// Get updated categories list and render the sidebar
	categories, _ := h.svc.GetAllCategories()
	var activeCategoryId int64
	if len(categories) > 0 {
		activeCategoryId = categories[0].Id
	}

	data := SidebarData{
		Categories:       categories,
		ActiveCategoryId: activeCategoryId,
	}

	c.HTML(http.StatusOK, "sidebar", data)
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
