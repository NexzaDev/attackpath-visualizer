package remediation

import (
	"testing"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
)

func TestRankPrioritizesChokePointByPathsBrokenAndRisk(t *testing.T) {
	graph := domain.Graph{
		Assets: []domain.Asset{
			{ID: "a", Name: "Entry", Kind: "service", Impact: 1},
			{ID: "b", Name: "Broker", Kind: "service", Impact: 2},
			{ID: "c", Name: "Worker", Kind: "compute", Impact: 3},
			{ID: "d", Name: "Customer data", Kind: "database", Impact: 5},
			{ID: "e", Name: "Backup data", Kind: "storage", Impact: 5},
		},
		Edges: []domain.Edge{
			{ID: "entry-broker", From: "a", To: "b", Relation: "connects", Confidence: 100, Evidence: evidence("entry")},
			{ID: "broker-data", From: "b", To: "d", Relation: "reads", Confidence: 90, Evidence: evidence("database")},
			{ID: "broker-worker", From: "b", To: "c", Relation: "invokes", Confidence: 95, Evidence: evidence("worker")},
			{ID: "worker-data", From: "c", To: "d", Relation: "reads", Confidence: 90, Evidence: evidence("database")},
			{ID: "worker-backup", From: "c", To: "e", Relation: "writes", Confidence: 80, Evidence: evidence("backup")},
		},
	}
	paths, err := graph.EnumeratePaths(8, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths.Paths) != 5 {
		t.Fatalf("expected five source-rooted path prefixes, got %d", len(paths.Paths))
	}
	paths.Truncated = true
	result := Rank(graph, paths)
	if len(result.Suggestions) != 5 || result.PathsConsidered != 5 || !result.PathsTruncated {
		t.Fatalf("unexpected remediation result: %#v", result)
	}
	first := result.Suggestions[0]
	if first.EdgeID != "entry-broker" || first.PathsAffected != 5 || first.WeightedRiskPoints == 0 {
		t.Fatalf("shared choke point was not ranked first: %#v", first)
	}
	if first.CriticalPaths != 3 || first.Recommendation == "" {
		t.Fatalf("critical paths/recommendation missing: %#v", first)
	}
}

func evidence(finding string) domain.Evidence {
	return domain.Evidence{Source: "fixture://test", Finding: finding, Observed: "synthetic"}
}
