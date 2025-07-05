package adapters

import (
	"testing"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/testutils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestGetAllCategoryRefs(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	t.Run("returns empty slice when no categories exist", func(t *testing.T) {
		refs, err := repo.GetAllCategoryRefs()
		require.NoError(t, err)
		require.Empty(t, refs)
	})

	t.Run("returns all categories ordered by position", func(t *testing.T) {
		cat1, err := repo.AddNewCategory("Cat1")
		require.NoError(t, err)
		cat2, err := repo.AddNewCategory("Cat2")
		require.NoError(t, err)

		refs, err := repo.GetAllCategoryRefs()
		require.NoError(t, err)
		require.Len(t, refs, 2)
		require.Equal(t, cat1.Name, refs[0].Name)
		require.Equal(t, cat1.Id, refs[0].Id)
		require.Equal(t, cat2.Name, refs[1].Name)
		require.Equal(t, cat2.Id, refs[1].Id)
	})
}

func TestAddNewCategory(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	t.Run("creates category with correct name and position", func(t *testing.T) {
		cat, err := repo.AddNewCategory("Test Category")
		require.NoError(t, err)
		require.Equal(t, "Test Category", string(cat.Name))
		require.NotZero(t, cat.Id)

		var name string
		var position int
		err = db.QueryRow(`SELECT name, position FROM categories WHERE id = ?`, cat.Id).Scan(&name, &position)
		require.NoError(t, err)
		require.Equal(t, "Test Category", name)
		require.Equal(t, 0, position)
	})

	t.Run("assigns sequential positions to multiple categories", func(t *testing.T) {
		cat1, err := repo.AddNewCategory("First")
		require.NoError(t, err)
		cat2, err := repo.AddNewCategory("Second")
		require.NoError(t, err)
		cat3, err := repo.AddNewCategory("Third")
		require.NoError(t, err)

		// Check positions in database
		var pos1, pos2, pos3 int
		err = db.QueryRow(`SELECT position FROM categories WHERE id = ?`, cat1.Id).Scan(&pos1)
		require.NoError(t, err)
		err = db.QueryRow(`SELECT position FROM categories WHERE id = ?`, cat2.Id).Scan(&pos2)
		require.NoError(t, err)
		err = db.QueryRow(`SELECT position FROM categories WHERE id = ?`, cat3.Id).Scan(&pos3)
		require.NoError(t, err)

		require.Equal(t, 1, pos1)
		require.Equal(t, 2, pos2)
		require.Equal(t, 3, pos3)
	})

	t.Run("handles empty category name", func(t *testing.T) {
		_, err := repo.AddNewCategory("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "title cannot be empty")
	})

	t.Run("handles category name exceeding max length", func(t *testing.T) {
		longName := ""
		for i := 0; i < 256; i++ {
			longName += "a"
		}
		_, err := repo.AddNewCategory(model.Title(longName))
		require.Error(t, err)
		require.Contains(t, err.Error(), "title too long")
	})
}

func TestReorderCategoriesCorrectly(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	cat1, _ := repo.AddNewCategory("First")
	cat2, _ := repo.AddNewCategory("Second")
	cat3, _ := repo.AddNewCategory("Third")

	positions := map[model.Id]int{
		cat1.Id: 2,
		cat2.Id: 1,
		cat3.Id: 0,
	}

	err := repo.ReorderCategories(positions)
	require.NoError(t, err)

	rows, err := db.Query(`SELECT name FROM categories ORDER BY position`)
	require.NoError(t, err)
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.Equal(t, []string{"Third", "Second", "First"}, names)
}

func TestReorderCategoriesWithPartialInput(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	cat1, _ := repo.AddNewCategory("First")
	_, _ = repo.AddNewCategory("Second")
	cat3, _ := repo.AddNewCategory("Third")

	positions := map[model.Id]int{
		cat1.Id: 2,
		cat3.Id: 0,
	}

	err := repo.ReorderCategories(positions)
	require.Error(t, err)
}

func TestReorderCategoriesWithNonExsitentCagegoryId(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	cat1, _ := repo.AddNewCategory("First")
	_, _ = repo.AddNewCategory("Second")

	positions := map[model.Id]int{
		cat1.Id:       0,
		model.Id(999): 2,
	}

	err := repo.ReorderCategories(positions)
	require.Error(t, err)
}

func TestReorderCategoriesWithMissingCategories(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	cat1, _ := repo.AddNewCategory("First")
	_, _ = repo.AddNewCategory("Second")

	positions := map[model.Id]int{
		cat1.Id: 0,
	}

	err := repo.ReorderCategories(positions)
	require.Error(t, err)
}
func TestReorderCategoriesWithDuplicatePositions(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	cat1, _ := repo.AddNewCategory("First")
	cat2, _ := repo.AddNewCategory("Second")

	positions := map[model.Id]int{
		cat1.Id: 0,
		cat2.Id: 0,
	}

	err := repo.ReorderCategories(positions)
	require.Error(t, err)
}

func TestReorderCategoriesWithOutOfRangePositions(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	cat1, _ := repo.AddNewCategory("First")
	cat2, _ := repo.AddNewCategory("Second")

	positions := map[model.Id]int{
		cat1.Id: 0,
		cat2.Id: 2,
	}

	err := repo.ReorderCategories(positions)
	require.Error(t, err)
}

func TestDeleteMiddleCategory(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	cat1, _ := repo.AddNewCategory("First")
	cat2, _ := repo.AddNewCategory("Second")
	cat3, _ := repo.AddNewCategory("Third")

	// Delete the middle category
	err := repo.DeleteCategory(cat2.Id)
	require.NoError(t, err)

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM categories WHERE id = ?`, cat2.Id).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	// Verify remaining categories are reordered
	rows, err := db.Query(`SELECT name FROM categories ORDER BY position`)
	require.NoError(t, err)
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.Equal(t, []string{"First", "Third"}, names)

	// Verify positions are sequential
	var pos1, pos3 int
	err = db.QueryRow(`SELECT position FROM categories WHERE id = ?`, cat1.Id).Scan(&pos1)
	require.NoError(t, err)
	err = db.QueryRow(`SELECT position FROM categories WHERE id = ?`, cat3.Id).Scan(&pos3)
	require.NoError(t, err)
	require.Equal(t, 0, pos1)
	require.Equal(t, 1, pos3)
}

func TestDeleteLastCategory(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	cat1, _ := repo.AddNewCategory("First")
	cat2, _ := repo.AddNewCategory("Second")

	err := repo.DeleteCategory(cat2.Id)
	require.NoError(t, err)

	rows, err := db.Query(`SELECT name FROM categories ORDER BY position`)
	require.NoError(t, err)
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.Equal(t, []string{"First"}, names)

	var pos int
	err = db.QueryRow(`SELECT position FROM categories WHERE id = ?`, cat1.Id).Scan(&pos)
	require.NoError(t, err)
	require.Equal(t, 0, pos)
}

func TestDeleteCategoryWithReferences(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteCategoryListRepository(db)

	cat, _ := repo.AddNewCategory("TestCat")
	testutils.CreateTestBookReference(t, db, cat.Id, "Book1", "123", "desc", false)
	testutils.CreateTestLinkReference(t, db, cat.Id, "Link1", "http://test", "desc", false)

	err := repo.DeleteCategory(cat.Id)
	require.NoError(t, err)

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM categories WHERE id = ?`, cat.Id).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	err = db.QueryRow(`SELECT COUNT(*) FROM base_references WHERE category_id = ?`, cat.Id).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}
