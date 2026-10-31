# Offline usage guide

AttackPath Visualizer analyzes relationships from synthetic or authorized inventory. It does not scan systems, exploit vulnerabilities, or contact remote services. Its scores are human-review prioritization aids, not proof of exploitability.

## Choose a front end

Launch with no arguments to choose Dashboard or Terminal. `--ui gui` and `--ui tui` select directly; `--ask` opens the chooser and `--no-prompt` uses the saved/default choice. `ATTACKPATH_VISUALIZER_UI=gui|tui` overrides the saved preference. In a headless session, the terminal front end is selected automatically. Explicit commands such as `overview`, `paths`, `assets`, `import`, `export`, `about`, and `diagnostics` do not open the launcher. Use `config reset-ui` to clear the remembered choice.

## Local data and imports

The first launch creates a local SQLite database under the OS user configuration directory and loads the clearly labeled, deterministic synthetic graph. Use Settings or the TUI Settings view to inspect the active database path. `--db PATH` chooses another database; `ATTACKPATH_VISUALIZER_DB` is the environment override.

Graph JSON contains `assets` and `edges` arrays. Asset impact is an integer from 1 to 5; edge confidence is 0–100. Edges connect existing, distinct asset IDs and can include evidence source, finding text, and observation time. Duplicate IDs, missing endpoints, self-edges, invalid ranges, malformed JSON, and documents over 8 MiB are rejected. Validation completes before graph replacement. In the CLI, use `import --yes FILE` for a confirmed non-interactive replacement; an interactive terminal asks before replacing. The desktop import dialog also asks for confirmation. A rejected import is recorded without partially replacing the current graph.

The CLI `export` command writes JSON to standard output. In the TUI, Reports asks for a destination and confirms before overwriting. In the dashboard Reports view, use export to a file or explicitly confirm copying graph/path JSON to the clipboard; those exports may contain inventory identifiers and evidence.

## Dashboard categories

- **Overview** summarizes stored assets, relationships, ranked paths, high-impact records, and import freshness.
- **Graph** shows a scrollable relationship canvas, accessible node list, edge evidence, zoom controls, and selected-node details.
- **Paths** ranks directed paths and shows confidence, impact, and the score breakdown; selecting a row opens its local evidence.
- **Assets** lists identity, type, impact, and description.
- **Evidence** links relationships to their source record and finding text.
- **Remediation** ranks relationships by the number of enumerated paths they participate in. Suggestions are advisory and never change inventory.
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
attackpath-visualizer import --yes inventory.json
attackpath-visualizer export --db ./graph.db
attackpath-visualizer notifier status
attackpath-visualizer diagnostics
```

`about` and `diagnostics` do not include imported file contents or secrets. Diagnostics report only application/build version, OS/architecture, Go runtime, and sanitized local configuration status.

## Scoring and limits

A path's impact is the highest impact among its assets. Confidence is compounded across edges in order (integer rounding is applied at each step); score is impact multiplied by the resulting confidence. Path enumeration is bounded to 32 edges and 1,000 results. Review inventory freshness, relationships, and evidence before acting.

## Notifications and privacy

In-app notifications are local. The background notifier is off until explicitly enabled in Settings or installed with `notifier install`. It creates a per-user startup entry without administrator rights and polls every 15 minutes. It opens SQLite in read-only mode; a separate private configuration file records which event IDs were delivered, so the graph database is not modified by the notifier. Severity thresholds, quiet hours, and minimal/detailed message preference are configurable per event type. Notification delivery makes no network calls. Use `notifier uninstall` to remove the startup entry; `notifier status` reports whether it is present.
