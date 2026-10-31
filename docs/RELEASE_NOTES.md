# AttackPath Visualizer v0.1.0-dev (dev beta)

This is an early development build. Not every planned feature is present, and desktop packages have not yet been opened on physical Windows, macOS, or Linux machines. Review the README and offline Usage guide before importing real inventory. Potential paths are prioritization leads, not proof of exploitability.

The draft contains native GUI packages for Windows x64, macOS ARM64, and Linux x64, CLI/TUI-only archives for servers, and a `SHA256SUMS` file covering every attached package. Builds are unsigned and not notarized.

- **Windows:** SmartScreen may warn. For a verified project release, choose **More info → Run anyway** only if you trust the downloaded file and have checked its checksum.
- **macOS:** Gatekeeper may block the unsigned app. After verifying the checksum, Control-click the app and choose **Open** the first time.
- **Linux:** Extract the archive and run `chmod +x <binary>` if the executable bit was not preserved.

Import only inventory you are authorized to use. Export a backup before replacing stored graph data.
