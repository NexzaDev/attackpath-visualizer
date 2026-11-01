package reports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
	"github.com/NexzaDev/attackpath-visualizer/internal/remediation"
)

const SchemaVersion = 1

type Bundle struct {
	SchemaVersion  int                `json:"schema_version"`
	PathCount      int                `json:"path_count"`
	PathsTruncated bool               `json:"paths_truncated"`
	Expansions     int                `json:"edge_expansions"`
	Paths          []domain.Path      `json:"paths"`
	Remediation    remediation.Result `json:"remediation"`
}

func Build(paths domain.PathResult, ranked remediation.Result) Bundle {
	return Bundle{SchemaVersion: SchemaVersion, PathCount: len(paths.Paths), PathsTruncated: paths.Truncated, Expansions: paths.Expanded, Paths: paths.Paths, Remediation: ranked}
}

func JSON(bundle Bundle) ([]byte, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(bundle); err != nil {
		return nil, fmt.Errorf("encode path and remediation report: %w", err)
	}
	return output.Bytes(), nil
}

func Markdown(bundle Bundle) string {
	var output strings.Builder
	output.WriteString("# AttackPath Visualizer — Path and Remediation Report\n\n")
	output.WriteString("Potential relationships are advisory and are not proof of exploitability. No changes are made to the inventory.\n\n")
	output.WriteString(fmt.Sprintf("- Paths included: %d\n- Enumeration expansions: %d\n- Results truncated: %t\n- Scoring: maximum node impact tier × compounded edge confidence; integer floor after each edge\n\n", bundle.PathCount, bundle.Expansions, bundle.PathsTruncated))
	output.WriteString("## Ranked paths\n\n")
	if len(bundle.Paths) == 0 {
		output.WriteString("No directed paths were found within the configured depth and expansion limits.\n\n")
	} else {
		output.WriteString("| Rank | Path | Impact | Edge confidence components | Confidence | Score |\n|---:|---|---:|---|---:|---:|\n")
		for index, path := range bundle.Paths {
			labels := make([]string, len(path.Nodes))
			for nodeIndex, node := range path.Nodes {
				labels[nodeIndex] = fmt.Sprintf("%s (`%s`)", node.Name, node.ID)
			}
			confidences := make([]string, len(path.Components.EdgeConfidences))
			for edgeIndex, confidence := range path.Components.EdgeConfidences {
				confidences[edgeIndex] = fmt.Sprintf("%d%%", confidence)
			}
			output.WriteString(fmt.Sprintf("| %d | %s | %d/5 | %s | %d%% | %d |\n", index+1, cell(strings.Join(labels, " → ")), path.Components.ImpactTier, cell(strings.Join(confidences, " × ")), path.Components.CompoundedConfidence, path.Components.Score))
		}
		output.WriteString("\n")
		for index, path := range bundle.Paths {
			output.WriteString(fmt.Sprintf("### Path %d evidence\n\n", index+1))
			output.WriteString(path.Explanation + "\n\n")
			for _, edge := range path.Edges {
				output.WriteString(fmt.Sprintf("- `%s`: `%s` → `%s` — %s (%d%% confidence). Source: %s. Finding: %s. Observed: %s.\n", edge.ID, edge.From, edge.To, edge.Relation, edge.Confidence, cell(edge.Evidence.Source), cell(edge.Evidence.Finding), cell(edge.Evidence.Observed)))
			}
			output.WriteString("\n")
		}
	}
	output.WriteString("## Ranked remediation suggestions\n\n")
	if len(bundle.Remediation.Suggestions) == 0 {
		output.WriteString("No path-affecting relationships are available for ranking.\n")
	} else {
		output.WriteString("Suggestions rank relationship choke points by enumerated paths affected, then weighted risk points. Validate with the asset owner before changing access.\n\n")
		output.WriteString("| Rank | Edge | Relationship | Paths affected | Critical paths | Weighted risk points | Recommendation |\n|---:|---|---|---:|---:|---:|---|\n")
		for index, item := range bundle.Remediation.Suggestions {
			output.WriteString(fmt.Sprintf("| %d | `%s` | %s | %d | %d | %d | %s |\n", index+1, item.EdgeID, cell(item.Relation), item.PathsAffected, item.CriticalPaths, item.WeightedRiskPoints, cell(item.Recommendation)))
		}
	}
	return output.String()
}

func cell(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " ")
}
