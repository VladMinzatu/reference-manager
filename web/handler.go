package web

import (
	"bytes"
	"encoding/json"
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
	CategoryId   int64
	CategoryName string
	References   []template.HTML
}

type AddReferenceFormData struct {
	CategoryId int64
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
			CategoryId:   activeCategoryId,
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
		CategoryId:   id,
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

	// Alternative here would be to just return c.HTML(.., "sidebar") and use custom client-side JS and HTMX event (triggered on both delete and add-form-sumbmit) to update the the references when the sidebar is updated.
	c.HTML(http.StatusOK, "body-fragment", gin.H{
		"sidebar": SidebarData{
			Categories:       categories,
			ActiveCategoryId: category.Id,
		},
		"references": ReferencesData{
			CategoryId:   category.Id,
			CategoryName: category.Name,
			References:   []template.HTML{},
		},
	})
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

	// Get updated categories list
	// TODO: error handling
	categories, _ := h.svc.GetAllCategories()
	var activeCategoryId int64
	var activeCategoryName string
	if len(categories) > 0 {
		activeCategoryId = categories[0].Id
		activeCategoryName = categories[0].Name
	}

	// TODO: error handling
	references, _ := h.svc.GetReferences(activeCategoryId, false)
	renderer := NewHTMLReferenceRenderer(h.template)
	for _, ref := range references {
		ref.Render(renderer)
	}

	// Alternative here would be to just return c.HTML(.., "sidebar") and use custom client-side JS and HTMX event (triggered on both delete and add-form-sumbmit) to update the the references when the sidebar is updated.
	c.HTML(http.StatusOK, "body-fragment", gin.H{
		"sidebar": SidebarData{
			Categories:       categories,
			ActiveCategoryId: activeCategoryId,
		},
		"references": ReferencesData{
			CategoryId:   activeCategoryId,
			CategoryName: activeCategoryName,
			References:   renderer.Collect(),
		},
	})
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

func (h *Handler) AddReferenceForm(c *gin.Context) {
	categoryId := c.Query("categoryId")
	if categoryId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "categoryId is required"})
		return
	}

	id, err := strconv.ParseInt(categoryId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid categoryId"})
		return
	}

	c.HTML(http.StatusOK, "add-reference-form", AddReferenceFormData{
		CategoryId: id,
	})
}

func (h *Handler) CreateReference(c *gin.Context) {
	categoryIdStr := c.PostForm("categoryId")
	categoryId, err := strconv.ParseInt(categoryIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid categoryId"})
		return
	}

	refType := c.PostForm("type")
	if refType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
		return
	}

	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	starred := c.PostForm("starred") == "on"

	switch refType {
	case "book":
		book, err := h.svc.AddBookReference(categoryId, title, c.PostForm("isbn"), c.PostForm("description"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create book reference"})
			return
		}
		if starred {
			if err := h.svc.UpdateBookReference(book.Id, book.Title, book.ISBN, book.Description, true); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update book reference starred status"})
				return
			}
		}

	case "link":
		url := c.PostForm("url")
		if url == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required for links"})
			return
		}
		link, err := h.svc.AddLinkReference(categoryId, title, url, c.PostForm("description"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link reference"})
			return
		}
		if starred {
			if err := h.svc.UpdateLinkReference(link.Id, link.Title, link.URL, link.Description, true); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update link reference starred status"})
				return
			}
		}

	case "note":
		content := c.PostForm("content")
		if content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content is required for notes"})
			return
		}
		note, err := h.svc.AddNoteReference(categoryId, title, content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create note reference"})
			return
		}
		if starred {
			if err := h.svc.UpdateNoteReference(note.Id, note.Title, note.Text, true); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note reference starred status"})
				return
			}
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reference type"})
		return
	}

	references := h.renderReferences(categoryId)
	c.HTML(http.StatusOK, "_references_list", references)
}

func (h *Handler) DeleteReference(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.String(http.StatusBadRequest, "Invalid path")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid reference id")
		return
	}

	if err := h.svc.DeleteReference(id); err != nil {
		slog.Error("failed to delete reference", "error", err, "id", id)
		c.String(http.StatusInternalServerError, "Failed to delete reference")
		return
	}

	// Return empty response since the reference will be removed from the DOM
	c.Status(http.StatusOK)
}

// Helper for rendering edit reference forms
func (h *Handler) EditBookForm(c *gin.Context) {
	renderEditReferenceForm(c, "_edit_book_form", map[string]interface{}{
		"Title":       c.Query("title"),
		"ISBN":        c.Query("isbn"),
		"Description": c.Query("description"),
		"Starred":     c.Query("starred") == "true" || c.Query("starred") == "1",
	})
}

func (h *Handler) EditLinkForm(c *gin.Context) {
	renderEditReferenceForm(c, "_edit_link_form", map[string]interface{}{
		"Title":       c.Query("title"),
		"URL":         c.Query("url"),
		"Description": c.Query("description"),
		"Starred":     c.Query("starred") == "true" || c.Query("starred") == "1",
	})
}

func (h *Handler) EditNoteForm(c *gin.Context) {
	renderEditReferenceForm(c, "_edit_note_form", map[string]interface{}{
		"Title":   c.Query("title"),
		"Text":    c.Query("content"),
		"Starred": c.Query("starred") == "true" || c.Query("starred") == "1",
	})
}

func renderEditReferenceForm(c *gin.Context, tmpl string, fields map[string]interface{}) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	fields["Id"] = id
	c.HTML(http.StatusOK, tmpl, fields)
}

func (h *Handler) UpdateBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid id")
		return
	}
	title := c.PostForm("title")
	isbn := c.PostForm("isbn")
	description := c.PostForm("description")
	starred := c.PostForm("starred") == "on"
	data := struct {
		Id          int64
		Title       string
		ISBN        string
		Description string
		Starred     bool
	}{
		Id:          id,
		Title:       title,
		ISBN:        isbn,
		Description: description,
		Starred:     starred,
	}
	err = h.svc.UpdateBookReference(id, title, isbn, description, starred)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update reference")
		return
	}
	c.HTML(http.StatusOK, "_book", data)
}

func (h *Handler) UpdateLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid id")
		return
	}
	title := c.PostForm("title")
	url := c.PostForm("url")
	description := c.PostForm("description")
	starred := c.PostForm("starred") == "on"
	data := struct {
		Id          int64
		Title       string
		URL         string
		Description string
		Starred     bool
	}{
		Id:          id,
		Title:       title,
		URL:         url,
		Description: description,
		Starred:     starred,
	}
	err = h.svc.UpdateLinkReference(id, title, url, description, starred)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update reference")
		return
	}
	c.HTML(http.StatusOK, "_link", data)
}

func (h *Handler) UpdateNote(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid id")
		return
	}
	title := c.PostForm("title")
	content := c.PostForm("content")
	starred := c.PostForm("starred") == "on"
	data := struct {
		Id      int64
		Title   string
		Text    string
		Starred bool
	}{
		Id:      id,
		Title:   title,
		Text:    content,
		Starred: starred,
	}
	err = h.svc.UpdateNoteReference(id, title, content, starred)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update reference")
		return
	}
	c.HTML(http.StatusOK, "_note", data)
}

func (h *Handler) EditCategoryForm(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	name := c.Query("name")
	data := struct {
		Id   int64
		Name string
	}{
		Id:   id,
		Name: name,
	}
	c.HTML(http.StatusOK, "_edit_category_form", data)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid id")
		return
	}
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "Category name required")
		return
	}
	err = h.svc.UpdateCategory(id, name)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update category")
		return
	}
	// Return updated sidebar
	categories, _ := h.svc.GetAllCategories()
	c.HTML(http.StatusOK, "sidebar", SidebarData{
		Categories:       categories,
		ActiveCategoryId: id,
	})
}

func (h *Handler) ReorderCategories(c *gin.Context) {
	positionsStr := c.PostForm("positions")
	if positionsStr == "" {
		c.String(http.StatusBadRequest, "Missing positions")
		return
	}
	var positions map[int64]int
	if err := json.Unmarshal([]byte(positionsStr), &positions); err != nil {
		c.String(http.StatusBadRequest, "Invalid positions format")
		return
	}
	if err := h.svc.ReorderCategories(positions); err != nil {
		c.String(http.StatusInternalServerError, "Failed to reorder categories")
		return
	}
	categories, _ := h.svc.GetAllCategories()
	c.HTML(http.StatusOK, "sidebar", SidebarData{
		Categories:       categories,
		ActiveCategoryId: 0,
	})
}
