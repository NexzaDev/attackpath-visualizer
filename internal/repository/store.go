package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	path string
}

func Open(path string) (*Store, error) {
	if path != ":memory:" {
		directory := filepath.Dir(path)
		if directory != "." {
			if err := os.MkdirAll(directory, 0700); err != nil {
				return nil, fmt.Errorf("create data directory: %w", err)
			}
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	for _, statement := range []string{"PRAGMA journal_mode=WAL", "PRAGMA busy_timeout=5000", "PRAGMA foreign_keys=ON"} {
		if _, err := db.Exec(statement); err != nil {
			db.Close()
			return nil, fmt.Errorf("configure database: %w", err)
		}
	}
	if err = runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}
	for _, rule := range []struct {
		typeName string
		severity int
	}{{"critical_path", 4}, {"validation_failed", 3}} {
		if _, err := db.Exec("INSERT OR IGNORE INTO notification_rules(event_type,enabled,minimum_severity,quiet_start,quiet_end,detailed_text) VALUES(?,1,?,'22:00','07:00',0)", rule.typeName, rule.severity); err != nil {
			db.Close()
			return nil, fmt.Errorf("initialize notification rules: %w", err)
		}
	}
	if path != ":memory:" {
		if err := os.Chmod(path, 0600); err != nil {
			db.Close()
			return nil, fmt.Errorf("restrict database permissions: %w", err)
		}
	}
	return &Store{db: db, path: path}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Path() string { return s.path }

func (s *Store) RecordFailedImport(ctx context.Context, source string) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO imports(source,asset_count,edge_count,status,error_message) VALUES(?,0,0,'failed','Graph document did not pass validation.')", source)
	if err != nil {
		return fmt.Errorf("record failed import: %w", err)
	}
	return nil
}

func (s *Store) ReplaceGraph(ctx context.Context, graph domain.Graph, source string) error {
	if err := graph.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin import: %w", err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "DELETE FROM edges"); err != nil {
		return fmt.Errorf("clear relationships: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM assets"); err != nil {
		return fmt.Errorf("clear assets: %w", err)
	}
	for _, asset := range graph.Assets {
		if _, err = tx.ExecContext(ctx, "INSERT INTO assets(id,name,kind,impact,description) VALUES(?,?,?,?,?)", asset.ID, asset.Name, asset.Kind, asset.Impact, asset.Description); err != nil {
			return fmt.Errorf("save asset %q: %w", asset.ID, err)
		}
	}
	for _, edge := range graph.Edges {
		evidence, marshalErr := json.Marshal(edge.Evidence)
		if marshalErr != nil {
			return fmt.Errorf("encode evidence for %q: %w", edge.ID, marshalErr)
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO edges(id,source_id,target_id,relation,confidence,evidence_json) VALUES(?,?,?,?,?,?)", edge.ID, edge.From, edge.To, edge.Relation, edge.Confidence, evidence); err != nil {
			return fmt.Errorf("save edge %q: %w", edge.ID, err)
		}
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO imports(source,asset_count,edge_count) VALUES(?,?,?)", source, len(graph.Assets), len(graph.Edges)); err != nil {
		return fmt.Errorf("record import: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit import: %w", err)
	}
	return nil
}

func (s *Store) Graph(ctx context.Context) (domain.Graph, error) {
	graph := domain.Graph{Assets: make([]domain.Asset, 0), Edges: make([]domain.Edge, 0)}
	rows, err := s.db.QueryContext(ctx, "SELECT id,name,kind,impact,description FROM assets ORDER BY name,id")
	if err != nil {
		return graph, fmt.Errorf("read assets: %w", err)
	}
	for rows.Next() {
		var asset domain.Asset
		if err = rows.Scan(&asset.ID, &asset.Name, &asset.Kind, &asset.Impact, &asset.Description); err != nil {
			rows.Close()
			return graph, fmt.Errorf("read asset: %w", err)
		}
		graph.Assets = append(graph.Assets, asset)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return graph, fmt.Errorf("read assets: %w", err)
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, "SELECT id,source_id,target_id,relation,confidence,evidence_json FROM edges ORDER BY id")
	if err != nil {
		return graph, fmt.Errorf("read relationships: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var edge domain.Edge
		var evidence string
		if err = rows.Scan(&edge.ID, &edge.From, &edge.To, &edge.Relation, &edge.Confidence, &evidence); err != nil {
			return graph, fmt.Errorf("read relationship: %w", err)
		}
		if err = json.Unmarshal([]byte(evidence), &edge.Evidence); err != nil {
			return graph, fmt.Errorf("decode evidence for %q: %w", edge.ID, err)
		}
		graph.Edges = append(graph.Edges, edge)
	}
	if err = rows.Err(); err != nil {
		return graph, fmt.Errorf("read relationships: %w", err)
	}
	return graph, nil
}

func (s *Store) ImportHistory(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT source,asset_count,edge_count,imported_at,status,error_message FROM imports ORDER BY id DESC LIMIT 50")
	if err != nil {
		return nil, fmt.Errorf("read import history: %w", err)
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var source, imported, status, message string
		var assets, edges int
		if err = rows.Scan(&source, &assets, &edges, &imported, &status, &message); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"source": source, "assets": assets, "edges": edges, "imported_at": imported, "status": status, "message": message})
	}
	return items, rows.Err()
}
