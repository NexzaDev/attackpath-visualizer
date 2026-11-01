//go:build !nogui

package gui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/NexzaDev/attackpath-visualizer/assets"
	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
	"github.com/NexzaDev/attackpath-visualizer/internal/launcher"
	"github.com/NexzaDev/attackpath-visualizer/internal/meta"
	"github.com/NexzaDev/attackpath-visualizer/internal/notifier"
	"github.com/NexzaDev/attackpath-visualizer/internal/remediation"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
	"github.com/NexzaDev/attackpath-visualizer/internal/service"
)

type Category struct {
	ID    string
	Label string
}

var Categories = []Category{
	{ID: "overview", Label: "Overview"},
	{ID: "graph", Label: "Graph"},
	{ID: "paths", Label: "Paths"},
	{ID: "assets", Label: "Assets"},
	{ID: "evidence", Label: "Evidence"},
	{ID: "remediation", Label: "Remediation"},
	{ID: "imports", Label: "Imports"},
	{ID: "reports", Label: "Reports"},
	{ID: "notifications", Label: "Notifications"},
	{ID: "about", Label: "About"},
	{ID: "settings", Label: "Settings"},
}

type Dashboard struct {
	app              fyne.App
	window           fyne.Window
	store            *repository.Store
	service          service.Service
	graph            domain.Graph
	paths            []domain.Path
	summary          service.Summary
	impactFilter     int
	confidenceFilter int
	history          []map[string]any
	notices          []repository.Notification
	search           string
	current          string
	collapsed        bool
	zoom             float32
	demo             bool
	switchMode       func(string) error
	content          *fyne.Container
	modeLabel        *widget.Label
	searchBox        *widget.Entry
}

func NewDashboard(app fyne.App, window fyne.Window, store *repository.Store, switchMode func(string) error) (*Dashboard, error) {
	dashboard := &Dashboard{app: app, window: window, store: store, service: service.Service{Store: store}, current: "overview", switchMode: switchMode}
	if err := dashboard.refresh(); err != nil {
		return nil, err
	}
	dashboard.collapsed = app.Preferences().BoolWithFallback("sidebar.collapsed", false)
	dashboard.zoom = float32(app.Preferences().FloatWithFallback("graph.zoom", 1))
	if dashboard.zoom < 0.5 || dashboard.zoom > 2 {
		dashboard.zoom = 1
	}
	return dashboard, nil
}

func (d *Dashboard) refresh() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	history, err := d.store.ImportHistory(ctx)
	if err != nil {
		return err
	}
	summary, err := d.service.Overview(ctx)
	if err != nil {
		return err
	}
	if len(summary.Graph.Assets)+len(summary.Graph.Identities) == 0 && len(history) == 0 {
		if err = d.service.LoadDemo(ctx); err != nil {
			return err
		}
		summary, err = d.service.Overview(ctx)
		if err != nil {
			return err
		}
		history, err = d.store.ImportHistory(ctx)
		if err != nil {
			return err
		}
	}
	notices, err := d.store.Notifications(ctx, 200)
	if err != nil {
		return err
	}
	d.summary = summary
	d.graph = summary.Graph
	d.paths = summary.Paths.Paths
	d.history = history
	d.notices = notices
	d.demo = len(history) > 0 && history[0]["source"] == "demo fixture"
	return nil
}

func (d *Dashboard) Content() fyne.CanvasObject {
	return d.shell()
}

func (d *Dashboard) Screen(id string) fyne.CanvasObject {
	for _, category := range Categories {
		if category.ID == id {
			return d.screen(category.ID)
		}
	}
	return d.emptyState("Screen unavailable", "Choose a category from the navigation.")
}

func (d *Dashboard) shell() fyne.CanvasObject {
	d.content = container.NewMax(d.screen(d.current))
	nav := d.navigation()
	top := d.topbar()
	return container.NewBorder(top, nil, nav, nil, d.content)
}

func (d *Dashboard) topbar() fyne.CanvasObject {
	d.searchBox = widget.NewEntry()
	d.searchBox.SetPlaceHolder("Search assets, paths and evidence")
	d.searchBox.OnChanged = func(value string) {
		d.search = strings.ToLower(strings.TrimSpace(value))
		d.renderCurrent()
	}
	d.modeLabel = widget.NewLabel("LOCAL DATA")
	if d.demo {
		d.modeLabel.SetText("DEMO DATA")
	}
	unread, _ := d.store.UnreadCount(context.Background())
	bell, _ := icon("notifications")
	notificationsButton := widget.NewButtonWithIcon(fmt.Sprintf("Notifications · %d", unread), bell, func() { d.show("notifications") })
	return container.NewBorder(nil, widget.NewSeparator(), nil, container.NewHBox(d.modeLabel, notificationsButton), container.NewPadded(d.searchBox))
}

func (d *Dashboard) navigation() fyne.CanvasObject {
	items := make([]fyne.CanvasObject, 0, len(Categories)+3)
	for i, category := range Categories {
		if i == 9 {
			items = append(items, widget.NewSeparator())
		}
		resource, _ := icon(category.ID)
		label := category.Label
		if d.collapsed {
			label = ""
		}
		button := widget.NewButtonWithIcon(label, resource, func(id string) func() { return func() { d.show(id) } }(category.ID))
		if category.ID == d.current {
			button.Importance = widget.HighImportance
		}
		items = append(items, button)
	}
	collapseLabel := "Collapse sidebar"
	if d.collapsed {
		collapseLabel = "Expand sidebar"
	}
	collapse := widget.NewButton(collapseLabel, func() {
		d.collapsed = !d.collapsed
		d.app.Preferences().SetBool("sidebar.collapsed", d.collapsed)
		d.window.SetContent(d.shell())
	})
	items = append(items, collapse)
	logoBytes, _ := assets.Files.ReadFile("icons/icon.png")
	logo := canvas.NewImageFromImage(decodePNG(logoBytes))
	logo.SetMinSize(fyne.NewSize(34, 34))
	brand := container.NewHBox(logo, widget.NewLabelWithStyle("ATTACKPATH", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	if d.collapsed {
		brand = container.NewHBox(logo)
	}
	return container.NewBorder(container.NewPadded(brand), container.NewPadded(collapse), nil, nil, container.NewVScroll(container.NewVBox(items...)))
}

func (d *Dashboard) renderCurrent() {
	if d.content == nil {
		return
	}
	d.content.Objects = []fyne.CanvasObject{d.screen(d.current)}
	d.content.Refresh()
}

func (d *Dashboard) show(id string) {
	d.current = id
	if d.window != nil {
		d.window.SetContent(d.shell())
		return
	}
	if d.content != nil {
		d.content.Objects = []fyne.CanvasObject{d.screen(id)}
		d.content.Refresh()
	}
}

func (d *Dashboard) screen(id string) fyne.CanvasObject {
	switch id {
	case "overview":
		return d.overview()
	case "graph":
		return d.graphScreen()
	case "paths":
		return d.pathsScreen()
	case "assets":
		return d.assetsScreen()
	case "evidence":
		return d.evidenceScreen()
	case "remediation":
		return d.remediationScreen()
	case "imports":
		return d.importsScreen()
	case "reports":
		return d.reportsScreen()
	case "notifications":
		return d.notificationsScreen()
	case "about":
		return d.aboutScreen()
	case "settings":
		return d.settingsScreen()
	default:
		return d.emptyState("Screen unavailable", "Choose a category from the navigation.")
	}
}

func (d *Dashboard) heading(title, subtitle string) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), widget.NewLabel(subtitle), widget.NewSeparator())
}

func (d *Dashboard) overview() fyne.CanvasObject {
	critical := d.summary.PathsByImpact[3] + d.summary.PathsByImpact[4]
	latest := "No imports recorded"
	if len(d.history) > 0 {
		latest = fmt.Sprintf("%v · %v assets · %v identities · %v relationships", d.history[0]["imported_at"], d.history[0]["assets"], d.history[0]["identities"], d.history[0]["edges"])
	}
	metrics := container.NewGridWithColumns(5,
		metric("Assets", fmt.Sprint(len(d.graph.Assets)), "Stored inventory"),
		metric("Identities", fmt.Sprint(len(d.graph.Identities)), "Stored principals"),
		metric("Relationships", fmt.Sprint(len(d.graph.Edges)), "Evidence-linked edges"),
		metric("Potential paths", fmt.Sprint(len(d.paths)), "Ranked simple paths"),
		metric("High impact", fmt.Sprint(critical), "Path impact tier 4–5"),
	)
	impactTiles := make([]fyne.CanvasObject, 5)
	for tier := 1; tier <= 5; tier++ {
		tier := tier
		impactTiles[tier-1] = widget.NewButton(fmt.Sprintf("Impact %d/5 · %d paths", tier, d.summary.PathsByImpact[tier-1]), func() {
			d.impactFilter, d.confidenceFilter = tier, 0
			d.show("paths")
		})
	}
	confidenceTiles := make([]fyne.CanvasObject, len(d.summary.Confidence))
	for index, band := range d.summary.Confidence {
		index, band := index, band
		confidenceTiles[index] = widget.NewButton(fmt.Sprintf("%s · %d paths", band.Label, band.Count), func() {
			d.confidenceFilter, d.impactFilter = index+1, 0
			d.show("paths")
		})
	}
	rows := make([][]string, 0, len(d.paths))
	visiblePaths := make([]domain.Path, 0, len(d.paths))
	for _, path := range d.paths {
		if len(rows) == 8 {
			break
		}
		if !d.matchesPath(path) {
			continue
		}
		visiblePaths = append(visiblePaths, path)
		rows = append(rows, []string{pathLabel(path), fmt.Sprintf("%d", path.Score), fmt.Sprintf("%d%%", path.Confidence), fmt.Sprintf("Impact %d/5", path.Impact)})
	}
	table := stringTable([]string{"Potential path", "Score", "Confidence", "Impact"}, rows, []float32{340, 70, 100, 100}, func(row int) {
		if row >= 0 && row < len(visiblePaths) {
			d.showPathDetails(visiblePaths[row])
		}
	})
	freshness := widget.NewCard("Import freshness", "Latest recorded input", widget.NewLabel(latest))
	choke := widget.NewCard("Relationship choke points", "Suggestions rank paths affected then risk points", d.chokePoints())
	confidence := widget.NewCard("Confidence distribution", "Enumerated paths by compounded confidence", container.NewGridWithColumns(5, confidenceTiles...))
	impact := widget.NewCard("Paths by impact tier", "Select a tier to filter ranked paths", container.NewGridWithColumns(5, impactTiles...))
	lower := container.NewGridWithColumns(2, freshness, choke)
	body := container.NewVBox(d.heading("Exposure overview", "Attention summary computed from the current local graph."), metrics, impact, confidence, container.NewHBox(widget.NewLabelWithStyle("Highest-ranked paths", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), widget.NewButton("View all paths", func() { d.impactFilter, d.confidenceFilter = 0, 0; d.show("paths") })), table, lower)
	return container.NewVScroll(container.NewPadded(body))
}

func metric(title, value, note string) fyne.CanvasObject {
	return widget.NewCard(title, note, widget.NewLabelWithStyle(value, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
}

func (d *Dashboard) graphScreen() fyne.CanvasObject {
	nodes := d.filteredNodes()
	selected := widget.NewLabel("Select a node to focus its direct relationships.")
	list := widget.NewList(func() int { return len(nodes) }, func() fyne.CanvasObject { return widget.NewLabel("template") }, func(id widget.ListItemID, object fyne.CanvasObject) {
		node := nodes[id]
		object.(*widget.Label).SetText(fmt.Sprintf("%s · %s/%s · impact %d/5", node.Name, node.EntityType, node.Kind, node.Impact))
	})
	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(nodes) {
			return
		}
		selected.SetText(d.nodeDetails(nodes[id]))
	}
	zoomLabel := widget.NewLabel(fmt.Sprintf("%d%%", int(d.zoom*100)))
	zoom := func(delta float32) {
		d.zoom += delta
		if d.zoom < 0.5 {
			d.zoom = 0.5
		}
		if d.zoom > 2 {
			d.zoom = 2
		}
		d.app.Preferences().SetFloat("graph.zoom", float64(d.zoom))
		d.renderCurrent()
	}
	zoomControls := container.NewHBox(widget.NewButton("−", func() { zoom(-0.1) }), zoomLabel, widget.NewButton("+", func() { zoom(0.1) }))
	mapPage := container.NewBorder(widget.NewLabel(fmt.Sprintf("Relationship map · %d visible nodes of %d · scroll to pan", min(len(nodes), 160), len(d.graph.Nodes()))), nil, nil, nil, container.NewScroll(d.graphCanvas(nodes)))
	mapPage = container.NewBorder(zoomControls, nil, nil, nil, mapPage)
	listPage := container.NewBorder(widget.NewLabel(fmt.Sprintf("Node inventory · %d assets and identities · global search filters this list", len(nodes))), nil, nil, nil, list)
	edges := d.filteredEdges()
	edgeList := widget.NewList(func() int { return len(edges) }, func() fyne.CanvasObject { return widget.NewLabel("template") }, func(id widget.ListItemID, object fyne.CanvasObject) {
		edge := edges[id]
		object.(*widget.Label).SetText(fmt.Sprintf("%s → %s · %s · confidence %d%%", edge.From, edge.To, edge.Relation, edge.Confidence))
	})
	edgeList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(edges) {
			edge := edges[id]
			dialog.ShowInformation("Relationship evidence", fmt.Sprintf("%s → %s\nRelationship: %s\nConfidence: %d%%\nSource: %s\nFinding: %s\nObserved: %s", edge.From, edge.To, edge.Relation, edge.Confidence, edge.Evidence.Source, edge.Evidence.Finding, edge.Evidence.Observed), d.window)
		}
	}
	edgePage := container.NewBorder(widget.NewLabel(fmt.Sprintf("All %d matching relationships · select a row for evidence", len(edges))), nil, nil, nil, edgeList)
	tabs := container.NewAppTabs(container.NewTabItem("Graph canvas", mapPage), container.NewTabItem("Accessible node list", listPage), container.NewTabItem("Edges & evidence", edgePage))
	if len(nodes) == 0 {
		return d.emptyState("No graph records", "Import a JSON inventory or restore a demo database to populate the graph.")
	}
	return container.NewPadded(container.NewBorder(d.heading("Relationship graph", "Pan by scrolling · zoom with + / − · select a node to focus relationships."), nil, nil, nil, container.NewGridWithColumns(2, tabs, container.NewVScroll(selected))))
}

func (d *Dashboard) graphCanvas(nodes []domain.Node) fyne.CanvasObject {
	const maxVisible = 160
	visible := nodes
	if len(visible) > maxVisible {
		visible = visible[:maxVisible]
	}
	const columns = 5
	baseWidth := float32(210) * d.zoom
	baseHeight := float32(80) * d.zoom
	gapX := float32(42) * d.zoom
	gapY := float32(66) * d.zoom
	mapWidth := float32(columns)*(baseWidth+gapX) + 36
	rows := (len(visible) + columns - 1) / columns
	mapHeight := float32(rows)*(baseHeight+gapY) + 40
	background := canvas.NewRectangle(color.NRGBA{R: 248, G: 250, B: 253, A: 255})
	if d.app != nil && d.app.Preferences().BoolWithFallback("appearance.dark", false) {
		background.FillColor = color.NRGBA{R: 20, G: 31, B: 47, A: 255}
	}
	background.SetMinSize(fyne.NewSize(mapWidth, mapHeight))
	background.Resize(fyne.NewSize(mapWidth, mapHeight))
	objects := []fyne.CanvasObject{background}
	positions := make(map[string]fyne.Position, len(visible))
	for index, node := range visible {
		column, row := index%columns, index/columns
		position := fyne.NewPos(18+float32(column)*(baseWidth+gapX), 18+float32(row)*(baseHeight+gapY))
		positions[node.ID] = position
	}
	for _, edge := range d.graph.Edges {
		from, fromOK := positions[edge.From]
		to, toOK := positions[edge.To]
		if !fromOK || !toOK {
			continue
		}
		line := canvas.NewLine(color.NRGBA{R: 105, G: 128, B: 157, A: 255})
		line.StrokeWidth = 1.5 * d.zoom
		line.Position1 = fyne.NewPos(from.X+baseWidth, from.Y+baseHeight/2)
		line.Position2 = fyne.NewPos(to.X, to.Y+baseHeight/2)
		objects = append(objects, line)
	}
	for index, node := range visible {
		position := positions[node.ID]
		status := fmt.Sprintf("Impact %d/5 · %s/%s", node.Impact, node.EntityType, node.Kind)
		button := widget.NewButton(node.Name+"\n"+status, func(record domain.Node) func() {
			return func() { dialog.ShowInformation(record.Name, d.nodeDetails(record), d.window) }
		}(node))
		button.Resize(fyne.NewSize(baseWidth, baseHeight))
		button.Move(position)
		if index%2 == 0 {
			button.Importance = widget.MediumImportance
		}
		objects = append(objects, button)
	}
	return container.NewWithoutLayout(objects...)
}

func (d *Dashboard) nodeDetails(node domain.Node) string {
	lines := []string{fmt.Sprintf("ID: %s", node.ID), fmt.Sprintf("Entity: %s", node.EntityType), fmt.Sprintf("Type: %s", node.Kind), fmt.Sprintf("Impact: %d/5", node.Impact), node.Description, "", "Direct relationships:"}
	related := 0
	for _, edge := range d.graph.Edges {
		if edge.From == node.ID || edge.To == node.ID {
			lines = append(lines, fmt.Sprintf("%s — %s → %s · %s · %d%%", edge.ID, edge.From, edge.To, edge.Relation, edge.Confidence))
			related++
		}
	}
	for _, item := range d.graph.Misconfigurations {
		if item.AssetID == node.ID || item.IdentityID == node.ID {
			lines = append(lines, fmt.Sprintf("Misconfiguration · severity %d · %s", item.Severity, item.Title))
			related++
		}
	}
	for _, item := range d.graph.Findings {
		if item.AssetID == node.ID || item.IdentityID == node.ID {
			lines = append(lines, fmt.Sprintf("Finding · severity %d · %s · %s", item.Severity, item.Title, item.Source))
			related++
		}
	}
	if related == 0 {
		lines = append(lines, "No direct relationships or findings are recorded.")
	}
	return strings.Join(lines, "\n")
}

func (d *Dashboard) pathsScreen() fyne.CanvasObject {
	rows := make([][]string, 0, len(d.paths))
	visible := make([]domain.Path, 0, len(d.paths))
	for _, path := range d.paths {
		if d.matchesPath(path) {
			visible = append(visible, path)
			rows = append(rows, []string{pathLabel(path), fmt.Sprint(path.Score), fmt.Sprintf("%d%%", path.Confidence), path.Explanation})
		}
	}
	table := stringTable([]string{"Directed path", "Score", "Confidence", "Score breakdown"}, rows, []float32{340, 70, 90, 320}, func(row int) {
		if row >= 0 && row < len(visible) {
			d.showPathDetails(visible[row])
		}
	})
	subtitle := "Impact tier × compounded edge confidence · select a row for exact components and evidence"
	if d.summary.Paths.Truncated {
		subtitle += " · result capped; rankings cover the bounded path set"
	}
	heading := []fyne.CanvasObject{d.heading("Ranked potential paths", subtitle)}
	if d.impactFilter > 0 || d.confidenceFilter > 0 {
		filterLabel := "Filtered path results"
		if d.impactFilter > 0 {
			filterLabel = fmt.Sprintf("Impact tier %d/5", d.impactFilter)
		}
		if d.confidenceFilter > 0 && d.confidenceFilter <= len(d.summary.Confidence) {
			filterLabel = d.summary.Confidence[d.confidenceFilter-1].Label + " confidence"
		}
		heading = append(heading, container.NewHBox(widget.NewLabel("Filter: "+filterLabel), widget.NewButton("Clear filter", func() { d.impactFilter, d.confidenceFilter = 0, 0; d.show("paths") })))
	}
	return container.NewPadded(container.NewBorder(container.NewVBox(heading...), nil, nil, nil, table))
}

func (d *Dashboard) assetsScreen() fyne.CanvasObject {
	nodes := d.filteredNodes()
	rows := make([][]string, len(nodes))
	for i, node := range nodes {
		rows[i] = []string{node.Name, node.EntityType, node.Kind, fmt.Sprintf("%d/5", node.Impact), node.Description}
	}
	table := stringTable([]string{"Node", "Entity", "Type", "Impact", "Description"}, rows, []float32{210, 80, 120, 70, 300}, func(row int) {
		if row >= 0 && row < len(nodes) {
			dialog.ShowInformation(nodes[row].Name, d.nodeDetails(nodes[row]), d.window)
		}
	})
	return container.NewPadded(container.NewBorder(d.heading("Assets and identities", "Inventory records and linked relationships from the current local graph."), nil, nil, nil, table))
}

func (d *Dashboard) evidenceScreen() fyne.CanvasObject {
	edges := d.filteredEdges()
	rows := make([][]string, len(edges))
	for i, edge := range edges {
		rows[i] = []string{edge.From + " → " + edge.To, edge.Relation, edge.Evidence.Source, edge.Evidence.Finding}
	}
	table := stringTable([]string{"Relationship", "Type", "Source", "Finding"}, rows, []float32{220, 150, 190, 340}, func(row int) {
		if row >= 0 && row < len(edges) {
			edge := edges[row]
			dialog.ShowInformation("Relationship evidence", fmt.Sprintf("Edge: %s\n%s → %s\nRelationship: %s\nConfidence: %d%%\nSource: %s\nFinding: %s\nObserved: %s", edge.ID, edge.From, edge.To, edge.Relation, edge.Confidence, edge.Evidence.Source, edge.Evidence.Finding, edge.Evidence.Observed), d.window)
		}
	})
	return container.NewPadded(container.NewBorder(d.heading("Evidence", "Source records attached to every relationship."), nil, nil, nil, table))
}

func (d *Dashboard) remediationScreen() fyne.CanvasObject {
	result := d.summary.Remediation
	visible := make([]remediation.Suggestion, 0, len(result.Suggestions))
	rows := make([][]string, 0, len(result.Suggestions))
	for _, item := range result.Suggestions {
		if d.search != "" && !strings.Contains(strings.ToLower(item.Relation+" "+item.EdgeID+" "+item.From+" "+item.To), d.search) {
			continue
		}
		visible = append(visible, item)
		rows = append(rows, []string{item.From + " → " + item.To, item.Relation, fmt.Sprint(item.PathsAffected), fmt.Sprint(item.CriticalPaths), fmt.Sprint(item.WeightedRiskPoints), item.EdgeID})
	}
	table := stringTable([]string{"Relationship", "Permission", "Paths affected", "Critical", "Risk points", "Edge ID"}, rows, []float32{210, 170, 100, 80, 100, 170}, func(row int) {
		if row >= 0 && row < len(visible) {
			item := visible[row]
			detail := fmt.Sprintf("%s\n\nPaths affected: %d\nCritical paths: %d\nWeighted risk points: %d\n\nThis is an advisory; the app does not modify inventory.", item.Recommendation, item.PathsAffected, item.CriticalPaths, item.WeightedRiskPoints)
			for _, edge := range d.graph.Edges {
				if edge.ID == item.EdgeID {
					detail += fmt.Sprintf("\n\nEvidence source: %s\nFinding: %s\nObserved: %s", edge.Evidence.Source, edge.Evidence.Finding, edge.Evidence.Observed)
					break
				}
			}
			dialog.ShowInformation("Advisory remediation", detail, d.window)
		}
	})
	note := "Suggestions rank choke points by the number of enumerated paths affected, then weighted risk points."
	if result.PathsTruncated {
		note += " Path enumeration was truncated; rankings cover the displayed bounded result only."
	}
	return container.NewPadded(container.NewBorder(d.heading("Remediation priorities", note), nil, nil, nil, table))
}

func (d *Dashboard) importsScreen() fyne.CanvasObject {
	open := widget.NewButton("Import graph JSON…", func() { d.openImport() })
	rows := make([][]string, 0, len(d.history))
	for _, item := range d.history {
		rows = append(rows, []string{fmt.Sprint(item["imported_at"]), fmt.Sprint(item["source"]), fmt.Sprint(item["assets"]), fmt.Sprint(item["edges"]), fmt.Sprint(item["status"]), fmt.Sprint(item["message"])})
	}
	table := stringTable([]string{"Imported at", "Source", "Assets", "Relationships", "Result", "Validation"}, rows, []float32{190, 190, 70, 110, 90, 290}, nil)
	if len(rows) == 0 {
		return container.NewPadded(container.NewVBox(d.heading("Imports", "Load synthetic or authorized JSON inventory."), open, d.emptyState("No import history", "The initial demo graph is loaded automatically on first run.")))
	}
	return container.NewPadded(container.NewBorder(container.NewVBox(d.heading("Imports", "Import history and validation status."), open), nil, nil, nil, table))
}

func (d *Dashboard) reportsScreen() fyne.CanvasObject {
	graphButton := widget.NewButton("Export graph JSON…", func() { d.saveReport("graph") })
	copyGraph := widget.NewButton("Copy graph JSON", func() { d.copyReport("graph") })
	jsonButton := widget.NewButton("Export analysis JSON…", func() { d.saveReport("report-json") })
	copyJSON := widget.NewButton("Copy analysis JSON", func() { d.copyReport("report-json") })
	markdownButton := widget.NewButton("Export Markdown…", func() { d.saveReport("markdown") })
	copyMarkdown := widget.NewButton("Copy Markdown", func() { d.copyReport("markdown") })
	return container.NewPadded(container.NewVBox(d.heading("Reports", "Export local inventory, scored paths, evidence, and remediation advice to JSON or Markdown."), widget.NewCard("Graph export", "Assets, identities, relationships, and source evidence", container.NewHBox(graphButton, copyGraph)), widget.NewCard("Analysis report", "Ranked paths with exact score components and evidence", container.NewHBox(jsonButton, copyJSON)), widget.NewCard("Markdown report", "Human-readable path and choke-point remediation report", container.NewHBox(markdownButton, copyMarkdown))))
}

func (d *Dashboard) copyReport(kind string) {
	var data []byte
	var err error
	label := "graph"
	switch kind {
	case "graph":
		data, err = json.MarshalIndent(d.graph, "", "  ")
	case "report-json":
		data, err = d.service.RenderReport(context.Background(), "json")
		label = "analysis JSON report"
	case "markdown":
		data, err = d.service.RenderReport(context.Background(), "markdown")
		label = "Markdown report"
	default:
		err = fmt.Errorf("unknown report kind %q", kind)
	}
	if err != nil {
		dialog.ShowError(err, d.window)
		return
	}
	dialog.ShowConfirm("Copy local data?", fmt.Sprintf("The %s includes local inventory data and evidence. Copy it to the clipboard?", label), func(ok bool) {
		if ok {
			d.window.Clipboard().SetContent(string(data))
			dialog.ShowInformation("Copied", fmt.Sprintf("The %s is on the clipboard.", label), d.window)
		}
	}, d.window)
}

func (d *Dashboard) notificationsScreen() fyne.CanvasObject {
	filtered := make([]repository.Notification, 0, len(d.notices))
	for _, item := range d.notices {
		if d.matchesText(item.Title + " " + item.Detail + " " + item.Type + " " + fmt.Sprint(item.Severity)) {
			filtered = append(filtered, item)
		}
	}
	markRead := widget.NewButton("Mark all as read", func() {
		if err := d.store.MarkAllNotificationsRead(context.Background()); err != nil {
			dialog.ShowError(err, d.window)
			return
		}
		_ = d.refresh()
		d.show("notifications")
	})
	settings := widget.NewButton("Notification rules…", func() { d.notificationRulesDialog() })
	if len(filtered) == 0 {
		return container.NewPadded(container.NewVBox(d.heading("Notifications", "Local events · no external destinations"), container.NewHBox(markRead, settings), d.emptyState("You're up to date", "New critical paths and rejected imports appear here.")))
	}
	list := widget.NewList(func() int { return len(filtered) }, func() fyne.CanvasObject {
		return widget.NewButtonWithIcon("Notification template", nil, nil)
	}, func(id widget.ListItemID, object fyne.CanvasObject) {
		if id < 0 || id >= len(filtered) {
			return
		}
		item := filtered[id]
		severityName, severityResource := severityIcon(item.Severity)
		state := "Unread"
		if item.Read {
			state = "Read"
		}
		button := object.(*widget.Button)
		button.Text = fmt.Sprintf("%s · %s · %s · %s", severityName, item.Title, state, item.CreatedAt.Local().Format("2006-01-02 15:04"))
		button.Icon = severityResource
		button.OnTapped = func() {
			dialog.ShowConfirm(item.Title, item.Detail+"\n\nOpen the related local record?", func(open bool) {
				_ = d.store.MarkNotificationRead(context.Background(), item.ID)
				_ = d.refresh()
				if open {
					if item.Type == "critical_path" {
						d.show("paths")
					} else {
						d.show("imports")
					}
				} else {
					d.show("notifications")
				}
			}, d.window)
		}
		button.Refresh()
	})
	return container.NewPadded(container.NewBorder(container.NewVBox(d.heading("Notifications", "Severity icon and text · local events · no external destinations"), container.NewHBox(markRead, settings)), nil, nil, nil, list))
}

func severityIcon(severity int) (string, fyne.Resource) {
	name := "info"
	label := "Informational"
	switch {
	case severity >= 5:
		name, label = "critical", "Critical"
	case severity == 4:
		name, label = "high", "High"
	case severity == 3:
		name, label = "medium", "Medium"
	case severity == 2:
		name, label = "low", "Low"
	}
	resource, _ := icon("severity-" + name)
	return fmt.Sprintf("%s %d", label, severity), resource
}

func (d *Dashboard) aboutScreen() fyne.CanvasObject {
	iconData, _ := assets.Files.ReadFile("icons/icon.png")
	logo := canvas.NewImageFromImage(decodePNG(iconData))
	logo.SetMinSize(fyne.NewSize(72, 72))
	diagnostics := diagnosticReport()
	diagPreview := widget.NewMultiLineEntry()
	diagPreview.SetText(diagnostics)
	diagPreview.Disable()
	copyReport := widget.NewButton("Copy diagnostic report", func() {
		dialog.ShowConfirm("Preview diagnostics", diagnostics+"\n\nOnly this sanitized report will be copied.", func(ok bool) {
			if ok {
				d.window.Clipboard().SetContent(diagnostics)
			}
		}, d.window)
	})
	overview := container.NewVBox(container.NewHBox(logo, container.NewVBox(widget.NewLabelWithStyle(meta.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), widget.NewLabel("Version "+meta.Version), widget.NewLabel(meta.Purpose))), widget.NewLabel(meta.License+" License · "+meta.Copyright), widget.NewLabel("Build: "+runtime.GOOS+"/"+runtime.GOARCH+" · "+runtime.Version()), widget.NewLabel("The application and third-party license texts are bundled for offline review."), widget.NewButton("Source Code", func() { d.confirmOpenURL(meta.Repository) }), widget.NewButton("Found a bug? Please report it on GitHub.", func() { d.previewBugReport() }))
	usage, _ := assets.Files.ReadFile("docs/USAGE.md")
	usageSearch := widget.NewEntry()
	usageSearch.SetPlaceHolder("Find in the offline usage guide")
	usageText := widget.NewMultiLineEntry()
	usageText.SetText(string(usage))
	usageText.Disable()
	usageSearch.OnChanged = func(search string) {
		if strings.TrimSpace(search) == "" {
			usageText.SetText(string(usage))
			return
		}
		lines := strings.Split(string(usage), "\n")
		matches := make([]string, 0)
		for index, line := range lines {
			if strings.Contains(strings.ToLower(line), strings.ToLower(search)) {
				start := index - 1
				if start < 0 {
					start = 0
				}
				end := index + 2
				if end > len(lines) {
					end = len(lines)
				}
				matches = append(matches, strings.Join(lines[start:end], "\n"))
			}
		}
		if len(matches) == 0 {
			usageText.SetText("No matching guidance was found.")
		} else {
			usageText.SetText(strings.Join(matches, "\n\n---\n\n"))
		}
	}
	usagePage := container.NewBorder(usageSearch, widget.NewButton("Open full documentation on GitHub", func() { d.confirmOpenURL(meta.Repository + "/blob/main/docs/USAGE.md") }), nil, nil, usageText)
	shortcuts := widget.NewMultiLineEntry()
	shortcuts.SetText("Ctrl/Cmd+1 … 9  Navigate categories\nCtrl/Cmd+F       Focus global search\nCtrl/Cmd+,       Open Settings\nF1               Open Usage guide\nEnter            Open selected table record")
	shortcuts.Disable()
	readmeText := widget.NewMultiLineEntry()
	readmeData, _ := assets.Files.ReadFile("docs/README.md")
	readmeText.SetText(string(readmeData))
	readmeText.Disable()
	licenseText := widget.NewMultiLineEntry()
	licenseData, _ := assets.Files.ReadFile("docs/LICENSES.md")
	licenseText.SetText(string(licenseData))
	licenseText.Disable()
	thirdPartyNotices, _ := assets.Files.ReadFile("docs/THIRD_PARTY_NOTICES.md")
	licenseOptions := []string{"Offline license summary", "Third-party dependency inventory"}
	_ = fs.WalkDir(assets.Files, "licenses", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			licenseOptions = append(licenseOptions, strings.TrimPrefix(path, "licenses/"))
		}
		return nil
	})
	licenseChoice := widget.NewSelect(licenseOptions, func(name string) {
		if name == "Offline license summary" {
			licenseText.SetText(string(licenseData))
			return
		}
		if name == "Third-party dependency inventory" {
			licenseText.SetText(string(thirdPartyNotices))
			return
		}
		data, err := assets.Files.ReadFile("licenses/" + name)
		if err != nil {
			licenseText.SetText("Could not read the bundled license: " + err.Error())
			return
		}
		licenseText.SetText(string(data))
	})
	licenseChoice.SetSelected("Offline license summary")
	licensePage := container.NewBorder(licenseChoice, nil, nil, nil, licenseText)
	tabs := container.NewAppTabs(container.NewTabItem("Overview", overview), container.NewTabItem("README", readmeText), container.NewTabItem("Usage guide", usagePage), container.NewTabItem("Shortcuts", shortcuts), container.NewTabItem("Diagnostics", container.NewVBox(diagPreview, copyReport)), container.NewTabItem("Licenses & notices", licensePage))
	return container.NewPadded(container.NewBorder(d.heading("About", "Application identity, offline help, and sanitized diagnostics."), nil, nil, nil, tabs))
}

func (d *Dashboard) settingsScreen() fyne.CanvasObject {
	dark := d.app.Preferences().BoolWithFallback("appearance.dark", false)
	appearance := widget.NewCheck("Use dark appearance", func(value bool) {
		d.app.Preferences().SetBool("appearance.dark", value)
		if value {
			d.app.Settings().SetTheme(productTheme{dark: true})
		} else {
			d.app.Settings().SetTheme(productTheme{})
		}
	})
	appearance.SetChecked(dark)
	compact := widget.NewCheck("Use compact table rows", func(value bool) {
		d.app.Preferences().SetBool("table.compact", value)
		d.show(d.current)
	})
	compact.SetChecked(d.app.Preferences().BoolWithFallback("table.compact", false))
	windowInfo := widget.NewLabel("Window size is restored between sessions; the platform window manager controls window position.")
	ask := widget.NewCheck("Ask me every time at launch", func(value bool) {
		if value {
			_ = launcher.SavePreference("")
		} else {
			_ = launcher.SavePreference("gui")
		}
	})
	ask.SetChecked(launcher.ReadPreference() == "")
	reset := widget.NewButton("Reset saved launch choice", func() {
		if err := launcher.SavePreference(""); err != nil {
			dialog.ShowError(err, d.window)
			return
		}
		dialog.ShowInformation("Launcher preference reset", "The next launch will ask you to choose a front end.", d.window)
		ask.SetChecked(true)
	})
	dbPath := d.store.Path()
	installedPath, installed, installErr := notifier.Status()
	if installErr != nil {
		installedPath = "Status unavailable: " + installErr.Error()
	}
	var background *widget.Check
	background = widget.NewCheck("Enable opt-in background OS notifications", func(value bool) {
		if value {
			dialog.ShowConfirm("Install background notifier?", "This installs a per-user startup entry. The notifier reads the local database every 15 minutes, sends OS notifications, and makes no network calls. Continue?", func(ok bool) {
				if !ok {
					background.SetChecked(false)
					return
				}
				executable, err := os.Executable()
				if err == nil {
					_, err = notifier.Install(executable, d.store.Path())
				}
				if err != nil {
					background.SetChecked(false)
					dialog.ShowError(err, d.window)
					return
				}
				dialog.ShowInformation("Notifier installed", "Background notifications are enabled for this user.", d.window)
			}, d.window)
		} else {
			dialog.ShowConfirm("Uninstall background notifier?", "The per-user startup entry will be removed. Existing in-app notifications will remain.", func(ok bool) {
				if !ok {
					background.SetChecked(true)
					return
				}
				_, err := notifier.Uninstall()
				if err != nil {
					background.SetChecked(true)
					dialog.ShowError(err, d.window)
					return
				}
				dialog.ShowInformation("Notifier uninstalled", "The per-user startup entry was removed.", d.window)
			}, d.window)
		}
	})
	background.SetChecked(installed)
	appearancePanel := widget.NewCard("Appearance", "Custom light and dark themes use the bundled Noto Sans family.", container.NewVBox(appearance, compact, windowInfo))
	launchPanel := widget.NewCard("Launch behavior", "Preferences are stored in the OS user configuration directory with private permissions.", container.NewVBox(ask, reset))
	dataPanel := widget.NewCard("Data location", "SQLite data stays on this device.", widget.NewLabel(dbPath))
	notifierPanel := widget.NewCard("Background notifications", "Opt-in only · 15-minute polling · local database read-only · no network calls", container.NewVBox(background, widget.NewLabel(installedPath)))
	privacy := widget.NewLabel("No telemetry or integrations are configured. Destructive inventory changes require an explicit import; imported graphs replace stored relationships atomically.")
	return container.NewPadded(container.NewVBox(d.heading("Settings", "Appearance, launcher defaults, local data and notification status."), appearancePanel, launchPanel, dataPanel, notifierPanel, privacy))
}

func (d *Dashboard) openImport() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, d.window)
			return
		}
		if reader == nil {
			return
		}
		data, readErr := io.ReadAll(io.LimitReader(reader, 8<<20+1))
		closeErr := reader.Close()
		if readErr == nil {
			readErr = closeErr
		}
		if readErr != nil {
			dialog.ShowError(readErr, d.window)
			return
		}
		dialog.ShowConfirm("Replace local graph?", "The selected JSON file will replace the currently stored assets and relationships after validation. Invalid input leaves the existing graph unchanged.", func(confirmed bool) {
			if !confirmed {
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := d.service.Import(ctx, bytes.NewReader(data), "user JSON import"); err != nil {
				dialog.ShowError(err, d.window)
				_ = d.refresh()
				d.show("imports")
				return
			}
			if err := d.refresh(); err != nil {
				dialog.ShowError(err, d.window)
				return
			}
			d.show("imports")
			dialog.ShowInformation("Import complete", "The graph was validated and saved atomically.", d.window)
		}, d.window)
	}, d.window)
	fileDialog.SetFilter(storageJSONFilter{})
	fileDialog.Show()
}

type storageJSONFilter struct{}

func (storageJSONFilter) Matches(uri fyne.URI) bool {
	return strings.EqualFold(filepath.Ext(uri.Name()), ".json")
}

func (storageJSONFilter) Add(uri fyne.URI) {
}

func (storageJSONFilter) Includes(uri fyne.URI) bool {
	return storageJSONFilter{}.Matches(uri)
}

func (d *Dashboard) saveReport(kind string) {
	var data []byte
	var err error
	extension := ".json"
	switch kind {
	case "graph":
		data, err = json.MarshalIndent(d.graph, "", "  ")
	case "report-json":
		data, err = d.service.RenderReport(context.Background(), "json")
	case "markdown":
		data, err = d.service.RenderReport(context.Background(), "markdown")
		extension = ".md"
	default:
		err = fmt.Errorf("unknown report kind %q", kind)
	}
	if err != nil {
		dialog.ShowError(err, d.window)
		return
	}
	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, d.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()
		if _, err := writer.Write(data); err != nil {
			dialog.ShowError(err, d.window)
			return
		}
		dialog.ShowInformation("Report exported", "The report was saved to the selected file.", d.window)
	}, d.window)
	fileDialog.SetFileName("attackpath-" + kind + extension)
	if extension == ".json" {
		fileDialog.SetFilter(storageJSONFilter{})
	}
	fileDialog.Show()
}

func (d *Dashboard) notificationRulesDialog() {
	ctx := context.Background()
	rules, err := d.store.NotificationRules(ctx)
	if err != nil {
		dialog.ShowError(err, d.window)
		return
	}
	critical := rules["critical_path"]
	validation := rules["validation_failed"]
	criticalCheck := widget.NewCheck("Notify in app for new critical paths", nil)
	criticalCheck.SetChecked(critical.Enabled)
	criticalSelect := widget.NewSelect([]string{"Severity 1+", "Severity 2+", "Severity 3+", "Severity 4+", "Severity 5"}, nil)
	criticalSelect.SetSelectedIndex(critical.MinimumSeverity - 1)
	validationCheck := widget.NewCheck("Notify in app for invalid imports", nil)
	validationCheck.SetChecked(validation.Enabled)
	validationSelect := widget.NewSelect([]string{"Severity 1+", "Severity 2+", "Severity 3+", "Severity 4+", "Severity 5"}, nil)
	validationSelect.SetSelectedIndex(validation.MinimumSeverity - 1)
	detailCheck := widget.NewCheck("Include extra detail in OS notification text", nil)
	detailCheck.SetChecked(critical.DetailedText)
	start := widget.NewEntry()
	start.SetText(critical.QuietStart)
	start.SetPlaceHolder("22:00")
	end := widget.NewEntry()
	end.SetText(critical.QuietEnd)
	end.SetPlaceHolder("07:00")
	quiet := container.NewGridWithColumns(2, start, end)
	content := container.NewVBox(widget.NewLabel("Severity combines a number and label. Quiet hours apply to background OS notices; in-app events remain visible."), criticalCheck, criticalSelect, validationCheck, validationSelect, detailCheck, widget.NewLabel("Quiet hours (local time, HH:MM)"), quiet)
	dialog.ShowCustomConfirm("Notification rules", "Save", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}
		if _, err := time.Parse("15:04", start.Text); err != nil {
			dialog.ShowInformation("Invalid quiet-hours start", "Enter a local time in HH:MM format.", d.window)
			return
		}
		if _, err := time.Parse("15:04", end.Text); err != nil {
			dialog.ShowInformation("Invalid quiet-hours end", "Enter a local time in HH:MM format.", d.window)
			return
		}
		critical.MinimumSeverity = criticalSelect.SelectedIndex() + 1
		critical.Enabled = criticalCheck.Checked
		critical.DetailedText = detailCheck.Checked
		critical.QuietStart, critical.QuietEnd = start.Text, end.Text
		validation.MinimumSeverity = validationSelect.SelectedIndex() + 1
		validation.Enabled = validationCheck.Checked
		validation.DetailedText = detailCheck.Checked
		validation.QuietStart, validation.QuietEnd = start.Text, end.Text
		if err := d.store.SaveNotificationRule(ctx, "critical_path", critical); err != nil {
			dialog.ShowError(err, d.window)
			return
		}
		if err := d.store.SaveNotificationRule(ctx, "validation_failed", validation); err != nil {
			dialog.ShowError(err, d.window)
			return
		}
		dialog.ShowInformation("Notification rules saved", "New events will follow the configured severity and quiet hours.", d.window)
	}, d.window)
}

func (d *Dashboard) previewBugReport() {
	diagnostics := diagnosticReport()
	title := url.QueryEscape("Bug report: AttackPath Visualizer")
	body := url.QueryEscape("## Description\n\n## Steps to reproduce\n\n## Sanitized diagnostics\n" + diagnostics)
	issueURL := meta.Issues + "?title=" + title + "&body=" + body
	if len(issueURL) > 7000 {
		d.window.Clipboard().SetContent(diagnostics)
		issueURL = meta.Issues
		dialog.ShowInformation("Diagnostics copied", "The report was too long for a safe URL. The sanitized diagnostics were copied; the blank issue form will open next.", d.window)
	}
	d.confirmOpenURL(issueURL)
}

func (d *Dashboard) confirmOpenURL(target string) {
	if !AllowedRepositoryURL(target) {
		dialog.ShowInformation("External link unavailable", "The configured repository link is invalid; no external page was opened.", d.window)
		return
	}
	dialog.ShowConfirm("Open external page?", "The operating-system browser will open:\n"+target, func(ok bool) {
		if ok {
			if err := openExternal(target); err != nil {
				dialog.ShowError(err, d.window)
			}
		}
	}, d.window)
}

func AllowedRepositoryURL(target string) bool {
	parsed, err := url.Parse(target)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "github.com" {
		return false
	}
	return parsed.Path == "/NexzaDev/attackpath-visualizer" || strings.HasPrefix(parsed.Path, "/NexzaDev/attackpath-visualizer/")
}

func openExternal(target string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "windows":
		command, args = "rundll32", []string{"url.dll,FileProtocolHandler", target}
	case "darwin":
		command, args = "open", []string{target}
	default:
		command, args = "xdg-open", []string{target}
	}
	if _, err := exec.LookPath(command); err != nil {
		return fmt.Errorf("no OS browser handler was found; open this URL manually: %s", target)
	}
	if err := exec.Command(command, args...).Start(); err != nil {
		return fmt.Errorf("open external link: %w", err)
	}
	return nil
}

func pathLabel(path domain.Path) string {
	labels := make([]string, len(path.Nodes))
	for index, node := range path.Nodes {
		labels[index] = fmt.Sprintf("%s [%s]", node.Name, node.ID)
	}
	return strings.Join(labels, " → ")
}

func (d *Dashboard) showPathDetails(path domain.Path) {
	lines := []string{path.Explanation, fmt.Sprintf("Impact component: tier %d/5", path.Components.ImpactTier), fmt.Sprintf("Edge confidence components: %v", path.Components.EdgeConfidences), fmt.Sprintf("Compounded confidence: %d%%", path.Components.CompoundedConfidence), fmt.Sprintf("Score: %d points", path.Components.Score), "", "Evidence by edge:"}
	for _, edge := range path.Edges {
		lines = append(lines, fmt.Sprintf("• %s → %s · %s · %d%%", edge.From, edge.To, edge.Relation, edge.Confidence), fmt.Sprintf("  Source: %s", edge.Evidence.Source), fmt.Sprintf("  Finding: %s", edge.Evidence.Finding), fmt.Sprintf("  Observed: %s", edge.Evidence.Observed))
	}
	dialog.ShowInformation("Potential path details", strings.Join(lines, "\n"), d.window)
}

func (d *Dashboard) chokePoints() fyne.CanvasObject {
	items := make([]fyne.CanvasObject, 0, 4)
	for index, suggestion := range d.summary.Remediation.Suggestions {
		if index == 4 {
			break
		}
		items = append(items, widget.NewLabel(fmt.Sprintf("%d paths · %s — %s → %s · %d risk points", suggestion.PathsAffected, suggestion.EdgeID, suggestion.From, suggestion.To, suggestion.WeightedRiskPoints)))
	}
	if len(items) == 0 {
		return widget.NewLabel("No path-affecting edges are currently available.")
	}
	return container.NewVBox(items...)
}

func (d *Dashboard) filteredNodes() []domain.Node {
	all := d.graph.Nodes()
	result := make([]domain.Node, 0, len(all))
	for _, node := range all {
		value := strings.ToLower(node.Name + " " + node.EntityType + " " + node.Kind + " " + node.ID + " " + node.Description)
		if d.search == "" || strings.Contains(value, d.search) {
			result = append(result, node)
		}
	}
	return result
}

func (d *Dashboard) filteredEdges() []domain.Edge {
	result := make([]domain.Edge, 0, len(d.graph.Edges))
	for _, edge := range d.graph.Edges {
		value := strings.ToLower(edge.ID + " " + edge.From + " " + edge.To + " " + edge.Relation + " " + edge.Evidence.Source + " " + edge.Evidence.Finding)
		if d.search == "" || strings.Contains(value, d.search) {
			result = append(result, edge)
		}
	}
	return result
}

func (d *Dashboard) matchesPath(path domain.Path) bool {
	if d.impactFilter > 0 && path.Impact != d.impactFilter {
		return false
	}
	if d.confidenceFilter > 0 && min(path.Confidence/20, 4)+1 != d.confidenceFilter {
		return false
	}
	value := pathLabel(path) + " " + path.Explanation
	for _, node := range path.Nodes {
		value += " " + node.Name + " " + node.ID
	}
	for _, edge := range path.Edges {
		value += " " + edge.Relation + " " + edge.Evidence.Finding + " " + edge.Evidence.Source
	}
	return d.matchesText(value)
}

func (d *Dashboard) matchesText(value string) bool {
	return d.search == "" || strings.Contains(strings.ToLower(value), d.search)
}

func (d *Dashboard) emptyState(title, detail string) fyne.CanvasObject {
	return container.NewCenter(widget.NewCard(title, "No records to show", widget.NewLabel(detail)))
}

func stringTable(headers []string, rows [][]string, widths []float32, selected func(int)) *widget.Table {
	rowOrder := make([]int, len(rows))
	for index := range rows {
		rowOrder[index] = index
	}
	table := widget.NewTable(func() (int, int) { return len(rows) + 1, len(headers) }, func() fyne.CanvasObject {
		return widget.NewLabel("template")
	}, func(id widget.TableCellID, object fyne.CanvasObject) {
		label := object.(*widget.Label)
		label.Wrapping = fyne.TextWrapOff
		if id.Row == 0 {
			label.SetText(headers[id.Col] + "  ↕")
			label.TextStyle = fyne.TextStyle{Bold: true}
			return
		}
		if id.Row-1 < 0 || id.Row-1 >= len(rowOrder) {
			label.SetText("")
			return
		}
		row := rowOrder[id.Row-1]
		if row < 0 || row >= len(rows) || id.Col >= len(rows[row]) {
			label.SetText("")
			return
		}
		label.TextStyle = fyne.TextStyle{}
		label.SetText(rows[row][id.Col])
	})
	for index, width := range widths {
		table.SetColumnWidth(index, width)
	}
	rowHeight := float32(38)
	if app := fyne.CurrentApp(); app != nil && app.Preferences().BoolWithFallback("table.compact", false) {
		rowHeight = 27
	}
	table.SetRowHeight(-1, rowHeight)
	sortedColumn := -1
	ascending := true
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 && id.Col >= 0 && id.Col < len(headers) {
			if sortedColumn == id.Col {
				ascending = !ascending
			} else {
				sortedColumn = id.Col
				ascending = true
			}
			sort.SliceStable(rowOrder, func(left, right int) bool {
				leftRow := rows[rowOrder[left]]
				rightRow := rows[rowOrder[right]]
				if id.Col >= len(leftRow) || id.Col >= len(rightRow) {
					return false
				}
				leftText := strings.ToLower(leftRow[id.Col])
				rightText := strings.ToLower(rightRow[id.Col])
				leftNumber, leftErr := strconv.ParseFloat(strings.TrimSuffix(leftText, "%"), 64)
				rightNumber, rightErr := strconv.ParseFloat(strings.TrimSuffix(rightText, "%"), 64)
				if leftErr == nil && rightErr == nil {
					if ascending {
						return leftNumber < rightNumber
					}
					return leftNumber > rightNumber
				}
				if ascending {
					return leftText < rightText
				}
				return leftText > rightText
			})
			table.UnselectAll()
			table.Refresh()
			return
		}
		if id.Row > 0 && selected != nil && id.Row-1 < len(rowOrder) {
			selected(rowOrder[id.Row-1])
		}
	}
	return table
}

func decodePNG(data []byte) image.Image {
	imageData, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return imageData
}

func icon(name string) (fyne.Resource, error) {
	data, err := assets.Files.ReadFile("icons/nav/" + name + ".svg")
	if err != nil {
		return nil, err
	}
	if app := fyne.CurrentApp(); app != nil && app.Preferences().BoolWithFallback("appearance.dark", false) {
		svg := string(data)
		for original, replacement := range map[string]string{"#334155": "#D4DFED", "#b4232e": "#FF9095", "#c44e28": "#FFAA7A", "#9a6b11": "#F4D078", "#3f6a9f": "#8FBCF7", "#52657e": "#BBC9DB"} {
			svg = strings.ReplaceAll(svg, original, replacement)
		}
		data = []byte(svg)
	}
	return fyne.NewStaticResource(name+".svg", data), nil
}

func diagnosticReport() string {
	return fmt.Sprintf("Application: %s\nVersion: %s\nOS: %s\nArchitecture: %s\nGo runtime: %s\nConfiguration: local SQLite enabled; no credentials configured\nTelemetry: disabled", meta.Name, meta.Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
