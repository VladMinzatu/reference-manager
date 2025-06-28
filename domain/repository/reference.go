package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

/*
This repository is used for operations that can be performed at the level of individual references in a concurrency safe way without the need to lock the entire category.
*/
type ReferencesRepository interface {
	UpdateReference(id model.Id, reference model.Reference) error
}
