package model

type Reference interface {
	GetId() Id
	Title() Title
	Starred() bool
	Render(renderer Renderer)                   // this is a classic Visitor pattern
	Persist(persistor ReferencePersistor) error // so is this
}

type Renderer interface {
	RenderBook(reference BookReference)
	RenderLink(reference LinkReference)
	RenderNote(reference NoteReference)
}

type ReferencePersistor interface {
	PersistBook(reference BookReference) error
	PersistLink(reference LinkReference) error
	PersistNote(reference NoteReference) error
}

type BaseReference struct {
	id      Id
	title   Title
	starred bool
}

func (b BaseReference) GetId() Id {
	return b.id
}

func (b BaseReference) Title() Title {
	return b.title
}

func (b BaseReference) Starred() bool {
	return b.starred
}

type BookReference struct {
	BaseReference
	ISBN        ISBN
	Description string
}

func NewBookReference(id Id, title Title, isbn ISBN, description string, starred bool) BookReference {
	return BookReference{
		BaseReference: BaseReference{
			id:      id,
			title:   title,
			starred: starred,
		},
		ISBN:        isbn,
		Description: description,
	}
}

func (b BookReference) Render(renderer Renderer) {
	renderer.RenderBook(b)
}

func (b BookReference) Persist(persistor ReferencePersistor) error {
	return persistor.PersistBook(b)
}

type LinkReference struct {
	BaseReference
	URL         URL
	Description string
}

func NewLinkReference(id Id, title Title, url URL, description string, starred bool) LinkReference {
	return LinkReference{
		BaseReference: BaseReference{
			id:      id,
			title:   title,
			starred: starred,
		},
		URL:         url,
		Description: description,
	}
}

func (l LinkReference) Render(renderer Renderer) {
	renderer.RenderLink(l)
}

func (l LinkReference) Persist(persistor ReferencePersistor) error {
	return persistor.PersistLink(l)
}

type NoteReference struct {
	BaseReference
	Text string
}

func NewNoteReference(id Id, title Title, text string, starred bool) NoteReference {
	return NoteReference{
		BaseReference: BaseReference{
			id:      id,
			title:   title,
			starred: starred,
		},
		Text: text,
	}
}

func (n NoteReference) Render(renderer Renderer) {
	renderer.RenderNote(n)
}

func (n NoteReference) Persist(persistor ReferencePersistor) error {
	return persistor.PersistNote(n)
}
