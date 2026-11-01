# Offline usage guide

AttackPath Visualizer analyzes relationships from synthetic or authorized inventory. It does not scan systems, exploit vulnerabilities, or contact remote services. Its scores are human-review prioritization aids, not proof of exploitability.

## Choose a front end

Launch with no arguments to choose Dashboard or Terminal. `--ui gui` and `--ui tui` select directly; `--ask` opens the chooser and `--no-prompt` uses the saved/default choice. `ATTACKPATH_VISUALIZER_UI=gui|tui` overrides the saved preference. In a headless session, the terminal front end is selected automatically. Explicit commands such as `overview`, `paths`, `assets`, `import`, `export`, `about`, and `diagnostics` do not open the launcher. Use `config reset-ui` to clear the remembered choice.

## Local data and imports

The first launch creates a local SQLite database under the OS user configuration directory and loads the clearly labeled, deterministic synthetic graph. Use Settings or the TUI Settings view to inspect the active database path. `--db PATH` chooses another database; `ATTACKPATH_VISUALIZER_DB` is the environment override.

Graph JSON supports `assets`, `identities`, `misconfigurations`, `findings`, and `edges`. Node impact and finding severity are integers from 1 to 5; edge confidence is 0–100. Edges connect existing, distinct asset or identity IDs and require source, finding, and observation evidence. Duplicate IDs, missing endpoints, self-edges, invalid ranges, malformed or multiple JSON documents, entity-count excesses, and documents over 8 MiB are rejected. Validation completes before graph replacement. In the CLI, use `import --yes FILE` for a confirmed non-interactive replacement; an interactive terminal asks before replacing. The desktop import dialog also asks for confirmation. A rejected import is recorded without partially replacing the current graph.

The CLI `export` command writes graph JSON to standard output. The CLI `report` command writes a JSON or Markdown path-and-remediation report (`--format markdown|json`, or `--json`). In the TUI, Reports asks which format and destination to use and confirms before overwriting. In the dashboard Reports view, export the graph or complete analysis report to a user-selected file, or explicitly confirm copying JSON or Markdown to the clipboard; these exports may contain inventory identifiers and evidence.

## Dashboard categories

- **Overview** summarizes stored assets, identities, relationships, impact and confidence distributions, ranked paths, remediation choke points, and import freshness. Distribution tiles filter the ranked paths.
- **Graph** shows a scrollable relationship canvas, accessible asset-and-identity list, edge evidence, zoom controls, and selected-node details including linked findings.
- **Paths** ranks directed paths and shows confidence, impact, exact score components, and per-edge evidence; selecting a row opens its local evidence. Filtering is available by impact and confidence band.
- **Assets** lists assets and identities with type, impact, description, and linked findings.
- **Evidence** links relationships to their source record and finding text.
- **Remediation** ranks relationships by paths affected, critical paths, and weighted risk points. Suggestions are advisory and never change inventory.
- **Imports** validates local JSON and lists recent import outcomes.
- **Reports** exports the graph or ranked paths to JSON files or, after confirmation, the clipboard.
- **Notifications** lists local app events with severity, read state, global-search filtering, and links to related records.
- **Settings** controls light/dark appearance, compact table rows, launcher preference, local database location, notification rules, and opt-in OS notification delivery.
- **About** bundles the README, this Usage guide, keyboard shortcuts, sanitized diagnostics, application identity, and offline license texts.

Click a table heading to sort by that column. The global search field filters relevant assets, paths, evidence, notifications, and graph nodes. Use the sidebar buttons for every category; `Ctrl/Cmd+1` through `Ctrl/Cmd+9` opens the first nine categories, `Ctrl/Cmd+F` focuses search, `Ctrl/Cmd+,` opens Settings, and `F1` opens About. On macOS, `Cmd` is used instead of `Ctrl`.

## CLI

Run `attackpath-visualizer help` for all commands. Examples:

```sh
attackpath-visualizer overview --json
attackpath-visualizer paths --db ./graph.db --json
attackpath-visualizer remediation --db ./graph.db --json
attackpath-visualizer report --format markdown --db ./graph.db
attackpath-visualizer report --format json --db ./graph.db
attackpath-visualizer import --yes inventory.json
attackpath-visualizer export --db ./graph.db
attackpath-visualizer notifier status
attackpath-visualizer diagnostics
```

`about` and `diagnostics` do not include imported file contents or secrets. Diagnostics report only application/build version, OS/architecture, Go runtime, and sanitized local configuration status.

## Scoring and limits

A path's impact component is the highest impact tier among its asset and identity nodes. Confidence is compounded across edges in path order, flooring the integer percentage after each edge; score is impact tier multiplied by the resulting confidence percentage. The explanation and report expose each edge confidence and the final components. Enumeration is cycle-safe, deduplicated, bounded to 32 edges and 1,000 results, and also has a 250,000-edge-expansion budget; limited results identify truncation. Review inventory freshness, relationships, and evidence before acting.

## Notifications and privacy

In-app notifications are local. The background notifier is off until explicitly enabled in Settings or installed with `notifier install`. It creates a per-user startup entry without administrator rights and polls every 15 minutes. It opens SQLite in read-only mode; a separate private configuration file records which event IDs were delivered, so the graph database is not modified by the notifier. Severity thresholds, quiet hours, and minimal/detailed message preference are configurable per event type. Notification delivery makes no network calls. Use `notifier uninstall` to remove the startup entry; `notifier status` reports whether it is present.
