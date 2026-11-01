package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
	"github.com/NexzaDev/attackpath-visualizer/internal/remediation"
	"github.com/NexzaDev/attackpath-visualizer/internal/reports"
)

type Analysis struct {
	Graph       domain.Graph       `json:"graph"`
	Paths       domain.PathResult  `json:"paths"`
	Remediation remediation.Result `json:"remediation"`
}

type ConfidenceBand struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type Summary struct {
	Analysis
	PathsByImpact [5]int           `json:"paths_by_impact"`
	Confidence    []ConfidenceBand `json:"confidence_distribution"`
}

func (s Service) Analyze(ctx context.Context, maxDepth, limit int) (Analysis, error) {
	graph, err := s.Store.Graph(ctx)
	if err != nil {
		return Analysis{}, err
	}
	paths, err := graph.EnumeratePaths(maxDepth, limit)
	if err != nil {
		return Analysis{}, fmt.Errorf("enumerate ranked paths: %w", err)
	}
	return Analysis{Graph: graph, Paths: paths, Remediation: remediation.Rank(graph, paths)}, nil
}

func (s Service) Overview(ctx context.Context) (Summary, error) {
	analysis, err := s.Analyze(ctx, 16, 1000)
	if err != nil {
		return Summary{}, err
	}
	summary := Summary{Analysis: analysis, Confidence: []ConfidenceBand{{Label: "0–19%", Count: 0}, {Label: "20–39%", Count: 0}, {Label: "40–59%", Count: 0}, {Label: "60–79%", Count: 0}, {Label: "80–100%", Count: 0}}}
	for _, path := range analysis.Paths.Paths {
		if path.Impact >= 1 && path.Impact <= 5 {
			summary.PathsByImpact[path.Impact-1]++
		}
		band := min(path.Confidence/20, 4)
		summary.Confidence[band].Count++
	}
	return summary, nil
}

func (s Service) RenderReport(ctx context.Context, format string) ([]byte, error) {
	analysis, err := s.Analyze(ctx, 16, 1000)
	if err != nil {
		return nil, err
	}
	bundle := reports.Build(analysis.Paths, analysis.Remediation)
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "json":
		return reports.JSON(bundle)
	case "markdown", "md":
		return []byte(reports.Markdown(bundle)), nil
	default:
		return nil, errors.New("report format must be json or markdown")
	}
}
