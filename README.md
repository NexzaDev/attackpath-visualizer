# AttackPath Visualizer

AttackPath Visualizer is a local-first defensive tool for exploring potential exposure paths across assets, identities, configuration findings, and evidence. It is an inventory relationship analyzer—not an attack executor, vulnerability scanner, or proof of exploitability. It makes no network requests at runtime and requires no account, paid service, API key, or cloud component.

The original product brief and acceptance criteria are preserved unchanged in [`docs/SPECIFICATION.md`](docs/SPECIFICATION.md). The implementation status and platform verification are tracked in [`docs/ROADMAP.md`](docs/ROADMAP.md) and the project pull request.

## Features

- Native Go command-line interface and Fyne desktop dashboard, plus a terminal menu; choose a front end at launch, remember a choice, or override it with `--ui=gui`, `--ui=tui`, or `--ask`.
- Directed asset graph with validation, explainable potential paths, evidence, confidence, impact, severity, remediation summaries, and deterministic offline demo data.
- Local SQLite storage, schema migrations, bounded JSON import, atomic graph replacement, import history, and JSON export.
- Overview, graph workspace, paths, assets, evidence, remediation, imports, reports, notifications, settings, About, usage guide, and sanitized diagnostics views.
- In-app event notifications and optional per-user OS notifications. The background notifier is off by default, uses a read-only database connection and a private local delivery ledger, and makes no network calls.
- Custom application icon, offline Noto Sans font assets, light/dark themes, keyboard shortcuts, and accessible text alternatives.
- CI coverage for Linux, macOS, and Windows, software-rendered dashboard screenshots, and a manually triggered draft-release workflow.

## Requirements

- Go 1.25 or newer.
- Native Fyne GUI build prerequisites. On Ubuntu, CI installs `bc`, `gcc`, `libgl1-mesa-dev`, `libwayland-dev`, `libx11-dev`, `libxkbcommon-dev`, `xorg-dev`, `xvfb`, and `language-pack-en`. These are build/test prerequisites only; the CLI-only build does not require the GUI stack.
- No network access is required to run the application after it has been built. First-time Go module downloads and release packaging do need network access.

## Run

From the repository root:

```sh
go run ./cmd/attackpath-visualizer --ask
go run ./cmd/attackpath-visualizer --ui=tui
go run ./cmd/attackpath-visualizer --ui=gui
go run ./cmd/attackpath-visualizer overview --json
go run ./cmd/attackpath-visualizer paths --json
go run ./cmd/attackpath-visualizer help
```

Launching with no subcommand opens the configured front end or asks which one to use when appropriate. A remembered choice can be cleared with `go run ./cmd/attackpath-visualizer config reset-ui`. `ATTACKPATH_VISUALIZER_UI=gui|tui` overrides the saved preference. Use `--db PATH` after a CLI subcommand (for example, `overview --db ./graph.db`) or after a launcher option (for example, `--ui=gui --db ./graph.db`) to choose a different SQLite file; `ATTACKPATH_VISUALIZER_DB` is the corresponding environment override.

The first launch initializes a clearly labeled synthetic demo graph. Demo data is deterministic and stays local. It is safe to replace it with authorized inventory through the Import view or the CLI. Imports accept JSON up to 8 MiB, validate the complete graph before changing stored data, and replace the current graph only after explicit confirmation. The CLI form is `attackpath-visualizer import --yes inventory.json`; omit `--yes` in an interactive terminal to confirm at the prompt. Export writes the local graph as JSON to standard output.

The graph schema has `assets` and `edges` arrays. An asset has a unique `id`, `name`, `kind`, impact from 1 to 5, and optional description. An edge has a unique `id`, `from`, `to`, relationship, confidence from 0 to 100, and optional source/finding/observed evidence. Self-edges, missing endpoints, duplicate IDs, invalid ranges, malformed JSON, and oversized imports are rejected. Paths are bounded simple directed paths, ranked by maximum asset impact multiplied by compounded edge confidence. Scores are prioritization aids, not exploitability claims.

## Build and verify

The commands below work from the repository root on each native OS. The default build includes the desktop GUI; the `nogui` build is for servers and terminals.

### Linux and macOS

```sh
mkdir -p bin
go test ./...
go vet ./...
go build -o ./bin/attackpath-visualizer ./cmd/attackpath-visualizer
go build -tags nogui -o ./bin/attackpath-visualizer-cli ./cmd/attackpath-visualizer
go run ./cmd/attackpath-visualizer --ask
```

On Ubuntu, install the Fyne native build prerequisites listed above before building the GUI.

### Windows PowerShell

```powershell
New-Item -ItemType Directory -Force bin | Out-Null
go test ./...
go vet ./...
go build -o .\bin\attackpath-visualizer.exe .\cmd\attackpath-visualizer
go build -tags nogui -o .\bin\attackpath-visualizer-cli.exe .\cmd\attackpath-visualizer
go run .\cmd\attackpath-visualizer --ask
```

For formatting in PowerShell, run `Get-ChildItem -Recurse -Filter *.go | ForEach-Object { gofmt -w $_.FullName }`. Native Fyne builds on Windows require a C compiler; CI installs MinGW-w64 GCC.

The common full check sequence is:

```sh
gofmt -w $(find . -name '*.go' -not -path './.git/*')
go vet ./...
go test ./...
go test -race ./...
go build -tags nogui ./...
go build ./...
go run ./cmd/render-screenshots -out artifacts/ui
```

`make check` runs the local formatting, analysis, test, and build gate. `go build -tags nogui ./...` builds the server-friendly CLI/TUI variant; `go build ./...` includes the native GUI. The GitHub Actions workflow runs the platform matrix on Go 1.27.3, uses Xvfb/software rendering for headless Linux UI checks, and uploads PNG screenshots as `attackpath-visualizer-ui-screenshots`. The branch-push lockfile workflow runs `go mod tidy` and commits any `go.mod`/`go.sum` updates. CI also checks module metadata and verifies downloaded modules. Screenshot artifacts are design evidence, not proof of desktop runtime behavior. Check the pull request for actual per-OS CI results before relying on a build.

## Notifications

Notifications are opt-in. In Settings, explicitly enable the background notifier, or run `attackpath-visualizer notifier install --db PATH`. The available commands are `notifier run`, `notifier install`, `notifier uninstall`, and `notifier status`. Installation creates a per-user startup entry without administrator privileges. The background process polls the local database every 15 minutes, observes configured severity and quiet-hour rules, and stores delivery de-duplication state in the OS user configuration directory. It does not write to the graph database or make network calls. In-app notices remain available when OS delivery is off.

## Release packages

Maintainers run the `Release packages` workflow manually from the default branch in GitHub Actions. It packages the GUI on native Windows, macOS ARM64, and Linux x64 runners, creates the CLI-only package, includes offline documentation and dependency license material, calculates `SHA256SUMS`, and attaches the results to a **draft pre-release**. The workflow does not publish a release automatically. See [`docs/RELEASING.md`](docs/RELEASING.md) for exact phone/desktop GitHub steps and unsigned-binary warnings. No signing or notarization is claimed.

## Data, security, and recovery

- The database and user preferences remain on the local machine. No telemetry or third-party integration is configured.
- Import validation completes before a transaction replaces stored graph rows. Keep an exported JSON backup before importing data you cannot recreate.
- Diagnostics contain only application/build version, OS/architecture, Go runtime, and a sanitized local configuration status. Review the report before sharing it.
- Use only systems and inventories you are authorized to assess. A relationship graph can be incomplete or stale and should be reviewed by a human.
- See [`docs/THREAT_MODEL.md`](docs/THREAT_MODEL.md), [`docs/USAGE.md`](docs/USAGE.md), and [`SECURITY.md`](SECURITY.md) for security boundaries, detailed workflows, and reporting guidance.

## License

The application is MIT-licensed; third-party and font license texts are listed in [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md), [`docs/LICENSES.md`](docs/LICENSES.md), and the `licenses/` directory. Release builds gather the platform-specific transitive Go module licenses. See [`LICENSE`](LICENSE).
