# Release procedure

The repository's **Release packages** GitHub Actions workflow is manual-only. It must be present on the default branch before it can be dispatched. The workflow builds on native operating-system runners and uses `fyne.io/tools/cmd/fyne` pinned to `v1.26.1`; it never publishes a release automatically.

The Ubuntu GUI prerequisite list was recorded from the Fyne v2.6.3 upstream `.github/workflows/platform_tests.yml`: `bc`, `gcc`, `libgl1-mesa-dev`, `libwayland-dev`, `libx11-dev`, `libxkbcommon-dev`, `xorg-dev`, `xvfb`, and `language-pack-en`. The CI and release workflows install the same list.

For a local native package test (the workflow remains the source of downloadable release assets), install the exact packager and run it from `cmd/attackpath-visualizer` on each supported OS:

```sh
go run fyne.io/tools/cmd/fyne@v1.26.1 package --target linux --release
go run fyne.io/tools/cmd/fyne@v1.26.1 package --target darwin --release
go run fyne.io/tools/cmd/fyne@v1.26.1 package --target windows --release
```

Run only the command matching the native runner/host. The macOS package is ARM64 on the configured Apple Silicon runner; the Windows and Linux outputs are x64 on their configured runners. These commands download Go modules and the pinned packaging tool.

## Run from a phone or desktop

1. Open `NexzaDev/attackpath-visualizer` in the GitHub website or mobile browser and select **Actions**.
2. Choose **Release packages** in the workflow list. If the workflow is absent, it has not been merged to the default branch yet.
3. Select **Run workflow**, leave the branch set to `main`, and confirm. The Windows, macOS ARM64, and Linux x64 jobs build in parallel.
4. Wait for all three package jobs and the **attach packages to a draft pre-release** job to finish green. A failed or cancelled matrix job means its platform output was not produced; do not describe the release as complete.
5. Open **Releases** and review the draft **AttackPath Visualizer v0.1.0-dev (dev beta)**. It must contain `attackpath-visualizer-windows-x64.exe`, `attackpath-visualizer-macos-arm64.tar.gz`, `attackpath-visualizer-linux-x64.tar.gz`, `attackpath-visualizer-cli-windows-x64.zip`, `attackpath-visualizer-cli-macos-arm64.tar.gz`, `attackpath-visualizer-cli-linux-x64.tar.gz`, and `SHA256SUMS`.
6. Download the draft assets and verify the checksum file before publication. Review the release notes and package contents; the GUI packages include the custom icon, README, offline Usage guide, application/third-party license files, and platform integration files.
7. Publish the draft only after reviewing green workflow results, package names and checksums, and the source commit. The workflow never publishes it automatically.

The build also uploads per-platform review artifacts named `attackpath-visualizer-release-windows-x64`, `attackpath-visualizer-release-macos-arm64`, and `attackpath-visualizer-release-linux-x64`. These are useful before a draft exists and are not a substitute for opening the application on the target OS.

## Unsigned build warnings

Packages are unsigned and macOS builds are not notarized. Windows SmartScreen may warn; for a verified build, use **More info → Run anyway** only after checking the checksum. macOS Gatekeeper may block the app; after checking the checksum, Control-click and choose **Open** the first time. On Linux, extract the archive and run `chmod +x <binary>` if necessary. Do not bypass platform protections for unverified files. No signing key or notarization credentials are used.

## Platform and verification limits

A successful runner job proves that its build and packaging commands completed on that runner. It does not prove that the package was installed or opened on a physical desktop. Record the actual workflow URL, commit, produced assets, checksum result, and any job that failed to start in release notes before publication. Never claim that a package was built or tested unless its workflow result exists. macOS x64 is not included until a supported Intel runner is available.
