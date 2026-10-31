# Releasing

Release packaging is not yet available. Once `.github/workflows/release.yml` exists, maintainers should run its `workflow_dispatch` from the default branch in GitHub Actions, review the produced Windows, macOS, and Linux assets and `SHA256SUMS`, then create or update the draft pre-release from a phone or desktop using the GitHub website. Verify asset names and checksum coverage before publication.

Unsigned builds may trigger Windows SmartScreen or macOS Gatekeeper. Do not bypass platform protections unless the artifact is from the verified project release. Windows users may use **More info → Run anyway** only for a verified release. macOS users may Control-click and choose **Open** after reviewing the source. On Linux, extract and run `chmod +x <binary>`. No signing or notarization is currently claimed.
