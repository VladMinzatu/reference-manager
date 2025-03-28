package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

type Repository interface { // this is a port, the adaptor of which is to be implemented outside our domain directory
	GetAllCategories() ([]model.Category, error)
	AddCategory(name string) (model.Category, error)
	GetRefereces(categoryId string) ([]model.Reference, error)
	AddBookReferece(categoryId string, title string, isbn string) (model.BookReference, error)
	AddLinkReferece(categoryId string, title string, url string, description string) (model.LinkReference, error)
	AddNoteReferece(categoryId string, title string, text string) (model.NoteReference, error)
	ReorderReferences(categoryId string, positions map[string]int) error
}
