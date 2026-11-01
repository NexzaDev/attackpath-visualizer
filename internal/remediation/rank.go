package remediation

import (
	"fmt"
	"sort"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
)

type Suggestion struct {
	EdgeID             string `json:"edge_id"`
	From               string `json:"from"`
	To                 string `json:"to"`
	Relation           string `json:"relation"`
	PathsAffected      int    `json:"paths_affected"`
	CriticalPaths      int    `json:"critical_paths"`
	WeightedRiskPoints int    `json:"weighted_risk_points"`
	Recommendation     string `json:"recommendation"`
}

type Result struct {
	Suggestions     []Suggestion `json:"suggestions"`
	PathsConsidered int          `json:"paths_considered"`
	PathsTruncated  bool         `json:"paths_truncated"`
}

type edgeImpact struct {
	edge     domain.Edge
	paths    int
	critical int
	points   int
}

func Rank(graph domain.Graph, pathSet domain.PathResult) Result {
	impacts := make(map[string]*edgeImpact, len(graph.Edges))
	for _, edge := range graph.Edges {
		impacts[edge.ID] = &edgeImpact{edge: edge}
	}
	for _, path := range pathSet.Paths {
		for _, edge := range path.Edges {
			impact := impacts[edge.ID]
			if impact == nil {
				continue
			}
			impact.paths++
			impact.points += path.Score
			if path.Impact >= 4 && path.Confidence >= 60 {
				impact.critical++
			}
		}
	}
	suggestions := make([]Suggestion, 0, len(impacts))
	for _, impact := range impacts {
		if impact.paths == 0 {
			continue
		}
		suggestions = append(suggestions, Suggestion{
			EdgeID: impact.edge.ID, From: impact.edge.From, To: impact.edge.To, Relation: impact.edge.Relation,
			PathsAffected: impact.paths, CriticalPaths: impact.critical, WeightedRiskPoints: impact.points,
			Recommendation: fmt.Sprintf("Review and, where appropriate, restrict or remove the %q relationship; it appears in %d enumerated paths.", impact.edge.Relation, impact.paths),
		})
	}
	sort.Slice(suggestions, func(i, j int) bool {
		left, right := suggestions[i], suggestions[j]
		if left.PathsAffected != right.PathsAffected {
			return left.PathsAffected > right.PathsAffected
		}
		if left.WeightedRiskPoints != right.WeightedRiskPoints {
			return left.WeightedRiskPoints > right.WeightedRiskPoints
		}
		if left.CriticalPaths != right.CriticalPaths {
			return left.CriticalPaths > right.CriticalPaths
		}
		return left.EdgeID < right.EdgeID
	})
	return Result{Suggestions: suggestions, PathsConsidered: len(pathSet.Paths), PathsTruncated: pathSet.Truncated}
}
