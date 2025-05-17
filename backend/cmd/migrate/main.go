package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"coffee-loyalty-system/pkg/storage"
)

const migrationsDir = "migrations"

func main() {
	cfg := storage.Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvIntOrDefault("DB_PORT", 5432),
		User:     getEnvOrDefault("DB_USER", "postgres"),
		Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("DB_NAME", "coffee_loyalty"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	db, err := storage.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	pool := db.Pool()

	// Create migrations table if it doesn't exist
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get list of migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Get applied migrations
	var appliedMigrations []string
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		log.Fatalf("Failed to query applied migrations: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			log.Fatalf("Failed to scan migration version: %v", err)
		}
		appliedMigrations = append(appliedMigrations, version)
	}

	// Create a map for quick lookup of applied migrations
	appliedMap := make(map[string]bool)
	for _, v := range appliedMigrations {
		appliedMap[v] = true
	}

	// Sort migration files by name
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply new migrations
	for _, file := range migrationFiles {
		version := strings.TrimSuffix(file, ".sql")
		if appliedMap[version] {
			continue
		}

		// Read and execute migration
		content, err := ioutil.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		// Start transaction
		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}

		// Execute migration
		_, err = tx.Exec(ctx, string(content))
		if err != nil {
			tx.Rollback(ctx)
			log.Fatalf("Failed to execute migration %s: %v", file, err)
		}

		// Record migration
		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			tx.Rollback(ctx)
			log.Fatalf("Failed to record migration %s: %v", file, err)
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("Failed to commit migration %s: %v", file, err)
		}

		fmt.Printf("Applied migration: %s\n", version)
	}

	fmt.Println("All migrations completed successfully!")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
