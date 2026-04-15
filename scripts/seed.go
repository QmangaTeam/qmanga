package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://qmanga:qmanga@localhost:5432/qmanga?sslmode=disable"
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Insert a sample genre and title
	if _, err := db.Exec(`INSERT INTO genres (slug, name_kz, name_en) VALUES ('action', 'Экшн', 'Action') ON CONFLICT DO NOTHING`); err != nil {
		log.Fatalf("insert genre: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO titles (title_kz, title_en, description_kz, type, is_published) VALUES ('Тестовая манхва', 'Test Manga', 'Тестовое описание', 'manhwa', true) ON CONFLICT DO NOTHING`); err != nil {
		log.Fatalf("insert title: %v", err)
	}

	fmt.Println("seed: done")
}
