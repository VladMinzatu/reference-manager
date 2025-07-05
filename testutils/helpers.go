package testutils

import (
	"database/sql"
	"testing"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/stretchr/testify/require"
)

func CreateTestCategory(t *testing.T, db *sql.DB, name string) (model.Id, model.Version) {
	res, err := db.Exec(`INSERT INTO categories (name, position, version) VALUES (?, 0, 1)`, name)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	catId, _ := model.NewId(id)
	return catId, 1
}

func CreateTestBookReference(t *testing.T, db *sql.DB, categoryId model.Id, title, isbn, description string, starred bool) model.Id {
	// Get the next position for this category
	var position int
	err := db.QueryRow(`SELECT COALESCE(MAX(position) + 1, 0) FROM base_references WHERE category_id = ?`, categoryId).Scan(&position)
	require.NoError(t, err)

	res, err := db.Exec(`INSERT INTO base_references (category_id, title, position, is_starred) VALUES (?, ?, ?, ?)`, categoryId, title, position, starred)
	require.NoError(t, err)
	refId, err := res.LastInsertId()
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO book_references (reference_id, isbn, description) VALUES (?, ?, ?)`, refId, isbn, description)
	require.NoError(t, err)

	return model.Id(refId)
}

func CreateTestLinkReference(t *testing.T, db *sql.DB, categoryId model.Id, title, url, description string, starred bool) model.Id {
	// Get the next position for this category
	var position int
	err := db.QueryRow(`SELECT COALESCE(MAX(position) + 1, 0) FROM base_references WHERE category_id = ?`, categoryId).Scan(&position)
	require.NoError(t, err)

	res, err := db.Exec(`INSERT INTO base_references (category_id, title, position, is_starred) VALUES (?, ?, ?, ?)`, categoryId, title, position, starred)
	require.NoError(t, err)
	refId, err := res.LastInsertId()
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO link_references (reference_id, url, description) VALUES (?, ?, ?)`, refId, url, description)
	require.NoError(t, err)

	return model.Id(refId)
}

func CreateTestNoteReference(t *testing.T, db *sql.DB, categoryId model.Id, title, text string, starred bool) model.Id {
	// Get the next position for this category
	var position int
	err := db.QueryRow(`SELECT COALESCE(MAX(position) + 1, 0) FROM base_references WHERE category_id = ?`, categoryId).Scan(&position)
	require.NoError(t, err)

	res, err := db.Exec(`INSERT INTO base_references (category_id, title, position, is_starred) VALUES (?, ?, ?, ?)`, categoryId, title, position, starred)
	require.NoError(t, err)
	refId, err := res.LastInsertId()
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO note_references (reference_id, text) VALUES (?, ?)`, refId, text)
	require.NoError(t, err)

	return model.Id(refId)
}
