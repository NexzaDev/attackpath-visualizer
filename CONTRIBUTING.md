# Contributing

Use focused branches and conventional commit subjects (`feat:`, `fix:`, `docs:`, `test:`, `build:`). Open an issue for substantial changes and a pull request with scope, test evidence, platform coverage, and risks. Never include real inventory, credentials, or customer data in fixtures.

Install Go 1.23+, then run `gofmt -w` on changed Go files, `go vet ./...`, and `go test -race ./...`. The full `make check` target also builds the CLI. Keep domain behavior independent from UI and persistence packages. Add deterministic tests for errors and boundaries. Do not add source comments; document behavior in package names, tests, and docs.

Security-sensitive changes must validate untrusted input, preserve local-only defaults, avoid logging content or secrets, and include threat-model updates. Review dependencies and licenses before adding modules. Pull requests should state any checks or platforms that could not be run.
