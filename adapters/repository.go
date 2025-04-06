package adapters

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/repository"
	_ "github.com/mattn/go-sqlite3"
)

type ReferenceType string

const (
	BookReferenceType ReferenceType = "book"
	LinkReferenceType ReferenceType = "link"
	NoteReferenceType ReferenceType = "note"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (repository.Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) GetAllCategories() ([]model.Category, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.name 
		FROM categories c
		LEFT JOIN category_positions cp ON c.id = cp.category_id
		ORDER BY cp.position`)
	if err != nil {
		return nil, fmt.Errorf("error querying categories: %v", err)
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var cat model.Category
		if err := rows.Scan(&cat.Id, &cat.Name); err != nil {
			return nil, fmt.Errorf("error scanning category: %v", err)
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *SQLiteRepository) AddCategory(name string) (model.Category, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.Category{}, err
	}
	defer tx.Rollback()

	result, err := tx.Exec("INSERT INTO categories (name) VALUES (?)", name)
	if err != nil {
		return model.Category{}, fmt.Errorf("error inserting category: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Category{}, fmt.Errorf("error getting last insert id: %v", err)
	}

	// Get next position from sequence and lock
	var position int
	err = tx.QueryRow(`
		INSERT INTO category_position_sequence (id, next_position) 
		VALUES (1, 0)
		ON CONFLICT (id) DO UPDATE 
		SET next_position = next_position + 1
		RETURNING next_position`).Scan(&position)
	if err != nil {
		return model.Category{}, fmt.Errorf("error getting next position: %v", err)
	}

	// Insert position
	_, err = tx.Exec(`
		INSERT INTO category_positions (category_id, position) 
		VALUES (?, ?)`, id, position)
	if err != nil {
		return model.Category{}, fmt.Errorf("error inserting category position: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return model.Category{}, fmt.Errorf("error committing transaction: %v", err)
	}

	return model.Category{Id: id, Name: name}, nil
}

func (r *SQLiteRepository) GetRefereces(categoryId string) ([]model.Reference, error) {
	rows, err := r.db.Query(`
		SELECT br.id, br.title, br.reference_type, 
			   COALESCE(bk.isbn, '') as isbn,
			   COALESCE(l.url, '') as url,
			   COALESCE(l.description, '') as description,
			   COALESCE(n.text, '') as text
		FROM base_references br
		LEFT JOIN book_references bk ON br.id = bk.reference_id
		LEFT JOIN link_references l ON br.id = l.reference_id
		LEFT JOIN note_references n ON br.id = n.reference_id
		LEFT JOIN reference_positions rp ON br.id = rp.reference_id
		WHERE br.category_id = ?
		ORDER BY rp.position`, categoryId)
	if err != nil {
		return nil, fmt.Errorf("error querying references: %v", err)
	}
	defer rows.Close()

	var references []model.Reference
	for rows.Next() {
		var id int64
		var title, refType, isbn, url, description, text string
		if err := rows.Scan(&id, &title, &refType, &isbn, &url, &description, &text); err != nil {
			return nil, fmt.Errorf("error scanning reference: %v", err)
		}

		switch ReferenceType(refType) {
		case BookReferenceType:
			references = append(references, model.BookReference{Id: id, Title: title, ISBN: isbn})
		case LinkReferenceType:
			references = append(references, model.LinkReference{Id: id, Title: title, URL: url, Description: description})
		case NoteReferenceType:
			references = append(references, model.NoteReference{Id: id, Title: title, Text: text})
		}
	}
	return references, nil
}

func (r *SQLiteRepository) AddBookReferece(categoryId string, title string, isbn string) (model.BookReference, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.BookReference{}, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO base_references (category_id, title, reference_type) 
		VALUES (?, ?, ?)`, categoryId, title, BookReferenceType)
	if err != nil {
		return model.BookReference{}, fmt.Errorf("error inserting base reference: %v", err)
	}

	refId, err := result.LastInsertId()
	if err != nil {
		return model.BookReference{}, fmt.Errorf("error getting last insert id: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO book_references (reference_id, isbn)
		VALUES (?, ?)`, refId, isbn)
	if err != nil {
		return model.BookReference{}, fmt.Errorf("error inserting book reference: %v", err)
	}

	// Get next position from sequence and lock
	var position int
	err = tx.QueryRow(`
		INSERT INTO reference_position_sequence (category_id, next_position) 
		VALUES (?, 1)
		ON CONFLICT (category_id) DO UPDATE 
		SET next_position = next_position + 1
		RETURNING next_position`, categoryId).Scan(&position)
	if err != nil {
		return model.BookReference{}, fmt.Errorf("error getting next position: %v", err)
	}

	// Insert position
	_, err = tx.Exec(`
		INSERT INTO reference_positions (reference_id, category_id, position)
		VALUES (?, ?, ?)`, refId, categoryId, position)
	if err != nil {
		return model.BookReference{}, fmt.Errorf("error inserting reference position: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return model.BookReference{}, fmt.Errorf("error committing transaction: %v", err)
	}

	return model.BookReference{Id: refId, Title: title, ISBN: isbn}, nil
}

func (r *SQLiteRepository) AddLinkReferece(categoryId string, title string, url string, description string) (model.LinkReference, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.LinkReference{}, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO base_references (category_id, title, reference_type)
		VALUES (?, ?, ?)`, categoryId, title, LinkReferenceType)
	if err != nil {
		return model.LinkReference{}, fmt.Errorf("error inserting base reference: %v", err)
	}

	refId, err := result.LastInsertId()
	if err != nil {
		return model.LinkReference{}, fmt.Errorf("error getting last insert id: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO link_references (reference_id, url, description)
		VALUES (?, ?, ?)`, refId, url, description)
	if err != nil {
		return model.LinkReference{}, fmt.Errorf("error inserting link reference: %v", err)
	}

	// Get next position from sequence
	var position int
	err = tx.QueryRow(`
		INSERT INTO reference_position_sequence (category_id, next_position) 
		VALUES (?, 1)
		ON CONFLICT (category_id) DO UPDATE 
		SET next_position = next_position + 1
		RETURNING next_position`, categoryId).Scan(&position)
	if err != nil {
		return model.LinkReference{}, fmt.Errorf("error getting next position: %v", err)
	}

	// Insert position
	_, err = tx.Exec(`
		INSERT INTO reference_positions (reference_id, category_id, position)
		VALUES (?, ?, ?)`, refId, categoryId, position)
	if err != nil {
		return model.LinkReference{}, fmt.Errorf("error inserting reference position: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return model.LinkReference{}, fmt.Errorf("error committing transaction: %v", err)
	}

	return model.LinkReference{Id: refId, Title: title, URL: url, Description: description}, nil
}

func (r *SQLiteRepository) AddNoteReferece(categoryId string, title string, text string) (model.NoteReference, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.NoteReference{}, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO base_references (category_id, title, reference_type)
		VALUES (?, ?, ?)`, categoryId, title, NoteReferenceType)
	if err != nil {
		return model.NoteReference{}, fmt.Errorf("error inserting base reference: %v", err)
	}

	refId, err := result.LastInsertId()
	if err != nil {
		return model.NoteReference{}, fmt.Errorf("error getting last insert id: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO note_references (reference_id, text)
		VALUES (?, ?)`, refId, text)
	if err != nil {
		return model.NoteReference{}, fmt.Errorf("error inserting note reference: %v", err)
	}

	// Get next position from sequence and lock
	var position int
	err = tx.QueryRow(`
		INSERT INTO reference_position_sequence (category_id, next_position) 
		VALUES (?, 0)
		ON CONFLICT (category_id) DO UPDATE 
		SET next_position = next_position + 1
		RETURNING next_position`, categoryId).Scan(&position)
	if err != nil {
		return model.NoteReference{}, fmt.Errorf("error getting next position: %v", err)
	}

	// Insert position
	_, err = tx.Exec(`
		INSERT INTO reference_positions (reference_id, category_id, position)
		VALUES (?, ?, ?)`, refId, categoryId, position)
	if err != nil {
		return model.NoteReference{}, fmt.Errorf("error inserting reference position: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return model.NoteReference{}, fmt.Errorf("error committing transaction: %v", err)
	}

	return model.NoteReference{Id: refId, Title: title, Text: text}, nil
}

func (r *SQLiteRepository) ReorderCategories(positions map[string]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	valueStrings := make([]string, len(positions))
	valueArgs := make([]interface{}, len(positions)*2)
	i := 0
	for id, pos := range positions {
		valueStrings[i] = "(?, ?)"
		valueArgs[i*2] = id
		valueArgs[i*2+1] = pos
		i++
	}

	query := fmt.Sprintf(`
		WITH updated_items (category_id, new_position) AS (
			VALUES %s
		),
		validation AS (
			SELECT COUNT(*) as unmatched_count
			FROM categories c
			LEFT JOIN updated_items ui ON c.id = ui.category_id
			WHERE ui.category_id IS NULL
		)
		UPDATE category_positions cp
		SET position = ui.new_position
		FROM updated_items ui, validation v
		WHERE cp.category_id = ui.category_id
		AND v.unmatched_count = 0`, strings.Join(valueStrings, ","))

	result, err := tx.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("error updating category positions: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no positions were updated - some categories may be missing")
	}

	return tx.Commit()
}

func (r *SQLiteRepository) ReorderReferences(categoryId string, positions map[string]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for id, pos := range positions {
		_, err := tx.Exec(`
			INSERT OR REPLACE INTO reference_positions (reference_id, category_id, position) 
			VALUES (?, ?, ?)`, id, categoryId, pos)
		if err != nil {
			return fmt.Errorf("error updating reference position: %v", err)
		}
	}

	return tx.Commit()
}
