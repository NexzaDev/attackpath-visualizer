package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestLegacySchemaMigratesWithoutLosingGraph(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE assets (id TEXT PRIMARY KEY, name TEXT NOT NULL, kind TEXT NOT NULL, impact INTEGER NOT NULL, description TEXT NOT NULL)`,
		`CREATE TABLE edges (id TEXT PRIMARY KEY, source_id TEXT NOT NULL, target_id TEXT NOT NULL, relation TEXT NOT NULL, confidence INTEGER NOT NULL, evidence_json TEXT NOT NULL)`,
		`CREATE TABLE imports (id INTEGER PRIMARY KEY, source TEXT NOT NULL, asset_count INTEGER NOT NULL, edge_count INTEGER NOT NULL, imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`INSERT INTO assets(id,name,kind,impact,description) VALUES('legacy','Legacy node','service',4,'preserved')`,
	} {
		if _, err = db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	graph, err := store.Graph(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Assets) != 1 || graph.Assets[0].ID != "legacy" {
		t.Fatalf("legacy graph was not preserved: %#v", graph)
	}
	if err = store.RecordFailedImport(context.Background(), "invalid input"); err != nil {
		t.Fatal(err)
	}
	history, err := store.ImportHistory(context.Background())
	if err != nil || len(history) != 1 || history[0]["status"] != "failed" {
		t.Fatalf("history=%v err=%v", history, err)
	}
	var version int
	if err = store.db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 2 {
		t.Fatalf("migration version = %d", version)
	}
}
