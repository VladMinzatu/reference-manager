package adapters

import (
	"database/sql"
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/repository"
)

type SQLiteReferencesRepository struct {
	db *sql.DB
}

func NewSQLiteReferencesRepository(db *sql.DB) repository.ReferencesRepository {
	return &SQLiteReferencesRepository{db: db}
}

func (r *SQLiteReferencesRepository) UpdateReference(id model.Id, reference model.Reference) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	// Update base_references (title, starred)
	result, err := tx.Exec(`UPDATE base_references SET title = ?, is_starred = ? WHERE id = ?`, string(reference.Title()), reference.Starred(), int64(id))
	if err != nil {
		return fmt.Errorf("error updating base reference: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("reference with id %d not found", id)
	}

	// Use SQLiteReferenceUpdatePersistor for type-specific update
	persistor := NewSQLiteReferenceUpdatePersistor(tx, int64(id))
	if err := reference.Persist(persistor); err != nil {
		return err
	}

	return tx.Commit()
}
