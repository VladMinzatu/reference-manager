package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/repository"
	"github.com/VladMinzatu/reference-manager/domain/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	categoryService        *service.CategoryService
	categoryListRepository repository.CategoryListRepository
	referenceRepo          repository.ReferencesRepository
	template               *template.Template
}

type SidebarData struct {
	Categories       []model.CategoryRef
	ActiveCategoryId model.Id
}

type ReferencesData struct {
	CategoryId   model.Id
	CategoryName model.Title
	References   []template.HTML
}

type AddReferenceFormData struct {
	CategoryId int64
}

func NewHandler(categoryService *service.CategoryService, categoryListRepository repository.CategoryListRepository, referenceRepo repository.ReferencesRepository) *Handler {
	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))
	return &Handler{categoryService: categoryService, categoryListRepository: categoryListRepository, referenceRepo: referenceRepo, template: tmpl}
}

func (h *Handler) Index(c *gin.Context) {
	categories, _ := h.categoryListRepository.GetAllCategoryRefs()
	var activeCategoryId model.Id
	var activeCategoryName model.Title
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
	catId, err := model.NewId(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid category id")
	}
	references := h.renderReferences(catId)

	// Get all categories for sidebar
	categories, _ := h.categoryListRepository.GetAllCategoryRefs()

	// Find the category name (from query or from categories list)
	categoryName, err := model.NewTitle(c.Query("categoryName"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid category name")
	}

	if categoryName == "" {
		for _, cat := range categories {
			if cat.Id == catId {
				categoryName = cat.Name
				break
			}
		}
	}

	c.HTML(http.StatusOK, "body-fragment", gin.H{
		"sidebar": SidebarData{
			Categories:       categories,
			ActiveCategoryId: catId,
		},
		"references": ReferencesData{
			CategoryId:   catId,
			CategoryName: categoryName,
			References:   references,
		},
	})
}

func (h *Handler) AddCategoryForm(c *gin.Context) {
	c.Status(http.StatusOK)
	h.template.ExecuteTemplate(c.Writer, "add-category-form.html", nil)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	name, err := model.NewTitle(c.PostForm("name"))
	if err != nil {
		c.String(http.StatusBadRequest, "Category name required")
		return
	}
	category, err := h.categoryListRepository.AddNewCategory(name)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create category")
		return
	}
	categories, _ := h.categoryListRepository.GetAllCategoryRefs()

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
	catId, err := model.NewId(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid category id")
	}

	if err := h.categoryListRepository.DeleteCategory(catId); err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete category")
		return
	}

	// Get updated categories list
	// TODO: error handling
	categories, _ := h.categoryListRepository.GetAllCategoryRefs()
	var activeCategoryId model.Id
	var activeCategoryName model.Title
	if len(categories) > 0 {
		activeCategoryId = categories[0].Id
		activeCategoryName = categories[0].Name
	}

	// TODO: error handling
	category, _ := h.categoryService.GetCategoryById(activeCategoryId)
	renderer := NewHTMLReferenceRenderer(h.template)
	for _, ref := range category.References {
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

func (h *Handler) renderReferences(categoryId model.Id) []template.HTML {
	category, _ := h.categoryService.GetCategoryById(categoryId)
	renderer := NewHTMLReferenceRenderer(h.template)
	for _, ref := range category.References {
		ref.Render(renderer)
	}
	return renderer.collected
}

type HTMLReferenceRenderer struct {
	tmpl      *template.Template
	collected []template.HTML
}

// DTOs for template rendering
type BookReferenceDTO struct {
	Id          int64
	Title       string
	ISBN        string
	Description string
	Starred     bool
}

type LinkReferenceDTO struct {
	Id          int64
	Title       string
	URL         string
	Description string
	Starred     bool
}

type NoteReferenceDTO struct {
	Id      int64
	Title   string
	Text    string
	Starred bool
}

func NewHTMLReferenceRenderer(tmpl *template.Template) *HTMLReferenceRenderer {
	return &HTMLReferenceRenderer{tmpl: tmpl, collected: make([]template.HTML, 0)}
}

func (r *HTMLReferenceRenderer) RenderBook(ref model.BookReference) {
	dto := BookReferenceDTO{
		Id:          int64(ref.GetId()),
		Title:       string(ref.Title()),
		ISBN:        string(ref.ISBN),
		Description: ref.Description,
		Starred:     ref.Starred(),
	}
	r.Render("_book", dto)
}

func (r *HTMLReferenceRenderer) RenderLink(ref model.LinkReference) {
	dto := LinkReferenceDTO{
		Id:          int64(ref.GetId()),
		Title:       string(ref.Title()),
		URL:         string(ref.URL),
		Description: ref.Description,
		Starred:     ref.Starred(),
	}
	r.Render("_link", dto)
}

func (r *HTMLReferenceRenderer) RenderNote(ref model.NoteReference) {
	dto := NoteReferenceDTO{
		Id:      int64(ref.GetId()),
		Title:   string(ref.Title()),
		Text:    ref.Text,
		Starred: ref.Starred(),
	}
	r.Render("_note", dto)
}

func (r *HTMLReferenceRenderer) Render(rendererName string, data interface{}) {
	var buf bytes.Buffer
	err := r.tmpl.ExecuteTemplate(&buf, rendererName, data)
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
	catId, err := model.NewId(categoryId)
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
		isbnStr := c.PostForm("isbn")
		description := c.PostForm("description")

		bookTitle, err := model.NewTitle(title)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid title: " + err.Error()})
			return
		}
		bookISBN, err := model.NewISBN(isbnStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid isbn: " + err.Error()})
			return
		}

		bookRef := model.NewBookReference(
			model.Id(0), // Id will be set by persistence layer
			bookTitle,
			bookISBN,
			description,
			starred,
		)

		_, err = h.categoryService.AddReference(model.Id(categoryId), bookRef)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create book reference"})
			return
		}

	case "link":
		url := model.URL(c.PostForm("url"))
		if url == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required for links"})
			return
		}
		description := c.PostForm("description")

		linkTitle, err := model.NewTitle(title)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid title: " + err.Error()})
			return
		}

		linkRef := model.NewLinkReference(
			model.Id(0), // Id will be set by persistence layer
			linkTitle,
			url,
			description,
			starred,
		)

		_, err = h.categoryService.AddReference(model.Id(categoryId), linkRef)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link reference"})
			return
		}

	case "note":
		noteTitle, err := model.NewTitle(title)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid title: " + err.Error()})
			return
		}

		content := c.PostForm("content")

		noteRef := model.NewNoteReference(
			model.Id(0), // Id will be set by persistence layer
			noteTitle,
			content,
			starred,
		)

		_, err = h.categoryService.AddReference(model.Id(categoryId), noteRef)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create note reference"})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reference type"})
		return
	}

	references := h.renderReferences(catId)
	c.HTML(http.StatusOK, "_references_list", references)
}

func (h *Handler) DeleteReference(c *gin.Context) {
	categoryIdStr := c.Param("categoryId") // TODO: pass the categoryId in here
	idStr := c.Param("id")
	if categoryIdStr == "" || idStr == "" {
		c.String(http.StatusBadRequest, "Invalid path")
		return
	}
	categoryId, err := strconv.ParseInt(categoryIdStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid category id")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid reference id")
		return
	}

	catId, err := model.NewId(categoryId)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid category id")
		return
	}
	refId, err := model.NewId(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid reference id")
		return
	}
	_, err = h.categoryService.RemoveReference(catId, refId)
	if err != nil {
		slog.Error("failed to delete reference", "error", err, "categoryId", categoryId, "id", id)
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

	refId, err := model.NewId(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid reference id")
		return
	}
	bookTitle, err := model.NewTitle(title)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid title")
		return
	}
	bookISBN, err := model.NewISBN(isbn)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ISBN")
		return
	}
	book := model.NewBookReference(
		refId,
		bookTitle,
		bookISBN,
		description,
		starred,
	)

	if err := h.referenceRepo.UpdateReference(refId, book); err != nil {
		c.String(http.StatusInternalServerError, "Failed to update reference")
		return
	}

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

	refId, err := model.NewId(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid reference id")
		return
	}
	linkTitle, err := model.NewTitle(title)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid title")
		return
	}
	linkURL, err := model.NewURL(url)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid URL")
		return
	}
	link := model.NewLinkReference(
		refId,
		linkTitle,
		linkURL,
		description,
		starred,
	)

	if err := h.referenceRepo.UpdateReference(refId, link); err != nil {
		c.String(http.StatusInternalServerError, "Failed to update reference")
		return
	}

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

	refId, err := model.NewId(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid reference id")
		return
	}
	noteTitle, err := model.NewTitle(title)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid title")
		return
	}
	note := model.NewNoteReference(
		refId,
		noteTitle,
		content,
		starred,
	)

	if err := h.referenceRepo.UpdateReference(refId, note); err != nil {
		c.String(http.StatusInternalServerError, "Failed to update reference")
		return
	}

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
	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid id")
		return
	}
	catId, err := model.NewId(idInt)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid category id")
		return
	}
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "Category name required")
		return
	}
	title, err := model.NewTitle(name)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid category name")
		return
	}
	_, err = h.categoryService.UpdateTitle(catId, title)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update category")
		return
	}
	// Return updated sidebar
	categories, _ := h.categoryListRepository.GetAllCategoryRefs()
	c.HTML(http.StatusOK, "sidebar", SidebarData{
		Categories:       categories,
		ActiveCategoryId: catId,
	})
}

func (h *Handler) ReorderCategories(c *gin.Context) {
	positionsStr := c.PostForm("positions")
	if positionsStr == "" {
		c.String(http.StatusBadRequest, "Missing positions")
		return
	}
	var positions map[model.Id]int
	if err := json.Unmarshal([]byte(positionsStr), &positions); err != nil {
		c.String(http.StatusBadRequest, "Invalid positions format")
		return
	}
	if err := h.categoryListRepository.ReorderCategories(positions); err != nil {
		c.String(http.StatusInternalServerError, "Failed to reorder categories")
		return
	}
	categories, _ := h.categoryListRepository.GetAllCategoryRefs()

	// Find the currently active category from the request or use the first one
	activeCategoryIdStr := c.PostForm("activeCategoryId")
	var activeCategoryId model.Id
	if activeCategoryIdStr != "" {
		id, err := strconv.ParseInt(activeCategoryIdStr, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid activeCategoryId")
			return
		}
		catId, err := model.NewId(id)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid activeCategoryId")
			return
		}
		activeCategoryId = catId
	} else if len(categories) > 0 {
		activeCategoryId = categories[0].Id
	}

	c.HTML(http.StatusOK, "sidebar", SidebarData{
		Categories:       categories,
		ActiveCategoryId: activeCategoryId,
	})
}

func (h *Handler) ReorderReferences(c *gin.Context) {
	categoryIdStr := c.PostForm("categoryId")
	positionsStr := c.PostForm("positions")
	if categoryIdStr == "" || positionsStr == "" {
		c.String(http.StatusBadRequest, "Missing categoryId or positions")
		return
	}
	categoryIdInt, err := strconv.ParseInt(categoryIdStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid categoryId")
		return
	}
	catId, err := model.NewId(categoryIdInt)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid categoryId")
		return
	}

	var rawPositions map[string]int
	if err := json.Unmarshal([]byte(positionsStr), &rawPositions); err != nil {
		c.String(http.StatusBadRequest, "Invalid positions format")
		return
	}
	positions := make(map[model.Id]int, len(rawPositions))
	for k, v := range rawPositions {
		idInt, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid reference id in positions")
			return
		}
		id, err := model.NewId(idInt)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid reference id in positions")
			return
		}
		positions[id] = v
	}

	_, err = h.categoryService.ReorderReferences(catId, positions)
	if err != nil {
		slog.Error("failed to reorder references", "error", err, "categoryId", categoryIdStr, "positions", positionsStr)
		c.String(http.StatusInternalServerError, "Failed to reorder references")
		return
	}
	references := h.renderReferences(catId)
	c.HTML(http.StatusOK, "_references_list", references)
}
