package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

type Repository interface { // this is a port, the adaptor of which is to be implemented outside our domain directory
	GetAllCategories() ([]model.Category, error)
	ReorderCategories(positions map[int64]int) error
	AddCategory(name string) (model.Category, error)
	UpdateCategory(id int64, name string) error
	DeleteCategory(id int64) error
	GetReferences(categoryId int64, starredOnly bool) ([]model.Reference, error)
	AddBookReference(categoryId int64, title string, isbn string, description string) (model.BookReference, error)
	UpdateBookReference(id int64, title string, isbn string, description string, starred bool) error
	AddLinkReference(categoryId int64, title string, url string, description string) (model.LinkReference, error)
	UpdateLinkReference(id int64, title string, url string, description string, starred bool) error
	AddNoteReference(categoryId int64, title string, text string) (model.NoteReference, error)
	UpdateNoteReference(id int64, title string, text string, starred bool) error
	ReorderReferences(categoryId int64, positions map[int64]int) error
	DeleteReference(id int64) error
}
