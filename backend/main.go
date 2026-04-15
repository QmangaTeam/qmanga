package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func runMigrations(databaseURL string) error {
	log.Printf("running migrations against %s", databaseURL)
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	// wait for db readiness
	deadline := time.Now().Add(30 * time.Second)
	for {
		if err := db.Ping(); err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("database not ready: %w", err)
		}
		time.Sleep(1 * time.Second)
	}

	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		sqlStr := string(content)
		if strings.TrimSpace(sqlStr) == "" {
			continue
		}
		log.Printf("applying migration %s", f)
		if _, err := db.Exec(sqlStr); err != nil {
			return fmt.Errorf("apply migration %s: %w", f, err)
		}
	}
	return nil
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://qmanga:qmanga@postgres:5432/qmanga?sslmode=disable"
	}

	// support running migrations via `./qmanga migrate` or `go run main.go migrate`
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := runMigrations(databaseURL); err != nil {
			log.Fatalf("migrations failed: %v", err)
		}
		log.Println("migrations applied")
		return
	}

	// attempt to run migrations at startup (best-effort)
	if err := runMigrations(databaseURL); err != nil {
		log.Printf("warning: migrations failed: %v", err)
	}

	// open persistent DB connection for handlers
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// wait for db readiness
	deadline := time.Now().Add(30 * time.Second)
	for {
		if err := db.Ping(); err == nil {
			break
		}
		if time.Now().After(deadline) {
			log.Fatalf("database not ready: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	r := gin.Default()
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// register auth and protected routes
	registerAuthRoutes(r, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
