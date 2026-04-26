# Kyn MVP v2 Finish Tasks

These are the remaining tasks required to close out v2 cleanly.

## 1. Reporter and Release Correctness

Status: completed

- fix Docker multi-arch build correctness for arm64 image publishing
- ensure machine reporters only emit actionable failing diagnostics
- add tests covering machine reporter filtering behavior

## 2. Distribution Verification and Release Docs

Status: in progress

- strengthen CI/release coverage for publishable artifacts
- clarify release/install story across binaries and container image usage
- close the remaining portability/documentation gaps for the v2 contract

## 3. Presets and Adoption Polish

Status: pending

- expand `kyn init --preset` beyond `web-ui`
- add docs/examples for `web-ui`, `api`, `proto`, and `iac`
- finish the final v2 adoption polish across README and docs
