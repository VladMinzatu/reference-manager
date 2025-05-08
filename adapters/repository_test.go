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

func TestUpdatingCategory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)

	// Add a category
	cat, err := repo.AddCategory("Original Name")
	require.NoError(t, err)

	// Update the category name
	err = repo.UpdateCategory(cat.Id, "Updated Name")
	require.NoError(t, err)

	// Retrieve categories and verify update
	categories, err := repo.GetAllCategories()
	require.NoError(t, err)
	require.Len(t, categories, 1)
	assert.Equal(t, cat.Id, categories[0].Id)
	assert.Equal(t, "Updated Name", categories[0].Name)

	// Try updating a non-existent category
	err = repo.UpdateCategory(999, "Should Fail")
	assert.Error(t, err)
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
	book, err := repo.AddBookReference(cat.Id, "Clean Code", "978-0132350884", "A handbook of agile software craftsmanship")
	require.NoError(t, err)
	assert.Equal(t, "Clean Code", book.Title)
	assert.Equal(t, "978-0132350884", book.ISBN)
	assert.Equal(t, "A handbook of agile software craftsmanship", book.Description)

	// Test book with no description
	bookNoDesc, err := repo.AddBookReference(cat.Id, "Design Patterns", "978-0201633610", "")
	require.NoError(t, err)
	assert.Equal(t, "Design Patterns", bookNoDesc.Title)
	assert.Equal(t, "978-0201633610", bookNoDesc.ISBN)
	assert.Equal(t, "", bookNoDesc.Description)

	link, err := repo.AddLinkReference(cat.Id, "Go Blog", "https://go.dev/blog", "The Go Programming Language Blog")
	require.NoError(t, err)
	assert.Equal(t, "Go Blog", link.Title)
	assert.Equal(t, "https://go.dev/blog", link.URL)
	assert.Equal(t, "The Go Programming Language Blog", link.Description)

	note, err := repo.AddNoteReference(cat.Id, "Embedded & LLVM", "Look up resources.")
	require.NoError(t, err)
	assert.Equal(t, "Embedded & LLVM", note.Title)
	assert.Equal(t, "Look up resources.", note.Text)

	// Get all references for the category
	refs, err := repo.GetReferences(cat.Id, false)
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
	refs, err = repo.GetReferences(999, false)
	require.NoError(t, err)
	assert.Empty(t, refs)
}

func TestUpdatingReferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(db)
	cat, err := repo.AddCategory("Update Test Category")
	require.NoError(t, err)

	// Add initial references
	book, err := repo.AddBookReference(cat.Id, "Original Book", "111-1111111111", "Original Description")
	require.NoError(t, err)
	link, err := repo.AddLinkReference(cat.Id, "Original Link", "https://original.com", "Original Link Description")
	require.NoError(t, err)
	note, err := repo.AddNoteReference(cat.Id, "Original Note", "Original note text")
	require.NoError(t, err)

	t.Run("update book reference", func(t *testing.T) {
		// Update the book reference and set it as non-starred
		err := repo.UpdateBookReference(book.Id, "Updated Book", "222-2222222222", "Updated Description", false)
		require.NoError(t, err)

		// Retrieve all references (non-starred included)
		refs, err := repo.GetReferences(cat.Id, false)
		require.NoError(t, err)

		// Find the updated book reference
		var found bool
		for _, ref := range refs {
			if b, ok := ref.(model.BookReference); ok && b.Id == book.Id {
				found = true
				assert.Equal(t, "Updated Book", b.Title)
				assert.Equal(t, "222-2222222222", b.ISBN)
				assert.Equal(t, "Updated Description", b.Description)
				assert.False(t, b.Starred, "Book reference should not be starred")
			}
		}
		assert.True(t, found, "Updated book reference not found")

		// Retrieve only starred references
		starredRefs, err := repo.GetReferences(cat.Id, true)
		require.NoError(t, err)

		// Ensure the updated book reference is not included in starred references
		for _, ref := range starredRefs {
			if b, ok := ref.(model.BookReference); ok && b.Id == book.Id {
				assert.Fail(t, "Non-starred book reference should not be included in starred references")
			}
		}
	})

	t.Run("star book reference", func(t *testing.T) {
		// Update the book reference and set it as starred
		err := repo.UpdateBookReference(book.Id, "Starred Book", "333-3333333333", "Starred Description", true)
		require.NoError(t, err)

		// Retrieve only starred references
		starredRefs, err := repo.GetReferences(cat.Id, true)
		require.NoError(t, err)

		// Find the updated book reference in starred references
		var found bool
		for _, ref := range starredRefs {
			if b, ok := ref.(model.BookReference); ok && b.Id == book.Id {
				found = true
				assert.Equal(t, "Starred Book", b.Title)
				assert.Equal(t, "333-3333333333", b.ISBN)
				assert.Equal(t, "Starred Description", b.Description)
				assert.True(t, b.Starred, "Book reference should be starred")
			}
		}
		assert.True(t, found, "Starred book reference not found")

		// Retrieve all references (non-starred included)
		allRefs, err := repo.GetReferences(cat.Id, false)
		require.NoError(t, err)

		// Ensure the starred book reference is included in all references
		var starredFound bool
		for _, ref := range allRefs {
			if b, ok := ref.(model.BookReference); ok && b.Id == book.Id {
				starredFound = true
				assert.Equal(t, "Starred Book", b.Title)
				assert.Equal(t, "333-3333333333", b.ISBN)
				assert.Equal(t, "Starred Description", b.Description)
				assert.True(t, b.Starred, "Book reference should be starred")
			}
		}
		assert.True(t, starredFound, "Starred book reference not found in all references")
	})

	t.Run("update link reference", func(t *testing.T) {
		// Update the link reference and set it as non-starred
		err := repo.UpdateLinkReference(link.Id, "Updated Link", "https://updated.com", "Updated Link Description", false)
		require.NoError(t, err)

		// Retrieve all references (non-starred included)
		refs, err := repo.GetReferences(cat.Id, false)
		require.NoError(t, err)

		var found bool
		for _, ref := range refs {
			if l, ok := ref.(model.LinkReference); ok && l.Id == link.Id {
				found = true
				assert.Equal(t, "Updated Link", l.Title)
				assert.Equal(t, "https://updated.com", l.URL)
				assert.Equal(t, "Updated Link Description", l.Description)
				assert.False(t, l.Starred, "Link reference should not be starred")
			}
		}
		assert.True(t, found, "Updated link reference not found")

		// Retrieve only starred references
		starredRefs, err := repo.GetReferences(cat.Id, true)
		require.NoError(t, err)

		// Ensure the updated link reference is not included in starred references
		for _, ref := range starredRefs {
			if l, ok := ref.(model.LinkReference); ok && l.Id == link.Id {
				assert.Fail(t, "Non-starred link reference should not be included in starred references")
			}
		}
	})

	t.Run("star link reference", func(t *testing.T) {
		// Update the link reference and set it as starred
		err := repo.UpdateLinkReference(link.Id, "Starred Link", "https://starred.com", "Starred Link Description", true)
		require.NoError(t, err)

		// Retrieve only starred references
		starredRefs, err := repo.GetReferences(cat.Id, true)
		require.NoError(t, err)

		// Find the updated link reference in starred references
		var found bool
		for _, ref := range starredRefs {
			if l, ok := ref.(model.LinkReference); ok && l.Id == link.Id {
				found = true
				assert.Equal(t, "Starred Link", l.Title)
				assert.Equal(t, "https://starred.com", l.URL)
				assert.Equal(t, "Starred Link Description", l.Description)
				assert.True(t, l.Starred, "Link reference should be starred")
			}
		}
		assert.True(t, found, "Starred link reference not found")

		// Retrieve all references (non-starred included)
		allRefs, err := repo.GetReferences(cat.Id, false)
		require.NoError(t, err)

		// Ensure the starred link reference is included in all references
		var starredFound bool
		for _, ref := range allRefs {
			if l, ok := ref.(model.LinkReference); ok && l.Id == link.Id {
				starredFound = true
				assert.Equal(t, "Starred Link", l.Title)
				assert.Equal(t, "https://starred.com", l.URL)
				assert.Equal(t, "Starred Link Description", l.Description)
				assert.True(t, l.Starred, "Link reference should be starred")
			}
		}
		assert.True(t, starredFound, "Starred link reference not found in all references")
	})

	t.Run("update note reference", func(t *testing.T) {
		err := repo.UpdateNoteReference(note.Id, "Updated Note", "Updated note text", false)
		require.NoError(t, err)

		refs, err := repo.GetReferences(cat.Id, false)
		require.NoError(t, err)

		var found bool
		for _, ref := range refs {
			if n, ok := ref.(model.NoteReference); ok && n.Id == note.Id {
				found = true
				assert.Equal(t, "Updated Note", n.Title)
				assert.Equal(t, "Updated note text", n.Text)
				assert.False(t, n.Starred, "Note reference should not be starred")
			}
		}
		assert.True(t, found, "Updated note reference not found")

		// Retrieve only starred references
		starredRefs, err := repo.GetReferences(cat.Id, true)
		require.NoError(t, err)

		// Ensure the updated note reference is not included in starred references
		for _, ref := range starredRefs {
			if n, ok := ref.(model.NoteReference); ok && n.Id == note.Id {
				assert.Fail(t, "Non-starred note reference should not be included in starred references")
			}
		}
	})

	t.Run("star note reference", func(t *testing.T) {
		// Update the note reference and set it as starred
		err := repo.UpdateNoteReference(note.Id, "Starred Note", "Starred note text", true)
		require.NoError(t, err)

		// Retrieve only starred references
		starredRefs, err := repo.GetReferences(cat.Id, true)
		require.NoError(t, err)

		// Find the updated note reference in starred references
		var found bool
		for _, ref := range starredRefs {
			if n, ok := ref.(model.NoteReference); ok && n.Id == note.Id {
				found = true
				assert.Equal(t, "Starred Note", n.Title)
				assert.Equal(t, "Starred note text", n.Text)
				assert.True(t, n.Starred, "Note reference should be starred")
			}
		}
		assert.True(t, found, "Starred note reference not found")

		// Retrieve all references (non-starred included)
		allRefs, err := repo.GetReferences(cat.Id, false)
		require.NoError(t, err)

		// Ensure the starred note reference is included in all references
		var starredFound bool
		for _, ref := range allRefs {
			if n, ok := ref.(model.NoteReference); ok && n.Id == note.Id {
				starredFound = true
				assert.Equal(t, "Starred Note", n.Title)
				assert.Equal(t, "Starred note text", n.Text)
				assert.True(t, n.Starred, "Note reference should be starred")
			}
		}
		assert.True(t, starredFound, "Starred note reference not found in all references")
	})

	t.Run("update non-existent book reference", func(t *testing.T) {
		err := repo.UpdateBookReference(9999, "Doesn't Exist", "000-0000000000", "No Desc", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not find entity with specified id")
	})

	t.Run("update non-existent link reference", func(t *testing.T) {
		err := repo.UpdateLinkReference(9999, "Doesn't Exist", "https://none.com", "No Desc", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not find entity with specified id")
	})

	t.Run("update non-existent note reference", func(t *testing.T) {
		err := repo.UpdateNoteReference(9999, "Doesn't Exist", "No Text", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not find entity with specified id")
	})
}

func TestDeletingReferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test category and references
	repo := NewSQLiteRepository(db)
	cat, err := repo.AddCategory("Test Category")
	require.NoError(t, err)

	book, err := repo.AddBookReference(cat.Id, "Clean Code", "978-0132350884", "A handbook of agile software craftsmanship")
	require.NoError(t, err)

	link, err := repo.AddLinkReference(cat.Id, "Go Blog", "https://go.dev/blog", "The Go Programming Language Blog")
	require.NoError(t, err)

	note, err := repo.AddNoteReference(cat.Id, "Embedded & LLVM", "Look up resources.")
	require.NoError(t, err)

	t.Run("delete middle reference", func(t *testing.T) {
		// Delete the link reference (middle one)
		err = repo.DeleteReference(link.Id)
		require.NoError(t, err)

		// Verify remaining references are reordered
		refs, err := repo.GetReferences(cat.Id, false)
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
		refs, err := repo.GetReferences(cat.Id, false)
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

	book, err := repo.AddBookReference(cat.Id, "Clean Code", "978-0132350884", "A handbook of agile software craftsmanship")
	require.NoError(t, err)

	link, err := repo.AddLinkReference(cat.Id, "Go Blog", "https://go.dev/blog", "The Go Programming Language Blog")
	require.NoError(t, err)

	note, err := repo.AddNoteReference(cat.Id, "Embedded & LLVM", "Look up resources.")
	require.NoError(t, err)

	// Initial order should be book -> link -> note
	refs, err := repo.GetReferences(cat.Id, false)
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
		refs, err = repo.GetReferences(cat.Id, false)
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
	otherBook, err := repo.AddBookReference(otherCat.Id, "Other Book", "123", "")
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
		refs, err = repo.GetReferences(cat.Id, false)
		require.NoError(t, err)
		assert.Equal(t, link.Id, refs[0].(model.LinkReference).Id)
		assert.Equal(t, note.Id, refs[1].(model.NoteReference).Id)
		assert.Equal(t, book.Id, refs[2].(model.BookReference).Id)
	})
}
