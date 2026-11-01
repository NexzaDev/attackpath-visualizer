# Architecture

The application is local-first. `internal/domain` validates assets and directed edges and enumerates bounded simple paths. The score is maximum impact (1–5) multiplied by compounded edge confidence (0–100); score components are retained as an explanation. `internal/service` applies input size limits and coordinates imports. `internal/repository` owns SQLite schema initialization, indexed adjacency columns, atomic graph replacement, and import history. `cmd/attackpath-visualizer` is the CLI boundary.

Data flows from a JSON fixture or authorized JSON import through validation, into a single SQLite transaction, then back into graph queries and JSON/text output. SQLite uses WAL mode, a five-second busy timeout, a single connection per process, and foreign-key references. The database is created under the user's configuration directory with a private parent directory. The engine has no listener and no external service dependencies.

The desktop and terminal interfaces are planned as separate front ends calling the same service. The CLI currently forms the implemented UI. The graph source/evidence fields are user-provided data and are not trusted as evidence of exploitability. Imports are bounded to 8 MiB and invalid data is rejected before stored state is changed.

## Decisions

- Go is the application language; `modernc.org/sqlite` provides a pure-Go SQLite driver to avoid a required C compiler for CLI builds.
- Directed simple paths are enumerated to bounded depth and count; cycles are suppressed per traversal. The current demo graph and realistic small inventories fit this model. A large-graph indexing/benchmark plan remains on the roadmap.
- No HTTP API is exposed. The interactive local clients need no listener, authentication surface, or network policy.
