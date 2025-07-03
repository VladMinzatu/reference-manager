package adapters

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/stretchr/testify/require"
)

func setupCategoryTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	err = goose.Up(db, "../db/migrations")
	require.NoError(t, err)
	return db
}

func createCategory(t *testing.T, db *sql.DB, name string) (model.Id, model.Version) {
	res, err := db.Exec(`INSERT INTO categories (name, position, version) VALUES (?, 0, 1)`, name)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	catId, _ := model.NewId(id)
	return catId, 1
}

func TestSQLiteCategoryRepository(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewSQLiteCategoryRepository(db)

	t.Run("GetCategoryById returns not found for missing", func(t *testing.T) {
		cat, err := repo.GetCategoryById(123456)
		require.Error(t, err)
		require.Nil(t, cat)
	})

	t.Run("AddReference and GetCategoryById", func(t *testing.T) {
		catId, version := createCategory(t, db, "TestCat")
		book := model.NewBookReference(0, "Book1", "123-456", "desc", true)
		err := repo.AddReference(catId, book, version)
		require.NoError(t, err)
		cat, err := repo.GetCategoryById(catId)
		require.NoError(t, err)
		require.Equal(t, "TestCat", string(cat.Name))
		require.Len(t, cat.References, 1)
		b, ok := cat.References[0].(model.BookReference)
		require.True(t, ok)
		require.Equal(t, "Book1", string(b.Title()))
		require.Equal(t, "123-456", string(b.ISBN))
		require.Equal(t, "desc", b.Description)
		require.True(t, b.Starred())
	})

	t.Run("UpdateTitle optimistic locking", func(t *testing.T) {
		catId, version := createCategory(t, db, "Cat2")
		err := repo.UpdateTitle(catId, "NewName", version)
		require.NoError(t, err)
		// Try again with old version
		err = repo.UpdateTitle(catId, "ShouldFail", version)
		require.Error(t, err)
		require.Contains(t, err.Error(), "version")
	})

	t.Run("ReorderReferences happy path", func(t *testing.T) {
		catId, version := createCategory(t, db, "Cat3")
		book1 := model.NewBookReference(0, "BookA", "111", "descA", false)
		book2 := model.NewBookReference(0, "BookB", "222", "descB", false)
		_ = repo.AddReference(catId, book1, version)
		_ = repo.AddReference(catId, book2, version+1)
		cat, err := repo.GetCategoryById(catId)
		require.NoError(t, err)
		require.Len(t, cat.References, 2)
		// Reverse order
		positions := map[model.Id]int{
			cat.References[0].GetId(): 1,
			cat.References[1].GetId(): 0,
		}
		err = repo.ReorderReferences(catId, positions, cat.Version)
		require.NoError(t, err)
		cat2, err := repo.GetCategoryById(catId)
		require.NoError(t, err)
		require.Equal(t, "BookB", string(cat2.References[0].Title()))
		require.Equal(t, "BookA", string(cat2.References[1].Title()))
	})

	t.Run("ReorderReferences with wrong version", func(t *testing.T) {
		catId, version := createCategory(t, db, "Cat4")
		book := model.NewBookReference(0, "BookC", "333", "descC", false)
		_ = repo.AddReference(catId, book, version)
		cat, _ := repo.GetCategoryById(catId)
		positions := map[model.Id]int{
			cat.References[0].GetId(): 0,
		}
		err := repo.ReorderReferences(catId, positions, 999)
		require.Error(t, err)
		require.Contains(t, err.Error(), "version")
	})

	t.Run("RemoveReference happy path", func(t *testing.T) {
		catId, version := createCategory(t, db, "Cat5")
		book := model.NewBookReference(0, "BookD", "444", "descD", false)
		_ = repo.AddReference(catId, book, version)
		cat, _ := repo.GetCategoryById(catId)
		require.Len(t, cat.References, 1)
		err := repo.RemoveReference(catId, cat.References[0].GetId(), cat.Version)
		require.NoError(t, err)
		cat2, _ := repo.GetCategoryById(catId)
		require.Len(t, cat2.References, 0)
	})

	t.Run("RemoveReference with wrong version", func(t *testing.T) {
		catId, version := createCategory(t, db, "Cat6")
		book := model.NewBookReference(0, "BookE", "555", "descE", false)
		_ = repo.AddReference(catId, book, version)
		cat, _ := repo.GetCategoryById(catId)
		err := repo.RemoveReference(catId, cat.References[0].GetId(), 999)
		require.Error(t, err)
		require.Contains(t, err.Error(), "version")
	})

	t.Run("AddReference with wrong version", func(t *testing.T) {
		catId, _ := createCategory(t, db, "Cat7")
		book := model.NewBookReference(0, "BookF", "666", "descF", false)
		err := repo.AddReference(catId, book, 999)
		require.Error(t, err)
		require.Contains(t, err.Error(), "version")
	})
}
