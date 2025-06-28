package service

import (
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/repository"
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

	newOrder := make([]model.Reference, len(category.References))
	seen := make(map[int]bool)
	for _, ref := range category.References {
		pos, ok := positions[ref.GetId()]
		if !ok {
			return nil, fmt.Errorf("reference %v missing in positions", ref.GetId())
		}
		if pos < 0 || pos >= len(category.References) {
			return nil, fmt.Errorf("invalid position %d for reference %v", pos, ref.GetId())
		}
		if seen[pos] {
			return nil, fmt.Errorf("duplicate position %d", pos)
		}
		newOrder[pos] = ref
		seen[pos] = true
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
