package adapters

import (
	"database/sql"
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/util"
)

type SQLiteCategoryListRepository struct {
	db *sql.DB
}

func NewSQLiteCategoryListRepository(db *sql.DB) *SQLiteCategoryListRepository {
	return &SQLiteCategoryListRepository{db: db}
}

func (r *SQLiteCategoryListRepository) GetAllCategoryRefs() ([]model.CategoryRef, error) {
	rows, err := r.db.Query(`SELECT id, name FROM categories ORDER BY position`)
	if err != nil {
		return nil, fmt.Errorf("error querying categories: %v", err)
	}
	defer rows.Close()

	var refs []model.CategoryRef
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("error scanning category: %v", err)
		}
		catId, err := model.NewId(id)
		if err != nil {
			return nil, fmt.Errorf("invalid id: %v", err)
		}
		title, err := model.NewTitle(name)
		if err != nil {
			return nil, fmt.Errorf("invalid title: %v", err)
		}
		refs = append(refs, model.CategoryRef{Id: catId, Name: title})
	}
	return refs, nil
}

func (r *SQLiteCategoryListRepository) AddNewCategory(name model.Title) (model.Category, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.Category{}, fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	if _, err := model.NewTitle(string(name)); err != nil {
		return model.Category{}, fmt.Errorf("invalid title: %v", err)
	}

	// Note: This logic is safe in SQLite because all writers are serialized.
	// In e.g. Postgres, we would need row/table-level locking via SELECT...FOR UPDATE prior to this statement
	// (sequences or separate table with table-level locking are also options, but with sqlite, we can keep it simple)
	result, err := tx.Exec(`INSERT INTO categories (name, position) SELECT ?, COALESCE(MAX(position) + 1, 0) FROM categories`, string(name))
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

	catId, _ := model.NewId(id)
	return model.Category{Id: catId, Name: name}, nil
}

func (r *SQLiteCategoryListRepository) ReorderCategories(positions map[model.Id]int) error {
	// This is the one place in the code where I violate my pledge to design with fine granularity of locking and concurrency in mind (see more details in README)
	// If we were using e.g. Postgres, we would start off the transaction with a SELECT...FOR UPDATE on our categories and that would be sufficient, even in case of new categories being added concurrently (due to how the reordering logic works).
	// Of course, even with SQLite, there are other options as well - we could use a separate table to lock or version our categories list for example.
	// But since this is just an exercise and the choice of SQLite iself here doesn't really match with my self-imposed restriction to ensure fine granularity,
	// I'm satisfied with taking this shortcuthere and know how it should ideally be done in another RDBMS.
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	rows, err := tx.Query(`SELECT id FROM categories`)
	if err != nil {
		return fmt.Errorf("error fetching category ids: %v", err)
	}
	var ids []model.Id
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("error scanning id: %v", err)
		}
		ids = append(ids, model.Id(id))
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating ids: %v", err)
	}

	if err := util.ValidatePositions(ids, positions); err != nil {
		return err
	}

	// Helper function to build the UPDATE statement for category positions
	buildUpdatePositionsQuery := func(positions map[model.Id]int) (string, []interface{}) {
		query := `UPDATE categories SET position = CASE id`
		args := []interface{}{}
		for id, pos := range positions {
			query += " WHEN ? THEN ?"
			args = append(args, int64(id), pos)
		}
		query += " END WHERE id IN ("
		first := true
		for id := range positions {
			if !first {
				query += ","
			}
			query += "?"
			args = append(args, int64(id))
			first = false
		}
		query += ")"
		return query, args
	}

	// Step 1: Set all positions to negative values to avoid unique constraint violation
	negPositions := make(map[model.Id]int, len(positions))
	for id, pos := range positions {
		negPositions[id] = -pos
	}
	negQuery, negArgs := buildUpdatePositionsQuery(negPositions)
	_, err = tx.Exec(negQuery, negArgs...)
	if err != nil {
		return fmt.Errorf("error setting negative positions: %v", err)
	}

	// Step 2: Set positions to their intended positive values
	posQuery, posArgs := buildUpdatePositionsQuery(positions)
	_, err = tx.Exec(posQuery, posArgs...)
	if err != nil {
		return fmt.Errorf("error updating category positions: %v", err)
	}

	return tx.Commit()
}

func (r *SQLiteCategoryListRepository) DeleteCategory(id model.Id) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`DELETE FROM categories WHERE id = ?`, int64(id))
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
	// again, as in the other methods, we're taking a shortcut here afforded by sqlite
	// we'd need to use e.g. row-level locking for this if we were using e.g. Postgres.
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
