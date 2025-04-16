package adapters

import (
	"database/sql"
	"testing"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/pressly/goose"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	migrationsDir := "../db/migrations"

	goose.SetDialect("sqlite3")
	err = goose.Up(db, migrationsDir)
	require.NoError(t, err)

	cleanup := func() {
		goose.Down(db, migrationsDir)
		db.Close()
	}

	return db, cleanup
}

func TestAddingAndRetrievingCategories(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)

	cat1, err := repo.AddCategory("Category 1")
	if err != nil {
		t.Fatalf("error adding category: %v", err)
	}

	cat2, err := repo.AddCategory("Category 2")
	if err != nil {
		t.Fatalf("error adding category: %v", err)
	}

	cat3, err := repo.AddCategory("Category 3")
	if err != nil {
		t.Fatalf("error adding category: %v", err)
	}

	//Test GetAllCategories
	categories, err := repo.GetAllCategories()
	if err != nil {
		t.Fatalf("error getting categories: %v", err)
	}

	// Verify results
	expectedCategories := []model.Category{cat1, cat2, cat3}
	assert.Equal(t, expectedCategories, categories)
}
func TestDeletingCategories(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)

	// Add initial categories
	cat1, err := repo.AddCategory("Category 1")
	require.NoError(t, err)

	cat2, err := repo.AddCategory("Category 2")
	require.NoError(t, err)

	cat3, err := repo.AddCategory("Category 3")
	require.NoError(t, err)

	// Verify initial state
	categories, err := repo.GetAllCategories()
	require.NoError(t, err)
	assert.Equal(t, []model.Category{cat1, cat2, cat3}, categories)

	t.Run("delete middle category", func(t *testing.T) {
		err = repo.DeleteCategory(cat2.Id)
		require.NoError(t, err)

		categories, err = repo.GetAllCategories()
		require.NoError(t, err)
		assert.Equal(t, []model.Category{cat1, cat3}, categories)
	})

	t.Run("delete non-existent category", func(t *testing.T) {
		err = repo.DeleteCategory(999)
		assert.Error(t, err)
	})

	t.Run("delete first category", func(t *testing.T) {
		err = repo.DeleteCategory(cat1.Id)
		require.NoError(t, err)

		categories, err = repo.GetAllCategories()
		require.NoError(t, err)
		assert.Equal(t, []model.Category{cat3}, categories)
	})

	t.Run("delete last remaining category", func(t *testing.T) {
		err = repo.DeleteCategory(cat3.Id)
		require.NoError(t, err)

		categories, err = repo.GetAllCategories()
		require.NoError(t, err)
		assert.Empty(t, categories)
	})
}

func TestReorderingCategories(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)

	// Add initial categories
	cat1, err := repo.AddCategory("Category 1")
	require.NoError(t, err)

	cat2, err := repo.AddCategory("Category 2")
	require.NoError(t, err)

	cat3, err := repo.AddCategory("Category 3")
	require.NoError(t, err)

	// Verify initial order
	categories, err := repo.GetAllCategories()
	require.NoError(t, err)
	assert.Equal(t, []model.Category{cat1, cat2, cat3}, categories)
	t.Run("successful reordering", func(t *testing.T) {
		// Reorder categories - move cat3 to first position
		positions := map[int64]int{
			cat3.Id: 0,
			cat1.Id: 1,
			cat2.Id: 2,
		}
		err = repo.ReorderCategories(positions)
		require.NoError(t, err)

		// Verify new order
		categories, err = repo.GetAllCategories()
		require.NoError(t, err)
		assert.Equal(t, []model.Category{cat3, cat1, cat2}, categories)
	})

	t.Run("missing category", func(t *testing.T) {
		positions := map[int64]int{
			cat1.Id: 0,
			cat2.Id: 1,
		}
		err = repo.ReorderCategories(positions)
		assert.Error(t, err)
	})

	t.Run("invalid category id", func(t *testing.T) {
		positions := map[int64]int{
			cat1.Id: 0,
			cat2.Id: 1,
			4:       2, // <- unknown category id
		}
		err = repo.ReorderCategories(positions)
		assert.Error(t, err)
	})

	t.Run("invalid category with all valid categories specified", func(t *testing.T) {
		positions := map[int64]int{
			cat1.Id: 0,
			cat2.Id: 1,
			cat3.Id: 2,
			4:       3, // <- unknown category id
		}
		err = repo.ReorderCategories(positions)
		assert.Error(t, err)
	})
}
func TestAddAndGetReferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test category
	repo := NewSQLiteRepository(db)
	cat, err := repo.AddCategory("Test Category")
	require.NoError(t, err)

	// Add references of different types
	book, err := repo.AddBookReferece(cat.Id, "Clean Code", "978-0132350884", "A handbook of agile software craftsmanship")
	require.NoError(t, err)
	assert.Equal(t, "Clean Code", book.Title)
	assert.Equal(t, "978-0132350884", book.ISBN)
	assert.Equal(t, "A handbook of agile software craftsmanship", book.Description)

	// Test book with no description
	bookNoDesc, err := repo.AddBookReferece(cat.Id, "Design Patterns", "978-0201633610", "")
	require.NoError(t, err)
	assert.Equal(t, "Design Patterns", bookNoDesc.Title)
	assert.Equal(t, "978-0201633610", bookNoDesc.ISBN)
	assert.Equal(t, "", bookNoDesc.Description)

	link, err := repo.AddLinkReferece(cat.Id, "Go Blog", "https://go.dev/blog", "The Go Programming Language Blog")
	require.NoError(t, err)
	assert.Equal(t, "Go Blog", link.Title)
	assert.Equal(t, "https://go.dev/blog", link.URL)
	assert.Equal(t, "The Go Programming Language Blog", link.Description)

	note, err := repo.AddNoteReferece(cat.Id, "Embedded & LLVM", "Look up resources.")
	require.NoError(t, err)
	assert.Equal(t, "Embedded & LLVM", note.Title)
	assert.Equal(t, "Look up resources.", note.Text)

	// Get all references for the category
	refs, err := repo.GetRefereces(cat.Id)
	require.NoError(t, err)
	assert.Len(t, refs, 4)

	// Verify references are returned in order of insertion
	assert.Equal(t, book.Id, refs[0].(model.BookReference).Id)
	assert.Equal(t, book.Title, refs[0].(model.BookReference).Title)
	assert.Equal(t, book.ISBN, refs[0].(model.BookReference).ISBN)
	assert.Equal(t, book.Description, refs[0].(model.BookReference).Description)

	assert.Equal(t, bookNoDesc.Id, refs[1].(model.BookReference).Id)
	assert.Equal(t, bookNoDesc.Title, refs[1].(model.BookReference).Title)
	assert.Equal(t, bookNoDesc.ISBN, refs[1].(model.BookReference).ISBN)
	assert.Equal(t, bookNoDesc.Description, refs[1].(model.BookReference).Description)

	assert.Equal(t, link.Id, refs[2].(model.LinkReference).Id)
	assert.Equal(t, link.Title, refs[2].(model.LinkReference).Title)
	assert.Equal(t, link.URL, refs[2].(model.LinkReference).URL)
	assert.Equal(t, link.Description, refs[2].(model.LinkReference).Description)

	assert.Equal(t, note.Id, refs[3].(model.NoteReference).Id)
	assert.Equal(t, note.Title, refs[3].(model.NoteReference).Title)
	assert.Equal(t, note.Text, refs[3].(model.NoteReference).Text)

	// Verify references for non-existent category returns empty slice
	refs, err = repo.GetRefereces(999)
	require.NoError(t, err)
	assert.Empty(t, refs)
}

func TestDeletingReferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test category and references
	repo := NewSQLiteRepository(db)
	cat, err := repo.AddCategory("Test Category")
	require.NoError(t, err)

	book, err := repo.AddBookReferece(cat.Id, "Clean Code", "978-0132350884", "A handbook of agile software craftsmanship")
	require.NoError(t, err)

	link, err := repo.AddLinkReferece(cat.Id, "Go Blog", "https://go.dev/blog", "The Go Programming Language Blog")
	require.NoError(t, err)

	note, err := repo.AddNoteReferece(cat.Id, "Embedded & LLVM", "Look up resources.")
	require.NoError(t, err)

	t.Run("delete middle reference", func(t *testing.T) {
		// Delete the link reference (middle one)
		err = repo.DeleteReference(link.Id)
		require.NoError(t, err)

		// Verify remaining references are reordered
		refs, err := repo.GetRefereces(cat.Id)
		require.NoError(t, err)
		assert.Len(t, refs, 2)
		assert.Equal(t, book.Id, refs[0].(model.BookReference).Id)
		assert.Equal(t, note.Id, refs[1].(model.NoteReference).Id)
	})

	t.Run("delete non-existent reference", func(t *testing.T) {
		err = repo.DeleteReference(999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reference with id 999 not found")
	})

	t.Run("delete remaining references", func(t *testing.T) {
		// Delete book reference
		err = repo.DeleteReference(book.Id)
		require.NoError(t, err)

		// Delete note reference
		err = repo.DeleteReference(note.Id)
		require.NoError(t, err)

		// Verify category is empty
		refs, err := repo.GetRefereces(cat.Id)
		require.NoError(t, err)
		assert.Empty(t, refs)
	})
}

func TestReorderReferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test category and references
	repo := NewSQLiteRepository(db)
	cat, err := repo.AddCategory("Test Category")
	require.NoError(t, err)

	book, err := repo.AddBookReferece(cat.Id, "Clean Code", "978-0132350884", "A handbook of agile software craftsmanship")
	require.NoError(t, err)

	link, err := repo.AddLinkReferece(cat.Id, "Go Blog", "https://go.dev/blog", "The Go Programming Language Blog")
	require.NoError(t, err)

	note, err := repo.AddNoteReferece(cat.Id, "Embedded & LLVM", "Look up resources.")
	require.NoError(t, err)

	// Initial order should be book -> link -> note
	refs, err := repo.GetRefereces(cat.Id)
	require.NoError(t, err)
	assert.Equal(t, book.Id, refs[0].(model.BookReference).Id)
	assert.Equal(t, link.Id, refs[1].(model.LinkReference).Id)
	assert.Equal(t, note.Id, refs[2].(model.NoteReference).Id)

	t.Run("valid reordering", func(t *testing.T) {
		// Reorder to note -> book -> link
		positions := map[int64]int{
			book.Id: 1,
			link.Id: 2,
			note.Id: 0,
		}
		err = repo.ReorderReferences(cat.Id, positions)
		require.NoError(t, err)

		// Verify new order
		refs, err = repo.GetRefereces(cat.Id)
		require.NoError(t, err)
		assert.Equal(t, note.Id, refs[0].(model.NoteReference).Id)
		assert.Equal(t, book.Id, refs[1].(model.BookReference).Id)
		assert.Equal(t, link.Id, refs[2].(model.LinkReference).Id)
	})

	t.Run("missing reference", func(t *testing.T) {
		invalidPositions := map[int64]int{
			book.Id: 0,
			link.Id: 1,
			999:     2, // non-existent reference
		}
		err = repo.ReorderReferences(cat.Id, invalidPositions)
		assert.Error(t, err)
	})
	t.Run("valid references plus non-existent", func(t *testing.T) {
		invalidPositions := map[int64]int{
			book.Id: 0,
			link.Id: 1,
			note.Id: 2,
			999:     3, // non-existent reference
		}
		err = repo.ReorderReferences(cat.Id, invalidPositions)
		assert.Error(t, err)
	})

	otherCat, err := repo.AddCategory("Other Category")
	require.NoError(t, err)
	otherBook, err := repo.AddBookReferece(otherCat.Id, "Other Book", "123", "")
	require.NoError(t, err)
	t.Run("reference from different category", func(t *testing.T) {
		invalidPositions := map[int64]int{
			book.Id:      0,
			link.Id:      1,
			otherBook.Id: 2, // reference from different category
		}
		err = repo.ReorderReferences(cat.Id, invalidPositions)
		assert.Error(t, err)
	})

	t.Run("all valid references plus one from different category", func(t *testing.T) {
		invalidPositions := map[int64]int{
			book.Id:      0,
			link.Id:      1,
			note.Id:      2,
			otherBook.Id: 3, // reference from different category
		}
		err = repo.ReorderReferences(cat.Id, invalidPositions)
		assert.Error(t, err)
	})

	t.Run("valid reordering when there are other categories and references", func(t *testing.T) {
		// Reorder to note -> book -> link
		positions := map[int64]int{
			book.Id: 2,
			link.Id: 0,
			note.Id: 1,
		}
		err = repo.ReorderReferences(cat.Id, positions)
		require.NoError(t, err)

		// Verify new order
		refs, err = repo.GetRefereces(cat.Id)
		require.NoError(t, err)
		assert.Equal(t, link.Id, refs[0].(model.LinkReference).Id)
		assert.Equal(t, note.Id, refs[1].(model.NoteReference).Id)
		assert.Equal(t, book.Id, refs[2].(model.BookReference).Id)
	})
}
