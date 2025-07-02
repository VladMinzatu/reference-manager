package adapters

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbFile := ":memory:"
	db, err := sql.Open("sqlite3", dbFile)
	require.NoError(t, err)
	db.SetMaxOpenConns(1)

	// Run migrations
	err = goose.Up(db, "../db/migrations")
	require.NoError(t, err)

	return db
}

func insertBaseReference(t *testing.T, db *sql.DB, title string, starred bool) int64 {
	res, err := db.Exec(`INSERT INTO categories (name, position) VALUES (?, 0)`, "TestCat")
	require.NoError(t, err)
	catId, err := res.LastInsertId()
	require.NoError(t, err)
	res, err = db.Exec(`INSERT INTO base_references (category_id, title, position, is_starred) VALUES (?, ?, 0, ?)`, catId, title, starred)
	require.NoError(t, err)
	refId, err := res.LastInsertId()
	require.NoError(t, err)
	return refId
}

func TestSQLiteReferencesRepository_UpdateReference(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteReferencesRepository(db)

	t.Run("book reference", func(t *testing.T) {
		refId := insertBaseReference(t, db, "Old Book", false)
		_, err := db.Exec(`INSERT INTO book_references (reference_id, isbn, description) VALUES (?, ?, ?)`, refId, "111-111", "desc")
		require.NoError(t, err)

		book := model.NewBookReference(model.Id(refId), "New Book", "222-222", "newdesc", true)
		err = repo.UpdateReference(model.Id(refId), book)
		require.NoError(t, err)

		var title, isbn, desc string
		var starred bool
		err = db.QueryRow(`SELECT br.title, bk.isbn, bk.description, br.is_starred FROM base_references br JOIN book_references bk ON br.id = bk.reference_id WHERE br.id = ?`, refId).Scan(&title, &isbn, &desc, &starred)
		require.NoError(t, err)
		require.Equal(t, "New Book", title)
		require.Equal(t, "222-222", isbn)
		require.Equal(t, "newdesc", desc)
		require.True(t, starred)
	})

	t.Run("link reference", func(t *testing.T) {
		refId := insertBaseReference(t, db, "Old Link", false)
		_, err := db.Exec(`INSERT INTO link_references (reference_id, url, description) VALUES (?, ?, ?)`, refId, "http://old", "olddesc")
		require.NoError(t, err)

		link := model.NewLinkReference(model.Id(refId), "New Link", "http://new", "newdesc", true)
		err = repo.UpdateReference(model.Id(refId), link)
		require.NoError(t, err)

		var title, url, desc string
		var starred bool
		err = db.QueryRow(`SELECT br.title, l.url, l.description, br.is_starred FROM base_references br JOIN link_references l ON br.id = l.reference_id WHERE br.id = ?`, refId).Scan(&title, &url, &desc, &starred)
		require.NoError(t, err)
		require.Equal(t, "New Link", title)
		require.Equal(t, "http://new", url)
		require.Equal(t, "newdesc", desc)
		require.True(t, starred)
	})

	t.Run("note reference", func(t *testing.T) {
		refId := insertBaseReference(t, db, "Old Note", false)
		_, err := db.Exec(`INSERT INTO note_references (reference_id, text) VALUES (?, ?)`, refId, "oldtext")
		require.NoError(t, err)

		note := model.NewNoteReference(model.Id(refId), "New Note", "newtext", true)
		err = repo.UpdateReference(model.Id(refId), note)
		require.NoError(t, err)

		var title, text string
		var starred bool
		err = db.QueryRow(`SELECT br.title, n.text, br.is_starred FROM base_references br JOIN note_references n ON br.id = n.reference_id WHERE br.id = ?`, refId).Scan(&title, &text, &starred)
		require.NoError(t, err)
		require.Equal(t, "New Note", title)
		require.Equal(t, "newtext", text)
		require.True(t, starred)
	})

	t.Run("non-existent reference id", func(t *testing.T) {
		nonExistentId := int64(99999)
		book := model.NewBookReference(model.Id(nonExistentId), "Doesn't Exist", "000", "none", false)
		err := repo.UpdateReference(model.Id(nonExistentId), book)
		require.Error(t, err)
		require.Contains(t, err.Error(), "reference with id")
	})

	t.Run("type mismatch: book update on link id", func(t *testing.T) {
		// Insert a link reference
		refId := insertBaseReference(t, db, "Link", false)
		_, err := db.Exec(`INSERT INTO link_references (reference_id, url, description) VALUES (?, ?, ?)`, refId, "http://link", "desc")
		require.NoError(t, err)

		// Try to update as a book
		book := model.NewBookReference(model.Id(refId), "Should Fail", "999", "faildesc", true)
		err = repo.UpdateReference(model.Id(refId), book)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no book reference found")

		// Ensure the link data is unchanged
		var title, url, desc string
		var starred bool
		err = db.QueryRow(`SELECT br.title, l.url, l.description, br.is_starred FROM base_references br JOIN link_references l ON br.id = l.reference_id WHERE br.id = ?`, refId).Scan(&title, &url, &desc, &starred)
		require.NoError(t, err)
		require.Equal(t, "Link", title)
		require.Equal(t, "http://link", url)
		require.Equal(t, "desc", desc)
		require.False(t, starred)
	})
}
