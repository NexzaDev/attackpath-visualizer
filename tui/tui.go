package tui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
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
	if len(graph.Assets) == 0 {
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
		history, historyErr := store.ImportHistory(ctx)
		if historyErr != nil {
			return historyErr
		}
		mode := "local data"
		if len(history) > 0 && history[0]["source"] == "demo fixture" {
			mode = "demo data"
		}
		fmt.Printf("\n%s %s · %s\n1 Overview  2 Graph  3 Paths  4 Assets  5 Evidence  6 Remediation  7 Imports  8 Reports  9 Notifications  10 Settings  11 About  12 Diagnostics  13 Switch to Dashboard  0 Quit\nSelect: ", meta.Name, meta.Version)
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
			if err := printRemediation(graph); err != nil {
				return err
			}
		case "7":
			if err := importGraph(reader, app, graph); err != nil {
				return err
			}
			printImports(store)
		case "8":
			if err := exportGraph(reader, graph); err != nil {
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
			fmt.Printf("%s\n%s\nVersion: %s\nLicense: MIT\nSource: %s\n", meta.Name, meta.Purpose, meta.Version, meta.Repository)
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
	for _, asset := range graph.Assets {
		fmt.Printf("[%d/5] %s (%s)\n", asset.Impact, asset.Name, asset.ID)
		for _, edge := range graph.Edges {
			if edge.From == asset.ID {
				fmt.Printf("  ── %s → %s\n", edge.Relation, edge.To)
			}
		}
	}
}

func printRemediation(graph domain.Graph) error {
	paths, err := graph.AllPaths(16, 1000)
	if err != nil {
		return err
	}
	counts := make(map[string]int)
	edges := make(map[string]domain.Edge)
	for _, path := range paths {
		for _, edge := range path.Edges {
			counts[edge.ID]++
			edges[edge.ID] = edge
		}
	}
	ids := make([]string, 0, len(counts))
	for id := range counts {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		if counts[ids[i]] == counts[ids[j]] {
			return ids[i] < ids[j]
		}
		return counts[ids[i]] > counts[ids[j]]
	})
	for _, id := range ids {
		edge := edges[id]
		fmt.Printf("%d paths · review %s: %s — %s → %s\n", counts[id], edge.ID, edge.Relation, edge.From, edge.To)
	}
	if len(ids) == 0 {
		fmt.Println("No path-affecting edges are currently available.")
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
	if len(graph.Assets) > 0 {
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
		fmt.Printf("%v · %v · assets %v · edges %v · %v\n", item["imported_at"], item["status"], item["assets"], item["edges"], item["message"])
	}
	if len(items) == 0 {
		fmt.Println("No imports are recorded.")
	}
}

func exportGraph(reader *bufio.Reader, graph domain.Graph) error {
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return err
	}
	fmt.Print("Export file [attackpath-graph.json]: ")
	file, _ := reader.ReadString('\n')
	file = strings.TrimSpace(file)
	if file == "" {
		file = "attackpath-graph.json"
	}
	if _, err := os.Stat(file); err == nil {
		fmt.Printf("%s exists. Replace it? [y/N]: ", file)
		answer, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(answer), "y") && !strings.EqualFold(strings.TrimSpace(answer), "yes") {
			fmt.Println("Export cancelled.")
			return nil
		}
	}
	if err := os.WriteFile(file, data, 0600); err != nil {
		return fmt.Errorf("export graph: %w", err)
	}
	fmt.Println("Exported to", file)
	return nil
}

func printOverview(graph domain.Graph) error {
	paths, err := graph.AllPaths(16, 1000)
	if err != nil {
		return err
	}
	fmt.Printf("Assets: %d · Relationships: %d · Potential paths: %d\n", len(graph.Assets), len(graph.Edges), len(paths))
	for index, path := range paths {
		if index >= 5 {
			break
		}
		fmt.Printf("%d. %s · score %d · confidence %d%% · %s\n", index+1, pathLabel(path), path.Score, path.Confidence, path.Explanation)
	}
	return nil
}

func printPaths(graph domain.Graph) error {
	paths, err := graph.AllPaths(16, 1000)
	if err != nil {
		return err
	}
	for index, path := range paths {
		fmt.Printf("%d. %s\n   %s · impact %d/5 · confidence %d%%\n", index+1, pathLabel(path), path.Explanation, path.Impact, path.Confidence)
		for _, edge := range path.Edges {
			fmt.Printf("   Evidence: %s · %s\n", edge.Evidence.Source, edge.Evidence.Finding)
		}
	}
	return nil
}

func printAssets(graph domain.Graph) error {
	for _, asset := range graph.Assets {
		fmt.Printf("%s · %s · impact %d/5 · %s\n", asset.Name, asset.Kind, asset.Impact, asset.Description)
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
	parts := make([]string, 0, len(path.Assets))
	for _, asset := range path.Assets {
		parts = append(parts, asset.Name)
	}
	return strings.Join(parts, " → ")
}
