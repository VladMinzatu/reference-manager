package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

/*
*
The operations in this repository are performed at the level of the entire category list.
Each operation is performed in a single transaction, with consistency enforced by row-level locking.
*/
type CategoryListRepository interface {
	GetAllCategoryNames() ([]model.Title, error)
	AddNewCategory(name model.Title) (model.Category, error)
	ReorderCategories(positions map[model.Id]int) error
	DeleteCategory(id model.Id) error
}
