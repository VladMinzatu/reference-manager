package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

type CategoriesRepository interface {
	GetAllCategories() ([]model.Category, error)
	AddCategory(name model.Title) (model.Category, error)
	DeleteCategory(id model.Id) error
	UpdateCategoryTitle(id model.Id, name model.Title) error
	ReorderCategories(positions map[model.Id]int) error
}
