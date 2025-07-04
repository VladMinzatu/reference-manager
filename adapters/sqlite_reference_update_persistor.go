package adapters

import (
	"database/sql"
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
)

// SQLiteReferenceUpdatePersistor implements ReferencePersistor for updating references in SQLite
// (assumes the base reference has already been updated)
type SQLiteReferenceUpdatePersistor struct {
	tx    *sql.Tx
	refId int64
}

func NewSQLiteReferenceUpdatePersistor(tx *sql.Tx, refId int64) *SQLiteReferenceUpdatePersistor {
	return &SQLiteReferenceUpdatePersistor{tx: tx, refId: refId}
}

func (p *SQLiteReferenceUpdatePersistor) PersistBook(reference model.BookReference) error {
	result, err := p.tx.Exec(`UPDATE book_references SET isbn = ?, description = ? WHERE reference_id = ?`, reference.ISBN, reference.Description, p.refId)
	if err != nil {
		return fmt.Errorf("error updating book reference: %v", err)
	}
	return checkRowsAffected(result, p.refId, "book")
}

func (p *SQLiteReferenceUpdatePersistor) PersistLink(reference model.LinkReference) error {
	result, err := p.tx.Exec(`UPDATE link_references SET url = ?, description = ? WHERE reference_id = ?`, reference.URL, reference.Description, p.refId)
	if err != nil {
		return fmt.Errorf("error updating link reference: %v", err)
	}
	return checkRowsAffected(result, p.refId, "link")
}

func (p *SQLiteReferenceUpdatePersistor) PersistNote(reference model.NoteReference) error {
	result, err := p.tx.Exec(`UPDATE note_references SET text = ? WHERE reference_id = ?`, reference.Text, p.refId)
	if err != nil {
		return fmt.Errorf("error updating note reference: %v", err)
	}
	return checkRowsAffected(result, p.refId, "note")
}

func checkRowsAffected(result sql.Result, refId int64, refType string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected for %s reference: %v", refType, err)
	}
	if rows == 0 {
		return fmt.Errorf("no %s reference found with id %d", refType, refId)
	}
	return nil
}
