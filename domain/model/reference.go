package model

type Reference interface {
	Render(renderer Renderer) // this is a classic Visitor pattern
}

type Renderer interface {
	RenderBook(reference BookReference)
	RenderLink(reference LinkReference)
	RenderNote(reference NoteReference)
}

type BookReference struct {
	Id          int64
	Title       string
	ISBN        string
	Description string
}

func (b BookReference) Render(renderer Renderer) {
	renderer.RenderBook(b)
}

type LinkReference struct {
	Id          int64
	Title       string
	URL         string
	Description string
}

func (l LinkReference) Render(renderer Renderer) {
	renderer.RenderLink(l)
}

type NoteReference struct {
	Id    int64
	Title string
	Text  string
}

func (n NoteReference) Render(renderer Renderer) {
	renderer.RenderNote(n)
}
