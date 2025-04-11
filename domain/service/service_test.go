package service

import (
	"testing"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetAllCategories() ([]model.Category, error) {
	args := m.Called()
	return args.Get(0).([]model.Category), args.Error(1)
}

func (m *MockRepository) AddCategory(name string) (model.Category, error) {
	args := m.Called(name)
	return args.Get(0).(model.Category), args.Error(1)
}

func (m *MockRepository) GetRefereces(categoryId int64) ([]model.Reference, error) {
	args := m.Called(categoryId)
	return args.Get(0).([]model.Reference), args.Error(1)
}

func (m *MockRepository) AddBookReferece(categoryId int64, title string, isbn string) (model.BookReference, error) {
	args := m.Called(categoryId, title, isbn)
	return args.Get(0).(model.BookReference), args.Error(1)
}

func (m *MockRepository) AddLinkReferece(categoryId int64, title string, url string, description string) (model.LinkReference, error) {
	args := m.Called(categoryId, title, url, description)
	return args.Get(0).(model.LinkReference), args.Error(1)
}

func (m *MockRepository) AddNoteReferece(categoryId int64, title string, text string) (model.NoteReference, error) {
	args := m.Called(categoryId, title, text)
	return args.Get(0).(model.NoteReference), args.Error(1)
}

func (m *MockRepository) ReorderReferences(categoryId int64, positions map[int64]int) error {
	args := m.Called(categoryId, positions)
	return args.Error(0)
}

func (m *MockRepository) ReorderCategories(positions map[int64]int) error {
	args := m.Called(positions)
	return args.Error(0)
}

func TestGetAllCategories(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	expectedCategories := []model.Category{
		{Id: 1, Name: "Category 1"},
		{Id: 2, Name: "Category 2"},
	}

	mockRepo.On("GetAllCategories").Return(expectedCategories, nil)

	categories, err := service.GetAllCategories()

	assert.NoError(t, err)
	assert.Equal(t, expectedCategories, categories)
	mockRepo.AssertExpectations(t)
}

func TestAddCategory(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	expectedCategory := model.Category{Id: 1, Name: "New Category"}
	mockRepo.On("AddCategory", "New Category").Return(expectedCategory, nil)

	category, err := service.AddCategory("New Category")

	assert.NoError(t, err)
	assert.Equal(t, expectedCategory, category)
	mockRepo.AssertExpectations(t)
}

func TestReorderCategories_ValidPositions(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	positions := map[int64]int{
		1: 0,
		2: 1,
		3: 2,
	}

	mockRepo.On("ReorderCategories", positions).Return(nil)

	err := service.ReorderCategories(positions)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestReorderCategories_InvalidPositions(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	positions := map[int64]int{
		1: 0,
		2: 0, // duplicate position
		3: 2,
	}

	err := service.ReorderCategories(positions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate position value")

	positions = map[int64]int{
		1: 0,
		2: 5, // position >= len(positions)
		3: 2,
	}

	err = service.ReorderCategories(positions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid position value")
}

func TestReorderReferences_ValidPositions(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	positions := map[int64]int{
		1: 0,
		2: 1,
		3: 2,
	}

	mockRepo.On("ReorderReferences", int64(1), positions).Return(nil)

	err := service.ReorderReferences(1, positions)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestReorderReferences_InvalidPositions(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	positions := map[int64]int{
		1: 0,
		2: 0, // duplicate position
		3: 2,
	}

	err := service.ReorderReferences(1, positions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate position value")

	positions = map[int64]int{
		1: 0,
		2: 5, // position >= len(positions)
		3: 2,
	}

	err = service.ReorderReferences(1, positions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid position value")
}
