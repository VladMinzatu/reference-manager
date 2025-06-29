package adapters

import (
	"database/sql"
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/repository"
)

type SQLiteCategoryRepository struct {
	db *sql.DB
}

func NewSQLiteCategoryRepository(db *sql.DB) repository.CategoryRepository {
	return &SQLiteCategoryRepository{db: db}
}

func (r *SQLiteCategoryRepository) GetCategoryById(id model.Id) (*model.Category, error) {
	const (
		BOOK_TYPE = "book"
		LINK_TYPE = "link"
		NOTE_TYPE = "note"
	)

	// Single query to get category and all references atomically
	query := `
		SELECT 
			c.id, c.name, c.version,
			br.id as ref_id, br.title as ref_title, br.position as ref_position, br.is_starred,
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
		FROM categories c
		LEFT JOIN base_references br ON c.id = br.category_id
		LEFT JOIN book_references bk ON br.id = bk.reference_id
		LEFT JOIN link_references l ON br.id = l.reference_id
		LEFT JOIN note_references n ON br.id = n.reference_id
		WHERE c.id = ?
		ORDER BY br.position`

	rows, err := r.db.Query(query, BOOK_TYPE, LINK_TYPE, NOTE_TYPE, id)
	if err != nil {
		return nil, fmt.Errorf("error querying category: %v", err)
	}
	defer rows.Close()

	var category *model.Category
	var references []model.Reference

	for rows.Next() {
		var catId int64
		var catName string
		var catVersion int64
		var refId sql.NullInt64
		var refTitle sql.NullString
		var refPosition sql.NullInt64
		var refStarred sql.NullBool
		var refType sql.NullString
		var isbn, bookDescription, url, linkDescription, text string

		err := rows.Scan(
			&catId, &catName, &catVersion,
			&refId, &refTitle, &refPosition, &refStarred,
			&refType, &isbn, &bookDescription, &url, &linkDescription, &text,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Initialize category on first row only
		if category == nil {
			id, _ := model.NewId(catId)
			version, _ := model.NewVersion(catVersion)
			category = &model.Category{
				Id:      id,
				Name:    model.Title(catName),
				Version: version,
			}
		}

		// Add reference if it exists (refId will be NULL if category has no references)
		if refId.Valid {
			switch refType.String {
			case BOOK_TYPE:
				bookId, _ := model.NewId(refId.Int64)
				bookTitle, _ := model.NewTitle(refTitle.String)
				bookISBN, _ := model.NewISBN(isbn)
				references = append(references, model.BookReference{
					BaseReference: model.BaseReference{
						Id:      bookId,
						Title:   bookTitle,
						Starred: refStarred.Bool,
					},
					ISBN:        bookISBN,
					Description: bookDescription,
				})
			case LINK_TYPE:
				linkId, _ := model.NewId(refId.Int64)
				linkTitle, _ := model.NewTitle(refTitle.String)
				linkURL, _ := model.NewURL(url)
				references = append(references, model.LinkReference{
					BaseReference: model.BaseReference{
						Id:      linkId,
						Title:   linkTitle,
						Starred: refStarred.Bool,
					},
					URL:         linkURL,
					Description: linkDescription,
				})
			case NOTE_TYPE:
				noteId, _ := model.NewId(refId.Int64)
				noteTitle, _ := model.NewTitle(refTitle.String)
				references = append(references, model.NoteReference{
					BaseReference: model.BaseReference{
						Id:      noteId,
						Title:   noteTitle,
						Starred: refStarred.Bool,
					},
					Text: text,
				})
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	if category == nil {
		return nil, fmt.Errorf("category with id %d not found", id)
	}

	category.References = references
	return category, nil
}

func (r *SQLiteCategoryRepository) UpdateTitle(id model.Id, title model.Title, version model.Version) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE categories 
		SET name = ?, version = version + 1 
		WHERE id = ? AND version = ?`, title, id, version)
	if err != nil {
		return fmt.Errorf("error updating category title: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category with id %d not found or version was out of date", id)
	}

	return tx.Commit()
}

func (r *SQLiteCategoryRepository) ReorderReferences(id model.Id, positions map[model.Id]int, version model.Version) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	// Use a single UPDATE statement for all position changes with version check - for atomicity and performance.
	query := `
		UPDATE base_references 
		SET position = CASE id 
	`
	var args []interface{}
	args = append(args, id, version)

	for refId, position := range positions {
		query += fmt.Sprintf(" WHEN %d THEN ?", refId)
		args = append(args, position)
	}
	query += ` END
		WHERE category_id = ? 
		AND EXISTS (
			SELECT 1 FROM categories 
			WHERE id = ? AND version = ?
		)`

	result, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error updating reference positions: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	expectedUpdates := int64(len(positions))
	if rowsAffected < expectedUpdates {
		return fmt.Errorf("expected to update %d references, but only updated %d; category with id %d may not exist or version was out of date", expectedUpdates, rowsAffected, id)
	}

	err = r.updateCategoryVersion(tx, id, version)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteCategoryRepository) AddReference(id model.Id, reference model.Reference, version model.Version) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	// Extract title from reference
	var title string
	switch ref := reference.(type) {
	case model.BookReference:
		title = string(ref.Title)
	case model.LinkReference:
		title = string(ref.Title)
	case model.NoteReference:
		title = string(ref.Title)
	default:
		return fmt.Errorf("unsupported reference type")
	}

	query := `
		INSERT INTO base_references (category_id, title, position, is_starred)
		SELECT ?, ?, COALESCE(MAX(position) + 1, 0), 0
		FROM base_references
		WHERE category_id = ?
		AND EXISTS (
			SELECT 1 FROM categories WHERE id = ? AND version = ?
		)`

	result, err := tx.Exec(query, id, title, id, id, version)
	if err != nil {
		return fmt.Errorf("error inserting base reference: %v", err)
	}

	refId, err := result.LastInsertId()
	if err != nil || refId == 0 {
		return fmt.Errorf("category with id %d not found or version was out of date", id)
	}

	// Add the specific reference type
	switch ref := reference.(type) {
	case model.BookReference:
		_, err = tx.Exec(`
			INSERT INTO book_references (reference_id, isbn, description) 
			SELECT ?, ?, ?
			WHERE EXISTS (
				SELECT 1 FROM categories WHERE id = ? AND version = ?
			)`, refId, ref.ISBN, ref.Description, id, version)
	case model.LinkReference:
		_, err = tx.Exec(`
			INSERT INTO link_references (reference_id, url, description) 
			SELECT ?, ?, ?
			WHERE EXISTS (
				SELECT 1 FROM categories WHERE id = ? AND version = ?
			)`, refId, ref.URL, ref.Description, id, version)
	case model.NoteReference:
		_, err = tx.Exec(`
			INSERT INTO note_references (reference_id, text) 
			SELECT ?, ?
			WHERE EXISTS (
				SELECT 1 FROM categories WHERE id = ? AND version = ?
			)`, refId, ref.Text, id, version)
	default:
		return fmt.Errorf("unsupported reference type")
	}

	if err != nil {
		return fmt.Errorf("error inserting specific reference: %v", err)
	}

	err = r.updateCategoryVersion(tx, id, version)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteCategoryRepository) RemoveReference(id model.Id, referenceId model.Id, version model.Version) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	query := `
		DELETE FROM base_references 
		WHERE id = ? AND category_id = ? 
		AND EXISTS (
			SELECT 1 FROM categories 
			WHERE id = ? AND version = ?
		)`

	result, err := tx.Exec(query, referenceId, id, id, version)
	if err != nil {
		return fmt.Errorf("error deleting reference: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category with id %d not found or version was out of date", id)
	}

	// Reorder remaining references to close any gaps
	_, err = tx.Exec(`
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY position) - 1 as new_pos
			FROM base_references
			WHERE category_id = ?
		)
		UPDATE base_references
		SET position = ranked.new_pos
		FROM ranked
		WHERE base_references.id = ranked.id AND base_references.category_id = ?`, id, id)
	if err != nil {
		return fmt.Errorf("error reordering remaining references: %v", err)
	}

	err = r.updateCategoryVersion(tx, id, version)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Helper method to update category version with optimistic locking
func (r *SQLiteCategoryRepository) updateCategoryVersion(tx *sql.Tx, id model.Id, version model.Version) error {
	result, err := tx.Exec("UPDATE categories SET version = version + 1 WHERE id = ? AND version = ?", id, version)
	if err != nil {
		return fmt.Errorf("error updating category version: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category with id %d not found or version was out of date", id)
	}

	return nil
}
