package ingestion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
)

const (
	MaxDocumentBytes = 8 << 20
	MaxAssets        = 10000
	MaxIdentities    = 10000
	MaxEdges         = 50000
	MaxFindings      = 50000
	MaxMisconfigs    = 50000
	CurrentVersion   = 1
)

type Validation struct {
	SchemaVersion     int `json:"schema_version"`
	Assets            int `json:"assets"`
	Identities        int `json:"identities"`
	Edges             int `json:"edges"`
	Findings          int `json:"findings"`
	Misconfigurations int `json:"misconfigurations"`
	DocumentBytes     int `json:"document_bytes"`
}

func Read(reader io.Reader) (domain.Graph, Validation, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxDocumentBytes+1))
	if err != nil {
		return domain.Graph{}, Validation{}, fmt.Errorf("read inventory: %w", err)
	}
	if len(data) == 0 {
		return domain.Graph{}, Validation{}, errors.New("inventory is empty; provide a JSON graph with at least one asset or identity")
	}
	if len(data) > MaxDocumentBytes {
		return domain.Graph{}, Validation{}, fmt.Errorf("inventory exceeds the %d byte limit", MaxDocumentBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var graph domain.Graph
	if err = decoder.Decode(&graph); err != nil {
		return domain.Graph{}, Validation{}, fmt.Errorf("inventory JSON schema: %w", err)
	}
	var trailing any
	if err = decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return domain.Graph{}, Validation{}, errors.New("inventory must contain exactly one JSON object")
		}
		return domain.Graph{}, Validation{}, fmt.Errorf("inventory has trailing data: %w", err)
	}
	if graph.SchemaVersion != 0 && graph.SchemaVersion != CurrentVersion {
		return domain.Graph{}, Validation{}, fmt.Errorf("unsupported schema_version %d; supported versions are 1 and legacy versionless JSON", graph.SchemaVersion)
	}
	if len(graph.Assets)+len(graph.Identities) == 0 {
		return domain.Graph{}, Validation{}, errors.New("inventory must contain at least one asset or identity")
	}
	for _, item := range []struct {
		name  string
		count int
		limit int
	}{
		{name: "assets", count: len(graph.Assets), limit: MaxAssets},
		{name: "identities", count: len(graph.Identities), limit: MaxIdentities},
		{name: "edges", count: len(graph.Edges), limit: MaxEdges},
		{name: "findings", count: len(graph.Findings), limit: MaxFindings},
		{name: "misconfigurations", count: len(graph.Misconfigurations), limit: MaxMisconfigs},
	} {
		if item.count > item.limit {
			return domain.Graph{}, Validation{}, fmt.Errorf("inventory contains %d %s; the limit is %d", item.count, item.name, item.limit)
		}
	}
	if err = graph.Validate(); err != nil {
		return domain.Graph{}, Validation{}, fmt.Errorf("validate inventory: %w", err)
	}
	graph.SchemaVersion = CurrentVersion
	return graph, Validation{SchemaVersion: CurrentVersion, Assets: len(graph.Assets), Identities: len(graph.Identities), Edges: len(graph.Edges), Findings: len(graph.Findings), Misconfigurations: len(graph.Misconfigurations), DocumentBytes: len(data)}, nil
}
