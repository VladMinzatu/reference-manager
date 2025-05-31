package main

import (
	"database/sql"

	"github.com/VladMinzatu/reference-manager/adapters"
	"github.com/VladMinzatu/reference-manager/domain/service"
	"github.com/VladMinzatu/reference-manager/web"
)

func main() {
	db, _ := sql.Open("sqlite3", "db/references.db")
	repo := adapters.NewSQLiteRepository(db)
	svc := service.NewReferenceService(repo)

	handler := web.NewHandler(svc)
	web.StartServer(handler)
}
