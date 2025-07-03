package service

import (
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/repository"
	"github.com/VladMinzatu/reference-manager/domain/util"
)

type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetCategoryById(categoryId model.Id) (*model.Category, error) {
	category, err := s.repo.GetCategoryById(categoryId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve category: %w", err)
	}
	if category == nil {
		return nil, fmt.Errorf("category with id %v not found", categoryId)
	}
	return category, nil
}

func (s *CategoryService) UpdateTitle(categoryId model.Id, title model.Title) (*model.Category, error) {
	category, err := s.GetCategoryById(categoryId)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateTitle(category.Id, title, category.Version); err != nil {
		return nil, err
	}
	category.Name = title
	category.Version++
	return category, nil
}

func (s *CategoryService) ReorderReferences(categoryId model.Id, positions map[model.Id]int) (*model.Category, error) {
	category, err := s.GetCategoryById(categoryId)
	if err != nil {
		return nil, err
	}

	if len(positions) != len(category.References) {
		return nil, fmt.Errorf("number of positions does not match number of references")
	}

	// Validate positions using shared domain utility
	ids := make([]model.Id, 0, len(category.References))
	for _, ref := range category.References {
		ids = append(ids, ref.GetId())
	}
	if err := util.ValidatePositions(ids, positions); err != nil {
		return nil, err
	}

	newOrder := make([]model.Reference, len(category.References))
	for _, ref := range category.References {
		pos := positions[ref.GetId()] // validation done above
		newOrder[pos] = ref
	}

	if err := s.repo.ReorderReferences(category.Id, positions, category.Version); err != nil {
		return nil, err
	}

	category.References = newOrder
	category.Version++
	return category, nil
}

func (s *CategoryService) AddReference(categoryId model.Id, reference model.Reference) (*model.Category, error) {
	category, err := s.GetCategoryById(categoryId)
	if err != nil {
		return nil, err
	}

	if err := s.repo.AddReference(category.Id, reference, category.Version); err != nil {
		return nil, fmt.Errorf("failed to add reference: %w", err)
	}
	category.References = append(category.References, reference)
	category.Version++
	return category, nil
}

func (s *CategoryService) RemoveReference(categoryId model.Id, referenceId model.Id) (*model.Category, error) {
	category, err := s.GetCategoryById(categoryId)
	if err != nil {
		return nil, err
	}

	if err := s.repo.RemoveReference(category.Id, referenceId, category.Version); err != nil {
		return nil, err
	}

	for i, ref := range category.References {
		if ref.GetId() == referenceId {
			category.References = append(category.References[:i], category.References[i+1:]...)
			break
		}
	}

	category.Version++
	return category, nil
}
