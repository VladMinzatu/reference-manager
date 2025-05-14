package service

import (
	"errors"
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

func (m *MockRepository) UpdateCategory(id int64, name string) error {
	args := m.Called(id, name)
	return args.Error(1)
}

func (m *MockRepository) DeleteCategory(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) GetReferences(categoryId int64, starredOnly bool) ([]model.Reference, error) {
	args := m.Called(categoryId, starredOnly)
	return args.Get(0).([]model.Reference), args.Error(1)
}

func (m *MockRepository) AddBookReference(categoryId int64, title string, isbn string, description string) (model.BookReference, error) {
	args := m.Called(categoryId, title, isbn, description)
	return args.Get(0).(model.BookReference), args.Error(1)
}

func (m *MockRepository) UpdateBookReference(id int64, title string, isbn string, description string, starred bool) error {
	args := m.Called(id, title, isbn, description, starred)
	return args.Error(1)
}

func (m *MockRepository) AddLinkReference(categoryId int64, title string, url string, description string) (model.LinkReference, error) {
	args := m.Called(categoryId, title, url, description)
	return args.Get(0).(model.LinkReference), args.Error(1)
}

func (m *MockRepository) UpdateLinkReference(id int64, title string, url string, description string, starred bool) error {
	args := m.Called(id, title, url, description, starred)
	return args.Error(1)
}

func (m *MockRepository) AddNoteReference(categoryId int64, title string, text string) (model.NoteReference, error) {
	args := m.Called(categoryId, title, text)
	return args.Get(0).(model.NoteReference), args.Error(1)
}

func (m *MockRepository) UpdateNoteReference(id int64, title string, text string, starred bool) error {
	args := m.Called(id, title, text, starred)
	return args.Error(1)
}

func (m *MockRepository) DeleteReference(id int64) error {
	args := m.Called(id)
	return args.Error(0)
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

func TestUpdateCategory(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	expectedCategory := model.Category{Id: 1, Name: "Updated Category"}
	mockRepo.On("UpdateCategory", int64(1), "Updated Category").Return(expectedCategory, nil)

	err := service.UpdateCategory(1, "Updated Category")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateCategory_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	mockRepo.On("UpdateCategory", int64(1), "Updated Category").Return(model.Category{}, errors.New("update error"))

	err := service.UpdateCategory(1, "Updated Category")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateBookReference(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	expectedBook := model.BookReference{BaseReference: model.BaseReference{Id: 1, Title: "Book", Starred: false}, ISBN: "123", Description: "desc"}
	mockRepo.On("UpdateBookReference", int64(1), "Book", "123", "desc", false).Return(expectedBook, nil)

	err := service.UpdateBookReference(1, "Book", "123", "desc", false)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateBookReference_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	mockRepo.On("UpdateBookReference", int64(1), "Book", "123", "desc", false).Return(model.BookReference{}, errors.New("update error"))

	err := service.UpdateBookReference(1, "Book", "123", "desc", false)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateLinkReference(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	expectedLink := model.LinkReference{BaseReference: model.BaseReference{Id: 1, Title: "Link", Starred: false}, URL: "http://a", Description: "desc"}
	mockRepo.On("UpdateLinkReference", int64(1), "Link", "http://a", "desc", false).Return(expectedLink, nil)

	err := service.UpdateLinkReference(1, "Link", "http://a", "desc", false)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateLinkReference_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	mockRepo.On("UpdateLinkReference", int64(1), "Link", "http://a", "desc", false).Return(model.LinkReference{}, errors.New("update error"))

	err := service.UpdateLinkReference(1, "Link", "http://a", "desc", false)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateNoteReference(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	expectedNote := model.NoteReference{BaseReference: model.BaseReference{Id: 1, Title: "Note", Starred: false}, Text: "text"}
	mockRepo.On("UpdateNoteReference", int64(1), "Note", "text", false).Return(expectedNote, nil)

	err := service.UpdateNoteReference(1, "Note", "text", false)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateNoteReference_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewReferenceService(mockRepo)

	mockRepo.On("UpdateNoteReference", int64(1), "Note", "text", false).Return(model.NoteReference{}, errors.New("update error"))

	err := service.UpdateNoteReference(1, "Note", "text", false)

	assert.Error(t, err)
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
