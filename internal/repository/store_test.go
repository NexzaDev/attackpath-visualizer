package repository

import (
	"context"
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
