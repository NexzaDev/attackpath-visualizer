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

func TestImportNotificationsAreMinimalAndDeduplicated(t *testing.T) {
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	app := Service{Store: store}
	invalid := `{"assets":[`
	if err = app.Import(context.Background(), strings.NewReader(invalid), "bad input"); err == nil {
		t.Fatal("expected invalid JSON")
	}
	if err = app.Import(context.Background(), strings.NewReader(invalid), "bad input"); err == nil {
		t.Fatal("expected invalid JSON")
	}
	items, err := store.Notifications(context.Background(), 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "Import failed validation" || strings.Contains(items[0].Detail, invalid) {
		t.Fatalf("unexpected validation notification: %#v", items)
	}
}

func TestImportRecordsNewHighImpactPaths(t *testing.T) {
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	app := Service{Store: store}
	graph := `{"assets":[{"id":"entry","name":"Entry","kind":"service","impact":1},{"id":"data","name":"Sensitive dataset","kind":"database","impact":5}],"edges":[{"id":"relationship","from":"entry","to":"data","relation":"has access","confidence":95,"evidence":{"source":"fixture://test","finding":"Access path exists","observed":"synthetic"}}]}`
	if err = app.Import(context.Background(), strings.NewReader(graph), "synthetic test"); err != nil {
		t.Fatal(err)
	}
	items, err := store.Notifications(context.Background(), 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "New critical potential path" {
		t.Fatalf("unexpected critical path notification: %#v", items)
	}
}

func TestImportNotifiesWhenExistingPathBecomesCritical(t *testing.T) {
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	app := Service{Store: store}
	lowRisk := `{"assets":[{"id":"entry","name":"Entry","kind":"service","impact":1},{"id":"data","name":"Sensitive dataset","kind":"database","impact":5}],"edges":[{"id":"relationship","from":"entry","to":"data","relation":"has access","confidence":50,"evidence":{"source":"fixture://test","finding":"Access path exists","observed":"synthetic"}}]}`
	if err = app.Import(context.Background(), strings.NewReader(lowRisk), "low risk"); err != nil {
		t.Fatal(err)
	}
	highRisk := strings.Replace(lowRisk, `"confidence":50`, `"confidence":95`, 1)
	if err = app.Import(context.Background(), strings.NewReader(highRisk), "increased confidence"); err != nil {
		t.Fatal(err)
	}
	items, err := store.Notifications(context.Background(), 20)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%v err=%v", items, err)
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
