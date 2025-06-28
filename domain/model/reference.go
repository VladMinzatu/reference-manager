package model

type Reference interface {
	GetId() Id
	Render(renderer Renderer) // this is a classic Visitor pattern
}

type Renderer interface {
	RenderBook(reference BookReference)
	RenderLink(reference LinkReference)
	RenderNote(reference NoteReference)
}

type BaseReference struct {
	Id      Id
	Title   Title
	Starred bool
}

func (b BaseReference) GetId() Id {
	return b.Id
}

type BookReference struct {
	BaseReference
	ISBN        ISBN
	Description string
}

func (b BookReference) Render(renderer Renderer) {
	renderer.RenderBook(b)
}

type LinkReference struct {
	BaseReference
	URL         URL
	Description string
}

func (l LinkReference) Render(renderer Renderer) {
	renderer.RenderLink(l)
}

type NoteReference struct {
	BaseReference
	Text string
}

func (n NoteReference) Render(renderer Renderer) {
	renderer.RenderNote(n)
}
