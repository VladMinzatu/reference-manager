package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

/*
This repostory contains operations that can be performed at the level of a single category.
All operations here affect a category and its references and are performed in a single transaction protected by the versioned optimistic locking at the category level.
*/
type CategoriesRepository interface {
	GetCategoryById(id model.Id) (model.Category, error)
	UpdateCategoryTitle(id model.Id, name model.Title) error
	ReorderReferences(categoryId model.Id, positions map[model.Id]int) (model.Category, error)
	AddReference(categoryId model.Id, reference model.Reference) (model.Category, error)
	DeleteReference(categoryId model.Id, id model.Id) error
}
