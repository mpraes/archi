# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-07-04
### Added
- Initial CLI: `archi [path]` scans a Go codebase and serves a local web UI.
- `export --format json|markdown` for CI/headless use.
- `check --max-distance <float>` exits non-zero on main-sequence violations.
- Embedded D3-based web UI with main-sequence scatter plot and module side panel.
- Resilient parser that logs and skips syntax-broken files (RNF-003).
- opt-in Gemini enrichment via `--ai` / `GEMINI_API_KEY` (RNF-007/008/009).