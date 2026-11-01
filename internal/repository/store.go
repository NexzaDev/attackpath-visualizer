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

type ImportResult struct {
	SchemaVersion         int
	IdentityCount         int
	FindingCount          int
	MisconfigurationCount int
	DocumentBytes         int
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
	return s.RecordImportFailure(ctx, source, "Graph document did not pass validation.")
}

func (s *Store) RecordImportFailure(ctx context.Context, source, message string) error {
	if source == "" {
		source = "local JSON import"
	}
	if message == "" {
		message = "Graph document did not pass validation."
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO imports(source,asset_count,edge_count,status,error_message) VALUES(?,0,0,'failed',?)", source, message)
	if err != nil {
		return fmt.Errorf("record failed import: %w", err)
	}
	return nil
}

func (s *Store) ReplaceGraph(ctx context.Context, graph domain.Graph, source string) error {
	return s.ReplaceGraphWithResult(ctx, graph, source, ImportResult{SchemaVersion: graph.SchemaVersion, IdentityCount: len(graph.Identities), FindingCount: len(graph.Findings), MisconfigurationCount: len(graph.Misconfigurations)})
}

func (s *Store) ReplaceGraphWithResult(ctx context.Context, graph domain.Graph, source string, result ImportResult) error {
	if err := graph.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin import: %w", err)
	}
	defer tx.Rollback()
	for _, statement := range []string{"DELETE FROM edges", "DELETE FROM findings", "DELETE FROM misconfigurations", "DELETE FROM identities", "DELETE FROM graph_nodes", "DELETE FROM assets"} {
		if _, err = tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("clear prior graph data: %w", err)
		}
	}
	for _, asset := range graph.Assets {
		if _, err = tx.ExecContext(ctx, "INSERT INTO assets(id,name,kind,impact,description) VALUES(?,?,?,?,?)", asset.ID, asset.Name, asset.Kind, asset.Impact, asset.Description); err != nil {
			return fmt.Errorf("save asset %q: %w", asset.ID, err)
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO graph_nodes(id,entity_type,name,kind,impact,description) VALUES(?,'asset',?,?,?,?)", asset.ID, asset.Name, asset.Kind, asset.Impact, asset.Description); err != nil {
			return fmt.Errorf("save graph node %q: %w", asset.ID, err)
		}
	}
	for _, identity := range graph.Identities {
		if _, err = tx.ExecContext(ctx, "INSERT INTO identities(id,name,kind,impact,description) VALUES(?,?,?,?,?)", identity.ID, identity.Name, identity.Kind, identity.Impact, identity.Description); err != nil {
			return fmt.Errorf("save identity %q: %w", identity.ID, err)
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO graph_nodes(id,entity_type,name,kind,impact,description) VALUES(?,'identity',?,?,?,?)", identity.ID, identity.Name, identity.Kind, identity.Impact, identity.Description); err != nil {
			return fmt.Errorf("save graph node %q: %w", identity.ID, err)
		}
	}
	for _, item := range graph.Misconfigurations {
		var identityID any
		if item.IdentityID != "" {
			identityID = item.IdentityID
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO misconfigurations(id,asset_id,identity_id,title,severity,description) VALUES(?,?,?,?,?,?)", item.ID, item.AssetID, identityID, item.Title, item.Severity, item.Description); err != nil {
			return fmt.Errorf("save misconfiguration %q: %w", item.ID, err)
		}
	}
	for _, item := range graph.Findings {
		var identityID, misconfigurationID any
		if item.IdentityID != "" {
			identityID = item.IdentityID
		}
		if item.MisconfigurationID != "" {
			misconfigurationID = item.MisconfigurationID
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO findings(id,asset_id,identity_id,misconfiguration_id,title,severity,source,observed) VALUES(?,?,?,?,?,?,?,?)", item.ID, item.AssetID, identityID, misconfigurationID, item.Title, item.Severity, item.Source, item.Observed); err != nil {
			return fmt.Errorf("save finding %q: %w", item.ID, err)
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
	if _, err = tx.ExecContext(ctx, "INSERT INTO imports(source,asset_count,edge_count,schema_version,identity_count,finding_count,misconfiguration_count,document_bytes) VALUES(?,?,?,?,?,?,?,?)", source, len(graph.Assets), len(graph.Edges), result.SchemaVersion, result.IdentityCount, result.FindingCount, result.MisconfigurationCount, result.DocumentBytes); err != nil {
		return fmt.Errorf("record import: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit import: %w", err)
	}
	return nil
}

func (s *Store) Graph(ctx context.Context) (domain.Graph, error) {
	graph := domain.Graph{SchemaVersion: 1, Assets: make([]domain.Asset, 0), Identities: make([]domain.Identity, 0), Misconfigurations: make([]domain.Misconfiguration, 0), Findings: make([]domain.Finding, 0), Edges: make([]domain.Edge, 0)}
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
	rows, err = s.db.QueryContext(ctx, "SELECT id,name,kind,impact,description FROM identities ORDER BY name,id")
	if err != nil {
		return graph, fmt.Errorf("read identities: %w", err)
	}
	for rows.Next() {
		var identity domain.Identity
		if err = rows.Scan(&identity.ID, &identity.Name, &identity.Kind, &identity.Impact, &identity.Description); err != nil {
			rows.Close()
			return graph, fmt.Errorf("read identity: %w", err)
		}
		graph.Identities = append(graph.Identities, identity)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return graph, fmt.Errorf("read identities: %w", err)
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, "SELECT id,asset_id,COALESCE(identity_id,''),title,severity,description FROM misconfigurations ORDER BY id")
	if err != nil {
		return graph, fmt.Errorf("read misconfigurations: %w", err)
	}
	for rows.Next() {
		var item domain.Misconfiguration
		if err = rows.Scan(&item.ID, &item.AssetID, &item.IdentityID, &item.Title, &item.Severity, &item.Description); err != nil {
			rows.Close()
			return graph, fmt.Errorf("read misconfiguration: %w", err)
		}
		graph.Misconfigurations = append(graph.Misconfigurations, item)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return graph, fmt.Errorf("read misconfigurations: %w", err)
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, "SELECT id,asset_id,COALESCE(identity_id,''),COALESCE(misconfiguration_id,''),title,severity,source,observed FROM findings ORDER BY id")
	if err != nil {
		return graph, fmt.Errorf("read findings: %w", err)
	}
	for rows.Next() {
		var item domain.Finding
		if err = rows.Scan(&item.ID, &item.AssetID, &item.IdentityID, &item.MisconfigurationID, &item.Title, &item.Severity, &item.Source, &item.Observed); err != nil {
			rows.Close()
			return graph, fmt.Errorf("read finding: %w", err)
		}
		graph.Findings = append(graph.Findings, item)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return graph, fmt.Errorf("read findings: %w", err)
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
	rows, err := s.db.QueryContext(ctx, "SELECT source,asset_count,edge_count,imported_at,status,error_message,schema_version,identity_count,finding_count,misconfiguration_count,document_bytes FROM imports ORDER BY id DESC LIMIT 50")
	if err != nil {
		return nil, fmt.Errorf("read import history: %w", err)
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var source, imported, status, message string
		var assets, edges, schema, identities, findings, misconfigurations, bytes int
		if err = rows.Scan(&source, &assets, &edges, &imported, &status, &message, &schema, &identities, &findings, &misconfigurations, &bytes); err != nil {
			return nil, fmt.Errorf("read import result: %w", err)
		}
		items = append(items, map[string]any{"source": source, "assets": assets, "edges": edges, "identities": identities, "findings": findings, "misconfigurations": misconfigurations, "schema_version": schema, "document_bytes": bytes, "imported_at": imported, "status": status, "message": message})
	}
	return items, rows.Err()
}
