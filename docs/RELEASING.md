# Release procedure

The repository provides two separate release paths. **Beta release** is triggered by pushing a new `v*` tag and publishes a public pre-release. **Release packages** is a manual workflow that attaches packages to a draft pre-release and never publishes it automatically. Both use native operating-system runners and pin `fyne.io/tools/cmd/fyne` to `v1.7.4`, which includes newer `golang.org/x/tools` support required by the configured Go toolchain. `cmd/attackpath-visualizer/FyneApp.toml` must place app fields under `[Details]`; its icon path is resolved relative to the app metadata.

The Ubuntu GUI prerequisite list comes from the Fyne v2.6.3 upstream `.github/workflows/platform_tests.yml`: `bc`, `gcc`, `libgl1-mesa-dev`, `libwayland-dev`, `libx11-dev`, `libxkbcommon-dev`, `xorg-dev`, `xvfb`, and `language-pack-en`.

## Public dev-beta release

Only tag a commit after CI has passed on Linux, macOS, and Windows and the Linux race detector is green. The tag workflow independently runs `go test ./...` and `go vet ./...` on each native runner before packaging. It verifies the app icon and package contents, uploads matrix artifacts, creates `SHA256SUMS`, then publishes a public GitHub pre-release.

For the initial release, use tag `v0.1.0-dev` on the verified PR head. Never force-update or reuse a tag. The publishing job checks that the tag resolves to the workflow commit and fails if a release with that tag already exists. If any package job fails, follow the project release instructions: remove a failed release and tag before pushing a corrected tag; do not bypass a failing check.

Expected public assets:

- `attackpath-visualizer-linux-x64.tar.gz`
- `attackpath-visualizer-macos-arm64.tar.gz`
- `attackpath-visualizer-windows-x64.zip` (contains the GUI executable and required license files)
- `SHA256SUMS` (checksums for all three packages)

Each GUI package includes the app and custom icon plus `LICENSE`, `THIRD_PARTY_NOTICES.md`, `licenses/`, README, and offline Usage documentation. The package job validates the icon resource and required license files before upload. The Windows archive is used so the executable and licenses travel together.

After publication, inspect the GitHub Release page and confirm the **Pre-release** badge, title `AttackPath Visualizer v0.1.0-dev (dev beta)`, all four asset names, plausible non-zero sizes, and checksum lines. Download each package on a supported machine, run `sha256sum -c SHA256SUMS`, inspect its contents, and record actual test results. Do not claim a package was opened on an OS unless it was actually run there.

## Manual draft release from a phone or desktop

The **Release packages** workflow is manual-only and must be present on the default branch before GitHub makes it dispatchable. Use this path to prepare a reviewable draft without publishing it:

1. Open `NexzaDev/attackpath-visualizer` in GitHub or its mobile browser and select **Actions**.
2. Choose **Release packages**. If it is absent, it is not present/enabled on the default branch.
3. Select **Run workflow**, choose `main`, and confirm.
4. Wait for the Windows x64, macOS ARM64, Linux x64, and draft-attachment jobs to finish. Any failed/cancelled job means its package was not produced.
5. Review the draft **AttackPath Visualizer v0.1.0-dev (dev beta)**, expected package names, `SHA256SUMS`, and release notes. Check the package contents and checksums before manually publishing.
6. The workflow never publishes the draft automatically. Publish only after review and maintainer approval.

## Unsigned build warnings

Packages are unsigned and macOS builds are not notarized. Verify checksums before opening downloaded files.

- **macOS:** Gatekeeper may warn. Right-click the app and choose **Open** the first time.
- **Windows:** SmartScreen may warn. Choose **More info → Run anyway** only if you trust the package and verified its checksum.
- **Linux:** Extract the archive and run `chmod +x <binary>` if necessary.

No signing key or notarization credentials are used.

## Verification limits

A green runner proves the commands completed on that runner; it does not prove installation or runtime behavior on a physical desktop. Record the actual workflow URL, commit, assets, sizes, checksums, and any failed-to-start jobs. macOS x64 is not included until a supported Intel runner is available.
