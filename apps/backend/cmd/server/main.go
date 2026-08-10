package main

import (
	"log"
	"os"

	"github.com/lopor-ai/lopor/internal/database"
	"github.com/lopor-ai/lopor/internal/server"
)

func main() {
	log.Println("Starting Lopor AI Workspace Backend Engine...")

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
		log.Printf("Warning: Database connection failed (running in offline mode): %v", err)
	} else {
		defer db.Close()
	}

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPass := getEnv("REDIS_PASSWORD", "")

	redisClient, err := database.ConnectRedis(redisHost, redisPort, redisPass, 0)
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		defer redisClient.Close()
	}

	jwtSecret := getEnv("JWT_SECRET", "super-secret-jwt-key-minimum-32-chars-change-in-prod")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")
	port := getEnv("APP_PORT", "8080")

	app := server.NewServer(server.Config{
		Port:        port,
		JWTSecret:   jwtSecret,
		FrontendURL: frontendURL,
		DB:          db,
		Redis:       redisClient,
	})

	log.Printf("Lopor API Server listening on port :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
