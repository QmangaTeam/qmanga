package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	identityCrypto "qmanga/internal/identity/adapters/crypto"
	identityPostgres "qmanga/internal/identity/adapters/postgres"
	"qmanga/internal/identity/application"
	"qmanga/internal/platform/postgres"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	database, failure := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	if failure != nil {
		log.Fatal("Database connection failed")
	}
	defer database.Close()
	if failure = database.Migrate(ctx); failure != nil {
		log.Fatal("Database migration failed")
	}
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" {
		log.Fatal("Set ADMIN_EMAIL")
	}
	if password == "" {
		fmt.Fprintln(os.Stderr, "Enter password through stdin (8–128 characters):")
		password, failure = bufio.NewReader(os.Stdin).ReadString('\n')
		if failure != nil && password == "" {
			log.Fatal("Password input failed")
		}
		password = strings.TrimRight(password, "\r\n")
	}
	service := application.New(identityPostgres.New(database.Pool()), identityCrypto.New(), nil, "")
	if _, failure = service.CreateBootstrapAdmin(ctx, email, password); failure != nil {
		log.Fatal("Admin creation failed; check email, password or existing account")
	}
	fmt.Println("Verified administrator created.")
}
