package main

import (
	"database/sql"
	"log"

	"github.com/VladMinzatu/reference-manager/adapters"
	"github.com/VladMinzatu/reference-manager/domain/service"
	"github.com/VladMinzatu/reference-manager/web"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "db/backup/vlad.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Failed to enable foreign keys:", err)
	}

	categoryRepo := adapters.NewSQLiteCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryListRepository := adapters.NewSQLiteCategoryListRepository(db)
	referenceRepo := adapters.NewSQLiteReferencesRepository(db)

	handler := web.NewHandler(categoryService, categoryListRepository, referenceRepo)
	web.StartServer(handler)
}
