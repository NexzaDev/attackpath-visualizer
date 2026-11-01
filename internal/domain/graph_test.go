package domain

import "testing"

func TestFindPathsRanksAndExplains(t *testing.T) {
	graph := Graph{
		Assets: []Asset{{ID: "a", Name: "Public gateway", Kind: "service", Impact: 2}, {ID: "b", Name: "Identity provider", Kind: "identity", Impact: 4}, {ID: "c", Name: "Data store", Kind: "database", Impact: 5}},
		Edges: []Edge{
			{ID: "ab", From: "a", To: "b", Relation: "assumes role", Confidence: 90, Evidence: testEvidence("trust policy")},
			{ID: "bc", From: "b", To: "c", Relation: "reads", Confidence: 80, Evidence: testEvidence("read permission")},
		},
	}
	paths, err := graph.FindPaths("a", "c", 8, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0].Confidence != 72 || paths[0].Score != 360 {
		t.Fatalf("unexpected path result: %#v", paths)
	}
	if len(paths[0].Components.EdgeConfidences) != 2 || paths[0].Components.ImpactTier != 5 {
		t.Fatalf("score components are incomplete: %#v", paths[0].Components)
	}
	if len(paths[0].Nodes) != 3 || len(paths[0].Edges) != 2 {
		t.Fatalf("path is missing nodes or edges: %#v", paths[0])
	}
}

func TestAllPathsDeduplicatesAndReportsBound(t *testing.T) {
	graph := Graph{
		Assets: []Asset{{ID: "a", Name: "A", Kind: "service", Impact: 1}, {ID: "b", Name: "B", Kind: "service", Impact: 2}, {ID: "c", Name: "C", Kind: "database", Impact: 5}},
		Edges: []Edge{
			{ID: "ab", From: "a", To: "b", Relation: "connects", Confidence: 100, Evidence: testEvidence("ab")},
			{ID: "bc", From: "b", To: "c", Relation: "connects", Confidence: 100, Evidence: testEvidence("bc")},
			{ID: "ac", From: "a", To: "c", Relation: "connects", Confidence: 80, Evidence: testEvidence("ac")},
		},
	}
	result, err := graph.EnumeratePaths(8, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Paths) != 1 || !result.Truncated {
		t.Fatalf("expected one bounded path and a truncation marker: %#v", result)
	}
	paths, err := graph.FindPaths("a", "c", 8, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || paths[0].Score != 500 || paths[1].Score != 400 {
		t.Fatalf("expected deterministic ranking of the two distinct paths: %#v", paths)
	}
}

func TestAllPathsIncludesDisconnectedCycleComponents(t *testing.T) {
	graph := Graph{
		Assets: []Asset{{ID: "root", Name: "Root", Kind: "service", Impact: 2}, {ID: "leaf", Name: "Leaf", Kind: "database", Impact: 3}, {ID: "cycle-a", Name: "Cycle A", Kind: "service", Impact: 2}, {ID: "cycle-b", Name: "Cycle B", Kind: "service", Impact: 3}},
		Edges: []Edge{
			{ID: "root-leaf", From: "root", To: "leaf", Relation: "connects", Confidence: 100, Evidence: testEvidence("root-leaf")},
			{ID: "cycle-ab", From: "cycle-a", To: "cycle-b", Relation: "connects", Confidence: 100, Evidence: testEvidence("cycle-ab")},
			{ID: "cycle-ba", From: "cycle-b", To: "cycle-a", Relation: "connects", Confidence: 100, Evidence: testEvidence("cycle-ba")},
		},
	}
	paths, err := graph.AllPaths(8, 20)
	if err != nil {
		t.Fatal(err)
	}
	foundCycle := false
	for _, path := range paths {
		if len(path.Nodes) == 2 && path.Nodes[0].ID == "cycle-a" && path.Nodes[1].ID == "cycle-b" {
			foundCycle = true
		}
	}
	if !foundCycle {
		t.Fatalf("a disconnected cyclic component was skipped: %#v", paths)
	}
}

func TestValidateRejectsDanglingNodesAndEvidence(t *testing.T) {
	graph := Graph{
		Assets: []Asset{{ID: "a", Name: "A", Kind: "service", Impact: 1}},
		Edges:  []Edge{{ID: "e", From: "a", To: "missing", Relation: "connects", Confidence: 100, Evidence: testEvidence("test")}},
	}
	if err := graph.Validate(); err == nil {
		t.Fatal("expected invalid edge error")
	}
	graph = Graph{Assets: []Asset{{ID: "a", Name: "A", Kind: "service", Impact: 1}, {ID: "b", Name: "B", Kind: "service", Impact: 2}}, Edges: []Edge{{ID: "e", From: "a", To: "b", Relation: "connects", Confidence: 100}}}
	if err := graph.Validate(); err == nil {
		t.Fatal("expected missing evidence error")
	}
}

func TestValidatesIdentityMisconfigurationAndFindingReferences(t *testing.T) {
	graph := Graph{
		Assets:     []Asset{{ID: "db", Name: "Database", Kind: "database", Impact: 5}},
		Identities: []Identity{{ID: "workload", Name: "Workload identity", Kind: "workload", Impact: 3}},
		Misconfigurations: []Misconfiguration{{ID: "misconfig", AssetID: "db", IdentityID: "workload", Title: "Overbroad access", Severity: 4, Description: "Synthetic access issue"}},
		Findings: []Finding{{ID: "finding", AssetID: "db", IdentityID: "workload", MisconfigurationID: "misconfig", Title: "Read access observed", Severity: 4, Source: "fixture://permissions", Observed: "synthetic"}},
		Edges: []Edge{{ID: "access", From: "workload", To: "db", Relation: "reads", Confidence: 90, Evidence: testEvidence("read access")}},
	}
	if err := graph.Validate(); err != nil {
		t.Fatal(err)
	}
	paths, err := graph.FindPaths("workload", "db", 4, 4)
	if err != nil || len(paths) != 1 {
		t.Fatalf("paths=%v err=%v", paths, err)
	}
	if paths[0].Nodes[0].EntityType != "identity" || paths[0].Nodes[0].Kind != "workload" || paths[0].Nodes[1].Kind != "database" {
		t.Fatalf("typed identities are missing from graph paths: %#v", paths[0].Nodes)
	}
	graph.Findings[0].MisconfigurationID = "missing"
	if err := graph.Validate(); err == nil {
		t.Fatal("expected missing misconfiguration reference error")
	}
}

func TestFindPathsRejectsCyclesAndInvalidLimits(t *testing.T) {
	graph := Graph{
		Assets: []Asset{{ID: "a", Name: "A", Kind: "service", Impact: 1}, {ID: "b", Name: "B", Kind: "service", Impact: 2}},
		Edges: []Edge{
			{ID: "ab", From: "a", To: "b", Relation: "connects", Confidence: 100, Evidence: testEvidence("ab")},
			{ID: "ba", From: "b", To: "a", Relation: "connects", Confidence: 100, Evidence: testEvidence("ba")},
		},
	}
	paths, err := graph.FindPaths("a", "", 4, 10)
	if err != nil || len(paths) != 1 {
		t.Fatalf("paths=%v err=%v", paths, err)
	}
	if _, err := graph.FindPaths("a", "", 0, 10); err == nil {
		t.Fatal("expected depth validation")
	}
}

func testEvidence(finding string) Evidence {
	return Evidence{Source: "fixture://test", Finding: finding, Observed: "synthetic"}
}
