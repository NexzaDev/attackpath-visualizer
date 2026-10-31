# Offline usage guide

AttackPath Visualizer analyzes relationships from synthetic or authorized inventory. It does not scan systems, exploit vulnerabilities, or contact remote services.

## Choose a front end

Launch with no arguments to choose Dashboard or Terminal. `--ui gui` and `--ui tui` select directly; `--ask` shows the chooser, `--no-prompt` uses the saved/default choice, and `ATTACKPATH_VISUALIZER_UI` can set a preference. A headless session uses the terminal interface. Explicit commands such as `overview`, `paths`, `assets`, `import`, `export`, `about`, and `diagnostics` do not show the launcher.

## Data and imports

The first launch creates a local SQLite database and loads labeled synthetic demo data. Use Settings to inspect the data location. JSON graph files contain an `assets` array and an `edges` array. Asset impacts are 1–5; edge confidence is 0–100. Imports are validated and replace the graph in one transaction. A malformed file is rejected without partial replacement. The maximum import size is 8 MiB.

## Dashboard categories

- **Overview** summarizes stored assets, edges, path rankings, high-impact records, and import freshness.
- **Graph** displays the relationship inventory and a keyboard-accessible node list; selecting a node focuses its direct relationships.
- **Paths** shows ranked directed paths, confidence, impact, and the exact score breakdown; selecting a path opens its edge evidence.
- **Assets** lists stored asset identity, type, impact, and description.
- **Evidence** links every relationship to its source record and finding text.
- **Remediation** ranks edges by how many calculated paths they participate in; suggestions are advisory and never modify inventory.
- **Imports** validates JSON files and lists recent import outcomes.
- **Reports** exports the stored graph or ranked paths to a user-selected JSON file.
- **Notifications** lists local app events, severity, read state, and links to related records.
- **About** provides this guide, shortcuts, diagnostics, source links, and dependency licenses.
- **Settings** controls appearance and launcher preference.

## CLI

Run `attackpath-visualizer help` for commands. `--json` prints machine-readable output. `--db PATH` selects a separate SQLite file. `import FILE` or `import --db PATH < inventory.json` imports local JSON. `export` writes the graph as JSON. `about` and `diagnostics` do not include imported file contents or secrets.

## Scoring and limits

A path's impact is its highest asset impact. Confidence is the product of edge confidences, compounded in order; score is impact multiplied by confidence. The score is a prioritization heuristic, not exploitability probability. Path enumeration is bounded by depth and result limits. Review all evidence and inventory context before acting.

## Notifications and privacy

In-app notices are local. Background notifications are disabled until explicitly enabled or installed from the CLI. The notifier reads the database read-only on a bounded interval, displays minimal local text, and makes no network requests. It does not place evidence, credentials, or indicators in notification text.
