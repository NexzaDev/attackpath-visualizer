# AttackPath Visualizer v0.1.0-dev (dev beta)

This is an early development pre-release. Not every planned feature is present, and the GUI packages have not been opened on physical Windows, macOS, or Linux machines. Review the README and offline Usage guide before importing real inventory. Potential paths are prioritization leads, not proof of exploitability.

The release contains native GUI packages for Linux x64, macOS ARM64, and Windows x64, plus a `SHA256SUMS` file covering each package. The Windows package is a ZIP containing the GUI executable and required license files. Builds are unsigned and not notarized.

- **macOS:** Gatekeeper may warn. Verify the checksum, then right-click the app and choose **Open** the first time.
- **Windows:** SmartScreen may warn. Verify the checksum, then choose **More info → Run anyway** only if you trust the downloaded package.
- **Linux:** Extract the archive, verify the checksum, and run `chmod +x <binary>` if the executable bit was not preserved.

Import only inventory you are authorized to use. Export a backup before replacing stored graph data.
