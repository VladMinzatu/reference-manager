package repository

import "github.com/VladMinzatu/reference-manager/domain/model"

type Repository interface { // this is a port, the adaptor of which is to be implemented outside our domain directory
	GetAllReferences() ([]model.Reference, error)
}
