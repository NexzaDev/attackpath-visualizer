package reports

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
	"github.com/NexzaDev/attackpath-visualizer/internal/remediation"
)

func TestReportsIncludeScoreComponentsEvidenceAndCoverage(t *testing.T) {
	path := domain.Path{
		Nodes:  []domain.Node{{ID: "entry", Name: "Entry", EntityType: "asset", Kind: "service", Impact: 1}, {ID: "data", Name: "Data", EntityType: "asset", Kind: "database", Impact: 5}},
		Edges:  []domain.Edge{{ID: "read-edge", From: "entry", To: "data", Relation: "reads", Confidence: 80, Evidence: domain.Evidence{Source: "fixture://access", Finding: "Read grant", Observed: "synthetic"}}},
		Impact: 5, Confidence: 80, Score: 400, Components: domain.ScoreBreakdown{ImpactTier: 5, EdgeConfidences: []int{80}, CompoundedConfidence: 80, Score: 400},
		Explanation: "Impact tier 5 × compounded confidence 80% (80%; floor after each edge) = 400 points",
	}
	paths := domain.PathResult{Paths: []domain.Path{path}, Truncated: true, Expanded: 25}
	ranked := remediation.Result{PathsConsidered: 1, PathsTruncated: true, Suggestions: []remediation.Suggestion{{EdgeID: "read-edge", From: "entry", To: "data", Relation: "reads", PathsAffected: 1, CriticalPaths: 1, WeightedRiskPoints: 400, Recommendation: "Review read access"}}}
	bundle := Build(paths, ranked)
	data, err := JSON(bundle)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Bundle
	if err = json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.PathCount != 1 || !decoded.PathsTruncated || decoded.Expansions != 25 || decoded.Paths[0].Components.EdgeConfidences[0] != 80 || decoded.Remediation.Suggestions[0].WeightedRiskPoints != 400 {
		t.Fatalf("JSON report lost ranked analysis data: %#v", decoded)
	}
	markdown := Markdown(bundle)
	for _, required := range []string{"## Ranked paths", "Edge confidence components", "80%", "fixture://access", "Read grant", "## Ranked remediation suggestions", "read-edge", "400", "Results truncated: true"} {
		if !strings.Contains(markdown, required) {
			t.Fatalf("Markdown report does not include %q:\n%s", required, markdown)
		}
	}
}

func TestMarkdownReportHandlesEmptyPathsAndEscapesCells(t *testing.T) {
	bundle := Build(domain.PathResult{}, remediation.Result{})
	markdown := Markdown(bundle)
	if !strings.Contains(markdown, "No directed paths were found") || !strings.Contains(markdown, "No path-affecting relationships") {
		t.Fatalf("empty report lacks useful states: %s", markdown)
	}
	if got := cell("alpha|beta\ngamma"); got != "alpha\\|beta gamma" {
		t.Fatalf("cell escaping = %q", got)
	}
}
