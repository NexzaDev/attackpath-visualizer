package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/NexzaDev/attackpath-visualizer/migrations"
)

func runMigrations(db *sql.DB) error {
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)"); err != nil {
		return fmt.Errorf("initialize schema migration table: %w", err)
	}
	files, err := migrations.Files.ReadDir(".")
	if err != nil {
		return fmt.Errorf("list embedded migrations: %w", err)
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		versionText := strings.SplitN(file.Name(), "_", 2)[0]
		version, parseErr := strconv.Atoi(versionText)
		if parseErr != nil || version < 1 {
			return fmt.Errorf("invalid migration filename %q", file.Name())
		}
		var applied int
		if err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version=?", version).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %d: %w", version, err)
		}
		if applied != 0 {
			continue
		}
		data, readErr := migrations.Files.ReadFile(file.Name())
		if readErr != nil {
			return fmt.Errorf("read migration %q: %w", file.Name(), readErr)
		}
		tx, beginErr := db.BeginTx(context.Background(), nil)
		if beginErr != nil {
			return fmt.Errorf("begin migration %d: %w", version, beginErr)
		}
		if version == 2 {
			if err = ensureColumnTx(tx, "imports", "status", "TEXT NOT NULL DEFAULT 'success'"); err == nil {
				err = ensureColumnTx(tx, "imports", "error_message", "TEXT NOT NULL DEFAULT ''")
			}
		}
		if err == nil {
			for _, statement := range strings.Split(string(data), ";") {
				statement = strings.TrimSpace(statement)
				if statement == "" {
					continue
				}
				if _, err = tx.Exec(statement); err != nil {
					break
				}
			}
		}
		if err == nil {
			_, err = tx.Exec("INSERT INTO schema_migrations(version) VALUES(?)", version)
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %q: %w", file.Name(), err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %q: %w", file.Name(), err)
		}
	}
	return nil
}

func ensureColumnTx(tx *sql.Tx, table, name, definition string) error {
	rows, err := tx.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var index, notNull, primary int
		var columnName, columnType string
		var defaultValue sql.NullString
		if err = rows.Scan(&index, &columnName, &columnType, &notNull, &defaultValue, &primary); err != nil {
			rows.Close()
			return err
		}
		if columnName == name {
			found = true
		}
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	if found {
		return nil
	}
	_, err = tx.Exec("ALTER TABLE " + table + " ADD COLUMN " + name + " " + definition)
	return err
}
