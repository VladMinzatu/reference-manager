package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

type Repository interface { // this is a port, the adaptor of which is to be implemented outside our domain directory
	GetAllCategories() ([]model.Category, error)
	ReorderCategories(positions map[int64]int) error
	AddCategory(name string) (model.Category, error)
	DeleteCategory(id int64) error
	GetRefereces(categoryId int64) ([]model.Reference, error)
	AddBookReferece(categoryId int64, title string, isbn string, description string) (model.BookReference, error)
	AddLinkReferece(categoryId int64, title string, url string, description string) (model.LinkReference, error)
	AddNoteReferece(categoryId int64, title string, text string) (model.NoteReference, error)
	ReorderReferences(categoryId int64, positions map[int64]int) error
	DeleteReference(id int64) error
}
