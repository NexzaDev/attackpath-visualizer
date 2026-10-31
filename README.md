# AttackPath Visualizer

A local-first defensive graph-analysis application for relating assets, identities, configuration findings, and evidence. It computes explainable potential exposure paths from synthetic or authorized inventory; it does not execute attacks or make network requests.

## Current status

The current implementation provides the Go graph engine, validated JSON import, durable SQLite storage, deterministic demo inventory, and CLI commands. The native dashboard, launcher, notifier, and release packages are under active development and are not yet considered verified. See [the roadmap](docs/ROADMAP.md) and [the original specification](docs/SPECIFICATION.md).

## Requirements and run

Install Go 1.23 or newer. No account, service, API key, or network connection is needed at runtime. From a clean checkout:

```sh
go test ./...
go run ./cmd/attackpath-visualizer overview
go run ./cmd/attackpath-visualizer paths --json
go run ./cmd/attackpath-visualizer import inventory.json
```

The first run creates an SQLite database beneath the OS user configuration directory and loads a clearly labeled synthetic sample graph. Use `--db PATH` to select a different database. `import` accepts a JSON file or standard input. Input is limited to 8 MiB and validated before the existing graph is replaced. `export` writes JSON to standard output. Commands include `help`, `about`, and `diagnostics`. CLI commands work on Linux, macOS, and Windows; the current GUI capability and platform verification status are tracked separately in the roadmap.

Build the CLI with `go build -tags nogui ./cmd/attackpath-visualizer`. Build/test all packages with `go test ./...`; run `go vet ./...` and `gofmt -w` as part of local checks. `make check` runs formatting, analysis, tests, and build. Windows users can run these commands in PowerShell after installing Go and Make (or invoke the corresponding Go commands directly).

## Data model

A graph JSON document has `assets` and `edges` arrays. Asset impact is an integer from 1 to 5. Edge confidence is 0–100 and each edge may hold a source, finding, and observed timestamp as evidence. Paths are simple directed paths, bounded to 32 edges, scored as maximum asset impact multiplied by the compounded edge confidence. Output contains the exact explanation and edge evidence. The demo fixture is in `testdata/demo.json`.

## Privacy and limitations

The application has no telemetry and no external integrations. SQLite is stored locally. Diagnostics intentionally include only application/build version, OS/architecture, runtime, and a sanitized local configuration status. Current imports replace the graph atomically; multi-source history and incremental merge are not yet implemented. Potential paths are graph relationships, not proof of exploitability. See [threat model](docs/THREAT_MODEL.md).

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md), [architecture](docs/ARCHITECTURE.md), and [release procedures](docs/RELEASING.md). CI checks Go formatting, static analysis, tests, native builds, and dependency vulnerabilities on Linux, macOS, and Windows.
