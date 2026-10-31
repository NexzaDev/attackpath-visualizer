# Roadmap

1. **Local graph core and CLI** — validate directed graph data, explain bounded path scores, import/export JSON, persist to SQLite, seed deterministic demo data, and pass unit/integration CI. This milestone is implemented in the current working tree; cross-OS evidence is pending CI.
2. **Native desktop and terminal launcher** — add the Fyne dashboard, shared navigation/views, accessible graph/path/evidence workflows, About and diagnostics, launcher decisions, icon assets, and headless rendered screenshots; build/test across supported OS runners.
3. **Notifications and releases** — durable in-app rules and deduplication, opt-in notifier install/status/uninstall, real OS packages and draft-release workflow with checksums.
4. **Scale and operability** — benchmark documented workloads, improve large graph traversal and table responsiveness, and add additional import formats only with complete validation and fixtures.

Completion evidence is recorded in CI and the pull request body. Unverified platform behavior must remain listed as unverified.
