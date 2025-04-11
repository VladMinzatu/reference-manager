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

	// Test invalid reordering (missing category)
	positions = map[int64]int{
		cat1.Id: 0,
		cat2.Id: 1,
	}
	err = repo.ReorderCategories(positions)
	assert.Error(t, err)

	// Test invalid reordering (invalid category)
	positions = map[int64]int{
		cat1.Id: 0,
		cat2.Id: 1,
		4:       2, // <- unknown category id
	}
	err = repo.ReorderCategories(positions)
	assert.Error(t, err)

	// Test invalid reordering (invalid category, with all valid categories specified)
	positions = map[int64]int{
		cat1.Id: 0,
		cat2.Id: 1,
		cat3.Id: 2,
		4:       3, // <- unknown category id
	}
	err = repo.ReorderCategories(positions)
	assert.Error(t, err)
}
func TestAddAndGetReferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test category
	repo := NewSQLiteRepository(db)
	cat, err := repo.AddCategory("Test Category")
	require.NoError(t, err)

	// Add references of different types
	book, err := repo.AddBookReferece(cat.Id, "Clean Code", "978-0132350884")
	require.NoError(t, err)
	assert.Equal(t, "Clean Code", book.Title)
	assert.Equal(t, "978-0132350884", book.ISBN)

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
	assert.Len(t, refs, 3)

	// Verify references are returned in order of insertion
	assert.Equal(t, book.Id, refs[0].(model.BookReference).Id)
	assert.Equal(t, book.Title, refs[0].(model.BookReference).Title)
	assert.Equal(t, book.ISBN, refs[0].(model.BookReference).ISBN)

	assert.Equal(t, link.Id, refs[1].(model.LinkReference).Id)
	assert.Equal(t, link.Title, refs[1].(model.LinkReference).Title)
	assert.Equal(t, link.URL, refs[1].(model.LinkReference).URL)
	assert.Equal(t, link.Description, refs[1].(model.LinkReference).Description)

	assert.Equal(t, note.Id, refs[2].(model.NoteReference).Id)
	assert.Equal(t, note.Title, refs[2].(model.NoteReference).Title)
	assert.Equal(t, note.Text, refs[2].(model.NoteReference).Text)

	// Verify references for non-existent category returns empty slice
	refs, err = repo.GetRefereces(999)
	require.NoError(t, err)
	assert.Empty(t, refs)
}
