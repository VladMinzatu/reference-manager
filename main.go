package main

import (
	"database/sql"

	"github.com/VladMinzatu/reference-manager/adapters"
	"github.com/VladMinzatu/reference-manager/domain/service"
	"github.com/VladMinzatu/reference-manager/web"
)

func main() {
	db, _ := sql.Open("sqlite3", "db/references.db")
	categoryRepo := adapters.NewSQLiteCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryListRepository := adapters.NewSQLiteCategoryListRepository(db)
	referenceRepo := adapters.NewSQLiteReferencesRepository(db)

	handler := web.NewHandler(categoryService, categoryListRepository, referenceRepo)
	web.StartServer(handler)
}
