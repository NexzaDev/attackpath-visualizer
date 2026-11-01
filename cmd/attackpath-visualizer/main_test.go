package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalysisCommandsUseStoredGraph(t *testing.T) {
	database := filepath.Join(t.TempDir(), "cli.db")
	var output bytes.Buffer
	if err := run([]string{"remediation", "--db", database, "--json"}, &output, &output); err != nil {
		t.Fatal(err)
	}
	var remediation struct {
		Suggestions []struct {
			EdgeID        string `json:"edge_id"`
			PathsAffected int    `json:"paths_affected"`
		} `json:"suggestions"`
	}
	if err := json.Unmarshal(output.Bytes(), &remediation); err != nil {
		t.Fatal(err)
	}
	if len(remediation.Suggestions) == 0 || remediation.Suggestions[0].EdgeID == "" || remediation.Suggestions[0].PathsAffected == 0 {
		t.Fatalf("remediation command returned no ranked choke points: %s", output.String())
	}

	output.Reset()
	if err := run([]string{"report", "--db", database, "--format", "markdown"}, &output, &output); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Ranked paths", "Ranked remediation suggestions", "demo://"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("Markdown report omitted %q", expected)
		}
	}

	output.Reset()
	if err := run([]string{"report", "--db", database, "--json"}, &output, &output); err != nil {
		t.Fatal(err)
	}
	var report struct {
		PathCount int `json:"path_count"`
	}
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.PathCount == 0 {
		t.Fatal("JSON report has no paths")
	}
}
