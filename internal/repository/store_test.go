package repository

import (
	"context"
	"reflect"
	"testing"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
)

func TestReplaceGraphAndImportHistory(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	graph := domain.Graph{Assets: []domain.Asset{{ID: "node", Name: "Synthetic node", Kind: "service", Impact: 3}}, Edges: []domain.Edge{}}
	if err = store.ReplaceGraph(context.Background(), graph, "test fixture"); err != nil {
		t.Fatal(err)
	}
	got, err := store.Graph(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Assets) != 1 || got.Assets[0].ID != "node" {
		t.Fatalf("unexpected graph: %#v", got)
	}
	history, err := store.ImportHistory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0]["source"] != "test fixture" {
		t.Fatalf("unexpected history: %#v", history)
	}
}

func TestReplaceGraphPersistsTypedEntitiesEvidenceAndValidationResult(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	graph := domain.Graph{
		SchemaVersion:     1,
		Assets:            []domain.Asset{{ID: "database", Name: "Synthetic database", Kind: "database", Impact: 5, Description: "Fixture"}},
		Identities:        []domain.Identity{{ID: "workload", Name: "Synthetic workload", Kind: "service", Impact: 3, Description: "Fixture"}},
		Misconfigurations: []domain.Misconfiguration{{ID: "misconfig", AssetID: "database", IdentityID: "workload", Title: "Overbroad access", Severity: 4, Description: "Fixture only"}},
		Findings:          []domain.Finding{{ID: "finding", AssetID: "database", IdentityID: "workload", MisconfigurationID: "misconfig", Title: "Read permission", Severity: 4, Source: "fixture://source", Observed: "synthetic"}},
		Edges:             []domain.Edge{{ID: "reads", From: "workload", To: "database", Relation: "reads", Confidence: 85, Evidence: domain.Evidence{Source: "fixture://evidence", Finding: "Read permission", Observed: "synthetic"}}},
	}
	result := ImportResult{SchemaVersion: 1, IdentityCount: 1, FindingCount: 1, MisconfigurationCount: 1, DocumentBytes: 512}
	if err = store.ReplaceGraphWithResult(context.Background(), graph, "typed fixture", result); err != nil {
		t.Fatal(err)
	}
	got, err := store.Graph(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(graph, got) {
		t.Fatalf("round trip mismatch\nwant: %#v\n got: %#v", graph, got)
	}
	history, err := store.ImportHistory(context.Background())
	if err != nil || len(history) != 1 || history[0]["identities"] != 1 || history[0]["findings"] != 1 || history[0]["misconfigurations"] != 1 || history[0]["schema_version"] != 1 || history[0]["document_bytes"] != 512 {
		t.Fatalf("validation history=%v err=%v", history, err)
	}
}

func TestReplaceGraphRejectsBeforeMutation(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	bad := domain.Graph{Assets: []domain.Asset{{ID: "bad", Name: "Bad", Kind: "host", Impact: 6}}}
	if err = store.ReplaceGraph(context.Background(), bad, "bad input"); err == nil {
		t.Fatal("expected validation error")
	}
	got, err := store.Graph(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Assets) != 0 {
		t.Fatalf("unexpected stored assets: %#v", got.Assets)
	}
}
