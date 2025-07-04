package adapters

import (
	"database/sql"
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
)

// SQLiteReferenceAddPersistor implements ReferencePersistor for adding references in SQLite
type SQLiteReferenceAddPersistor struct {
	categoryId model.Id
	version    model.Version
	tx         *sql.Tx
	baseRefId  int64
}

func NewSQLiteReferenceAddPersistor(categoryId model.Id, version model.Version, tx *sql.Tx, baseRefId int64) *SQLiteReferenceAddPersistor {
	return &SQLiteReferenceAddPersistor{
		categoryId: categoryId,
		version:    version,
		tx:         tx,
		baseRefId:  baseRefId,
	}
}

func (p *SQLiteReferenceAddPersistor) PersistBook(reference model.BookReference) error {
	_, err := p.tx.Exec(`
		INSERT INTO book_references (reference_id, isbn, description) 
		SELECT ?, ?, ?
		WHERE EXISTS (
			SELECT 1 FROM categories WHERE id = ? AND version = ?
		)`, p.baseRefId, reference.ISBN, reference.Description, p.categoryId, p.version)
	if err != nil {
		return fmt.Errorf("error inserting book reference: %v", err)
	}
	return nil
}

func (p *SQLiteReferenceAddPersistor) PersistLink(reference model.LinkReference) error {
	_, err := p.tx.Exec(`
		INSERT INTO link_references (reference_id, url, description) 
		SELECT ?, ?, ?
		WHERE EXISTS (
			SELECT 1 FROM categories WHERE id = ? AND version = ?
		)`, p.baseRefId, reference.URL, reference.Description, p.categoryId, p.version)
	if err != nil {
		return fmt.Errorf("error inserting link reference: %v", err)
	}
	return nil
}

func (p *SQLiteReferenceAddPersistor) PersistNote(reference model.NoteReference) error {
	_, err := p.tx.Exec(`
		INSERT INTO note_references (reference_id, text) 
		SELECT ?, ?
		WHERE EXISTS (
			SELECT 1 FROM categories WHERE id = ? AND version = ?
		)`, p.baseRefId, reference.Text, p.categoryId, p.version)
	if err != nil {
		return fmt.Errorf("error inserting note reference: %v", err)
	}
	return nil
}
