# Security policy

## Supported versions

The project is in development. Until a stable release is declared, only the latest `v0.1.0-dev` pre-release and the current default-branch source are considered for security fixes. Older development builds are unsupported; update to the latest available version before reporting.

## Private vulnerability reporting

GitHub private vulnerability reporting is the preferred channel: [Report a vulnerability](https://github.com/NexzaDev/attackpath-visualizer/security/advisories/new). Please do not disclose suspected vulnerabilities in public issues, discussions, or pull requests.

### Backup channel: email

If you cannot use the GitHub form, send the report to **SuperMcDuck@proton.me**.

## What to include

Include the affected version or commit, affected component, clear reproduction steps using synthetic data, expected and actual behavior, impact assessment, and a proposed mitigation if known. Do not send real asset inventories, credentials, tokens, or personal data. Redact sensitive details and attach only the minimum synthetic proof needed to reproduce the issue.

## Response expectations

For this volunteer-run project, maintainers target acknowledging reports within 7 days and providing an initial assessment within 14 days. These are targets, not guarantees. The assessment will identify severity, affected versions, next steps, and an expected timeline where possible.

## Scope

In scope are local inventory parsing and validation, SQLite persistence and migrations, path/scoring/remediation logic, data exported by the application, diagnostics, desktop/terminal launch behavior, release packaging, and the opt-in background notifier. The application is not a vulnerability scanner, does not execute attack techniques, and does not contact assessed systems. Third-party dependencies should be reported to their upstream maintainers as well; include a project report when the dependency issue affects this application.

## Disclosure

Please allow maintainers reasonable time to investigate, prepare a fix, and coordinate disclosure before publishing technical details. We will coordinate a disclosure date with the reporter and credit them if they want attribution. Do not access, alter, or retain data that is not yours while validating a report.
