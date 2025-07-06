package adapters

import (
	"database/sql"
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
)

type SQLiteCategoryRepository struct {
	db *sql.DB
}

func NewSQLiteCategoryRepository(db *sql.DB) *SQLiteCategoryRepository {
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

		// Next, add reference if it exists (refId will be NULL if category has no references)

		// I need to add a note here for the curious reader: yes, this is a switch statement, but because it's coming from the persistence, it's not a switch on type
		// and cannot be removed via double dispatch. Since it's the only place where it happens, I think adding an abstract Factory here is overkill, as it would require
		// the same kind of update when adding a new reference type.
		// On the plus side, the impact of forgetting to add support for a new type here is not big - the new references wold just not show up.
		// Almost certainly something that will not cause more than 5 min of head scratching during development at worst.
		if refId.Valid {
			var ref model.Reference
			switch refType.String {
			case BOOK_TYPE:
				ref = buildBookReference(refId, refTitle, refStarred, isbn, bookDescription)
			case LINK_TYPE:
				ref = buildLinkReference(refId, refTitle, refStarred, url, linkDescription)
			case NOTE_TYPE:
				ref = buildNoteReference(refId, refTitle, refStarred, text)
			}
			if ref != nil {
				references = append(references, ref)
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

func buildBookReference(refId sql.NullInt64, refTitle sql.NullString, refStarred sql.NullBool, isbn, bookDescription string) model.Reference {
	bookId, _ := model.NewId(refId.Int64)
	bookTitle, _ := model.NewTitle(refTitle.String)
	bookISBN, _ := model.NewISBN(isbn)
	return model.NewBookReference(bookId, bookTitle, bookISBN, bookDescription, refStarred.Bool)
}

func buildLinkReference(refId sql.NullInt64, refTitle sql.NullString, refStarred sql.NullBool, url, linkDescription string) model.Reference {
	linkId, _ := model.NewId(refId.Int64)
	linkTitle, _ := model.NewTitle(refTitle.String)
	linkURL, _ := model.NewURL(url)
	return model.NewLinkReference(linkId, linkTitle, linkURL, linkDescription, refStarred.Bool)
}

func buildNoteReference(refId sql.NullInt64, refTitle sql.NullString, refStarred sql.NullBool, text string) model.Reference {
	noteId, _ := model.NewId(refId.Int64)
	noteTitle, _ := model.NewTitle(refTitle.String)
	return model.NewNoteReference(noteId, noteTitle, text, refStarred.Bool)
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

	if len(positions) == 0 {
		return fmt.Errorf("no references to reorder")
	}

	// Step 1: Set all positions to negative values to avoid unique constraint violation
	negPositions := make(map[model.Id]int, len(positions))
	for refId, pos := range positions {
		negPositions[refId] = -pos - 1 // -1 to avoid unique constraint violation for position=0
	}
	negQuery, negArgs := buildUpdateReferencePositionsQuery(id, negPositions, version)
	result, err := tx.Exec(negQuery, negArgs...)
	if err != nil {
		return fmt.Errorf("error setting negative positions: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected (neg): %v", err)
	}
	expectedUpdates := int64(len(positions))
	if rowsAffected < expectedUpdates {
		return fmt.Errorf("expected to update %d references, but only updated %d; category with id %d may not exist or version was out of date", expectedUpdates, rowsAffected, id)
	}

	// Step 2: Set positions to their intended positive values
	posQuery, posArgs := buildUpdateReferencePositionsQuery(id, positions, version)
	result, err = tx.Exec(posQuery, posArgs...)
	if err != nil {
		return fmt.Errorf("error updating reference positions: %v", err)
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected (pos): %v", err)
	}
	if rowsAffected < expectedUpdates {
		return fmt.Errorf("expected to update %d references, but only updated %d; category with id %d may not exist or version was out of date", expectedUpdates, rowsAffected, id)
	}

	err = r.updateCategoryVersion(tx, id, version)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// buildUpdateReferencePositionsQuery builds a single SQL update statement for reference positions using CASE WHEN.
func buildUpdateReferencePositionsQuery(categoryId model.Id, positions map[model.Id]int, version model.Version) (string, []interface{}) {
	caseStmt := ""
	var caseArgs []interface{}
	var refIds []interface{}
	for refId, position := range positions {
		caseStmt += " WHEN ? THEN ?"
		caseArgs = append(caseArgs, int64(refId), position)
		refIds = append(refIds, int64(refId))
	}

	inClause := ""
	for i := range refIds {
		if i > 0 {
			inClause += ","
		}
		inClause += "?"
	}

	query := fmt.Sprintf(`
		UPDATE base_references
		SET position = CASE id %s END
		WHERE category_id = ?
		AND id IN (%s)
		AND EXISTS (
			SELECT 1 FROM categories
			WHERE id = ? AND version = ?
		)`, caseStmt, inClause)

	args := append(caseArgs, int64(categoryId))
	args = append(args, refIds...)
	args = append(args, int64(categoryId), int64(version))
	return query, args
}

func (r *SQLiteCategoryRepository) AddReference(id model.Id, reference model.Reference, version model.Version) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO base_references (category_id, title, position, is_starred)
		SELECT ?, ?, COALESCE(MAX(position) + 1, 0), 0
		FROM base_references
		WHERE category_id = ?
		AND EXISTS (
			SELECT 1 FROM categories WHERE id = ? AND version = ?
		)`

	result, err := tx.Exec(query, id, string(reference.Title()), id, id, version)
	if err != nil {
		return fmt.Errorf("error inserting base reference: %v", err)
	}

	refId, err := result.LastInsertId()
	if err != nil || refId == 0 {
		return fmt.Errorf("category with id %d not found or version was out of date", id)
	}

	persistor := NewSQLiteReferenceAddPersistor(id, version, tx, refId)
	err = reference.Persist(persistor)
	if err != nil {
		return err
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
