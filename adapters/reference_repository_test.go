package adapters

import (
	"testing"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/testutils"
	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/require"
)

func TestUpdatingBookReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteReferencesRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	refId := testutils.CreateTestBookReference(t, db, catId, "Old Book", "111-111", "desc", false)

	book := model.NewBookReference(model.Id(refId), "New Book", "222-222", "newdesc", true)
	err := repo.UpdateReference(model.Id(refId), book)
	require.NoError(t, err)

	var title, isbn, desc string
	var starred bool
	err = db.QueryRow(`SELECT br.title, bk.isbn, bk.description, br.is_starred FROM base_references br JOIN book_references bk ON br.id = bk.reference_id WHERE br.id = ?`, refId).Scan(&title, &isbn, &desc, &starred)
	require.NoError(t, err)
	require.Equal(t, "New Book", title)
	require.Equal(t, "222-222", isbn)
	require.Equal(t, "newdesc", desc)
	require.True(t, starred)
}

func TestUpdatingLinkReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteReferencesRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	refId := testutils.CreateTestLinkReference(t, db, catId, "Old Link", "http://old", "olddesc", false)

	link := model.NewLinkReference(model.Id(refId), "New Link", "http://new", "newdesc", true)
	err := repo.UpdateReference(model.Id(refId), link)
	require.NoError(t, err)

	var title, url, desc string
	var starred bool
	err = db.QueryRow(`SELECT br.title, l.url, l.description, br.is_starred FROM base_references br JOIN link_references l ON br.id = l.reference_id WHERE br.id = ?`, refId).Scan(&title, &url, &desc, &starred)
	require.NoError(t, err)
	require.Equal(t, "New Link", title)
	require.Equal(t, "http://new", url)
	require.Equal(t, "newdesc", desc)
	require.True(t, starred)
}

func TestUpdatingNoteReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteReferencesRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	refId := testutils.CreateTestNoteReference(t, db, catId, "Old Note", "oldtext", false)

	note := model.NewNoteReference(model.Id(refId), "New Note", "newtext", true)
	err := repo.UpdateReference(model.Id(refId), note)
	require.NoError(t, err)

	var title, text string
	var starred bool
	err = db.QueryRow(`SELECT br.title, n.text, br.is_starred FROM base_references br JOIN note_references n ON br.id = n.reference_id WHERE br.id = ?`, refId).Scan(&title, &text, &starred)
	require.NoError(t, err)
	require.Equal(t, "New Note", title)
	require.Equal(t, "newtext", text)
	require.True(t, starred)
}

func TestUpdatingNonExistentReference(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteReferencesRepository(db)

	nonExistentId := int64(99999)
	book := model.NewBookReference(model.Id(nonExistentId), "Doesn't Exist", "000", "none", false)
	err := repo.UpdateReference(model.Id(nonExistentId), book)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reference with id")
}

func TestUpdatingReferenceWithWrongType(t *testing.T) {
	db, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	repo := NewSQLiteReferencesRepository(db)

	catId, _ := testutils.CreateTestCategory(t, db, "TestCat")
	refId := testutils.CreateTestLinkReference(t, db, catId, "Link", "http://link", "desc", false)

	book := model.NewBookReference(model.Id(refId), "Should Fail", "999", "faildesc", true)
	err := repo.UpdateReference(model.Id(refId), book)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no book reference found")

	var title, url, desc string
	var starred bool
	err = db.QueryRow(`SELECT br.title, l.url, l.description, br.is_starred FROM base_references br JOIN link_references l ON br.id = l.reference_id WHERE br.id = ?`, refId).Scan(&title, &url, &desc, &starred)
	require.NoError(t, err)
	require.Equal(t, "Link", title)
	require.Equal(t, "http://link", url)
	require.Equal(t, "desc", desc)
	require.False(t, starred)
}
