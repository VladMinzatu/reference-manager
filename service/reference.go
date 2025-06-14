package service

import (
	"github.com/jmoiron/sqlx"
	"github.com/vladminzatu/reference-manager/model"
)

type ReferenceService struct {
	db *sqlx.DB
}

func NewReferenceService(db *sqlx.DB) *ReferenceService {
	return &ReferenceService{db: db}
}

func (s *ReferenceService) GetBookReference(id int64) (model.BookReference, error) {
	var book model.BookReference
	err := s.db.Get(&book, "SELECT * FROM book_references WHERE id = ?", id)
	return book, err
}

func (s *ReferenceService) GetLinkReference(id int64) (model.LinkReference, error) {
	var link model.LinkReference
	err := s.db.Get(&link, "SELECT * FROM link_references WHERE id = ?", id)
	return link, err
}

func (s *ReferenceService) GetNoteReference(id int64) (model.NoteReference, error) {
	var note model.NoteReference
	err := s.db.Get(&note, "SELECT * FROM note_references WHERE id = ?", id)
	return note, err
}

func (s *ReferenceService) DeleteReference(id int64) error {
	// Try to delete from each table - one will succeed
	_, err := s.db.Exec("DELETE FROM book_references WHERE id = ?", id)
	if err == nil {
		return nil
	}
	_, err = s.db.Exec("DELETE FROM link_references WHERE id = ?", id)
	if err == nil {
		return nil
	}
	_, err = s.db.Exec("DELETE FROM note_references WHERE id = ?", id)
	return err
}
