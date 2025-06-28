package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

/*
Defines operations that are performed at the level of the entire category list.
Each operation is meant to be performed in a single transaction, with consistency enforced at the infrastructure level. (through e.g. row-level locking)
*/
type CategoryListRepository interface {
	GetAllCategoryNames() ([]model.Title, error)
	AddNewCategory(name model.Title) (model.Category, error)
	ReorderCategories(positions map[model.Id]int) error
	DeleteCategory(id model.Id) error
}
