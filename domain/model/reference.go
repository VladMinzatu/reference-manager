package model

type ReferenceType string

const (
	BookReferenceType ReferenceType = "book"
	LinkReferenceType ReferenceType = "link"
	NoteReferenceType ReferenceType = "note"
)

type Reference interface {
	Render(renderer Renderer) // this is a classic Visitor pattern
}

type Renderer interface {
	RenderBook(reference BookReference)
	RenderLink(reference LinkReference)
	RenderNote(reference NoteReference)
}

type BookReference struct {
	Id    string
	Title string
	ISBN  string
}

func (b BookReference) Render(renderer Renderer) {
	renderer.RenderBook(b)
}

type LinkReference struct {
	Id          string
	Title       string
	URL         string
	Description string
}

func (l LinkReference) Render(renderer Renderer) {
	renderer.RenderLink(l)
}

type NoteReference struct {
	Id    string
	Title string
	Text  string
}

func (n NoteReference) Render(renderer Renderer) {
	renderer.RenderNote(n)
}
