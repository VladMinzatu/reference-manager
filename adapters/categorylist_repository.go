package adapters

import (
	"database/sql"
	"fmt"

	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/repository"
)

type SQLiteCategoryListRepository struct {
	db *sql.DB
}

func NewSQLiteCategoryListRepository(db *sql.DB) repository.CategoryListRepository {
	return &SQLiteCategoryListRepository{db: db}
}

func (r *SQLiteCategoryListRepository) GetAllCategoryNames() ([]model.Title, error) {
	rows, err := r.db.Query(`SELECT name FROM categories ORDER BY position`)
	if err != nil {
		return nil, fmt.Errorf("error querying category names: %v", err)
	}
	defer rows.Close()

	var names []model.Title
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("error scanning category name: %v", err)
		}
		title, err := model.NewTitle(name)
		if err != nil {
			return nil, fmt.Errorf("invalid category name: %v", err)
		}
		names = append(names, title)
	}
	return names, nil
}

func (r *SQLiteCategoryListRepository) AddNewCategory(name model.Title) (model.Category, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.Category{}, fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

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
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	// Note: Normally we would have a SELECT...FOR UPDATE here, locking all the category rows.
	// But SQLite does not support SELECT ... FOR UPDATE or row-level locking.
	// Concurrency is handled by database-level locks: once a transaction writes, it blocks other writers.
	// This ensures that reordering and other multi-step operations are safe from concurrent modification.
	// Therefore, we move straight to updating the positions.

	for id, pos := range positions {
		_, err := tx.Exec(`UPDATE categories SET position = ? WHERE id = ?`, pos, int64(id))
		if err != nil {
			return fmt.Errorf("error updating category position: %v", err)
		}
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
