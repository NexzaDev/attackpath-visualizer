package service

import (
	"context"
	"strings"
	"testing"

	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
)

func TestImportRejectsInvalidInputWithoutReplacingState(t *testing.T) {
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	app := Service{Store: store}
	valid := `{"assets":[{"id":"a","name":"A","kind":"service","impact":1}],"edges":[]}`
	if err = app.Import(context.Background(), strings.NewReader(valid), "fixture"); err != nil {
		t.Fatal(err)
	}
	if err = app.Import(context.Background(), strings.NewReader(`{"assets":[`), "bad"); err == nil {
		t.Fatal("expected malformed JSON error")
	}
	graph, err := store.Graph(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Assets) != 1 || graph.Assets[0].ID != "a" {
		t.Fatalf("existing graph was changed: %#v", graph)
	}
}

func TestImportRejectsOversizedDocument(t *testing.T) {
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	app := Service{Store: store}
	if err := app.Import(context.Background(), strings.NewReader(strings.Repeat("x", maxImportBytes+1)), "large"); err == nil {
		t.Fatal("expected size error")
	}
}
