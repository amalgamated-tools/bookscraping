package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

func SetupDatabase() (Querier, error) {
	slog.Debug("Setting up database")

	// Determine database path: prefer mounted /data folder, fall back to ./db
	var dbFilePath string
	if _, err := os.Stat("/data"); err == nil {
		dbFilePath = "/data/bookscraping.db"
		slog.Debug("Using mounted /data folder", slog.String("path", dbFilePath))
	} else {
		dbFilePath = "./db/bookscraping.db"
		slog.Debug("Using local db folder", slog.String("path", dbFilePath))
	}

	// Ensure parent directory exists
	dbDir := filepath.Dir(dbFilePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		slog.Error("Failed to create database directory", slog.String("path", dbDir), slog.Any("error", err))
		return nil, fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
	}

	// Open database with modernc.org/sqlite pure Go driver
	slog.Debug("Opening database", slog.String("path", dbFilePath))
	sqlDB, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		slog.Error("Failed to open database", slog.String("path", dbFilePath), slog.Any("error", err))
		return nil, fmt.Errorf("failed to open database at %s: %w", dbFilePath, err)
	}

	// Run migrations
	if err := runMigrations(sqlDB); err != nil {
		slog.Error("Failed to run migrations", slog.Any("error", err))
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.Error("Failed to close database", slog.Any("error", closeErr))
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	slog.Debug("Database created and migrated successfully", slog.String("path", dbFilePath))

	// Create queries instance
	queries := New(sqlDB)
	count, err := queries.CountBooks(context.Background())
	if err != nil {
		slog.Error("Failed to count books in database", slog.Any("error", err))
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.Error("Failed to close database", slog.Any("error", closeErr))
		}
		return nil, fmt.Errorf("failed to count books in database: %w", err)
	}
	slog.Debug("Database connected", slog.Int64("book_count", count))
	return queries, nil
}

// runMigrations reads and executes all SQL migration files
// Supports dbmate format with '-- migrate:up' and '-- migrate:down' markers
func runMigrations(sqlDB *sql.DB) error {
	ctx := context.Background()

	// Create migrations table if it doesn't exist
	if _, err := sqlDB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Find and sort migration files
	migrationsDir := "db/migrations"
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []fs.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrations = append(migrations, entry)
		}
	}

	// Sort migrations by filename (timestamp-based)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name() < migrations[j].Name()
	})

	// Execute each migration
	for _, migration := range migrations {
		filename := migration.Name()
		version := strings.TrimSuffix(filename, ".sql")

		// Check if migration has already been applied
		var applied int
		err := sqlDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&applied)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if applied > 0 {
			slog.Debug("Migration already applied", slog.String("version", version))
			continue
		}

		// Read migration file
		migrationPath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Extract the "up" SQL from the migration file (dbmate format)
		upSQL := extractUpSQL(string(content))
		if upSQL == "" {
			return fmt.Errorf("migration %s has no '-- migrate:up' section", filename)
		}

		// Run this migration in a transaction to ensure all-or-nothing execution
		tx, err := sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", filename, err)
		}

		// Execute migration - split by semicolon to handle multiple statements
		statements := splitStatements(upSQL)
		for _, stmt := range statements {
			if strings.TrimSpace(stmt) == "" {
				continue
			}

			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				// Any error means the migration did not fully apply; roll back.
				if rbErr := tx.Rollback(); rbErr != nil {
					return fmt.Errorf("failed to execute migration %s: %v (rollback error: %w)", filename, err, rbErr)
				}
				return fmt.Errorf("failed to execute migration %s: %w", filename, err)
			}
		}

		// Record migration within the same transaction
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("failed to record migration %s: %v (rollback error: %w)", filename, err, rbErr)
			}
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}
		slog.Info("Migration applied", slog.String("version", version))
	}

	return nil
}

// extractUpSQL extracts the SQL between '-- migrate:up' and '-- migrate:down' markers
func extractUpSQL(content string) string {
	lines := strings.Split(content, "\n")
	var upLines []string
	inUpBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "-- migrate:up" {
			inUpBlock = true
			continue
		}

		if trimmed == "-- migrate:down" {
			break
		}

		if inUpBlock && trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			upLines = append(upLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(upLines, "\n"))
}

// splitStatements splits SQL by semicolon, handling strings and inline comments properly
func splitStatements(sql string) []string {
	// First, remove inline comments (-- to end of line)
	sql = removeInlineComments(sql)

	var statements []string
	var current strings.Builder
	inString := false
	var stringChar rune
	var i int
	runes := []rune(sql)

	for i < len(runes) {
		char := runes[i]

		if !inString && (char == '\'' || char == '"') {
			inString = true
			stringChar = char
			current.WriteRune(char)
		} else if inString && char == stringChar {
			if i+1 < len(runes) && runes[i+1] == stringChar {
				// Escaped quote
				current.WriteRune(char)
				current.WriteRune(char)
				i++
			} else {
				inString = false
				current.WriteRune(char)
			}
		} else if !inString && char == ';' {
			statements = append(statements, current.String())
			current.Reset()
		} else {
			current.WriteRune(char)
		}

		i++
	}

	// Add any remaining statement
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}

// removeInlineComments removes SQL inline comments (-- to end of line) while preserving strings
func removeInlineComments(sql string) string {
	var result strings.Builder
	runes := []rune(sql)
	inString := false
	var stringChar rune

	for i := 0; i < len(runes); i++ {
		char := runes[i]

		// Handle string delimiters
		if !inString && (char == '\'' || char == '"') {
			inString = true
			stringChar = char
			result.WriteRune(char)
		} else if inString && char == stringChar {
			if i+1 < len(runes) && runes[i+1] == stringChar {
				// Escaped quote
				result.WriteRune(char)
				result.WriteRune(runes[i+1])
				i++
			} else {
				inString = false
				result.WriteRune(char)
			}
		} else if !inString && char == '-' && i+1 < len(runes) && runes[i+1] == '-' {
			// Found inline comment, skip until end of line (but don't skip the newline itself)
			i += 2
			for i < len(runes) && runes[i] != '\n' {
				i++
			}
			// Don't increment i here, so the newline will be written in the next iteration
			i--
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}
