# Roadmap

1. **Local graph core and CLI** — graph validation, bounded explainable path scoring, local SQLite persistence, versioned SQL migrations, JSON import/export, import history, and deterministic offline demo data are implemented in this branch. Unit and cross-platform integration evidence is pending CI.
2. **Native desktop and terminal front ends** — Fyne dashboard, terminal menu, chooser, graph/path/evidence/remediation workflows, imports/reports/settings, accessibility aids, offline About/diagnostics, custom icon/theme, and headless screenshot capture are implemented in this branch. Native builds, screenshots, and desktop runtime behavior are not verified until CI and real-device checks complete.
3. **Notifications and releases** — in-app rules and event history, opt-in read-only background notifier, platform startup entries, and manually triggered draft-release packaging are implemented in this branch. Notification delivery, release packages, and real platform behavior remain unverified until workflow runs and package review.
4. **Scale and operability** — benchmark documented workloads, improve large graph traversal and table responsiveness, broaden import formats only with complete validation/fixtures, and record user feedback.

Completion evidence is recorded in CI and the pull request body. Do not treat code presence or source inspection as proof of builds, tests, package outputs, screenshot quality, or runtime behavior.
