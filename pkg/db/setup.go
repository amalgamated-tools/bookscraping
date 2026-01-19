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
	slog.Info("Setting up database")

	// Determine database path: prefer mounted /data folder, fall back to ./db
	var dbFilePath string
	if _, err := os.Stat("/data"); err == nil {
		dbFilePath = "/data/bookscraping.db"
		slog.Info("Using mounted /data folder", slog.String("path", dbFilePath))
	} else {
		dbFilePath = "./db/bookscraping.db"
		slog.Info("Using local db folder", slog.String("path", dbFilePath))
	}

	// Ensure parent directory exists
	dbDir := filepath.Dir(dbFilePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		slog.Error("Failed to create database directory", slog.String("path", dbDir), slog.Any("error", err))
		return nil, fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
	}

	// Open database with modernc.org/sqlite pure Go driver
	slog.Info("Opening database", slog.String("path", dbFilePath))
	sqlDB, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		slog.Error("Failed to open database", slog.String("path", dbFilePath), slog.Any("error", err))
		return nil, fmt.Errorf("failed to open database at %s: %w", dbFilePath, err)
	}

	// Run migrations
	if err := runMigrations(sqlDB); err != nil {
		slog.Error("Failed to run migrations", slog.Any("error", err))
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	slog.Info("Database created and migrated successfully", slog.String("path", dbFilePath))

	// Create queries instance
	queries := New(sqlDB)
	count, err := queries.CountBooks(context.Background())
	if err != nil {
		slog.Error("Failed to count books in database", slog.Any("error", err))
		sqlDB.Close()
		return nil, fmt.Errorf("failed to count books in database: %w", err)
	}
	slog.Info("Database connected", slog.Int64("book_count", count))
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
		filepath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Extract the "up" SQL from the migration file (dbmate format)
		upSQL := extractUpSQL(string(content))
		if upSQL == "" {
			return fmt.Errorf("migration %s has no '-- migrate:up' section", filename)
		}

		// Execute migration - split by semicolon to handle multiple statements
		statements := splitStatements(upSQL)
		for _, stmt := range statements {
			trimmedStmt := strings.TrimSpace(stmt)
			if trimmedStmt == "" {
				continue
			}

			// Check if this is an ALTER TABLE ADD COLUMN statement
			// If so, verify the column doesn't already exist before executing
			if shouldSkipStatement(ctx, sqlDB, trimmedStmt) {
				slog.Debug("Skipping already-applied statement", slog.String("migration", filename), slog.String("statement", trimmedStmt))
				continue
			}

			_, err := sqlDB.ExecContext(ctx, trimmedStmt)
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", filename, err)
			}
		}

		// Record migration
		if _, err := sqlDB.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		slog.Info("Migration applied", slog.String("version", version))
	}

	return nil
}

// shouldSkipStatement checks if a statement should be skipped because it would fail due to already-existing structures
func shouldSkipStatement(ctx context.Context, db *sql.DB, stmt string) bool {
	stmt = strings.TrimSpace(stmt)
	stmtUpper := strings.ToUpper(stmt)

	// Check for ALTER TABLE ADD COLUMN
	if strings.Contains(stmtUpper, "ALTER TABLE") && strings.Contains(stmtUpper, "ADD COLUMN") {
		// Extract table name and column name from statement
		// Expected format: ALTER TABLE table_name ADD COLUMN column_name ...
		// Note: This parser expects simple, unquoted identifiers as used in our migrations
		tableName, columnName := parseAlterTableAddColumn(stmt)
		if tableName != "" && columnName != "" {
			if columnExists(ctx, db, tableName, columnName) {
				return true
			}
		}
	}

	// CREATE INDEX statements with IF NOT EXISTS are already idempotent,
	// so we don't need to check for their existence

	return false
}

// parseAlterTableAddColumn extracts table and column names from ALTER TABLE ADD COLUMN statement
// Note: This parser assumes simple, unquoted identifiers as used in our migrations.
// It will not correctly handle quoted identifiers (e.g., "my table", [my column]) or
// complex SQL with embedded quotes. For our use case with straightforward migration files,
// this simple parser is sufficient.
func parseAlterTableAddColumn(stmt string) (tableName, columnName string) {
	const (
		alterTableLen = len("ALTER TABLE")
		addColumnLen  = len("ADD COLUMN")
	)

	stmtUpper := strings.ToUpper(stmt)

	// Find "ALTER TABLE"
	alterTableIdx := strings.Index(stmtUpper, "ALTER TABLE")
	if alterTableIdx == -1 {
		return "", ""
	}

	// Find "ADD COLUMN"
	addColumnIdx := strings.Index(stmtUpper, "ADD COLUMN")
	if addColumnIdx == -1 {
		return "", ""
	}

	// Extract table name (between ALTER TABLE and ADD COLUMN)
	tableNamePart := strings.TrimSpace(stmt[alterTableIdx+alterTableLen : addColumnIdx])
	tableNameFields := strings.Fields(tableNamePart)
	if len(tableNameFields) > 0 {
		tableName = tableNameFields[0]
	}

	// Extract column name (after ADD COLUMN, before space or type definition)
	columnNamePart := strings.TrimSpace(stmt[addColumnIdx+addColumnLen:])
	columnNameFields := strings.Fields(columnNamePart)
	if len(columnNameFields) > 0 {
		columnName = columnNameFields[0]
	}

	return tableName, columnName
}

// columnExists checks if a column exists in a table using PRAGMA table_info
func columnExists(ctx context.Context, db *sql.DB, tableName, columnName string) bool {
	// Validate table name to prevent SQL injection
	// SQLite table names should only contain alphanumeric characters and underscores
	if !isValidSQLiteIdentifier(tableName) {
		slog.Warn("Invalid table name", slog.String("table", tableName))
		return false
	}

	// PRAGMA table_info doesn't support parameterized queries, but we've validated the input
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		// If we can't check, assume it doesn't exist and let the migration fail if needed
		slog.Warn("Failed to check if column exists", slog.String("table", tableName), slog.String("column", columnName), slog.Any("error", err))
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var typeStr string
		var notNull int
		var dfltValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &typeStr, &notNull, &dfltValue, &pk); err != nil {
			slog.Warn("Failed to scan table_info row", slog.String("table", tableName), slog.Any("error", err))
			continue
		}

		if strings.EqualFold(name, columnName) {
			return true
		}
	}

	return false
}

// isValidSQLiteIdentifier checks if a string is a valid SQLite identifier (table/column name)
// Note: This only validates unquoted identifiers. Quoted identifiers (e.g., "my table", [my table])
// are not supported by this validator but are also not used in our migrations.
// Valid unquoted identifiers:
// - Start with a letter (a-z, A-Z) or underscore
// - Contain only letters, digits, and underscores
func isValidSQLiteIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// First character must be a letter or underscore (not a digit)
	firstChar := rune(s[0])
	if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z') || firstChar == '_') {
		return false
	}

	// Remaining characters can be letters, digits, or underscores
	for _, ch := range s {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}
	return true
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

// splitStatements splits SQL by semicolon, handling strings properly
func splitStatements(sql string) []string {
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
		} else if inString && rune(char) == stringChar {
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
