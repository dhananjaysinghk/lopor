package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/lopor-ai/lopor/internal/database"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/migrate/main.go [up|down]")
	}

	command := os.Args[1]

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "lopor_db")
	dbSSL := getEnv("DB_SSLMODE", "disable")

	db, err := database.ConnectPostgres(database.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPass,
		DBName:   dbName,
		SSLMode:  dbSSL,
	})
	if err != nil {
		log.Fatalf("Migration DB Connection Failed: %v", err)
	}
	defer db.Close()

	fmt.Printf("Executing migration command '%s' against %s...\n", command, dbName)

	var sqlFile string
	switch command {
	case "up":
		sqlFile = "migrations/000001_init_schema.up.sql"
	case "down":
		sqlFile = "migrations/000001_init_schema.down.sql"
	default:
		log.Fatalf("Unknown command: %s. Use 'up' or 'down'", command)
	}

	content, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("Failed to read migration file %s: %v", sqlFile, err)
	}

	ctx := context.Background()
	_, err = db.Pool.Exec(ctx, string(content))
	if err != nil {
		log.Fatalf("Migration execution failed: %v", err)
	}

	log.Printf("Successfully executed migration %s!", command)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
