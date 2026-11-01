package tui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/NexzaDev/attackpath-visualizer/internal/config"
	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
	"github.com/NexzaDev/attackpath-visualizer/internal/launcher"
	"github.com/NexzaDev/attackpath-visualizer/internal/meta"
	"github.com/NexzaDev/attackpath-visualizer/internal/notifier"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
	"github.com/NexzaDev/attackpath-visualizer/internal/service"
)

func Run(requestedDB string) error {
	path, err := config.DatabasePath(requestedDB)
	if err != nil {
		return err
	}
	store, err := repository.Open(path)
	if err != nil {
		return err
	}
	defer store.Close()
	ctx := context.Background()
	app := service.Service{Store: store}
	graph, err := store.Graph(ctx)
	if err != nil {
		return err
	}
	if len(graph.Assets)+len(graph.Identities) == 0 {
		history, historyErr := store.ImportHistory(ctx)
		if historyErr != nil {
			return historyErr
		}
		if len(history) == 0 {
			if err = app.LoadDemo(ctx); err != nil {
				return err
			}
		}
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		graph, err = store.Graph(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("\n%s %s\n1 Overview  2 Graph  3 Paths  4 Assets  5 Evidence  6 Remediation  7 Imports  8 Reports  9 Notifications  10 Settings  11 About  12 Diagnostics  13 Switch to Dashboard  0 Quit\nSelect: ", meta.Name, meta.Version)
		choice, readErr := reader.ReadString('\n')
		if readErr != nil {
			if strings.TrimSpace(choice) == "" {
				return nil
			}
		}
		switch strings.TrimSpace(choice) {
		case "1":
			if err := printOverview(graph); err != nil {
				return err
			}
		case "2":
			printGraph(graph)
		case "3":
			if err := printPaths(graph); err != nil {
				return err
			}
		case "4":
			if err := printAssets(graph); err != nil {
				return err
			}
		case "5":
			printEvidence(graph)
		case "6":
			if err := printRemediation(app); err != nil {
				return err
			}
		case "7":
			if err := importGraph(reader, app, graph); err != nil {
				return err
			}
			printImports(store)
		case "8":
			if err := exportReports(reader, app, graph); err != nil {
				return err
			}
		case "9":
			printNotifications(store, reader)
		case "10":
			fmt.Printf("Database: %s\nNotifier status: ", path)
			notifierPath, installed, statusErr := notifier.Status()
			if statusErr != nil {
				fmt.Println(statusErr)
			} else {
				fmt.Printf("%s (%s)\n", notifierPath, map[bool]string{true: "installed", false: "off"}[installed])
			}
			fmt.Print("Type e to enable notifier, d to disable it, r to reset launch mode, or Enter: ")
			setting, _ := reader.ReadString('\n')
			switch strings.ToLower(strings.TrimSpace(setting)) {
			case "e":
				executable, err := os.Executable()
				if err != nil {
					return err
				}
				path, err := notifier.Install(executable, path)
				if err != nil {
					return err
				}
				fmt.Println("Notifier installed at", path)
			case "d":
				path, err := notifier.Uninstall()
				if err != nil {
					return err
				}
				fmt.Println("Notifier removed from", path)
			case "r":
				if err := launcher.SavePreference(""); err != nil {
					return err
				}
				fmt.Println("Saved launcher mode cleared.")
			}
		case "11":
			fmt.Printf("%s\n%s\nVersion: %s\nLicense: MIT\nSource: %s\nSecurity vulnerability reports: GitHub private advisory form (preferred): %s\nBackup channel: email. If you cannot use the GitHub form, send the report to SuperMcDuck@proton.me.\n", meta.Name, meta.Purpose, meta.Version, meta.Repository, meta.SecurityAdvisory)
		case "12":
			fmt.Printf("OS/architecture: %s/%s\nRuntime: %s\nConfiguration: local-only; no credentials configured\n", runtime.GOOS, runtime.GOARCH, runtime.Version())
		case "13":
			return launcher.StartGUI()
		case "0":
			return nil
		default:
			fmt.Println("Choose a number from 0 to 13.")
		}

	}
}
func printGraph(graph domain.Graph) {
	for _, node := range graph.Nodes() {
		fmt.Printf("[%d/5] %s (%s · %s)\n", node.Impact, node.Name, node.EntityType, node.ID)
		for _, edge := range graph.Edges {
			if edge.From == node.ID {
				fmt.Printf("  ── %s → %s\n", edge.Relation, edge.To)
			}
		}
	}
}

func printRemediation(app service.Service) error {
	analysis, err := app.Analyze(context.Background(), 16, 1000)
	if err != nil {
		return err
	}
	for index, item := range analysis.Remediation.Suggestions {
		fmt.Printf("%d. %d paths · %d critical · %d risk points · review %s: %s — %s → %s\n   %s\n", index+1, item.PathsAffected, item.CriticalPaths, item.WeightedRiskPoints, item.EdgeID, item.Relation, item.From, item.To, item.Recommendation)
	}
	if len(analysis.Remediation.Suggestions) == 0 {
		fmt.Println("No path-affecting edges are currently available.")
	}
	if analysis.Paths.Truncated {
		fmt.Println("Note: rankings cover the bounded path set; path enumeration was truncated.")
	}
	return nil
}

func importGraph(reader *bufio.Reader, app service.Service, graph domain.Graph) error {
	fmt.Print("Import JSON file path (empty to cancel): ")
	path, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read import path: %w", err)
	}
	path = strings.TrimSpace(path)
	if path == "" {
		fmt.Println("Import cancelled.")
		return nil
	}
	if len(graph.Assets)+len(graph.Identities) > 0 {
		fmt.Print("This replaces the current local graph. Continue? [y/N]: ")
		answer, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return fmt.Errorf("read import confirmation: %w", readErr)
		}
		if !strings.EqualFold(strings.TrimSpace(answer), "y") && !strings.EqualFold(strings.TrimSpace(answer), "yes") {
			fmt.Println("Import cancelled.")
			return nil
		}
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open import file: %w", err)
	}
	defer file.Close()
	if err = app.Import(context.Background(), file, "user JSON import"); err != nil {
		return err
	}
	fmt.Println("Import validated and saved.")
	return nil
}

func printImports(store *repository.Store) {
	items, err := store.ImportHistory(context.Background())
	if err != nil {
		fmt.Println("Could not read import history:", err)
		return
	}
	for _, item := range items {
		fmt.Printf("%v · %v · assets %v · identities %v · misconfigurations %v · findings %v · edges %v · %v\n", item["imported_at"], item["status"], item["assets"], item["identities"], item["misconfigurations"], item["findings"], item["edges"], item["message"])
	}
	if len(items) == 0 {
		fmt.Println("No imports are recorded.")
	}
}

func exportReports(reader *bufio.Reader, app service.Service, graph domain.Graph) error {
	fmt.Print("Export [g]raph JSON, analysis [j]SON, or [m]arkdown (Enter cancels): ")
	choice, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read report choice: %w", err)
	}
	choice = strings.ToLower(strings.TrimSpace(choice))
	if choice == "" {
		fmt.Println("Export cancelled.")
		return nil
	}
	var data []byte
	filename := ""
	switch choice {
	case "g":
		data, err = json.MarshalIndent(graph, "", "  ")
		filename = "attackpath-graph.json"
	case "j":
		data, err = app.RenderReport(context.Background(), "json")
		filename = "attackpath-analysis.json"
	case "m":
		data, err = app.RenderReport(context.Background(), "markdown")
		filename = "attackpath-report.md"
	default:
		return fmt.Errorf("choose g, j, or m")
	}
	if err != nil {
		return fmt.Errorf("render report: %w", err)
	}
	fmt.Printf("Export file [%s]: ", filename)
	file, readErr := reader.ReadString('\n')
	if readErr != nil && readErr != io.EOF {
		return fmt.Errorf("read report filename: %w", readErr)
	}
	file = strings.TrimSpace(file)
	if file == "" {
		file = filename
	}
	if _, err := os.Stat(file); err == nil {
		fmt.Printf("%s exists. Replace it? [y/N]: ", file)
		answer, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return fmt.Errorf("read overwrite confirmation: %w", readErr)
		}
		if !strings.EqualFold(strings.TrimSpace(answer), "y") && !strings.EqualFold(strings.TrimSpace(answer), "yes") {
			fmt.Println("Export cancelled.")
			return nil
		}
	}
	if err := os.WriteFile(file, data, 0600); err != nil {
		return fmt.Errorf("export report: %w", err)
	}
	fmt.Println("Exported to", file)
	return nil
}

func printOverview(graph domain.Graph) error {
	result, err := graph.EnumeratePaths(16, 1000)
	if err != nil {
		return err
	}
	fmt.Printf("Assets: %d · Identities: %d · Relationships: %d · Potential paths: %d · expansions: %d · truncated: %t\n", len(graph.Assets), len(graph.Identities), len(graph.Edges), len(result.Paths), result.Expanded, result.Truncated)
	for index, path := range result.Paths {
		if index >= 5 {
			break
		}
		fmt.Printf("%d. %s · score %d · confidence %d%% · %s\n", index+1, pathLabel(path), path.Score, path.Confidence, path.Explanation)
	}
	return nil
}

func printPaths(graph domain.Graph) error {
	result, err := graph.EnumeratePaths(16, 1000)
	if err != nil {
		return err
	}
	for index, path := range result.Paths {
		fmt.Printf("%d. %s\n   %s\n   Components: impact tier %d/5 × edge confidences %v = compounded %d%%; score %d\n", index+1, pathLabel(path), path.Explanation, path.Components.ImpactTier, path.Components.EdgeConfidences, path.Components.CompoundedConfidence, path.Components.Score)
		for _, edge := range path.Edges {
			fmt.Printf("   %s · %s · %d%% · source %s · finding %s · observed %s\n", edge.ID, edge.Relation, edge.Confidence, edge.Evidence.Source, edge.Evidence.Finding, edge.Evidence.Observed)
		}
	}
	if result.Truncated {
		fmt.Printf("Path listing truncated after %d expansions; rankings cover this bounded result.\n", result.Expanded)
	}
	return nil
}

func printAssets(graph domain.Graph) error {
	for _, node := range graph.Nodes() {
		fmt.Printf("%s · %s/%s · impact %d/5 · %s\n", node.Name, node.EntityType, node.Kind, node.Impact, node.Description)
		for _, item := range graph.Misconfigurations {
			if item.AssetID == node.ID || item.IdentityID == node.ID {
				fmt.Printf("  Misconfiguration S%d · %s\n", item.Severity, item.Title)
			}
		}
		for _, item := range graph.Findings {
			if item.AssetID == node.ID || item.IdentityID == node.ID {
				fmt.Printf("  Finding S%d · %s · %s\n", item.Severity, item.Title, item.Source)
			}
		}
	}
	return nil
}

func printEvidence(graph domain.Graph) {
	for _, edge := range graph.Edges {
		fmt.Printf("%s: %s → %s · %s · %d%%\n   %s: %s · %s\n", edge.ID, edge.From, edge.To, edge.Relation, edge.Confidence, edge.Evidence.Source, edge.Evidence.Finding, edge.Evidence.Observed)
	}
}

func printNotifications(store *repository.Store, reader *bufio.Reader) {
	items, err := store.Notifications(context.Background(), 50)
	if err != nil {
		fmt.Println("Could not read notifications:", err)
		return
	}
	if len(items) == 0 {
		fmt.Println("No notifications recorded.")
		return
	}
	for _, item := range items {
		state := "unread"
		if item.Read {
			state = "read"
		}
		fmt.Printf("S%d · %s · %s · %s\n", item.Severity, item.Title, state, item.CreatedAt.Local().Format(time.RFC822))
	}
	fmt.Print("Type a to mark all read, or press Enter: ")
	answer, _ := reader.ReadString('\n')
	if strings.EqualFold(strings.TrimSpace(answer), "a") {
		if err := store.MarkAllNotificationsRead(context.Background()); err != nil {
			fmt.Println("Could not mark notifications read:", err)
		} else {
			fmt.Println("Notifications marked as read.")
		}
	}
}

func pathLabel(path domain.Path) string {
	parts := make([]string, 0, len(path.Nodes))
	for _, node := range path.Nodes {
		parts = append(parts, fmt.Sprintf("%s [%s]", node.Name, node.ID))
	}
	return strings.Join(parts, " → ")
}
