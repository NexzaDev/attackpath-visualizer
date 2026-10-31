package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/NexzaDev/attackpath-visualizer/internal/meta"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
	"github.com/NexzaDev/attackpath-visualizer/internal/service"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		args = []string{"help"}
	}
	if args[0] == "--version" || args[0] == "version" {
		fmt.Fprintf(stdout, "%s %s\n", meta.Name, meta.Version)
		return nil
	}
	if args[0] == "--help" || args[0] == "help" {
		printHelp(stdout)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if args[0] == "about" {
		fmt.Fprintf(stdout, "%s\nVersion: %s\nLicense: %s\nRepository: %s\nBuild: %s/%s; Go %s\n", meta.Name, meta.Version, meta.License, meta.Repository, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return nil
	}
	if args[0] == "diagnostics" {
		fmt.Fprintf(stdout, "App: %s\nVersion: %s\nOS/architecture: %s/%s\nGo runtime: %s\nConfiguration: local SQLite; no credentials configured\n", meta.Name, meta.Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return nil
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	dbPath := fs.String("db", "", "SQLite database path")
	jsonOutput := fs.Bool("json", false, "write machine-readable JSON")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if args[0] != "import" && fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	path := *dbPath
	if path == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("find user configuration directory: %w", err)
		}
		path = filepath.Join(base, "AttackPathVisualizer", "graph.db")
	}
	store, err := repository.Open(path)
	if err != nil {
		return err
	}
	defer store.Close()
	app := service.Service{Store: store}
	graph, err := store.Graph(ctx)
	if err != nil {
		return err
	}
	if len(graph.Assets) == 0 {
		fixture, err := filepath.Abs("testdata/demo.json")
		if err != nil {
			return err
		}
		if err = app.LoadDemo(ctx, fixture); err != nil {
			return fmt.Errorf("initialize demo data: %w", err)
		}
		graph, err = store.Graph(ctx)
		if err != nil {
			return err
		}
	}
	switch args[0] {
	case "overview":
		paths, err := graph.FindPaths(graph.Assets[0].ID, "", 8, 100)
		if err != nil {
			return err
		}
		return printValue(stdout, *jsonOutput, map[string]any{"mode": "demo data", "assets": len(graph.Assets), "relationships": len(graph.Edges), "paths": paths})
	case "assets":
		return printValue(stdout, *jsonOutput, graph.Assets)
	case "paths":
		if len(graph.Assets) == 0 {
			return errors.New("no assets are available")
		}
		paths, err := graph.FindPaths(graph.Assets[0].ID, "", 8, 100)
		if err != nil {
			return err
		}
		return printValue(stdout, *jsonOutput, paths)
	case "import":
		var input io.Reader = os.Stdin
		if file := fs.Arg(0); file != "" {
			f, openErr := os.Open(file)
			if openErr != nil {
				return openErr
			}
			defer f.Close()
			input = f
		}
		if err := app.Import(ctx, input, "user import"); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "Import validated and saved.")
		return nil
	case "export":
		return printValue(stdout, true, graph)
	default:
		return fmt.Errorf("unknown command %q; run 'attackpath-visualizer help'", args[0])
	}
}

func printValue(w io.Writer, jsonOutput bool, value any) error {
	if jsonOutput {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func printHelp(w io.Writer) {
	fmt.Fprintf(w, "%s %s\n%s\n\nUsage: attackpath-visualizer <command> [options]\n\nCommands:\n  overview      Summarize local demo or imported data\n  paths         List ranked potential paths\n  assets        List inventoried assets\n  import [file] Import a graph JSON file or stdin\n  export        Export the stored graph as JSON\n  about         Show version and project information\n  diagnostics   Show a sanitized local runtime report\n  help          Show this help\n\nOptions: --db PATH, --json\n", meta.Name, meta.Version, meta.Purpose)
}
