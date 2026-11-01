package ingestion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
)

func TestReadAcceptsTypedSyntheticInventory(t *testing.T) {
	input := `{"schema_version":1,"assets":[{"id":"db","name":"Data","kind":"database","impact":5}],"identities":[{"id":"workload","name":"Builder","kind":"workload","impact":3}],"misconfigurations":[{"id":"m1","asset_id":"db","identity_id":"workload","title":"Wide read access","severity":4,"description":"Fixture finding"}],"findings":[{"id":"f1","asset_id":"db","identity_id":"workload","misconfiguration_id":"m1","title":"Read access observed","severity":4,"source":"fixture://inventory","observed":"synthetic"}],"edges":[{"id":"e1","from":"workload","to":"db","relation":"reads","confidence":85,"evidence":{"source":"fixture://permissions","finding":"Read permission present","observed":"synthetic"}}]}`
	graph, result, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Assets) != 1 || len(graph.Identities) != 1 || len(graph.Misconfigurations) != 1 || len(graph.Findings) != 1 || len(graph.Edges) != 1 {
		t.Fatalf("incomplete inventory: %#v", graph)
	}
	if result.SchemaVersion != 1 || result.DocumentBytes != len(input) || result.Identities != 1 {
		t.Fatalf("unexpected validation result: %#v", result)
	}
}

func TestReadRejectsMalformedAndOutOfSchemaDocuments(t *testing.T) {
	for _, test := range []struct {
		name string
		data string
	}{
		{name: "malformed JSON", data: `{"assets":[`},
		{name: "unknown field", data: `{"assets":[{"id":"a","name":"A","kind":"service","impact":1,"secret":"x"}],"edges":[]}`},
		{name: "trailing JSON object", data: `{"assets":[{"id":"a","name":"A","kind":"service","impact":1}],"edges":[]}{}`},
		{name: "unsupported version", data: `{"schema_version":2,"assets":[{"id":"a","name":"A","kind":"service","impact":1}],"edges":[]}`},
		{name: "empty inventory", data: `{"assets":[],"edges":[]}`},
		{name: "missing evidence", data: `{"assets":[{"id":"a","name":"A","kind":"service","impact":1},{"id":"b","name":"B","kind":"database","impact":3}],"edges":[{"id":"e","from":"a","to":"b","relation":"reads","confidence":100}]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := Read(strings.NewReader(test.data)); err == nil {
				t.Fatal("expected inventory validation error")
			}
		})
	}
}

func TestReadRejectsOversizedDocumentAndRecordCount(t *testing.T) {
	if _, _, err := Read(strings.NewReader(strings.Repeat(" ", MaxDocumentBytes+1))); err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("oversized input error = %v", err)
	}
	assets := make([]domain.Asset, MaxAssets+1)
	for index := range assets {
		assets[index] = domain.Asset{ID: fmt.Sprintf("a-%d", index), Name: fmt.Sprintf("Asset %d", index), Kind: "service", Impact: 1}
	}
	data, err := json.Marshal(domain.Graph{Assets: assets})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = Read(bytes.NewReader(data)); err == nil || !strings.Contains(err.Error(), "10000") || !strings.Contains(err.Error(), "assets") {
		t.Fatalf("record-count error = %v", err)
	}
}
