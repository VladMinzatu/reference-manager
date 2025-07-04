package adapters

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/stretchr/testify/require"
)

func setupCategoryListTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	err = goose.Up(db, "../db/migrations")
	require.NoError(t, err)
	return db
}

func TestSQLiteCategoryListRepository(t *testing.T) {
	db := setupCategoryListTestDB(t)
	repo := NewSQLiteCategoryListRepository(db)

	t.Run("Add and GetAllCategoryRefs", func(t *testing.T) {
		cat1, err := repo.AddNewCategory("Cat1")
		require.NoError(t, err)
		cat2, err := repo.AddNewCategory("Cat2")
		require.NoError(t, err)
		refs, err := repo.GetAllCategoryRefs()
		require.NoError(t, err)
		require.Equal(t, []model.CategoryRef{{Id: cat1.Id, Name: cat1.Name}, {Id: cat2.Id, Name: cat2.Name}}, refs)
		_ = cat1
		_ = cat2
	})

	t.Run("ReorderCategories", func(t *testing.T) {
		_, err := repo.AddNewCategory("Cat3")
		require.NoError(t, err)
		refs, err := repo.GetAllCategoryRefs()
		require.NoError(t, err)
		// Get ids
		ids := make([]model.Id, 0, 3)
		for _, ref := range refs {
			ids = append(ids, ref.Id)
		}
		// Reverse order
		positions := map[model.Id]int{
			ids[0]: 2,
			ids[1]: 1,
			ids[2]: 0,
		}
		err = repo.ReorderCategories(positions)
		require.NoError(t, err)
		// Check new order
		rows, err := db.Query(`SELECT name FROM categories ORDER BY position`)
		require.NoError(t, err)
		var reordered []string
		for rows.Next() {
			var name string
			require.NoError(t, rows.Scan(&name))
			reordered = append(reordered, name)
		}
		require.Equal(t, []string{"Cat3", "Cat2", "Cat1"}, reordered)
	})

	t.Run("DeleteCategory", func(t *testing.T) {
		// Get id of Cat2
		var idToDelete int64
		err := db.QueryRow(`SELECT id FROM categories WHERE name = ?`, "Cat2").Scan(&idToDelete)
		require.NoError(t, err)
		err = repo.DeleteCategory(model.Id(idToDelete))
		require.NoError(t, err)
		// Check remaining
		rows, err := db.Query(`SELECT name FROM categories ORDER BY position`)
		require.NoError(t, err)
		var names []string
		for rows.Next() {
			var name string
			require.NoError(t, rows.Scan(&name))
			names = append(names, name)
		}
		require.Equal(t, []string{"Cat3", "Cat1"}, names)
	})

	t.Run("Delete non-existent category", func(t *testing.T) {
		err := repo.DeleteCategory(model.Id(99999))
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})

	t.Run("Reorder with missing id", func(t *testing.T) {
		// Only provide one id
		rows, err := db.Query(`SELECT id FROM categories ORDER BY position`)
		require.NoError(t, err)
		var ids []model.Id
		for rows.Next() {
			var id int64
			require.NoError(t, rows.Scan(&id))
			ids = append(ids, model.Id(id))
		}
		positions := map[model.Id]int{
			ids[0]: 0, // missing ids[1]
		}
		err = repo.ReorderCategories(positions)
		// Should not error, but will only update one row
		require.NoError(t, err)
		// Check order: only first category should be at position 0
		var pos int
		err = db.QueryRow(`SELECT position FROM categories WHERE id = ?`, ids[0]).Scan(&pos)
		require.NoError(t, err)
		require.Equal(t, 0, pos)
	})

	t.Run("Reorder with extra id", func(t *testing.T) {
		positions := map[model.Id]int{
			model.Id(123456): 0, // non-existent id
		}
		err := repo.ReorderCategories(positions)
		// Should not error, but will not update any rows
		require.NoError(t, err)
	})

	t.Run("Reorder with duplicate position", func(t *testing.T) {
		rows, err := db.Query(`SELECT id FROM categories ORDER BY position`)
		require.NoError(t, err)
		var ids []model.Id
		for rows.Next() {
			var id int64
			require.NoError(t, rows.Scan(&id))
			ids = append(ids, model.Id(id))
		}
		positions := map[model.Id]int{
			ids[0]: 0,
			ids[1]: 1,
			ids[2]: 1, // duplicate position
		}
		err = repo.ReorderCategories(positions)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate position")
	})

	t.Run("Reorder with out-of-range position", func(t *testing.T) {
		rows, err := db.Query(`SELECT id FROM categories ORDER BY position`)
		require.NoError(t, err)
		var ids []model.Id
		for rows.Next() {
			var id int64
			require.NoError(t, rows.Scan(&id))
			ids = append(ids, model.Id(id))
		}
		positions := map[model.Id]int{
			ids[0]: 0,
			ids[1]: 1,
			ids[2]: 3, // out of range
		}
		err = repo.ReorderCategories(positions)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid position")
	})

	t.Run("Reorder with empty map", func(t *testing.T) {
		err := repo.ReorderCategories(map[model.Id]int{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "entries")
	})

	t.Run("Reorder with all ids missing", func(t *testing.T) {
		// Use completely wrong ids
		positions := map[model.Id]int{
			model.Id(9999): 0,
			model.Id(8888): 1,
			model.Id(7777): 2,
		}
		err := repo.ReorderCategories(positions)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing position for id")
	})
}
