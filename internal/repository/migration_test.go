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
		`INSERT INTO assets(id,name,kind,impact,description) VALUES('legacy-target','Legacy target','database',5,'preserved')`,
		`INSERT INTO edges(id,source_id,target_id,relation,confidence,evidence_json) VALUES('legacy-edge','legacy','legacy-target','reads',70,'{}')`,
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
	if len(graph.Assets) != 2 || graph.Assets[0].ID != "legacy" || len(graph.Edges) != 1 || graph.Edges[0].Evidence.Source != "legacy://no-recorded-evidence" {
		t.Fatalf("legacy graph was not preserved and normalized: %#v", graph)
	}
	if err = graph.Validate(); err != nil {
		t.Fatalf("migrated legacy graph is not valid: %v", err)
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
	if version != 4 {
		t.Fatalf("migration version = %d", version)
	}
}
