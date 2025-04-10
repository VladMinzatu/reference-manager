package adapters

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/repository"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) repository.Repository {
	return &SQLiteRepository{db: db}
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
		INSERT INTO category_position_sequences (id, next_position) 
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
	const (
		BOOK_TYPE = "book"
		LINK_TYPE = "link"
		NOTE_TYPE = "note"
	)

	rows, err := r.db.Query(`
		SELECT br.id, br.title,
			   CASE 
				   WHEN bk.reference_id IS NOT NULL THEN ?
				   WHEN l.reference_id IS NOT NULL THEN ?
				   WHEN n.reference_id IS NOT NULL THEN ?
			   END as ref_type,
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
		ORDER BY rp.position`, BOOK_TYPE, LINK_TYPE, NOTE_TYPE, categoryId)
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

		switch refType {
		case BOOK_TYPE:
			references = append(references, model.BookReference{Id: id, Title: title, ISBN: isbn})
		case LINK_TYPE:
			references = append(references, model.LinkReference{Id: id, Title: title, URL: url, Description: description})
		case NOTE_TYPE:
			references = append(references, model.NoteReference{Id: id, Title: title, Text: text})
		}
	}
	return references, nil
}

func (r *SQLiteRepository) AddBookReferece(categoryId string, title string, isbn string) (model.BookReference, error) {
	refId, err := r.addBaseReference(categoryId, title)
	if err != nil {
		return model.BookReference{}, err
	}

	_, err = r.db.Exec(`
		INSERT INTO book_references (reference_id, isbn)
		VALUES (?, ?)`, refId, isbn)
	if err != nil {
		return model.BookReference{}, fmt.Errorf("error inserting book reference: %v", err)
	}

	return model.BookReference{Id: refId, Title: title, ISBN: isbn}, nil
}

func (r *SQLiteRepository) AddLinkReferece(categoryId string, title string, url string, description string) (model.LinkReference, error) {
	refId, err := r.addBaseReference(categoryId, title)
	if err != nil {
		return model.LinkReference{}, err
	}

	_, err = r.db.Exec(`
		INSERT INTO link_references (reference_id, url, description)
		VALUES (?, ?, ?)`, refId, url, description)
	if err != nil {
		return model.LinkReference{}, fmt.Errorf("error inserting link reference: %v", err)
	}

	return model.LinkReference{Id: refId, Title: title, URL: url, Description: description}, nil
}

func (r *SQLiteRepository) AddNoteReferece(categoryId string, title string, text string) (model.NoteReference, error) {
	refId, err := r.addBaseReference(categoryId, title)
	if err != nil {
		return model.NoteReference{}, err
	}

	_, err = r.db.Exec(`
		INSERT INTO note_references (reference_id, text)
		VALUES (?, ?)`, refId, text)
	if err != nil {
		return model.NoteReference{}, fmt.Errorf("error inserting note reference: %v", err)
	}

	return model.NoteReference{Id: refId, Title: title, Text: text}, nil
}

func (r *SQLiteRepository) addBaseReference(categoryId string, title string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO base_references (category_id, title)
		VALUES (?, ?)`, categoryId, title)
	if err != nil {
		return 0, fmt.Errorf("error inserting base reference: %v", err)
	}

	refId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert id: %v", err)
	}

	// Get next position from sequence and lock
	var position int
	err = tx.QueryRow(`
		INSERT INTO reference_position_sequences (category_id, next_position) 
		VALUES (?, 0)
		ON CONFLICT (category_id) DO UPDATE 
		SET next_position = next_position + 1
		RETURNING next_position`, categoryId).Scan(&position)
	if err != nil {
		return 0, fmt.Errorf("error getting next position: %v", err)
	}

	// Insert position
	_, err = tx.Exec(`
		INSERT INTO reference_positions (reference_id, category_id, position)
		VALUES (?, ?, ?)`, refId, categoryId, position)
	if err != nil {
		return 0, fmt.Errorf("error inserting reference position: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("error committing transaction: %v", err)
	}

	return refId, nil
}

func (r *SQLiteRepository) ReorderCategories(positions map[int64]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	valueStrings, valueArgs := r.preparePositionArgs(positions)

	missingIdsQuery := `
    SELECT COUNT(*) AS missing_categories
      FROM categories c
      LEFT JOIN input_values ui ON c.id = ui.category_id
      WHERE ui.category_id IS NULL`

	invalidIdsQuery := `
		SELECT COUNT(*) AS invalid_categories
      FROM input_values ui
      LEFT JOIN categories c ON c.id = ui.category_id
      WHERE c.id IS NULL`

	updateTable := "category_positions AS cp"
	updateConditions := `cp.category_id = ui.category_id
		AND missing_categories = 0
		AND invalid_categories = 0`

	query := r.buildPositionUpdateQuery(valueStrings, missingIdsQuery, invalidIdsQuery, updateTable, updateConditions)
	//log.Fatal(query)
	result, err := tx.Exec(query, append(valueArgs, valueArgs...)...)
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

func (r *SQLiteRepository) ReorderReferences(categoryId string, positions map[int64]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	valueStrings, valueArgs := r.preparePositionArgs(positions)

	missingIdsQuery := `
    SELECT COUNT(*) AS missing_references
      FROM base_references r
      LEFT JOIN input_values ui ON r.id = ui.reference_id
      WHERE r.category_id = ? AND ui.reference_id IS NULL `
	invalidIdsQuery := `
		SELECT COUNT(*) AS invalid_references
      FROM input_values ui
      LEFT JOIN base_references r ON r.id = ui.reference_id
      WHERE r.category_id != ? OR r.id IS NULL`

	updateTable := "reference_positions AS rp"
	updateConditions := `rp.reference_id = ui.reference_id
    AND rp.category_id = ?
    AND missing_references = 0
    AND invalid_references = 0`

	query := r.buildPositionUpdateQuery(valueStrings, missingIdsQuery, invalidIdsQuery, updateTable, updateConditions)

	result, err := tx.Exec(query, append(append(valueArgs, categoryId, categoryId), append(valueArgs, categoryId)...)...)
	if err != nil {
		return fmt.Errorf("error updating reference positions: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no positions were updated - some references may be missing or invalid")
	}

	return tx.Commit()
}

func (r *SQLiteRepository) buildPositionUpdateQuery(valueStrings []string, missingIdsQuery string, invalidIdsQuery string, updateTable string, updateConditions string) string {
	// Since we are using sqlite, having nested SELECT validation queries is safe due to sqlite's db-level concurrency model
	// However, if we were using PG, this query would not be free of race conditions with concurrent transactions adding/removing references/categories.
	// We'd have to lock the affected rows while their positions are updated, and that can be done with SELECT...FOR UPDATE statements. Would also allow us to split this into separate queries and make the code more readable.
	// But since we don't have the option to write it that way in sqlite (and we don't need to either), here goes...
	return fmt.Sprintf(`
    WITH input_values(category_id, new_position) AS (
    VALUES %s
    ),
    missing_ids_validations AS (
        %s
    ),
		invalid_ids_validation AS (
				%s
		)
    UPDATE %s
    SET position = ui.new_position
    FROM input_values ui, missing_ids_validations, invalid_ids_validation
    WHERE %s`, strings.Join(valueStrings, ","), missingIdsQuery, invalidIdsQuery, updateTable, updateConditions)
}

func (r *SQLiteRepository) preparePositionArgs(positions map[int64]int) ([]string, []interface{}) {
	valueStrings := make([]string, len(positions))
	valueArgs := make([]interface{}, len(positions)*2)
	i := 0
	for id, pos := range positions {
		valueStrings[i] = "(?, ?)"
		valueArgs[i*2] = id
		valueArgs[i*2+1] = pos
		i++
	}
	return valueStrings, valueArgs
}
