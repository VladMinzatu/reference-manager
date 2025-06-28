package repository

import (
	"errors"

	"github.com/VladMinzatu/reference-manager/domain/model"
)

/*
This repostory contains operations that can be performed at the level of a single category.
All operations here affect a category and its references and are performed in a single transaction protected by the versioned optimistic locking at the category level.
*/

// Sentinel error for optimistic locking/version mismatch
var ErrCategoryVersionConflict = errors.New("category version conflict: concurrent update detected")

type CategoryRepository interface {
	GetCategoryById(id model.Id) (*model.Category, error)

	UpdateTitle(id model.Id, title model.Title, version model.Version) error
	ReorderReferences(id model.Id, positions map[model.Id]int, version model.Version) error
	AddReference(id model.Id, reference model.Reference, version model.Version) error
	RemoveReference(id model.Id, referenceId model.Id, version model.Version) error
}
