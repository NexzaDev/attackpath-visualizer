package domain

import (
	"errors"
	"fmt"
	"sort"
)

type Asset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Impact      int    `json:"impact"`
	Description string `json:"description"`
}

type Evidence struct {
	Source   string `json:"source"`
	Finding  string `json:"finding"`
	Observed string `json:"observed"`
}

type Edge struct {
	ID         string   `json:"id"`
	From       string   `json:"from"`
	To         string   `json:"to"`
	Relation   string   `json:"relation"`
	Confidence int      `json:"confidence"`
	Evidence   Evidence `json:"evidence"`
}

type Graph struct {
	Assets []Asset `json:"assets"`
	Edges  []Edge  `json:"edges"`
}

type Path struct {
	Assets      []Asset `json:"assets"`
	Edges       []Edge  `json:"edges"`
	Impact      int     `json:"impact"`
	Confidence  int     `json:"confidence"`
	Score       int     `json:"score"`
	Explanation string  `json:"explanation"`
}

func (g Graph) Validate() error {
	assets := make(map[string]Asset, len(g.Assets))
	for _, asset := range g.Assets {
		if asset.ID == "" || asset.Name == "" || asset.Kind == "" || asset.Impact < 1 || asset.Impact > 5 {
			return fmt.Errorf("asset %q requires an id, name, kind, and impact from 1 to 5", asset.ID)
		}
		if _, exists := assets[asset.ID]; exists {
			return fmt.Errorf("duplicate asset id %q", asset.ID)
		}
		assets[asset.ID] = asset
	}
	edges := make(map[string]struct{}, len(g.Edges))
	for _, edge := range g.Edges {
		if edge.ID == "" || edge.Relation == "" || edge.Confidence < 0 || edge.Confidence > 100 {
			return fmt.Errorf("edge %q requires an id, relation, and confidence from 0 to 100", edge.ID)
		}
		if _, exists := edges[edge.ID]; exists {
			return fmt.Errorf("duplicate edge id %q", edge.ID)
		}
		edges[edge.ID] = struct{}{}
		if _, exists := assets[edge.From]; !exists {
			return fmt.Errorf("edge %q references missing source asset %q", edge.ID, edge.From)
		}
		if _, exists := assets[edge.To]; !exists {
			return fmt.Errorf("edge %q references missing target asset %q", edge.ID, edge.To)
		}
		if edge.From == edge.To {
			return fmt.Errorf("edge %q cannot connect an asset to itself", edge.ID)
		}
	}
	return nil
}

func (g Graph) AllPaths(maxDepth, limit int) ([]Path, error) {
	if maxDepth < 1 || maxDepth > 32 || limit < 1 || limit > 1000 {
		return nil, errors.New("max depth must be 1–32 and limit must be 1–1000")
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	incoming := make(map[string]bool, len(g.Assets))
	neighbors := make(map[string][]string, len(g.Assets))
	for _, edge := range g.Edges {
		incoming[edge.To] = true
		neighbors[edge.From] = append(neighbors[edge.From], edge.To)
		neighbors[edge.To] = append(neighbors[edge.To], edge.From)
	}
	assets := make([]string, 0, len(g.Assets))
	for _, asset := range g.Assets {
		assets = append(assets, asset.ID)
	}
	sort.Strings(assets)
	sources := make([]string, 0)
	visited := make(map[string]bool, len(g.Assets))
	for _, first := range assets {
		if visited[first] {
			continue
		}
		component := []string{first}
		visited[first] = true
		for index := 0; index < len(component); index++ {
			for _, neighbor := range neighbors[component[index]] {
				if !visited[neighbor] {
					visited[neighbor] = true
					component = append(component, neighbor)
				}
			}
		}
		componentRoots := make([]string, 0)
		for _, id := range component {
			if !incoming[id] {
				componentRoots = append(componentRoots, id)
			}
		}
		if len(componentRoots) == 0 {
			sources = append(sources, component...)
		} else {
			sources = append(sources, componentRoots...)
		}
	}
	sort.Strings(sources)
	paths := make([]Path, 0)
	for _, source := range sources {
		remaining := limit - len(paths)
		if remaining <= 0 {
			break
		}
		found, err := g.findPaths(source, "", maxDepth, remaining)
		if err != nil {
			return nil, err
		}
		paths = append(paths, found...)
	}
	sort.SliceStable(paths, func(i, j int) bool {
		if paths[i].Score == paths[j].Score {
			return paths[i].Confidence > paths[j].Confidence
		}
		return paths[i].Score > paths[j].Score
	})
	return paths, nil
}

func (g Graph) FindPaths(source, target string, maxDepth, limit int) ([]Path, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	return g.findPaths(source, target, maxDepth, limit)
}

func (g Graph) findPaths(source, target string, maxDepth, limit int) ([]Path, error) {
	if maxDepth < 1 || maxDepth > 32 || limit < 1 || limit > 1000 {
		return nil, errors.New("max depth must be 1–32 and limit must be 1–1000")
	}
	assets := make(map[string]Asset, len(g.Assets))
	for _, asset := range g.Assets {
		assets[asset.ID] = asset
	}
	if _, exists := assets[source]; !exists {
		return nil, fmt.Errorf("unknown source asset %q", source)
	}
	if target != "" {
		if _, exists := assets[target]; !exists {
			return nil, fmt.Errorf("unknown target asset %q", target)
		}
	}
	adjacency := make(map[string][]Edge)
	for _, edge := range g.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge)
	}
	for id := range adjacency {
		sort.Slice(adjacency[id], func(i, j int) bool { return adjacency[id][i].ID < adjacency[id][j].ID })
	}
	paths := make([]Path, 0)
	visited := map[string]bool{source: true}
	chainAssets := []Asset{assets[source]}
	chainEdges := make([]Edge, 0, maxDepth)
	var walk func(string)
	walk = func(at string) {
		if len(paths) >= limit {
			return
		}
		if at != source && (target == "" || at == target) {
			impact := 0
			for _, asset := range chainAssets {
				if asset.Impact > impact {
					impact = asset.Impact
				}
			}
			confidence := 100
			for _, edge := range chainEdges {
				confidence = confidence * edge.Confidence / 100
			}
			score := impact * confidence
			paths = append(paths, Path{Assets: append([]Asset(nil), chainAssets...), Edges: append([]Edge(nil), chainEdges...), Impact: impact, Confidence: confidence, Score: score, Explanation: fmt.Sprintf("impact %d/5 × confidence %d%% = %d", impact, confidence, score)})
			if target != "" {
				return
			}
		}
		if len(chainEdges) == maxDepth {
			return
		}
		for _, edge := range adjacency[at] {
			if visited[edge.To] {
				continue
			}
			visited[edge.To] = true
			chainEdges = append(chainEdges, edge)
			chainAssets = append(chainAssets, assets[edge.To])
			walk(edge.To)
			chainAssets = chainAssets[:len(chainAssets)-1]
			chainEdges = chainEdges[:len(chainEdges)-1]
			delete(visited, edge.To)
		}
	}
	walk(source)
	sort.SliceStable(paths, func(i, j int) bool {
		if paths[i].Score == paths[j].Score {
			return paths[i].Confidence > paths[j].Confidence
		}
		return paths[i].Score > paths[j].Score
	})
	return paths, nil
}
