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
		SELECT id, name 
		FROM categories
		ORDER BY position`)
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

	// Concurrency note: this is safe under the assumption that we use sqlite, because this whole transaction is protected by SQLite's db-level lock
	// If we were using Postgres, this would have to be modified to lock something, e.g. WITH next_pos AS (SELECT...FOR UPDATE) INSERT INTO categories...
	// Or with a position_sequence table pattern, which has some performance benefits, but needs more complex management. We'd have to consider the specific use case (here, probably not worth it).
	result, err := tx.Exec(`
		INSERT INTO categories (name, position) 
		SELECT ?, COALESCE(MAX(position) + 1, 0) 
		FROM categories`, name)
	if err != nil {
		return model.Category{}, fmt.Errorf("error inserting category: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Category{}, fmt.Errorf("error getting last insert id: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return model.Category{}, fmt.Errorf("error committing transaction: %v", err)
	}

	return model.Category{Id: id, Name: name}, nil
}

func (r *SQLiteRepository) UpdateCategory(id int64, name string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec("UPDATE categories SET name = ? WHERE id = ?", name, id)
	if err != nil {
		return fmt.Errorf("error updating category: %v", err)
	}

	if err := r.checkUpdateResult(result); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (r *SQLiteRepository) DeleteCategory(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Again, we're taking a shortcut here that sqlite's transaction handling allows us to take.
	// If we were using e.g. Postgres, we'd have to reorder the statemets or organise them slightly differently and use SELECT...FOR UPDATE to achieve the necessary locking to make this transaction concurrency safe
	result, err := tx.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting category: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category with id %d not found", id)
	}

	// Reorder remaining categories to close any gaps
	_, err = tx.Exec(`
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY position) - 1 as new_pos
			FROM categories
		)
		UPDATE categories
		SET position = ranked.new_pos
		FROM ranked
		WHERE categories.id = ranked.id`)
	if err != nil {
		return fmt.Errorf("error reordering remaining categories: %v", err)
	}

	return tx.Commit()
}

func (r *SQLiteRepository) GetRefereces(categoryId int64) ([]model.Reference, error) {
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
			   COALESCE(bk.description, '') as book_description,
			   COALESCE(l.url, '') as url,
			   COALESCE(l.description, '') as link_description,
			   COALESCE(n.text, '') as text
		FROM base_references br
		LEFT JOIN book_references bk ON br.id = bk.reference_id
		LEFT JOIN link_references l ON br.id = l.reference_id
		LEFT JOIN note_references n ON br.id = n.reference_id
		WHERE br.category_id = ?
		ORDER BY br.position`, BOOK_TYPE, LINK_TYPE, NOTE_TYPE, categoryId)
	if err != nil {
		return nil, fmt.Errorf("error querying references: %v", err)
	}
	defer rows.Close()

	var references []model.Reference
	for rows.Next() {
		var id int64
		var title, refType, isbn, bookDescription, url, linkDescription, text string
		if err := rows.Scan(&id, &title, &refType, &isbn, &bookDescription, &url, &linkDescription, &text); err != nil {
			return nil, fmt.Errorf("error scanning reference: %v", err)
		}

		switch refType {
		case BOOK_TYPE:
			references = append(references, model.BookReference{Id: id, Title: title, ISBN: isbn, Description: bookDescription})
		case LINK_TYPE:
			references = append(references, model.LinkReference{Id: id, Title: title, URL: url, Description: linkDescription})
		case NOTE_TYPE:
			references = append(references, model.NoteReference{Id: id, Title: title, Text: text})
		}
	}
	return references, nil
}

func (r *SQLiteRepository) AddBookReferece(categoryId int64, title string, isbn string, description string) (model.BookReference, error) {
	refId, err := r.addBaseReference(categoryId, title)
	if err != nil {
		return model.BookReference{}, err
	}

	_, err = r.db.Exec(`
		INSERT INTO book_references (reference_id, isbn, description)
		VALUES (?, ?, ?)`, refId, isbn, description)
	if err != nil {
		return model.BookReference{}, fmt.Errorf("error inserting book reference: %v", err)
	}

	return model.BookReference{Id: refId, Title: title, ISBN: isbn, Description: description}, nil
}

func (r *SQLiteRepository) UpdateBookReference(id int64, title string, isbn string, description string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update base_references table
	result, err := tx.Exec("UPDATE base_references SET title = ? WHERE id = ?", title, id)
	if err != nil {
		return fmt.Errorf("error updating base reference: %v", err)
	}
	if err := r.checkUpdateResult(result); err != nil {
		return err
	}

	// Update book_references table
	result, err = tx.Exec("UPDATE book_references SET isbn = ?, description = ? WHERE reference_id = ?", isbn, description, id)
	if err != nil {
		return fmt.Errorf("error updating book reference: %v", err)
	}
	if err := r.checkUpdateResult(result); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (r *SQLiteRepository) AddLinkReferece(categoryId int64, title string, url string, description string) (model.LinkReference, error) {
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

func (r *SQLiteRepository) UpdateLinkReference(id int64, title string, url string, description string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update base_references table
	result, err := tx.Exec("UPDATE base_references SET title = ? WHERE id = ?", title, id)
	if err != nil {
		return fmt.Errorf("error updating base reference: %v", err)
	}
	if err := r.checkUpdateResult(result); err != nil {
		return err
	}

	// Update link_references table
	result, err = tx.Exec("UPDATE link_references SET url = ?, description = ? WHERE reference_id = ?", url, description, id)
	if err != nil {
		return fmt.Errorf("error updating link reference: %v", err)
	}
	if err := r.checkUpdateResult(result); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (r *SQLiteRepository) AddNoteReferece(categoryId int64, title string, text string) (model.NoteReference, error) {
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

func (r *SQLiteRepository) UpdateNoteReference(id int64, title string, text string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update base_references table
	result, err := tx.Exec("UPDATE base_references SET title = ? WHERE id = ?", title, id)
	if err != nil {
		return fmt.Errorf("error updating base reference: %v", err)
	}
	if err := r.checkUpdateResult(result); err != nil {
		return err
	}

	// Update note_references table
	result, err = tx.Exec("UPDATE note_references SET text = ? WHERE reference_id = ?", text, id)
	if err != nil {
		return fmt.Errorf("error updating note reference: %v", err)
	}
	if err := r.checkUpdateResult(result); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (r *SQLiteRepository) checkUpdateResult(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("could not find entity with specified id")
	}
	return nil
}

func (r *SQLiteRepository) addBaseReference(categoryId int64, title string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		WITH next_pos AS (
			SELECT COALESCE(MAX(position) + 1, 0) as pos 
			FROM base_references 
			WHERE category_id = ?
		)
		INSERT INTO base_references (category_id, title, position)
		SELECT ?, ?, pos FROM next_pos`, categoryId, categoryId, title)
	if err != nil {
		return 0, fmt.Errorf("error inserting base reference: %v", err)
	}

	refId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert id: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("error committing transaction: %v", err)
	}

	return refId, nil
}

func (r *SQLiteRepository) DeleteReference(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the category_id before deleting (needed for reordering)
	var categoryId int64
	err = tx.QueryRow("SELECT category_id FROM base_references WHERE id = ?", id).Scan(&categoryId)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("reference with id %d not found", id)
		}
		return fmt.Errorf("error getting category id: %v", err)
	}

	// Delete the reference
	result, err := tx.Exec("DELETE FROM base_references WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting reference: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("reference with id %d not found", id)
	}

	// Reorder remaining references in the category to close any gaps
	_, err = tx.Exec(`
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY position) - 1 as new_pos
			FROM base_references
			WHERE category_id = ?
		)
		UPDATE base_references
		SET position = ranked.new_pos
		FROM ranked
		WHERE base_references.id = ranked.id`, categoryId)
	if err != nil {
		return fmt.Errorf("error reordering remaining references: %v", err)
	}

	return tx.Commit()
}

func (r *SQLiteRepository) ReorderCategories(positions map[int64]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	valueStrings, valueArgs := r.preparePositionArgs(positions)

	// For Postgres we could use "SELECT id FROM categories FOR UPDATE" to lock the rows
	// until transaction completion. With SQLite we get db-level locking by default.
	var categoryIds []int64
	rows, err := tx.Query("SELECT id FROM categories")
	if err != nil {
		return fmt.Errorf("error querying categories: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("error scanning category id: %v", err)
		}
		categoryIds = append(categoryIds, id)
	}

	if err := r.validatePositions(categoryIds, positions); err != nil {
		return err
	}

	// Update positions directly in categories table since we've validated the input
	query := `
		WITH input_values (input_id, new_position) AS (
			VALUES ` + strings.Join(valueStrings, ",") + `
		)
		UPDATE categories 
		SET position = ui.new_position
		FROM input_values ui
		WHERE categories.id = ui.input_id`

	result, err := tx.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("error updating category positions: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no positions were updated")
	}

	return tx.Commit()
}

func (r *SQLiteRepository) ReorderReferences(categoryId int64, positions map[int64]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get all reference IDs for this category
	var referenceIds []int64
	// For Postgres we could use "SELECT id FROM categories FOR UPDATE" to lock the rows
	// until transaction completion. With SQLite we get transaction-level locking by default.
	rows, err := tx.Query("SELECT id FROM base_references WHERE category_id = ?", categoryId)
	if err != nil {
		return fmt.Errorf("error querying references: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("error scanning reference id: %v", err)
		}
		referenceIds = append(referenceIds, id)
	}

	if err := r.validatePositions(referenceIds, positions); err != nil {
		return err
	}

	valueStrings, valueArgs := r.preparePositionArgs(positions)

	// Update positions directly in base_references table since we've validated the input
	query := `
		WITH input_values (input_id, new_position) AS (
			VALUES ` + strings.Join(valueStrings, ",") + `
		)
		UPDATE base_references 
		SET position = ui.new_position
		FROM input_values ui
		WHERE base_references.id = ui.input_id
		AND base_references.category_id = ?`

	result, err := tx.Exec(query, append(valueArgs, categoryId)...)
	if err != nil {
		return fmt.Errorf("error updating reference positions: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no positions were updated")
	}

	return tx.Commit()
}

func (r *SQLiteRepository) validatePositions(existingIds []int64, positions map[int64]int) error {
	// Validate no items are missing from the positions map
	for _, id := range existingIds {
		if _, exists := positions[id]; !exists {
			return fmt.Errorf("missing position for id %d", id)
		}
	}

	// Validate no unknown items in positions map
	for id := range positions {
		found := false
		for _, knownId := range existingIds {
			if id == knownId {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid id in positions: %d", id)
		}
	}
	return nil
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
