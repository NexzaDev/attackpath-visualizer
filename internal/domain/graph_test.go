package domain

import "testing"

func TestFindPathsRanksAndExplains(t *testing.T) {
	graph := Graph{Assets: []Asset{{ID: "a", Name: "Public gateway", Kind: "service", Impact: 2}, {ID: "b", Name: "Identity provider", Kind: "identity", Impact: 4}, {ID: "c", Name: "Data store", Kind: "database", Impact: 5}}, Edges: []Edge{{ID: "ab", From: "a", To: "b", Relation: "assumes role", Confidence: 90}, {ID: "bc", From: "b", To: "c", Relation: "reads", Confidence: 80}}}
	paths, err := graph.FindPaths("a", "c", 8, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0].Confidence != 72 || paths[0].Score != 360 {
		t.Fatalf("unexpected path result: %#v", paths)
	}
	if paths[0].Explanation != "impact 5/5 × confidence 72% = 360" {
		t.Fatalf("unexpected explanation: %q", paths[0].Explanation)
	}
}

func TestValidateRejectsDanglingEdges(t *testing.T) {
	graph := Graph{Assets: []Asset{{ID: "a", Name: "A", Kind: "service", Impact: 1}}, Edges: []Edge{{ID: "e", From: "a", To: "missing", Relation: "connects", Confidence: 100}}}
	if err := graph.Validate(); err == nil {
		t.Fatal("expected invalid edge error")
	}
}

func TestFindPathsRejectsCyclesAndInvalidLimits(t *testing.T) {
	graph := Graph{Assets: []Asset{{ID: "a", Name: "A", Kind: "service", Impact: 1}, {ID: "b", Name: "B", Kind: "service", Impact: 2}}, Edges: []Edge{{ID: "ab", From: "a", To: "b", Relation: "connects", Confidence: 100}, {ID: "ba", From: "b", To: "a", Relation: "connects", Confidence: 100}}}
	paths, err := graph.FindPaths("a", "", 4, 10)
	if err != nil || len(paths) != 1 {
		t.Fatalf("paths=%v err=%v", paths, err)
	}
	if _, err := graph.FindPaths("a", "", 0, 10); err == nil {
		t.Fatal("expected limit validation")
	}
}
