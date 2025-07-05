package adapters

import (
	"testing"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/testutils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestGetCategoryById_NotFound(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	cat, err := repo.GetCategoryById(123456)
	require.Error(t, err)
	require.Nil(t, cat)
	require.Contains(t, err.Error(), "not found")
}

func TestGetCategoryById_WithoutReferences(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")

	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Equal(t, "TestCat", string(cat.Name))
	require.Equal(t, catId, cat.Id)
	require.Equal(t, model.Version(1), cat.Version)
	require.Empty(t, cat.References)
}

func TestGetCategoryById_WithBookReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	refId := testutils.CreateTestBookReference(t, db, catId, "Test Book", "123-456", "Test description", true)

	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Equal(t, "TestCat", string(cat.Name))
	require.Len(t, cat.References, 1)

	book, ok := cat.References[0].(model.BookReference)
	require.True(t, ok)
	require.Equal(t, refId, book.GetId())
	require.Equal(t, "Test Book", string(book.Title()))
	require.Equal(t, "123-456", string(book.ISBN))
	require.Equal(t, "Test description", book.Description)
	require.True(t, book.Starred())
}

func TestGetCategoryById_WithLinkReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	refId := testutils.CreateTestLinkReference(t, db, catId, "Test Link", "http://example.com", "Test description", false)

	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Equal(t, "TestCat", string(cat.Name))
	require.Len(t, cat.References, 1)

	link, ok := cat.References[0].(model.LinkReference)
	require.True(t, ok)
	require.Equal(t, refId, link.GetId())
	require.Equal(t, "Test Link", string(link.Title()))
	require.Equal(t, "http://example.com", string(link.URL))
	require.Equal(t, "Test description", link.Description)
	require.False(t, link.Starred())
}

func TestGetCategoryById_WithNoteReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	refId := testutils.CreateTestNoteReference(t, db, catId, "Test Note", "Test note content", true)

	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Equal(t, "TestCat", string(cat.Name))
	require.Len(t, cat.References, 1)

	note, ok := cat.References[0].(model.NoteReference)
	require.True(t, ok)
	require.Equal(t, refId, note.GetId())
	require.Equal(t, "Test Note", string(note.Title()))
	require.Equal(t, "Test note content", note.Text)
	require.True(t, note.Starred())
}

func TestGetCategoryById_WithMultipleReferencesInOrder(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	bookId := testutils.CreateTestBookReference(t, db, catId, "Book 1", "111", "desc1", false)
	linkId := testutils.CreateTestLinkReference(t, db, catId, "Link 1", "http://1", "desc2", true)
	noteId := testutils.CreateTestNoteReference(t, db, catId, "Note 1", "content1", false)

	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Equal(t, "TestCat", string(cat.Name))
	require.Len(t, cat.References, 3)

	// Check order and types
	require.Equal(t, bookId, cat.References[0].GetId())
	require.Equal(t, linkId, cat.References[1].GetId())
	require.Equal(t, noteId, cat.References[2].GetId())

	_, ok1 := cat.References[0].(model.BookReference)
	require.True(t, ok1)
	_, ok2 := cat.References[1].(model.LinkReference)
	require.True(t, ok2)
	_, ok3 := cat.References[2].(model.NoteReference)
	require.True(t, ok3)
}

func TestUpdateTitle_UpdatesTitleSuccessfully(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, version := testutils.CreateTestCategory(t, db, "Old Name")

	err := repo.UpdateTitle(catId, "New Name", version)
	require.NoError(t, err)

	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Equal(t, "New Name", string(cat.Name))
	require.Equal(t, model.Version(2), cat.Version)
}

func TestUpdateTitle_FailsWithWrongVersion(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, version := testutils.CreateTestCategory(t, db, "Test Cat")

	// First update should succeed
	err := repo.UpdateTitle(catId, "Updated Name", version)
	require.NoError(t, err)

	// Second update with old version should fail
	err = repo.UpdateTitle(catId, "Should Fail", version)
	require.Error(t, err)
}

func TestUpdateTitle_FailsWithNonExistentCategory(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	err := repo.UpdateTitle(99999, "New Name", 1)
	require.Error(t, err)
}

func TestReorderReferences_ReordersReferencesCorrectly(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	bookId := testutils.CreateTestBookReference(t, db, catId, "Book 1", "111", "desc1", false)
	linkId := testutils.CreateTestLinkReference(t, db, catId, "Link 1", "http://1", "desc2", false)
	noteId := testutils.CreateTestNoteReference(t, db, catId, "Note 1", "content1", false)

	// Get current category to get updated version
	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)

	// Reverse the order
	positions := map[model.Id]int{
		bookId: 2,
		linkId: 1,
		noteId: 0,
	}

	err = repo.ReorderReferences(catId, positions, cat.Version)
	require.NoError(t, err)

	cat2, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Len(t, cat2.References, 3)
	require.Equal(t, noteId, cat2.References[0].GetId())
	require.Equal(t, linkId, cat2.References[1].GetId())
	require.Equal(t, bookId, cat2.References[2].GetId())
}

func TestReorderReferences_FailsWithWrongVersion(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	bookId := testutils.CreateTestBookReference(t, db, catId, "Book 1", "111", "desc1", false)

	positions := map[model.Id]int{
		bookId: 0,
	}

	err := repo.ReorderReferences(catId, positions, 999) // Wrong version
	require.Error(t, err)
}

func TestReorderReferences_FailsWithNonExistentCategory(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	positions := map[model.Id]int{
		model.Id(1): 0,
	}

	err := repo.ReorderReferences(99999, positions, 1)
	require.Error(t, err)
}

func TestReorderReferences_FailsWithNonExistentReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")

	positions := map[model.Id]int{
		model.Id(999): 0, // Non-existent reference
	}

	err := repo.ReorderReferences(catId, positions, 1)
	require.Error(t, err)
}

func TestAddBookReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, version := testutils.CreateTestCategory(t, db, "TestCat")
	book := model.NewBookReference(0, "New Book", "123-456", "Test description", true)

	err := repo.AddReference(catId, book, version)
	require.NoError(t, err)

	// Verify the reference was added
	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Len(t, cat.References, 1)

	addedBook, ok := cat.References[0].(model.BookReference)
	require.True(t, ok)
	require.Equal(t, "New Book", string(addedBook.Title()))
	require.Equal(t, "123-456", string(addedBook.ISBN))
	require.Equal(t, "Test description", addedBook.Description)
}

func TestAddLinkReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, version := testutils.CreateTestCategory(t, db, "TestCat")
	link := model.NewLinkReference(0, "New Link", "http://example.com", "Test description", false)

	err := repo.AddReference(catId, link, version)
	require.NoError(t, err)

	// Verify the reference was added
	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Len(t, cat.References, 1)

	addedLink, ok := cat.References[0].(model.LinkReference)
	require.True(t, ok)
	require.Equal(t, "New Link", string(addedLink.Title()))
	require.Equal(t, "http://example.com", string(addedLink.URL))
	require.Equal(t, "Test description", addedLink.Description)
	require.False(t, addedLink.Starred())
}

func TestAddNoteReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, version := testutils.CreateTestCategory(t, db, "TestCat")
	note := model.NewNoteReference(0, "New Note", "Test note content", true)

	err := repo.AddReference(catId, note, version)
	require.NoError(t, err)

	// Verify the reference was added
	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Len(t, cat.References, 1)

	addedNote, ok := cat.References[0].(model.NoteReference)
	require.True(t, ok)
	require.Equal(t, "New Note", string(addedNote.Title()))
	require.Equal(t, "Test note content", addedNote.Text)
}

func TestAddReferenceAssignsSequentialPositions(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, version := testutils.CreateTestCategory(t, db, "TestCat")
	book1 := model.NewBookReference(0, "Book 1", "111", "desc1", false)
	book2 := model.NewBookReference(0, "Book 2", "222", "desc2", false)

	err := repo.AddReference(catId, book1, version)
	require.NoError(t, err)

	// Get updated version for second add
	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)

	err = repo.AddReference(catId, book2, cat.Version)
	require.NoError(t, err)

	// Verify both references were added with correct positions
	cat2, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Len(t, cat2.References, 2)
	require.Equal(t, "Book 1", string(cat2.References[0].Title()))
	require.Equal(t, "Book 2", string(cat2.References[1].Title()))
}

func TestAddReferenceFailsWithWrongVersion(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	book := model.NewBookReference(0, "New Book", "123-456", "Test description", true)

	err := repo.AddReference(catId, book, 999) // Wrong version
	require.Error(t, err)
	require.Contains(t, err.Error(), "version")
}

func TestAddReferenceFailsWithNonExistentCategory(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	book := model.NewBookReference(0, "New Book", "123-456", "Test description", true)

	err := repo.AddReference(99999, book, 1)
	require.Error(t, err)
}

func TestRemoveReference_RemovesReferenceAndReordersRemaining(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	bookId := testutils.CreateTestBookReference(t, db, catId, "Book 1", "111", "desc1", false)
	linkId := testutils.CreateTestLinkReference(t, db, catId, "Link 1", "http://1", "desc2", false)
	noteId := testutils.CreateTestNoteReference(t, db, catId, "Note 1", "content1", false)

	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)

	err = repo.RemoveReference(catId, linkId, cat.Version)
	require.NoError(t, err)

	cat2, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Len(t, cat2.References, 2)
	require.Equal(t, bookId, cat2.References[0].GetId())
	require.Equal(t, noteId, cat2.References[1].GetId())
}

func TestRemoveReference_RemovesLastReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	bookId := testutils.CreateTestBookReference(t, db, catId, "Book 1", "111", "desc1", false)

	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)

	err = repo.RemoveReference(catId, bookId, cat.Version)
	require.NoError(t, err)

	cat2, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Empty(t, cat2.References)
}

func TestRemoveReference_RemovesFirstReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	bookId := testutils.CreateTestBookReference(t, db, catId, "Book 1", "111", "desc1", false)
	linkId := testutils.CreateTestLinkReference(t, db, catId, "Link 1", "http://1", "desc2", false)

	cat, err := repo.GetCategoryById(catId)
	require.NoError(t, err)

	err = repo.RemoveReference(catId, bookId, cat.Version)
	require.NoError(t, err)

	cat2, err := repo.GetCategoryById(catId)
	require.NoError(t, err)
	require.Len(t, cat2.References, 1)
	require.Equal(t, linkId, cat2.References[0].GetId())
}

func TestRemoveReference_FailsWithWrongVersion(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	bookId := testutils.CreateTestBookReference(t, db, catId, "Book 1", "111", "desc1", false)

	err := repo.RemoveReference(catId, bookId, 999) // Wrong version
	require.Error(t, err)
	require.Contains(t, err.Error(), "version")
}

func TestRemoveReference_FailsWithNonExistentCategory(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	err := repo.RemoveReference(99999, model.Id(1), 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestRemoveReference_FailsWithNonExistentReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")

	err := repo.RemoveReference(catId, model.Id(999), 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}
