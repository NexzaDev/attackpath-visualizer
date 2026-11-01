package domain

import (
	"container/heap"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/NexzaDev/attackpath-visualizer/internal/scoring"
)

const (
	maxPathDepth = 32
	maxPathCount = 1000
)

type Asset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Impact      int    `json:"impact"`
	Description string `json:"description"`
}

type Identity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Impact      int    `json:"impact"`
	Description string `json:"description"`
}

type Misconfiguration struct {
	ID          string `json:"id"`
	AssetID     string `json:"asset_id"`
	IdentityID  string `json:"identity_id,omitempty"`
	Title       string `json:"title"`
	Severity    int    `json:"severity"`
	Description string `json:"description"`
}

type Finding struct {
	ID                 string `json:"id"`
	AssetID            string `json:"asset_id"`
	IdentityID         string `json:"identity_id,omitempty"`
	MisconfigurationID string `json:"misconfiguration_id,omitempty"`
	Title              string `json:"title"`
	Severity           int    `json:"severity"`
	Source             string `json:"source"`
	Observed           string `json:"observed"`
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
	Assets            []Asset            `json:"assets"`
	Identities        []Identity         `json:"identities,omitempty"`
	Misconfigurations []Misconfiguration `json:"misconfigurations,omitempty"`
	Findings          []Finding          `json:"findings,omitempty"`
	Edges             []Edge             `json:"edges"`
}

type Node struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	EntityType  string `json:"entity_type"`
	Kind        string `json:"kind"`
	Impact      int    `json:"impact"`
	Description string `json:"description"`
}

type ScoreBreakdown struct {
	ImpactTier           int   `json:"impact_tier"`
	EdgeConfidences      []int `json:"edge_confidences"`
	CompoundedConfidence int   `json:"compounded_confidence"`
	Score                int   `json:"score"`
}

type Path struct {
	Nodes       []Node         `json:"nodes"`
	Assets      []Asset        `json:"assets,omitempty"`
	Edges       []Edge         `json:"edges"`
	Impact      int            `json:"impact"`
	Confidence  int            `json:"confidence"`
	Score       int            `json:"score"`
	Components  ScoreBreakdown `json:"components"`
	Explanation string         `json:"explanation"`
}

type PathResult struct {
	Paths     []Path `json:"paths"`
	Truncated bool   `json:"truncated"`
	Expanded  int    `json:"expanded"`
}

func (g Graph) Validate() error {
	nodes := make(map[string]Node, len(g.Assets)+len(g.Identities))
	for _, asset := range g.Assets {
		if err := validateNode(asset.ID, asset.Name, asset.Kind, asset.Impact); err != nil {
			return fmt.Errorf("asset %q: %w", asset.ID, err)
		}
		if err := addNode(nodes, Node{ID: asset.ID, Name: asset.Name, EntityType: "asset", Kind: asset.Kind, Impact: asset.Impact, Description: asset.Description}); err != nil {
			return fmt.Errorf("asset %q: %w", asset.ID, err)
		}
	}
	for _, identity := range g.Identities {
		if err := validateNode(identity.ID, identity.Name, identity.Kind, identity.Impact); err != nil {
			return fmt.Errorf("identity %q: %w", identity.ID, err)
		}
		if err := addNode(nodes, Node{ID: identity.ID, Name: identity.Name, EntityType: "identity", Kind: identity.Kind, Impact: identity.Impact, Description: identity.Description}); err != nil {
			return fmt.Errorf("identity %q: %w", identity.ID, err)
		}
	}
	ids := make(map[string]string, len(nodes)+len(g.Misconfigurations)+len(g.Findings)+len(g.Edges))
	for id := range nodes {
		ids[id] = "node"
	}
	misconfigurations := make(map[string]Misconfiguration, len(g.Misconfigurations))
	for _, item := range g.Misconfigurations {
		if err := validateID(item.ID); err != nil || !validText(item.Title) || !validText(item.Description) || item.Severity < 1 || item.Severity > 5 {
			return fmt.Errorf("misconfiguration %q requires an id, title, description, and severity from 1 to 5", item.ID)
		}
		if err := addID(ids, item.ID, "misconfiguration"); err != nil {
			return err
		}
		if _, ok := nodeKind(nodes, item.AssetID, "asset"); !ok {
			return fmt.Errorf("misconfiguration %q references missing asset %q", item.ID, item.AssetID)
		}
		if item.IdentityID != "" {
			if _, ok := nodeKind(nodes, item.IdentityID, "identity"); !ok {
				return fmt.Errorf("misconfiguration %q references missing identity %q", item.ID, item.IdentityID)
			}
		}
		misconfigurations[item.ID] = item
	}
	for _, item := range g.Findings {
		if err := validateID(item.ID); err != nil || !validText(item.Title) || item.Severity < 1 || item.Severity > 5 || !validText(item.Source) || !validText(item.Observed) {
			return fmt.Errorf("finding %q requires an id, title, source, observed value, and severity from 1 to 5", item.ID)
		}
		if err := addID(ids, item.ID, "finding"); err != nil {
			return err
		}
		if _, ok := nodeKind(nodes, item.AssetID, "asset"); !ok {
			return fmt.Errorf("finding %q references missing asset %q", item.ID, item.AssetID)
		}
		if item.IdentityID != "" {
			if _, ok := nodeKind(nodes, item.IdentityID, "identity"); !ok {
				return fmt.Errorf("finding %q references missing identity %q", item.ID, item.IdentityID)
			}
		}
		if item.MisconfigurationID != "" {
			if _, ok := misconfigurations[item.MisconfigurationID]; !ok {
				return fmt.Errorf("finding %q references missing misconfiguration %q", item.ID, item.MisconfigurationID)
			}
		}
	}
	for _, edge := range g.Edges {
		if err := validateID(edge.ID); err != nil || !validText(edge.Relation) || edge.Confidence < 0 || edge.Confidence > 100 {
			return fmt.Errorf("edge %q requires an id, relation, and confidence from 0 to 100", edge.ID)
		}
		if err := addID(ids, edge.ID, "edge"); err != nil {
			return err
		}
		if _, exists := nodes[edge.From]; !exists {
			return fmt.Errorf("edge %q references missing source node %q", edge.ID, edge.From)
		}
		if _, exists := nodes[edge.To]; !exists {
			return fmt.Errorf("edge %q references missing target node %q", edge.ID, edge.To)
		}
		if edge.From == edge.To {
			return fmt.Errorf("edge %q cannot connect a node to itself", edge.ID)
		}
		if !validText(edge.Evidence.Source) || !validText(edge.Evidence.Finding) || !validText(edge.Evidence.Observed) {
			return fmt.Errorf("edge %q requires source, finding, and observed evidence", edge.ID)
		}
	}
	return nil
}

func (g Graph) AllPaths(maxDepth, limit int) ([]Path, error) {
	result, err := g.EnumeratePaths(maxDepth, limit)
	return result.Paths, err
}

func (g Graph) EnumeratePaths(maxDepth, limit int) (PathResult, error) {
	if maxDepth < 1 || maxDepth > maxPathDepth || limit < 1 || limit > maxPathCount {
		return PathResult{}, errors.New("max depth must be 1–32 and limit must be 1–1000")
	}
	if err := g.Validate(); err != nil {
		return PathResult{}, err
	}
	nodes := g.nodes()
	adjacency := g.adjacency()
	incoming := make(map[string]bool, len(nodes))
	neighbors := make(map[string][]string, len(nodes))
	for _, edge := range g.Edges {
		incoming[edge.To] = true
		neighbors[edge.From] = append(neighbors[edge.From], edge.To)
		neighbors[edge.To] = append(neighbors[edge.To], edge.From)
	}
	ids := make([]string, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	sources := make([]string, 0, len(ids))
	componentVisited := make(map[string]bool, len(ids))
	for _, first := range ids {
		if componentVisited[first] {
			continue
		}
		component := []string{first}
		componentVisited[first] = true
		for index := 0; index < len(component); index++ {
			for _, neighbor := range neighbors[component[index]] {
				if !componentVisited[neighbor] {
					componentVisited[neighbor] = true
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
	candidates := &pathHeap{}
	heap.Init(candidates)
	seen := make(map[string]struct{})
	result := PathResult{}
	for _, source := range sources {
		if err := g.findPaths(nodes, adjacency, source, "", maxDepth, limit, &result.Expanded, &result.Truncated, seen, candidates); err != nil {
			return PathResult{}, err
		}
	}
	result.Paths = append([]Path(nil), (*candidates)...)
	sortPaths(result.Paths)
	return result, nil
}

func (g Graph) FindPaths(source, target string, maxDepth, limit int) ([]Path, error) {
	if maxDepth < 1 || maxDepth > maxPathDepth || limit < 1 || limit > maxPathCount {
		return nil, errors.New("max depth must be 1–32 and limit must be 1–1000")
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	nodes := g.nodes()
	if _, exists := nodes[source]; !exists {
		return nil, fmt.Errorf("unknown source node %q", source)
	}
	if target != "" {
		if _, exists := nodes[target]; !exists {
			return nil, fmt.Errorf("unknown target node %q", target)
		}
	}
	candidates := &pathHeap{}
	heap.Init(candidates)
	expanded := 0
	truncated := false
	err := g.findPaths(nodes, g.adjacency(), source, target, maxDepth, limit, &expanded, &truncated, make(map[string]struct{}), candidates)
	paths := append([]Path(nil), (*candidates)...)
	sortPaths(paths)
	return paths, err
}

func (g Graph) adjacency() map[string][]Edge {
	adjacency := make(map[string][]Edge)
	for _, edge := range g.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge)
	}
	for id := range adjacency {
		sort.Slice(adjacency[id], func(i, j int) bool { return adjacency[id][i].ID < adjacency[id][j].ID })
	}
	return adjacency
}

func (g Graph) findPaths(nodes map[string]Node, adjacency map[string][]Edge, source, target string, maxDepth, limit int, expanded *int, truncated *bool, seen map[string]struct{}, candidates *pathHeap) error {
	visited := map[string]bool{source: true}
	chainNodes := []Node{nodes[source]}
	chainEdges := make([]Edge, 0, maxDepth)
	var walk func(string) error
	walk = func(at string) error {
		if at != source && (target == "" || at == target) {
			signature := pathSignature(chainEdges)
			if _, duplicate := seen[signature]; !duplicate {
				seen[signature] = struct{}{}
				path, err := makePath(chainNodes, chainEdges)
				if err != nil {
					return err
				}
				if candidates.Len() == limit {
					*truncated = true
				}
				candidates.add(path, limit)
			}
			if target != "" {
				return nil
			}
		}
		if len(chainEdges) == maxDepth {
			return nil
		}
		for _, edge := range adjacency[at] {
			if visited[edge.To] {
				continue
			}
			*expanded = *expanded + 1
			if *expanded > 250000 {
				return errors.New("path enumeration exceeded the 250000 edge-expansion budget; reduce graph size or path depth")
			}
			visited[edge.To] = true
			chainEdges = append(chainEdges, edge)
			chainNodes = append(chainNodes, nodes[edge.To])
			if err := walk(edge.To); err != nil {
				return err
			}
			chainNodes = chainNodes[:len(chainNodes)-1]
			chainEdges = chainEdges[:len(chainEdges)-1]
			delete(visited, edge.To)
		}
		return nil
	}
	return walk(source)
}

func makePath(nodes []Node, edges []Edge) (Path, error) {
	impact := 1
	confidences := make([]int, len(edges))
	assets := make([]Asset, 0, len(nodes))
	for _, node := range nodes {
		if node.Impact > impact {
			impact = node.Impact
		}
		if node.EntityType == "asset" {
			assets = append(assets, Asset{ID: node.ID, Name: node.Name, Kind: node.Kind, Impact: node.Impact, Description: node.Description})
		}
	}
	for index, edge := range edges {
		confidences[index] = edge.Confidence
	}
	result, err := scoring.Score(impact, confidences)
	if err != nil {
		return Path{}, err
	}
	return Path{Nodes: append([]Node(nil), nodes...), Assets: assets, Edges: append([]Edge(nil), edges...), Impact: impact, Confidence: result.CompoundedConfidence, Score: result.Score, Components: ScoreBreakdown{ImpactTier: result.ImpactTier, EdgeConfidences: result.EdgeConfidences, CompoundedConfidence: result.CompoundedConfidence, Score: result.Score}, Explanation: result.Explanation}, nil
}

func (g Graph) nodes() map[string]Node {
	nodes := make(map[string]Node, len(g.Assets)+len(g.Identities))
	for _, asset := range g.Assets {
		nodes[asset.ID] = Node{ID: asset.ID, Name: asset.Name, EntityType: "asset", Kind: asset.Kind, Impact: asset.Impact, Description: asset.Description}
	}
	for _, identity := range g.Identities {
		nodes[identity.ID] = Node{ID: identity.ID, Name: identity.Name, EntityType: "identity", Kind: identity.Kind, Impact: identity.Impact, Description: identity.Description}
	}
	return nodes
}

func pathSignature(edges []Edge) string {
	ids := make([]string, len(edges))
	for index, edge := range edges {
		ids[index] = edge.ID
	}
	return strings.Join(ids, "\x00")
}

func sortPaths(paths []Path) {
	sort.SliceStable(paths, func(i, j int) bool { return pathBetter(paths[i], paths[j]) })
}

func pathBetter(left, right Path) bool {
	if left.Score != right.Score {
		return left.Score > right.Score
	}
	if left.Confidence != right.Confidence {
		return left.Confidence > right.Confidence
	}
	return pathSignature(left.Edges) < pathSignature(right.Edges)
}

func validateNode(id, name, kind string, impact int) error {
	if err := validateID(id); err != nil || !validText(name) || !validText(kind) || impact < 1 || impact > 5 {
		return errors.New("requires an id, name, kind, and impact from 1 to 5")
	}
	return nil
}

func validateID(value string) error {
	if !validText(value) || len(value) > 256 {
		return errors.New("id must contain 1 to 256 non-control characters")
	}
	return nil
}

func validText(value string) bool {
	if strings.TrimSpace(value) == "" || len(value) > 4096 {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return false
		}
	}
	return true
}

func addNode(nodes map[string]Node, node Node) error {
	if _, exists := nodes[node.ID]; exists {
		return fmt.Errorf("duplicate node id %q", node.ID)
	}
	nodes[node.ID] = node
	return nil
}

func addID(ids map[string]string, id, kind string) error {
	if previous, exists := ids[id]; exists {
		return fmt.Errorf("duplicate id %q used by %s and %s", id, previous, kind)
	}
	ids[id] = kind
	return nil
}

func nodeKind(nodes map[string]Node, id, kind string) (Node, bool) {
	node, exists := nodes[id]
	if !exists {
		return Node{}, false
	}
	if kind == "asset" && node.EntityType != "asset" {
		return Node{}, false
	}
	if kind == "identity" && node.EntityType != "identity" {
		return Node{}, false
	}
	return node, true
}
