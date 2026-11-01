# AttackPath-Visualizer — Exposure Relationship Graph

**Repository:** `attackpath-visualizer`

**GitHub description:** A defensive graph-analysis tool that connects assets, identities, misconfigurations, and exposure findings to explain potential attack paths.

## Language stack

| Layer | Choice | Responsibility |
|---|---|---|
| Core/backend | Go | Domain logic, performance-sensitive work, APIs, integrations, and background tasks |
| Interface | Go CLI/TUI + native desktop dashboard GUI with Fyne (required) | Native desktop application plus terminal UI, chosen at launch; no browser frontend |
| Persistence | SQLite with indexed relational tables and adjacency lists | Durable application data and queryable history |

Go owns ingestion, graph query orchestration, explainable scoring, and APIs. SQLite stores relationships; the terminal UI renders paths.

## What this repository is

This README contains the project specification and the complete prompt intended for an AI coding agent. It does not claim the application has already been implemented.

## Product scope

Build graph-based risk exploration that ingests synthetic or authorized inventory, models assets and relationships, computes explainable potential paths, ranks them by impact/confidence, and links edges to source evidence. Include demo data and remediation suggestions without executing attacks.

## Non-negotiable delivery constraints

- This is **not a website, web app, SaaS product, or hosted dashboard**. Do not create a browser frontend, React app, or public-facing web server.
- Deliver a local-first application with two front ends over one service layer: a polished **native desktop dashboard GUI** and a **CLI/TUI**, chosen at launch. No website or browser frontend. Any optional listener must be disabled by default and bind to loopback only unless explicitly configured.
- The complete core workflow must run for **€0** on a normal developer computer. Use free/open-source tools and libraries, with no required paid API keys, subscriptions, cloud accounts, hosted databases, or commercial services.
- Default to local files and SQLite where persistence is needed. External integrations must be optional adapters, never required for core functionality; provide fixtures and a fully usable offline/demo mode.
- Do not require paid CI, hosting, domains, premium IDEs, or paid threat-intelligence feeds. Document optional free-tier limits honestly and never make a free-tier account mandatory.
- Use a free, OS-supported build/test workflow. State the exact local prerequisites and commands needed to build, test, and run the project.

## Cross-platform and product-quality requirements

- **Supported operating systems:** Linux, macOS, and Windows. The project must have documented build, test, and run instructions for all three, including architecture-specific notes where relevant.
- **Fast, suitable implementation languages:** Go is the default for network tools, collectors, services, and concurrent workloads; Rust is preferred for memory-sensitive, parsing-heavy, or performance-critical components. Do not use Python as the main implementation language. Use it only for optional developer scripts if there is a clear reason.
- **No website:** Do not create a website, hosted dashboard, SaaS product, browser frontend, or public web server. This is a local application/tool.
- **Professional interface:** Provide a polished CLI/TUI with consistent command naming, useful help, clear errors, structured output, progress indicators, and accessible color handling, and a required native desktop dashboard GUI built with Fyne, not a browser-based UI. Follow the mandatory desktop application UI section for the launcher, navigation, icons, About, and design rules. Visual tools must have a real interface, not just a placeholder screen.
- **Distinctive, production-quality UX:** Use a coherent design system, deliberate spacing and typography, restrained colors, meaningful empty/loading/error/success states, keyboard navigation, sensible defaults, accessible contrast, responsive layouts, and realistic sample data. Avoid generic AI-generated dashboard templates, gratuitous gradients, meaningless charts, fake metrics, placeholder buttons, lorem ipsum, and decorative features that do not work.
- **Real functionality only:** Every visible control must work. Use representative local fixtures for demos; label demo data clearly. Do not invent integrations, results, benchmarks, or completed tests.
- **Cross-platform verification:** CI must build and test on Linux, macOS, and Windows where supported. Use portable filesystem paths, signal handling, process management, permissions, and networking APIs. Test path handling and platform-specific behavior. Do not claim a platform is supported until its build/test path is documented and verified.
- **€0 operation:** Core functionality must run locally using free/open-source dependencies without paid API keys, subscriptions, cloud accounts, hosted databases, domains, or commercial services. External integrations are optional and must have an offline/demo alternative.
- **Performance discipline:** Profile before optimizing; use bounded concurrency, cancellation, timeouts, streaming for large inputs, and benchmarks for genuinely hot paths. Avoid unnecessary allocations and heavyweight dependencies, but do not trade correctness or maintainability for synthetic benchmark numbers.
- **No source-code comments:** Keep source files free of comments as requested. Express design intent through clear names, small cohesive modules, tests, and documentation outside source files.
- **No AI slop:** Do not fabricate features or inflate scope/LOC. Prefer a small set of complete, tested workflows over many incomplete screens. Use consistent naming, idiomatic language patterns, explicit error handling, deterministic tests, dependency/license review, and a clean release process.

## Product polish, branding, and support features

- **Custom icon and identity:** Create an original project-specific application icon (required; no generic placeholder). Provide editable SVG source artwork plus platform assets (Windows `.ico`, macOS `.icns`, Linux SVG/PNG and desktop-entry assets) and use it for the window, taskbar or dock, packaging metadata, and About. Do not use third-party trademarks.
- **Source access:** GUI applications must have a visible **Source Code** action in About that opens the real public GitHub repository URL configured for the project. Keep the URL in one central metadata/config location. If it is not configured, show an honest fallback rather than a fabricated URL. CLI-only tools must provide an `about` command with version, license, project URL when configured, OS and build details.
- **About page/dialog:** Every GUI application must include an About screen/dialog showing app name, icon, version, purpose, license, maintainer/copyright only when known, repository link, dependency notices, and build/runtime information where useful. Include a **“Found a bug? Please report it on GitHub.”** action that opens the repository issue tracker when configured. Never fabricate repository URLs.
- **Support and diagnostics:** Include a previewable, copyable diagnostic report with app version, OS, architecture, runtime/build versions and sanitized configuration status. Do not include credentials, tokens, personal file contents, or sensitive host/network data by default.
- **Useful product features:** Contextual help, searchable help for complex tools, keyboard shortcuts, copy/export actions, safe settings defaults, first-run guidance, actionable errors, cancellation, confirmation for destructive actions, and clear empty/loading/error/success states.
- **Accessibility:** Keyboard navigation, readable contrast, scalable text, visible focus indicators, and screen-reader labels where supported. Do not rely on color alone to communicate severity.
- **Security and privacy:** No telemetry by default. Disclose external network requests, redact secrets from logs, use OS-appropriate config directories and restrictive permissions, and provide dry-run mode for potentially disruptive operations.
- **Release quality:** Versioning, changelog, checksums and signed release artifacts where feasible with free tools, reproducible build guidance, dependency/license inventory, uninstall/reset instructions, and platform-specific troubleshooting.
- **Honest completeness:** No fake buttons, fake charts, fabricated performance numbers, placeholder links, or screens disconnected from real functionality. Label demo data clearly and omit features that are not implemented.

## Application interface: native desktop dashboard and terminal UI

AttackPath Visualizer ships as one local application with two front ends over the same service layer: a native desktop dashboard with a graph workspace (built with Fyne) and a terminal UI. At launch the user chooses which one to open. There is no website, browser frontend, or hosted service. The full rules are in the mandatory desktop application UI section of the prompt below.

| Category | Icon (Lucide-style name) | What it shows |
|---|---|---|
| Overview | `layout-dashboard` | Exposure summary: potential paths by impact tier, top-ranked paths, choke points, import freshness |
| Graph | `waypoints` | Interactive relationship graph of assets and edges with pan, zoom, filter, and focus |
| Paths | `route` | Ranked potential paths with an explainable score breakdown |
| Assets | `server` | Asset inventory and relationships |
| Evidence | `file-search` | Source evidence behind every edge |
| Remediation | `wrench` | Ranked suggestions, for example the edge whose removal breaks the most paths |
| Imports | `upload` | Ingestion history, validation results, demo-data label |
| Reports | `file-text` | Exported path and remediation reports |
| About | `info` | App identity, offline Usage guide, Source Code, bug reporting |
| Settings | `settings-2` | Appearance, density, launch preference, data location, safe defaults |

## Engineering quality, correctness, and release gates

- **No promise of zero bugs:** Software cannot be guaranteed bug-free. The engineering process must systematically reduce defects, catch regressions, and document known limitations instead of claiming perfection.
- **Requirements before implementation:** Convert every feature into acceptance criteria and test cases. Keep a traceable checklist linking requirements to code, tests, documentation, and release status.
- **Small, complete increments:** Implement vertical slices that compile and work end-to-end. Do not generate the whole application in one unverified pass or leave placeholder modules and stubbed workflows.
- **Compile and test continuously:** After each meaningful change, run formatter, compiler, static analysis/linter, unit tests, integration tests, and the relevant platform build. Fix failures before moving forward; never suppress a failure just to get a green result.
- **Defensive engineering:** Validate all external input, handle malformed files and partial network failures, set explicit timeouts, cancel work cleanly, bound concurrency and memory, avoid races, and return actionable errors. Never silently swallow errors.
- **Test design:** Cover normal behavior, boundary cases, invalid input, empty input, large inputs, permission failures, cancellation, timeouts, corrupt state, offline mode, and regression cases. Use deterministic tests, fixtures, fuzz/property testing where appropriate, and race/concurrency checks for Go or Rust tools as applicable.
- **Cross-platform correctness:** Build and test on Linux, macOS, and Windows. Test paths, permissions, signals, line endings, process handling, sockets, and packaging on the actual target platforms. If a platform cannot be tested, disclose it instead of claiming full verification.
- **Security assurance:** Perform threat modeling, dependency vulnerability review, secret scanning, static analysis, safe parsing, least-privilege design, secure defaults, and checks for path traversal, command injection, unsafe deserialization, data leakage, and unintended network exposure.
- **Performance validation:** Set realistic performance goals and benchmark representative workloads with reproducible commands and hardware/context notes. Profile bottlenecks before optimization. Never invent benchmark numbers or sacrifice correctness for synthetic speed.
- **Code review and maintainability:** Use idiomatic Go/Rust, clear module boundaries, explicit ownership and error contracts, minimal dependencies, no dead code, no duplicated logic, no unexplained magic values, and no artificial LOC padding. Source code must contain no comments as requested; use descriptive names, tests, and external docs.
- **Release blocking gates:** A release must not be marked ready if required builds/tests/lints/security checks fail, core flows are untested, documentation is inconsistent, or critical/high-severity known defects remain unresolved without explicit risk acceptance.
- **Honest reporting:** Report exactly which commands ran, which platforms were tested, what passed/failed, any skipped checks, known issues, and remaining risks. Never claim a test, scan, benchmark, or build ran unless it actually ran.

## Proposed codebase layout

```text
attackpath-visualizer/
├── README.md
├── LICENSE
├── CONTRIBUTING.md
├── SECURITY.md
├── CODE_OF_CONDUCT.md
├── CHANGELOG.md
├── .editorconfig
├── .gitignore
├── .env.example
├── Makefile
├── cmd/
│   └── attackpath-visualizer/
├── internal/
│   ├── config/
│   ├── domain/
│   ├── service/
│   ├── repository/
│   ├── transport/
│   └── observability/
├── api/
├── tui/
│   ├── views/
│   ├── components/
│   ├── formatting/
│   └── tests/
├── migrations/
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── testdata/
├── scripts/
├── assets/
│   ├── icons/
│   │   ├── icon.svg
│   │   ├── icon.png
│   │   ├── icon.ico
│   │   └── icon.icns
│   └── branding/
├── docs/
│   ├── ARCHITECTURE.md
│   ├── THREAT_MODEL.md
│   └── ROADMAP.md
└── .github/
    ├── workflows/
    ├── ISSUE_TEMPLATE/
    └── PULL_REQUEST_TEMPLATE.md
```

## Engineering and open-source requirements

The implementation must include:
- `README.md`, `LICENSE` (MIT, subject to dependency/asset license review), `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, and `CHANGELOG.md`.
- `.github/workflows/` CI for formatting, linting/static analysis, tests, and builds on Linux, macOS, and Windows where supported, plus issue and pull request templates.
- `docs/ARCHITECTURE.md`, `docs/THREAT_MODEL.md`, and `docs/ROADMAP.md`.
- A `Makefile` or equivalent cross-platform task runner, `.editorconfig`, `.gitignore`, `.env.example`, tests, sanitized fixtures, centralized app/repository metadata, and documented release packaging for Linux/macOS/Windows.
- No comments in source code, no hardcoded secrets, no filler code, and no claims of passing tests unless the tests were run.

## Full prompt for Arena.ai/Agent

Copy the entire prompt below into Arena.ai/Agent:

---

# AttackPath-Visualizer — Exposure Relationship Graph

**Suggested GitHub repository:** `attackpath-visualizer`

**GitHub description:** A defensive graph-analysis tool that connects assets, identities, misconfigurations, and exposure findings to explain potential attack paths.

## Languages and technology decisions

- **Primary language:** Go for the core domain, concurrent network work, ingestion, and performance-sensitive processing. Choose Rust instead when memory safety or parsing throughput makes it a better fit. Do not use Python as the core application language.
- **User interface:** Go CLI/TUI + native desktop dashboard GUI with Fyne (required). Build a local application whose front end, desktop dashboard or terminal UI, is chosen at launch; do not build a website, browser frontend, SaaS, or hosted service.
- **Persistence:** SQLite with indexed relational tables and adjacency lists.
- **Language rationale:** Go owns ingestion, graph query orchestration, explainable scoring, and APIs. SQLite stores relationships; the terminal UI renders paths.
- Prefer compiled, statically typed, performance-conscious code for the core. Do not make Python the primary application language. Python may be used only for a clearly justified auxiliary script or offline experiment; it must not be required to run the main product.
- Choose mature, maintained libraries rather than writing cryptography, protocol implementations, parsers, or database abstractions from scratch.
- Pin or bound dependencies, explain major choices, and avoid adding frameworks or dependencies without a concrete need.

## Your role

Act as a principal software engineer, systems architect, cybersecurity specialist, network engineer, performance engineer, QA lead, and open-source maintainer. Work with the discipline expected of a production engineering team. Be persistent and confident, but prove progress with working code, measured behavior, reproducible tests, and accurate documentation. The goal is a high-quality open-source project, not an impressive-looking scaffold.

## Product mission

Build graph-based risk exploration that ingests synthetic or authorized inventory, models assets and relationships, computes explainable potential paths, ranks them by impact/confidence, and links edges to source evidence. Include demo data and remediation suggestions without executing attacks.

## Mandatory product constraints: no website and €0 operating cost

- Do not build a website, web application, browser frontend, SaaS, or hosted service. The deliverable is a local application with a native desktop dashboard GUI and a CLI/TUI, built only with free tools.
- All essential functionality must work locally for €0. Do not require paid APIs, paid cloud services, subscriptions, commercial databases, paid scanners, paid feeds, or paid developer tools.
- Use free/open-source dependencies and local storage (SQLite/files) by default. Optional integrations must have an offline alternative using documented fixtures; never make credentials mandatory for the basic workflow.
- Do not make the user deploy or host anything. No required Docker image registry, domain, VPS, cloud account, or externally hosted service.
- If a local HTTP endpoint is genuinely necessary for an integration, it must be optional, disabled by default, and bind to 127.0.0.1 by default. Prefer CLI/TUI interaction over HTTP.
- Document a zero-cost setup, build, test, and run path. Be explicit about optional integrations and any limits of free tiers.

## Mandatory cross-platform, professional UX, and zero-cost constraints

1. Support **Linux, macOS, and Windows**. Provide exact build, test, package, and run commands for each OS. Add CI jobs for all three platforms where technically supported, and report platform gaps honestly.
2. Use **Go** as the default primary language for network/security tooling and concurrent collectors. Use **Rust** when memory safety, parsing throughput, or performance makes it the better fit. Python must not be the core application language. Do not add multiple languages without a clear architectural reason.
3. Do not build a website, browser frontend, SaaS, hosted service, or public web server. Deliver a local-first application with a professional CLI/TUI and a required cross-platform **native desktop dashboard GUI** (Fyne); do not use a browser-based UI.
4. The UI must look intentionally designed and production-ready: consistent typography and spacing, a restrained palette, clear hierarchy, useful navigation, keyboard support, accessible contrast, real loading/empty/error/success states, useful filtering/search, and coherent interactions. Avoid generic AI dashboard templates, gratuitous gradients, fake charts/metrics, lorem ipsum, placeholder buttons, and visual clutter.
5. Every visible control must perform a real action. Demo data must be clearly labeled. All core workflows must work offline with local files/SQLite and included fixtures. Do not require paid API keys, paid services, subscriptions, cloud accounts, hosting, domains, or a hosted database.
6. Performance must be demonstrated with reproducible benchmarks and realistic workloads, not unsupported speed claims. Use bounded concurrency, cancellation, timeouts, streaming for large inputs, and profiling where appropriate.
7. Use portable filesystem, process, signal, permission, and networking behavior. Add tests for Windows path semantics and platform-specific code. Do not claim OS support until the relevant build/test workflow passes.
8. No source-code comments, no hardcoded secrets, no fabricated capabilities, no fake test results, and no artificial LOC padding. Prefer fewer complete features with high reliability over a large number of unfinished features.
9. Include an accessible local demo mode, deterministic fixtures, dependency and license review, release packaging instructions for Linux/macOS/Windows, and clear troubleshooting docs.
10. The core workflow must cost **€0**. Optional external integrations must be clearly marked optional and must have a usable offline alternative.

## Mandatory application identity, About, source, and reporting features

1. Create a **custom original application icon** for this project (required). Supply editable SVG source and platform assets: `.ico` for Windows, `.icns` for macOS, and SVG/PNG plus desktop-entry integration for Linux. No generic placeholder icon or third-party trademark.
2. Every GUI application must include an **About** screen/dialog accessible from the main menu or settings. Show app name, icon, version, purpose, license, maintainer/copyright only when known, build/runtime details where useful, and dependency/license notices.
3. Add a **Source Code** action in About that opens the actual GitHub repository URL configured for the project. Add a separate **“Found a bug? Please report it on GitHub.”** action that opens the repository issue tracker/new-issue page when configured. Store URLs centrally. Never invent or hardcode a fake URL; if the repository is not configured, show an honest fallback and disable the external action.
4. For CLI-only tools, provide an `about` command displaying name, version, license, repository URL if configured, OS/architecture, and build information. Provide `--help`, `--version`, stable exit codes, and machine-readable JSON output where useful.
5. Include a safe **Copy diagnostic report** feature for GUI apps and a `diagnostics` command for CLI tools. Preview before saving/copying. Include app version, OS/architecture, runtime/build details and sanitized settings status only. Never include secrets, tokens, personal file contents, or sensitive network data by default.
6. Add contextual help, searchable help where useful, keyboard shortcuts, copy/export, safe settings defaults, first-run guidance, actionable errors with next steps, cancellation, and confirmation for destructive actions.
7. No telemetry or analytics by default. Disclose external network requests, redact secrets from logs, use OS-appropriate config directories and restrictive file permissions, and provide dry-run mode for risky operations.
8. Every UI action must have real behavior and tests. No decorative controls, fake metrics, placeholder links, fabricated test results, or UI shell disconnected from functionality.
9. Add version/changelog and platform-specific install, upgrade, uninstall, reset, and troubleshooting instructions. Provide checksums and signed artifacts where feasible using free tools.
10. Keep all existing constraints: Linux/macOS/Windows support, high-performance Go/Rust core, no website/browser frontend, no required paid services, €0 core operation, no source-code comments, MIT license, SECURITY.md, CONTRIBUTING.md, tests, and professional visual design.

## Mandatory professional engineering workflow: correctness first

1. Do not attempt to generate the entire product in one pass. First write requirements, acceptance criteria, threat model, architecture decisions, and a prioritized implementation plan. Then implement small, complete, testable increments.
2. Before each feature is considered complete, ensure it compiles, passes its tests, handles failure cases, is documented, and works end-to-end. No TODO-only implementation, fake functionality, empty handlers, stubbed return values, or buttons that only look functional.
3. Maintain a requirements-to-tests checklist. Every requirement must map to an implementation and a test, or be explicitly marked out of scope. Add regression tests for every discovered bug.
4. Run and fix the full relevant quality suite: formatter, compiler/build, linter/static analysis, unit tests, integration tests, end-to-end tests where appropriate, dependency audit, secret scan, and platform builds. Do not suppress warnings/errors just to make CI pass.
5. Test valid, invalid, empty, boundary, malformed, oversized, interrupted, timed-out, offline, permission-denied, and partial-failure scenarios. Use deterministic fixtures. Add fuzz/property tests for parsers and input-heavy code; use race/concurrency checks where supported.
6. Validate all untrusted inputs and paths. Use timeouts, cancellation, bounded concurrency, explicit resource limits, safe defaults, least privilege, redacted logs, and actionable errors. Do not silently swallow errors or expose secrets in output.
7. Support Linux, macOS, and Windows with actual CI jobs where possible. Test OS-specific path handling, process management, permissions, signals, networking and packaging. State clearly when a platform or check has not been tested.
8. Benchmark only after correctness is established. Include reproducible benchmarks and describe the test workload and environment. Never invent speed claims, coverage percentages, test results, or security scan results.
9. Before release, require passing builds, tests, lint/static analysis and security checks; no unresolved critical/high-severity defects unless explicitly documented with an accepted risk and mitigation. Include release notes, known issues, rollback/recovery guidance, and checksums where feasible.
10. At the end of the work, provide a truthful quality report listing the exact commands run, platforms tested, pass/fail counts, skipped checks, known limitations, and remaining risks. Do not claim “bug-free” or “fully tested” without evidence.
11. Keep all product constraints: high-performance Go/Rust core, professional functional UI, custom icon and About/source/bug-report links, zero-cost local-first core, no website, Linux/macOS/Windows support, no source-code comments, MIT license, SECURITY.md, CONTRIBUTING.md, and well-structured documentation.

## Required codebase structure

Use the following as the starting point and adapt it to the application’s real needs. Every included directory must have a real purpose; do not create empty scaffolding.

```text
attackpath-visualizer/
├── README.md
├── LICENSE
├── CONTRIBUTING.md
├── SECURITY.md
├── CODE_OF_CONDUCT.md
├── CHANGELOG.md
├── .editorconfig
├── .gitignore
├── .env.example
├── Makefile
├── cmd/
│   └── attackpath-visualizer/
├── internal/
│   ├── config/
│   ├── domain/
│   ├── service/
│   ├── repository/
│   ├── transport/
│   └── observability/
├── api/
├── tui/
│   ├── views/
│   ├── components/
│   ├── formatting/
│   └── tests/
├── migrations/
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── testdata/
├── scripts/
├── docs/
│   ├── ARCHITECTURE.md
│   ├── THREAT_MODEL.md
│   └── ROADMAP.md
└── .github/
    ├── workflows/
    ├── ISSUE_TEMPLATE/
    └── PULL_REQUEST_TEMPLATE.md
```

### Structural rules
- `cmd/` contains executable entry points only.
- `internal/domain/` contains core models and business rules independent of transport/UI.
- `internal/service/` contains use cases and orchestration.
- `internal/repository/` contains persistence and external data adapters behind clear interfaces.
- `internal/transport/` contains HTTP, CLI, or protocol transport boundaries.
- `tui/` contains the terminal UI and must not duplicate authoritative business rules.
- `api/` contains OpenAPI or protocol definitions where applicable.
- `migrations/` contains versioned database migrations.
- `tests/` contains integration and end-to-end tests; keep unit tests close to their packages if that is idiomatic.
- `testdata/` contains synthetic, sanitized fixtures only.
- `scripts/` contains reproducible developer/CI helpers, not hidden deployment side effects.
- Keep modules cohesive, avoid giant files and circular dependencies, and document deviations from this layout.

## Mandatory desktop application UI (dashboard), launcher, icon, and About

This section is mandatory. It overrides any earlier wording that calls the desktop GUI optional or conditional. The dashboard described here is a native window of a local application built with Fyne; it is not a website. Hosted dashboards, browser frontends, and web servers remain forbidden.

### 1. One application, two front ends, chosen at launch

- Build one application with two front ends, a native desktop dashboard with a graph workspace and a terminal UI, both calling the same service layer. Neither front end contains business rules. Add `gui/`, `launcher/`, `assets/` (icons, fonts, embedded docs), and one central metadata module `internal/meta/meta.go` next to the existing `tui/` code.
- **Launcher.** Running the application with no arguments must let the user choose how to open it:
  - Started from a desktop or start menu (no terminal attached): show a small native chooser window titled "How do you want to open AttackPath Visualizer?" with two clearly labeled choices, **Dashboard (graphical window)** and **Terminal (text interface)**, a **Remember my choice** checkbox, and a Quit button.
  - Choosing Terminal opens the platform terminal emulator running the TUI: Windows Terminal with a `cmd.exe` fallback, macOS Terminal via `open`, Linux via `$TERMINAL`, `x-terminal-emulator`, then a documented list of common emulators. If none can be found, show the exact command to run manually; do not fail silently.
  - Started from a terminal with no arguments: print a short numbered prompt (1 Dashboard window, 2 Terminal UI) with the remembered or default choice on Enter. On a headless session (SSH, no display server), skip the prompt, open the TUI, and print one line saying why.
  - Flags and environment: `--ui gui|tui`, `--ask` to force the chooser, `--no-prompt` to use the remembered or default front end, and an `ATTACKPATH_VISUALIZER_UI` environment variable. Any explicit CLI subcommand (scan, import, export, about, diagnostics, and so on) never shows the chooser. Non-interactive stdin and CI never prompt.
  - The remembered choice lives in the OS user config directory with restrictive permissions. Settings has an **Ask me every time** switch, and `config reset-ui` clears it.
  - Both front ends offer a menu action to relaunch in the other mode.
  - Both front ends share one SQLite database with WAL mode and a busy timeout; if the other front end holds the database, show a clear notice instead of failing.
- Provide the `nogui` build tag (a build with `-tags nogui` omits the GUI and does not need cgo). Document the GUI build prerequisites for each OS, and do not claim the GUI is verified on a platform unless it was built and tested there.

### 2. Application shell and navigation

- Left navigation sidebar. Every category has an **icon next to its label**. The sidebar collapses to icons with tooltips, remembers its state, and marks the active item by shape and weight as well as color. About and Settings sit at the bottom.
- Top bar: global search, a data-mode chip when relevant (Demo data, Fixture, Simulation, Read-only), and a help entry.
- Keyboard: Ctrl/Cmd+1 to 9 jump to categories, Ctrl/Cmd+F search, Ctrl/Cmd+, Settings, F1 Usage guide. Show shortcuts in tooltips and in a shortcuts list.
- Persist window size, position, sidebar state, selected theme, and table preferences.

### 3. AttackPath Visualizer-specific interface (dashboard with a graph workspace)

| Category | Icon (Lucide-style name) | What it shows |
|---|---|---|
| Overview | `layout-dashboard` | Exposure summary: potential paths by impact tier, top-ranked paths, choke points, import freshness |
| Graph | `waypoints` | Interactive relationship graph of assets and edges with pan, zoom, filter, and focus |
| Paths | `route` | Ranked potential paths with an explainable score breakdown |
| Assets | `server` | Asset inventory and relationships |
| Evidence | `file-search` | Source evidence behind every edge |
| Remediation | `wrench` | Ranked suggestions, for example the edge whose removal breaks the most paths |
| Imports | `upload` | Ingestion history, validation results, demo-data label |
| Reports | `file-text` | Exported path and remediation reports |
| About | `info` | App identity, offline Usage guide, Source Code, bug reporting |
| Settings | `settings-2` | Appearance, density, launch preference, data location, safe defaults |

**Overview screen (first screen after launch).** Answer "what needs attention now" with dense, information-first content computed from the real data store:

- Potential paths per impact tier (each tile opens the filtered list)
- Top-ranked paths table with score and confidence
- Choke points that break the most paths
- Confidence distribution and data freshness of the last import

**Signature views.**

- Graph canvas with a keyboard-accessible node list as an alternative view
- Path detail: edge-by-edge evidence links and the exact score components

### 4. Visual design system (professional, not template-looking)

- Define a custom theme in code (tokens for color, spacing, type scale, radius, and elevation) and do not ship the toolkit's default look. Provide light and dark themes and follow the OS setting by default.
- Use a 4/8 px spacing scale, one accent color, and a restrained neutral palette. Use 1 px hairline borders, one small consistent corner radius, and elevation only for overlays. Use bundled open-license fonts (record their licenses) with a system fallback and a defined type scale.
- Severity and status always combine an icon and a text label with color. Meet WCAG AA contrast (4.5:1 text, 3:1 UI components) and keep colors distinguishable for common color-vision deficiencies.
- Tables are the primary surface: sortable, filterable, column chooser, row detail side panel, keyboard row navigation, copy and export actions, a compact and comfortable density switch, and virtualization for large data sets.
- Charts show real stored data only, have labeled axes and units, a text or table alternative, and are never used where a number or a table is clearer. Build them from toolkit canvas primitives or a permissively licensed library.
- Every screen has meaningful loading, empty (with the next action), error (with a fix), and success states.
- Avoid: gradients, glow, glassmorphism, emoji as icons, stock illustrations, identical rounded card grids on every page, decorative sparklines, random or hardcoded numbers, donut charts for two categories, placeholder text, and controls that do nothing.

### 5. Icons

- Use one consistent icon set for navigation, actions, and status: either a permissively licensed set embedded as SVG (Lucide under ISC or Tabler under MIT; record the license in `THIRD_PARTY_NOTICES`) or original icons drawn to the same grid and stroke width. Never mix sets and never use emoji. The names in the table above are Lucide-style suggestions; choose the nearest available icon.
- **Custom application icon (required).** Create an original icon for AttackPath Visualizer. Concept: three nodes joined by a bent path that ends in a target ring. Supply editable SVG source, a simplified version that stays legible at 16 px, `.ico` (Windows), `.icns` (macOS), PNG sizes from 16 to 1024, and Linux SVG/PNG with a desktop entry. Use it for the window, taskbar or dock, the launcher chooser, and About. No third-party trademarks, no generic placeholder, no stock art.

### 6. About, Usage guide, and bug reporting

- Reachable from the sidebar (info icon) and from the Help menu. About has these tabs: **Overview** (icon, name, version, build info, license, purpose, maintainer or copyright only when known), **Usage guide**, **Shortcuts**, **Diagnostics**, and **Licenses & notices**.
- **Usage guide** renders the bundled `docs/USAGE.md` offline inside the app (embedded at build time), with a table of contents and search, and an **Open full documentation on GitHub** link. It covers the launcher, the dashboard, the terminal UI, and every category above. Add screenshots to the docs only if they were captured from the running application with demo data.
- Central metadata in `internal/meta/meta.go` holds the repository URL `https://github.com/NexzaDev/attackpath-visualizer`; it is the single source for every link. About provides **Source Code** (opens the repository) and **Found a bug? Please report it on GitHub.** (opens `https://github.com/NexzaDev/attackpath-visualizer/issues/new`).
- The bug report action first shows a preview of exactly what will be included (app version, OS, architecture, build info, sanitized settings status), then opens the new-issue page with a prefilled title and body. If the prefilled body would exceed a safe URL length, copy it to the clipboard and open the blank new-issue page. Never attach logs, evidence, credentials, tokens, host data, or file contents automatically.
- External links open only on an explicit click, only for URLs under the configured repository, using the OS default handler, after a confirmation showing the full URL.

### 7. Accessibility and quality

- Full keyboard operation, visible focus indicators, text scaling, reduced-motion support, labeled controls, and no information conveyed by color alone. Test with a screen reader where the toolkit supports it and state the limits honestly where it does not.
- Every visible control performs a real action. Demo or simulated data is labeled wherever it appears.

### 8. UI structure and tests

- View models and formatting live outside the toolkit code so they are unit-testable. Add table-driven tests for the launcher decision (TTY or not, flags, environment, saved preference, headless), a test that About links are limited to the configured repository, tests that every navigation category resolves to a screen, and GUI smoke tests with the toolkit's headless test driver. In CI, build the GUI on the operating systems the runners support, and document what was and was not verified.

### 9. UI acceptance criteria

- The chooser and both front ends work from a clean checkout using demo data, with no paid service and no network access.
- All categories above exist, have icons, and show real data.
- About shows the offline Usage guide, Source Code, and the bug-report action pointing to `https://github.com/NexzaDev/attackpath-visualizer`.
- The graph stays usable at 1,000 nodes; document the measured interaction performance
- Every edge opens its source evidence record; no control executes or simulates an attack

### 10. Building the GUI when the sandbox is limited

Your execution sandbox may be unable to download the toolkit's dependencies, install system graphics libraries, or open a display. If so, do not skip the GUI and do not fake verification. Use GitHub Actions for everything the sandbox cannot do:

- Add a lockfile workflow that runs in CI with network access and commits `go.sum` back to the branch. A workflow_dispatch-only workflow cannot be started from a non-default branch, so trigger it on push to your own branch (limited to changes in `go.mod` or the workflow file) with `contents: write` as the only elevated permission, commit only when the file actually changed so it cannot loop, and push a follow-up commit afterwards because pushes made with the workflow token do not start other workflows. Remove or disable it at the end and say so in the PR body; if you cannot disable it, tell the maintainer exactly how. Never add vanity-host replace directives or vendored hacks to work around the sandbox.
- Keep the CLI/TUI buildable and fully tested without the GUI through the `nogui` build tag, so a GUI problem never blocks the rest of the product.
- In CI, build the GUI on Ubuntu, macOS, and Windows where the runners allow it, and run headless GUI tests. Take the system package list for Linux from the toolkit's official documentation at the time of the work and record it.
- Render every main screen headlessly with Fyne's software renderer from its test packages, using demo data only, and upload the PNG files as workflow artifacts. List the artifact names in the PR body so the maintainer can review the design from a phone. Treat these screenshots as design evidence, not as proof that the application runs on a real desktop.
- Report exactly what CI proved and what remains unverified, for example real window behavior, OS-specific launch, icon display in the taskbar, and the terminal-emulator launch path. Do not claim visual or platform verification that CI did not perform.
- <!-- FINAL-LESSONS --> If every CI job fails at startup with a billing or spending-limit message, stop immediately, say so in one line, and do not keep pushing. If only macOS jobs cannot start, continue and mark macOS as unverified everywhere it is mentioned. Never describe a platform as verified because the job was defined.
- Read CI failures through check-run annotations: make failing steps print each finding as a GitHub error annotation so you can read them without log access.
- The custom theme must genuinely replace the toolkit defaults for colors, type, and spacing as described in section 4; a theme that only wraps the default theme and adds a few colors does not satisfy this section. Review your own rendered screenshots for clipped text, overlaps, empty areas, and weak contrast before finishing.
- <!-- GREEN-CI-RULE --> Do not finish while any CI job on the final commit is red. After the last push, check every job on every operating system. For each red job, read its annotations, fix the real cause, push, and check again. If a job still fails, say so at the top of the final report with the job name and the exact error, and do not describe the project as finished.
- If the GUI cannot be completed this way, deliver the CLI/TUI fully, document the blocker and a precise plan in `docs/ROADMAP.md`, and state that the GUI is not delivered.

### 11. Notifications (in the app and outside it)

Add notifications where they fit this product. Events for attackpath-visualizer: a new critical path appears after an import; an import fails validation. Add a **Notifications** category (icon `bell`) to the navigation table.

- **Inside the app:** a bell in the top bar with an unread count and a Notifications category listing every notification with a severity icon plus text, filters, mark-as-read, and a link to the relevant record. Notification rules are configurable per event type with a severity threshold, quiet hours, and de-duplication so a repeated condition does not flood the list.
- **Outside the app (opt-in):** a background notifier that reports while the dashboard is closed: `attackpath-visualizer notifier run|install|uninstall|status`. `install` creates a per-user autostart entry without administrator rights (Windows Startup or Task Scheduler, macOS LaunchAgent, Linux XDG autostart). It is off by default and needs explicit consent in Settings or on the command line. It reads the local database read-only on a bounded interval, makes no network calls of its own, and sends operating-system notifications.
- **Privacy:** notification text is minimal by default (for example "2 new critical findings") with a setting for more detail. Never include secrets, credentials, evidence content, or full indicators in notification text, and never send a notification to any external service.
- **Tests:** test rule evaluation, de-duplication, quiet hours, and install/uninstall logic with an injected notifier and a temporary directory. State that real OS notification display is unverified on each platform unless CI proves it.
- Use one small permissively licensed notification library if one is needed, and record it in the dependency decision record, `THIRD_PARTY_NOTICES`, and the threat model.

### 12. Release packages for Windows, macOS, and Linux

Deliver real downloadable packages built by GitHub Actions, not only source code. The agent does not build these on its own machine; the workflow does.

- Add a manually triggered release workflow (workflow_dispatch) on the default branch's workflow set that builds the GUI application on each operating system's own runner with the `fyne package` tool from `fyne.io/tools/cmd/fyne` pinned to an exact version, so the icon is embedded in the Windows `.exe` and in the macOS `.app` bundle.
- Outputs (follow the pattern already proven on the Argus ASM release, <!-- PROVEN-PATTERN -->and keep these exact asset names): `attackpath-visualizer-windows-x64.exe`, `attackpath-visualizer-macos-arm64.tar.gz` (add `attackpath-visualizer-macos-x64.tar.gz` when a runner for it is available), and `attackpath-visualizer-linux-x64.tar.gz`, plus `SHA256SUMS`. Each package contains the application with the custom icon embedded where the platform supports it (Windows `.exe` resource, macOS `.app` bundle inside the archive when the packaging tool produces one, Linux desktop entry and icon in the archive), the license files of the application and of every third-party module in that platform's build, a short `README`, and the Usage guide. The Windows `.exe` may be a single file.
- Mark the release as a **pre-release** named "<App name> v0.1.0-dev (dev beta)" with release notes that say it is an early build, that not every planned feature exists, and how to open unsigned binaries (macOS: right-click and choose Open the first time; Windows: choose More info and Run anyway; Linux: run `chmod +x <binary>` after extracting).
- The background notifier is part of the same binary (`attackpath-visualizer notifier ...`); no separate service or installer that needs administrator rights.
- Write a `SHA256SUMS` file covering every package, and attach the packages and checksums to a **draft** GitHub Release. Never publish it automatically.
- Document the unsigned-binary warnings (Windows SmartScreen, macOS Gatekeeper) and how a user opens an unsigned app on each system. Do not claim code signing or notarization unless it was really done with a key the maintainer controls.
- Document, in `docs/RELEASING.md`, the exact steps for the maintainer to run the workflow and publish the draft release using only a phone and the GitHub website.
- Build the `nogui` or headless variant too and attach it, so a CLI/TUI-only package exists for servers.
- Report honestly which packages the workflow produced, which job could not start (for example macOS because of billing), and which results were never run or never opened on a real machine.

## Engineering standards: no vibe coding
- Make evidence-based engineering decisions. Do not invent features, test results, benchmarks, integrations, or security claims.
- Do not generate filler code to inflate line count. Favor complete vertical slices, maintainability, performance, and correctness.
- Inspect the repository first; preserve valid existing work and identify the current state before changing files.
- Write a short plan, then implement and verify the product. Do not stop at a plan, mockup, or architecture diagram.
- Use clear module boundaries, typed interfaces, cohesive packages, explicit error handling, input validation, structured logging, migrations, and reproducible builds.
- Avoid giant files, god objects, circular dependencies, duplicated business logic, unnecessary abstractions, dead code, fake data in production paths, and placeholder buttons.
- Add unit, integration, API, and end-to-end tests appropriate to the system. Test invalid inputs, edge cases, permissions, failure modes, and security boundaries.
- Run formatter, linter, static analysis/type checks, tests, dependency checks, and production build. Fix failures and report only checks actually run.
- No comments in application or test source code: no inline comments, block comments, or commented-out code. Express intent through names, types, tests, and external documentation. Keep required license headers only if legally necessary and place licensing in LICENSE rather than adding routine source comments.
- Never claim the code is bug-free. Provide verified test evidence and list remaining limitations honestly.
- Do not optimize for a target line count. Do not claim 457,000+ LOC unless measured by a reproducible tool and supported by real, non-generated functionality.

## Required repository files
- `README.md`: purpose, scope, features, architecture summary, supported platforms, prerequisites, setup, configuration, commands, screenshots placeholders only if no real screenshots exist, API/CLI usage, tests, troubleshooting, limitations, and roadmap.
- `LICENSE`: MIT License, only if the author intends to distribute this project under MIT. Verify that all bundled dependencies and copied assets are compatible; do not claim third-party code is MIT.
- `CONTRIBUTING.md`: development setup, branch/commit conventions, issue and PR expectations, testing requirements, code style, security-sensitive changes, and review checklist.
- `SECURITY.md`: supported versions, private vulnerability reporting instructions using a repository-appropriate contact/channel placeholder, response expectations, scope, and responsible disclosure guidance. Do not invent an email address.
- `CODE_OF_CONDUCT.md`: Contributor Covenant or a clearly documented equivalent.
- `CHANGELOG.md`: versioned user-visible changes, starting with an initial unreleased section.
- `docs/ARCHITECTURE.md`: components, data flow, trust boundaries, storage, deployment, and design decisions.
- `docs/THREAT_MODEL.md`: assets, adversaries, trust boundaries, abuse cases, mitigations, residual risks, and assumptions.
- `docs/ROADMAP.md`: prioritized milestones with measurable completion criteria.
- `.github/`: CI workflows for formatting/linting, tests, static checks, build, and dependency/security scanning where supported; issue and pull request templates.
- `.editorconfig`, `.gitignore`, and `.env.example` with placeholders only.
- Tests, fixtures, migrations, and scripts organized in predictable directories.

## Required architecture and repository layout
Use this as the default shape, adapting it to the chosen language and project. Do not create empty directories merely to match a template.

```text
.
├── README.md
├── LICENSE
├── CONTRIBUTING.md
├── SECURITY.md
├── CODE_OF_CONDUCT.md
├── CHANGELOG.md
├── .editorconfig
├── .gitignore
├── .env.example
├── Makefile
├── cmd/
├── internal/
├── pkg/
├── api/
├── tui/
├── migrations/
├── tests/
├── testdata/
├── scripts/
├── docs/
│   ├── ARCHITECTURE.md
│   ├── THREAT_MODEL.md
│   └── ROADMAP.md
└── .github/
    ├── workflows/
    ├── ISSUE_TEMPLATE/
    └── PULL_REQUEST_TEMPLATE.md
```

Adapt directories where the framework requires another convention; explain important architectural choices in documentation. Keep core business logic independent from HTTP handlers and UI components. Do not expose internal packages as public APIs without a reason.

## Security and open-source release bar
- Secure defaults, least privilege, authentication and authorization where applicable, parameterized database access, safe output encoding, rate limiting, audit logging, and explicit secret redaction.
- No hardcoded credentials, production tokens, personal data, unsafe default scan targets, or external network calls in tests unless explicitly mocked.
- Use pinned or bounded dependencies, generate a dependency inventory/SBOM where practical, review licenses, and document update policy.
- CI must run on a clean checkout. Provide deterministic seed data and a local demo mode.
- Include a `Makefile` or equivalent task runner with documented targets such as `test`, `lint`, `format`, `typecheck`, `build`, and `check`.
- Provide OpenAPI documentation for HTTP APIs or complete CLI help for CLI-first tools.
- Keep UI accessible and responsive; all displayed operational metrics must come from real application data or clearly labeled demo fixtures.
- Use conventional commits or document the selected commit convention. Make pull requests reviewable and changes focused.


## Project-specific acceptance criteria
- Deliver a locally runnable end-to-end MVP using synthetic demo data without paid accounts or production credentials.
- Provide the core domain workflow, persistence where appropriate, API or CLI, and a usable responsive UI when a UI is relevant.
- Implement actual behavior behind every visible control; no fake buttons or hardcoded production metrics.
- Include migrations, seed data, health/readiness checks, structured logs, graceful shutdown, and sensible timeouts where appropriate.
- Add API contract tests and document API authentication, authorization, and rate limits.
- Include a `.env.example` with placeholders only and fail safely when required configuration is missing.
- Include the required open-source governance and security files listed above. The MIT license must be a valid standard license text and must not misrepresent third-party dependencies or assets.
- Use GitHub Actions CI to run formatting, linting/static checks, tests, build, and dependency/license checks where practical. Use least-privilege workflow permissions.
- Document how to run each verification command locally and make CI reproducible.
- Include a short architecture decision record for significant choices such as storage, queueing, parsing, or provider abstractions.

## Execution protocol
1. Inspect the repository and report its state in your working notes; do not overwrite useful work.
2. Create a prioritized plan and implement the highest-value vertical slice immediately.
3. Compile/build early and often. Run tests after each meaningful subsystem; do not defer all validation to the end.
4. Perform a security review of trust boundaries, auth, validation, secret handling, dependencies, and dangerous defaults.
5. Review the complete diff. Search for TODO/FIXME, placeholders, commented-out code, leaked secrets, duplicated logic, and unsupported claims.
6. Run formatting, lint/static analysis, tests, type checks if relevant, production build, and dependency/license checks. Fix discovered issues and rerun.
7. Update README and docs to match the implementation, not the intended future state.
8. Finish with an honest report of completed functionality, exact commands and outcomes, known limitations, and prioritized next milestones.

## Final rule

Do not settle for vibe-coded output. Do not generate filler to inflate LOC, do not pretend incomplete features are finished, do not guarantee zero bugs, and do not report checks that were not actually run. If the full product is too large for one execution, deliver a correct and testable foundation first and then proceed through the roadmap in ranked milestones. Keep all active and test source code free of comments, as requested.


---

## Release checklist

- [ ] Main workflow works end to end with local demo data.
- [ ] Clean-checkout setup instructions have been tested.
- [ ] Formatting, lint/static checks, tests, and production build pass.
- [ ] CI workflows run with least privilege.
- [ ] `SECURITY.md` provides a real maintainer-controlled reporting path before public release.
- [ ] MIT licensing and third-party license compatibility have been reviewed.
- [ ] Documentation describes implemented behavior and known limitations accurately.
- [ ] No secrets, private data, commented-out code, or unexplained stubs are present.
- [ ] Launcher decision tests pass; the chooser works from a desktop launch and from a terminal; headless fallback is tested.
- [ ] Every navigation category has an icon from one consistent set, and icon licenses are recorded.
- [ ] Custom application icon assets (SVG, `.ico`, `.icns`, PNG sizes, Linux desktop entry) are present and used by the window, taskbar or dock, and About.
- [ ] About shows the offline Usage guide, Source Code, and the bug-report action for the configured repository.
- [ ] The GUI was built and smoke-tested on every operating system claimed; other platforms are documented as unverified.
- [ ] If the GUI was built through CI, the lockfile workflow is documented or removed, and screenshot artifacts are listed in the PR body.
- [ ] Notifications work in the app, and the opt-in background notifier installs and uninstalls cleanly; real OS display is reported as verified only where it was actually shown.
- [ ] The release workflow has produced Windows, macOS, and Linux packages with the icon embedded, plus SHA256SUMS, as a draft release; anything not produced or not run is listed as unverified.
