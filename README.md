# archi

> Static analysis with architectural focus and actionable diagnostics.

`archi` scans a codebase, computes per-module architectural metrics, and serves an embedded
local dashboard with comparison against the previous scan. It supports Go, JavaScript/TypeScript,
and Python projects (or mixed repositories) through auto-detection or `--lang`.

- Single static binary (Go 1.26+), with frontend assets embedded via `go:embed`.
- Resilient parsing: syntax-broken files are warned and skipped, not fatal.
- Offline-first flow: metrics UI loads immediately; AI insights stream in separately.
- Privacy by design: AI receives metrics only, never source code.

## Install

```sh
# Homebrew (macOS or Linux)
brew tap mpraes/tap && brew install archi

# AUR (Arch Linux) - prebuilt binary
yay -S archi-bin

# mise
mise use -g github:mpraes/archi

# Direct binary download (Linux)
curl -fsSL "https://github.com/mpraes/archi/releases/download/vX.Y.Z/archi-vX.Y.Z-linux-amd64.tar.gz" | tar xz

# Direct binary download (Windows)
# Download archi-vX.Y.Z-windows-amd64.zip from GitHub Releases, extract archi.exe, then:
# .\archi.exe . --no-browser

# DEB / RPM
sudo dpkg -i archi_X.Y.Z_amd64.deb
# or: sudo rpm -i archi-X.Y.Z-1.x86_64.rpm

# Build from source
git clone https://github.com/mpraes/archi && cd archi
(cd web && npm install && npm run build) && go build -o archi ./cmd/archi
```

Pre-built binaries are published for **linux/amd64** and **windows/amd64**.
Checksums are attached to every release.

## Update to a new version

### Upgrade (recommended)

Use in-place upgrade first (faster and safer):

- Homebrew: `brew upgrade archi`
- AUR: `yay -Syu archi-bin`
- RPM: `sudo rpm -Uvh archi-X.Y.Z-1.x86_64.rpm`
- DEB: `sudo apt update && sudo apt install --only-upgrade archi` (when installed from apt repo)
- mise: `mise use -g github:mpraes/archi@latest`

### Clean reinstall (troubleshooting)

Use this path only when upgrade fails, files are corrupted, or paths were manually changed.

```sh
# Homebrew
brew uninstall archi
brew update && brew install archi
# or: brew upgrade archi

# AUR (yay)
yay -Rns archi-bin
yay -S archi-bin
# or: yay -Syu archi-bin

# mise
mise uninstall -g github:mpraes/archi
mise use -g github:mpraes/archi@latest

# Direct binary install (Linux)
sudo rm -f /usr/local/bin/archi
curl -fsSL "https://github.com/mpraes/archi/releases/download/vX.Y.Z/archi-vX.Y.Z-linux-amd64.tar.gz" | tar xz
sudo install -m 0755 archi-vX.Y.Z-linux-amd64/archi /usr/local/bin/archi
archi --version

# DEB
sudo apt remove archi
sudo dpkg -i archi_X.Y.Z_amd64.deb

# RPM
sudo rpm -e archi
sudo rpm -i archi-X.Y.Z-1.x86_64.rpm
# or upgrade in place: sudo rpm -Uvh archi-X.Y.Z-1.x86_64.rpm
```

Replace `X.Y.Z` with the release number from [GitHub Releases](https://github.com/mpraes/archi/releases), for example `0.1.0`.

## Quickstart

```sh
archi .                         # scan current dir, open browser at :8080
archi /path/to/project --port 3000 --no-browser
archi . --exclude "internal/migrations,*.pb.go"
archi . --lang all              # force mixed-language parsing
archi . --ai                    # force AI path (expects API key)
```

Headless modes for scripts and CI:

```sh
archi export --format json     > report.json
archi export --format markdown > ARCHITECTURE.md
archi check   --max-distance 0.8   # exits 1 on violation - put this in CI
```

## Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-a, --ai` | off | Forces AI insights route; with missing key, API returns unavailable |
| `--api-key` | env | Gemini key (alternative to `GEMINI_API_KEY`), never persisted |
| `-p, --port` | 8080 | Local web server port |
| `--no-browser` | off | Don't auto-open the browser (WSL, SSH, VMs) |
| `-l, --lang` | auto | Force parser language (`go`, `js`, `ts`, `py`, `all`) |
| `--exclude` | builtin | Extra globs to skip in addition to defaults |
| `-h, --help`, `--version` | | Built-in |

## AI enrichment (optional)

The AI path runs as a second pass after metrics are computed:

- Enabled automatically when `GEMINI_API_KEY` or `--api-key` is present.
- `--ai` forces the AI route to be exposed (useful for validation/debugging).
- Never blocks the initial UI render — a skeleton shows first, insights stream in.
- Only metric payloads are sent to the model. **No source code or API keys leave the machine.**

## Supported languages and scope

- **Go**: parsed with stdlib `go/parser` + `go/ast`.
- **JS/TS**: parsed with Tree-sitter grammars for JavaScript/TypeScript.
- **Python**: parsed with Tree-sitter Python grammar.
- **Auto mode** (`--lang` omitted): detects repositories and can parse mixed-language projects.

Built-in excludes include: `node_modules`, `.git`, `vendor`, `dist`, `build`, `out`,
`venv`, `.venv`, `__pycache__`, `*_test.go`, `*.spec.ts`, `*.spec.js`, `*.spec.tsx`, `*.pyc`.

## Why this exists

Most architecture tools expose disconnected metrics. `archi` prioritizes diagnostic flow:
top-level risk KPIs, comparative deltas from the previous scan, module ranking tables, and a
visual map (main-sequence chart) as secondary support. Clicking a module opens focused context
for coupling, connascence, orphan/god blocks, and complexity.

## Documentation

Full design specs (in Portuguese) live in [`docs/`](./docs):

- [`docs/func_requirements.md`](./docs/func_requirements.md) — RFD-001..016 feature list
- [`docs/non_func_requirements.md`](./docs/non_func_requirements.md) — RNF-001..010 hard constraints
- [`docs/cli_config.md`](./docs/cli_config.md) — CLI surface and flags
- [`docs/stack.md`](./docs/stack.md) — mandated libraries
- [`docs/layout.md`](./docs/layout.md) — web UI layout / design principles

[`CHANGELOG.md`](./CHANGELOG.md) records every released version in Keep-a-Changelog format.

## Releasing

Releases are tag-driven. CI runs on every push/PR; a `v*.*.*` tag (or manual dispatch with a
`version` input) triggers the release job, which cross-compiles linux/amd64 and windows/amd64
binaries from Ubuntu, slices the matching `## [<version>]` section out of `CHANGELOG.md` into
the GitHub Release body, and fans out to nfpm-packaged DEB/RPM for Linux.

```sh
git tag v0.1.0 && git push origin v0.1.0
gh run watch
```

See [`.github/workflows/CI.yml`](./.github/workflows/CI.yml) for the full pipeline.

## License

MIT — see [`LICENSE`](./LICENSE).