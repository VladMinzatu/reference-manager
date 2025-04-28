package service

import (
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/repository"
)

type ReferenceService struct {
	repo repository.Repository
}

func NewReferenceService(repo repository.Repository) *ReferenceService {
	srv := &ReferenceService{repo: repo}
	return srv
}

func (s *ReferenceService) GetAllCategories() ([]model.Category, error) {
	return s.repo.GetAllCategories()
}

func (s *ReferenceService) AddCategory(name string) (model.Category, error) {
	return s.repo.AddCategory(name)
}

func (s *ReferenceService) UpdateCategory(id int64, name string) error {
	return s.repo.UpdateCategory(id, name)
}

func (s *ReferenceService) DeleteCategory(id int64) error {
	return s.repo.DeleteCategory(id)
}

func (s *ReferenceService) GetReferences(categoryId int64) ([]model.Reference, error) {
	return s.repo.GetRefereces(categoryId)
}

func (s *ReferenceService) AddBookReference(categoryId int64, title string, isbn string, description string) (model.BookReference, error) {
	return s.repo.AddBookReferece(categoryId, title, isbn, description)
}

func (s *ReferenceService) UpdateBookReference(id int64, title string, isbn string, description string) error {
	return s.repo.UpdateBookReference(id, title, isbn, description)
}

func (s *ReferenceService) AddLinkReference(categoryId int64, title string, url string, description string) (model.LinkReference, error) {
	return s.repo.AddLinkReferece(categoryId, title, url, description)
}

func (s *ReferenceService) UpdateLinkReference(id int64, title string, url string, description string) error {
	return s.repo.UpdateLinkReference(id, title, url, description)
}

func (s *ReferenceService) AddNoteReference(categoryId int64, title string, text string) (model.NoteReference, error) {
	return s.repo.AddNoteReferece(categoryId, title, text)
}

func (s *ReferenceService) UpdateNoteReference(id int64, title string, text string) error {
	return s.repo.UpdateNoteReference(id, title, text)
}

func (s *ReferenceService) DeleteReference(id int64) error {
	return s.repo.DeleteReference(id)
}

func (s *ReferenceService) ReorderCategories(positions map[int64]int) error {
	if err := validatePositions(positions); err != nil {
		return err
	}
	return s.repo.ReorderCategories(positions)
}

func (s *ReferenceService) ReorderReferences(categoryId int64, positions map[int64]int) error {
	if err := validatePositions(positions); err != nil {
		return err
	}
	return s.repo.ReorderReferences(categoryId, positions)
}

func validatePositions(positions map[int64]int) error {
	n := len(positions)
	seen := make(map[int]struct{})
	for _, pos := range positions {
		if pos < 0 || pos >= n {
			return fmt.Errorf("invalid position value: positions must be in range 0..%d", n-1)
		}
		if _, exists := seen[pos]; exists {
			return fmt.Errorf("duplicate position value: %d. Each position must be unique", pos)
		}
		seen[pos] = struct{}{}
	}
	return nil
}
