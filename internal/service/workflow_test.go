package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/NexzaDev/attackpath-visualizer/internal/reports"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
)

func TestImportRankRemediationAndReportsEndToEnd(t *testing.T) {
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	app := Service{Store: store}
	inventory := `{"schema_version":1,"assets":[{"id":"entry","name":"Synthetic entry","kind":"service","impact":1},{"id":"broker","name":"Synthetic broker","kind":"service","impact":2},{"id":"customer-data","name":"Synthetic data","kind":"database","impact":5},{"id":"backup","name":"Synthetic backup","kind":"storage","impact":4}],"edges":[{"id":"entry-broker","from":"entry","to":"broker","relation":"connects","confidence":100,"evidence":{"source":"fixture://topology","finding":"Entry reaches broker","observed":"synthetic"}},{"id":"broker-data","from":"broker","to":"customer-data","relation":"reads","confidence":95,"evidence":{"source":"fixture://permissions","finding":"Broker can read data","observed":"synthetic"}},{"id":"broker-backup","from":"broker","to":"backup","relation":"writes","confidence":80,"evidence":{"source":"fixture://permissions","finding":"Broker can write backups","observed":"synthetic"}},{"id":"entry-direct","from":"entry","to":"customer-data","relation":"legacy reads","confidence":50,"evidence":{"source":"fixture://review","finding":"Low confidence direct relationship","observed":"synthetic"}}]}`
	if err = app.Import(context.Background(), strings.NewReader(inventory), "authorized synthetic E2E"); err != nil {
		t.Fatal(err)
	}
	analysis, err := app.Analyze(context.Background(), 16, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Paths.Paths) != 4 || analysis.Paths.Truncated {
		t.Fatalf("unexpected ranked path result: %#v", analysis.Paths)
	}
	if analysis.Paths.Paths[0].Score != 475 || analysis.Paths.Paths[0].Components.EdgeConfidences[0] != 100 || analysis.Paths.Paths[0].Components.EdgeConfidences[1] != 95 {
		t.Fatalf("path ranking or exact scoring components incorrect: %#v", analysis.Paths.Paths[0])
	}
	summary, err := app.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	impactCount := 0
	confidenceCount := 0
	for _, count := range summary.PathsByImpact {
		impactCount += count
	}
	for _, band := range summary.Confidence {
		confidenceCount += band.Count
	}
	if impactCount != 4 || confidenceCount != 4 {
		t.Fatalf("overview distributions omit enumerated paths: impacts=%v confidence=%v", summary.PathsByImpact, summary.Confidence)
	}
	if len(analysis.Remediation.Suggestions) != 4 || analysis.Remediation.Suggestions[0].EdgeID != "entry-broker" || analysis.Remediation.Suggestions[0].PathsAffected != 3 {
		t.Fatalf("choke point remediation ranking incorrect: %#v", analysis.Remediation)
	}
	jsonReport, err := app.RenderReport(context.Background(), "json")
	if err != nil {
		t.Fatal(err)
	}
	var bundle reports.Bundle
	if err = json.Unmarshal(jsonReport, &bundle); err != nil {
		t.Fatal(err)
	}
	if bundle.PathCount != 4 || bundle.Remediation.Suggestions[0].EdgeID != "entry-broker" || bundle.Paths[0].Components.Score != 475 {
		t.Fatalf("JSON report omitted the computed results: %#v", bundle)
	}
	markdownReport, err := app.RenderReport(context.Background(), "markdown")
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"Synthetic entry", "customer-data", "fixture://permissions", "entry-broker", "2", "475"} {
		if !strings.Contains(string(markdownReport), part) {
			t.Fatalf("Markdown report does not include %q:\n%s", part, markdownReport)
		}
	}
	if _, err = app.RenderReport(context.Background(), "html"); err == nil {
		t.Fatal("unsupported report format was accepted")
	}
	history, err := store.ImportHistory(context.Background())
	if err != nil || len(history) != 1 || history[0]["source"] != "authorized synthetic E2E" || history[0]["status"] != "success" || history[0]["schema_version"] != 1 || history[0]["document_bytes"] != len(inventory) {
		t.Fatalf("import validation history=%v err=%v", history, err)
	}
}
