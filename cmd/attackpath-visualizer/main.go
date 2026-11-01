package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/NexzaDev/attackpath-visualizer/internal/config"
	"github.com/NexzaDev/attackpath-visualizer/internal/launcher"
	"github.com/NexzaDev/attackpath-visualizer/internal/meta"
	"github.com/NexzaDev/attackpath-visualizer/internal/notifier"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
	"github.com/NexzaDev/attackpath-visualizer/internal/service"
	"github.com/NexzaDev/attackpath-visualizer/tui"
)

func main() {
	args := os.Args[1:]
	launch, options, dbPath, err := launcher.ParseLaunchArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	if launch {
		options.EnvUI = os.Getenv("ATTACKPATH_VISUALIZER_UI")
		options.Saved = launcher.ReadPreference()
		options.TTY = launcher.InteractiveStdin()
		options.Headless = launcher.Headless() || os.Getenv("CI") != ""
		if os.Getenv("CI") != "" {
			options.NoPrompt = true
		}
		decision, decideErr := launcher.Decide(options)
		if decideErr != nil {
			fmt.Fprintln(os.Stderr, "error:", decideErr)
			os.Exit(2)
		}
		if decision.Prompt {
			decision.Mode, err = launcher.Prompt(os.Stdin, os.Stdout, options.Saved)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(2)
			}
		}
		if decision.Message != "" {
			fmt.Fprintln(os.Stderr, decision.Message)
		}
		if decision.Chooser {
			err = launchChooser(dbPath)
		} else if decision.Mode == "tui" {
			err = tui.Run(dbPath)
		} else {
			err = launchGUI(dbPath)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}
	if err := run(args, os.Stdout, os.Stderr); err != nil {
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
		fmt.Fprintf(stdout, "%s\nVersion: %s\nLicense: %s\n%s\nRepository: %s\nBuild: %s/%s; Go %s\n", meta.Name, meta.Version, meta.License, meta.Copyright, meta.Repository, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return nil
	}
	if args[0] == "diagnostics" {
		fmt.Fprintf(stdout, "App: %s\nVersion: %s\nOS/architecture: %s/%s\nGo runtime: %s\nConfiguration: local SQLite; no credentials configured\n", meta.Name, meta.Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return nil
	}
	if args[0] == "config" {
		if len(args) != 2 || args[1] != "reset-ui" {
			return errors.New("usage: attackpath-visualizer config reset-ui")
		}
		if err := launcher.SavePreference(""); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "Saved launch preference cleared.")
		return nil
	}
	if args[0] == "notifier" {
		return runNotifier(args[1:], stdout)
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	dbPath := fs.String("db", "", "SQLite database path")
	jsonOutput := fs.Bool("json", false, "write machine-readable JSON")
	formatOutput := fs.String("format", "markdown", "report format: markdown or json")
	confirmImport := fs.Bool("yes", false, "confirm replacing the stored graph during import")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if args[0] != "import" && fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	path, err := config.DatabasePath(*dbPath)
	if err != nil {
		return err
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
	if len(graph.Assets)+len(graph.Identities) == 0 {
		history, historyErr := store.ImportHistory(ctx)
		if historyErr != nil {
			return historyErr
		}
		if len(history) == 0 {
			if err = app.LoadDemo(ctx); err != nil {
				return fmt.Errorf("initialize demo data: %w", err)
			}
			graph, err = store.Graph(ctx)
			if err != nil {
				return err
			}
		}
	}
	history, err := store.ImportHistory(ctx)
	if err != nil {
		return err
	}
	mode := "local data"
	if len(history) > 0 && history[0]["source"] == "demo fixture" {
		mode = "demo data"
	}
	switch args[0] {
	case "overview":
		summary, err := app.Overview(ctx)
		if err != nil {
			return err
		}
		return printValue(stdout, *jsonOutput, map[string]any{"mode": mode, "assets": len(graph.Assets), "identities": len(graph.Identities), "relationships": len(graph.Edges), "paths": summary.Paths, "paths_by_impact": summary.PathsByImpact, "confidence_distribution": summary.Confidence, "remediation": summary.Remediation})
	case "assets":
		return printValue(stdout, *jsonOutput, graph.Nodes())
	case "paths":
		if len(graph.Assets)+len(graph.Identities) == 0 {
			return errors.New("no graph nodes are available")
		}
		analysis, err := app.Analyze(ctx, 8, 100)
		if err != nil {
			return err
		}
		return printValue(stdout, *jsonOutput, analysis.Paths)
	case "remediation":
		analysis, err := app.Analyze(ctx, 16, 1000)
		if err != nil {
			return err
		}
		return printValue(stdout, *jsonOutput, analysis.Remediation)
	case "report":
		format := *formatOutput
		if *jsonOutput {
			format = "json"
		}
		data, err := app.RenderReport(ctx, format)
		if err != nil {
			return err
		}
		_, err = stdout.Write(data)
		return err
	case "import":
		if fs.NArg() > 1 {
			return fmt.Errorf("import accepts at most one file argument, got %d", fs.NArg())
		}
		input := io.Reader(os.Stdin)
		if len(graph.Assets)+len(graph.Identities) > 0 && !*confirmImport {
			if !launcher.InteractiveStdin() {
				return errors.New("import replaces the current graph; pass --yes to confirm in non-interactive use")
			}
			fmt.Fprint(stdout, "Import replaces the current graph. Continue? [y/N]: ")
			confirmationReader := bufio.NewReader(os.Stdin)
			answer, readErr := confirmationReader.ReadString('\n')
			if fs.NArg() == 0 {
				input = confirmationReader
			}
			if readErr != nil && readErr != io.EOF {
				return fmt.Errorf("read import confirmation: %w", readErr)
			}
			if !strings.EqualFold(strings.TrimSpace(answer), "y") && !strings.EqualFold(strings.TrimSpace(answer), "yes") {
				fmt.Fprintln(stdout, "Import cancelled.")
				return nil
			}
		}
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

func runNotifier(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: attackpath-visualizer notifier run|install|uninstall|status")
	}
	switch args[0] {
	case "install":
		flags := flag.NewFlagSet("notifier install", flag.ContinueOnError)
		flags.SetOutput(os.Stderr)
		dbPath := flags.String("db", "", "SQLite database path")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return fmt.Errorf("unexpected notifier install arguments: %s", strings.Join(flags.Args(), " "))
		}
		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("find application executable: %w", err)
		}
		resolvedDB, err := config.DatabasePath(*dbPath)
		if err != nil {
			return err
		}
		if _, err = os.Stat(resolvedDB); err != nil {
			return fmt.Errorf("initialize the local database before installing background notifications: %w", err)
		}
		path, err := notifier.Install(executable, resolvedDB)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Opt-in background notifier installed at %s. It reads the local database and makes no network calls.\n", path)
		return nil
	case "uninstall":
		path, err := notifier.Uninstall()
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Background notifier entry removed: %s\n", path)
		return nil
	case "status":
		path, enabled, err := notifier.Status()
		if err != nil {
			return err
		}
		state := "disabled"
		if enabled {
			state = "installed"
		}
		fmt.Fprintf(stdout, "Notifier %s · %s\n", state, path)
		return nil
	case "run":
		flags := flag.NewFlagSet("notifier run", flag.ContinueOnError)
		flags.SetOutput(os.Stderr)
		dbPath := flags.String("db", "", "SQLite database path")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return fmt.Errorf("unexpected notifier arguments: %s", strings.Join(flags.Args(), " "))
		}
		return notifier.RunUntilSignal(*dbPath)
	default:
		return fmt.Errorf("unknown notifier action %q; use run, install, uninstall, or status", args[0])
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
	fmt.Fprintf(w, "%s %s\n%s\n\nUsage: attackpath-visualizer <command> [options]\n\nCommands:\n  overview      Summarize local demo or imported data\n  paths         List ranked potential paths\n  remediation   Rank path-breaking relationship suggestions\n  report        Write a JSON or Markdown analysis report\n  assets        List inventoried assets and identities\n  import [file] Import a graph JSON file or stdin\n  export        Export the stored graph as JSON\n  about         Show version and project information\n  diagnostics   Show a sanitized local runtime report\n  notifier      Run or manage the opt-in background notifier\n  config reset-ui  Clear the remembered launcher choice\n  help          Show this help\n\nOptions: --db PATH, --json, --format markdown|json (report), --yes (confirm graph replacement)\n", meta.Name, meta.Version, meta.Purpose)
}
