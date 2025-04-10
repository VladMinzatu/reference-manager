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
