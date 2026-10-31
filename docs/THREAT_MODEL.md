# Threat model

## Assets and trust boundaries

Protected assets are the user's local inventory database, imported configuration/evidence documents, application settings, and integrity of graph explanations. Trust boundaries are the command line/import file to parser, parser to graph validator, graph service to SQLite, and local outputs/diagnostics to the user's terminal. Inputs may be malformed or misleading. The program does not access a production environment or execute attack techniques.

## Adversaries and abuse cases

A malicious or malformed inventory could exceed memory limits, contain duplicate identifiers, point edges at nonexistent assets, or attempt to alter database relationships. A local user could set a path to an unintended database. A future notifier could disclose sensitive data if notification text becomes overly detailed. Dependency compromise is also in scope.

## Mitigations

Imports are capped at 8 MiB, decoded into typed records, validated for identifiers, ranges, duplicates, self-edges, and dangling references, then persisted in one transaction using parameterized SQL. SQLite uses indexed foreign-key columns and local WAL mode. Runtime diagnostics do not read imported files or reveal paths/host data. There is no telemetry, default listener, credential requirement, or outbound network operation. CI runs tests, vet, and vulnerability checks.

## Residual risks and assumptions

Inventory text can contain false or sensitive information and remains the operator's responsibility. Local users/processes with access to the account can read or change the database. Platform permissions depend on the OS configuration directory. The graph score is a prioritization heuristic, not a calibrated probability or proof of reachability. Packaging/signing and real-desktop behavior are not yet verified.
