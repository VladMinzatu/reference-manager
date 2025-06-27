package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

type ReferencesRepository interface {
	GetAllReferences(categoryId model.Id) ([]model.Reference, error)
	AddReference(categoryId model.Id, reference model.Reference) (model.Reference, error)
	UpdateReference(id model.Id, reference model.Reference) error
	DeleteReference(id model.Id) error
	ReorderReferences(positions map[model.Id]int) error
}
