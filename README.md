# archi

> Static code analysis you can actually look at. Point it at a codebase, get an architectural
> MRI — coupling, instability, abstraction, main-sequence distance, cyclomatic complexity —
> served in a local web UI built around the main-sequence scatter plot.

`archi` scans a project, computes architectural metrics per module, and renders them in an
embedded dark-theme UI (D3 scatter plot + side panel). It runs 100% offline by default; an
optional Gemini enrichment streams in as a second pass and never blocks the initial render.
For CI it ships a `check` subcommand that exits non-zero on main-sequence violations, so a
growing codebase can't quietly drift into the "zone of pain" without you noticing.

- Single static binary, no runtime deps. Go 1.26+, `embed`-ded frontend.
- Resilient parser: logs and skips syntax-broken files, never aborts the scan.
- Source code and API keys never leave the machine. The LLM payload carries metrics only.

## Install

```sh
# Homebrew (macOS or Linux)
brew tap mpraes/tap && brew install archi

# AUR (Arch Linux) - prebuilt binary
yay -S archi-bin

# mise
mise use -g github:mpraes/archi

# Direct binary download
curl -fsSL https://github.com/mpraes/archi/releases/latest/download/archi-v0.1.0-linux-amd64.tar.gz | tar xz

# DEB / RPM
sudo dpkg -i archi_0.1.0_amd64.deb
# or: sudo rpm -i archi-0.1.0-1.x86_64.rpm

# Build from source
git clone https://github.com/mpraes/archi && cd archi
(cd web && npm install && npm run build) && go build -o archi ./cmd/archi
```

Pre-built binaries are published for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64,
and windows/amd64. Checksums are attached to every release.

## Quickstart

```sh
archi .                         # scan current dir, open browser at :8080
archi /path/to/project --port 3000 --no-browser
archi . --exclude "internal/migrations,*.pb.go"
archi . --ai                    # opt-in Gemini insights (needs GEMINI_API_KEY)
```

Headless modes for scripts and CI:

```sh
archi export --format json     > report.json
archi export --format markdown > ARCHITECTURE.md
archi check   --max-distance 0.7   # exits 1 on violation - put this in CI
```

## Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-a, --ai` | off | Activate opt-in Gemini insights |
| `--api-key` | env | Gemini key (alternative to `GEMINI_API_KEY`); never persisted |
| `-p, --port` | 8080 | Local web server port |
| `--no-browser` | off | Don't auto-open the browser (WSL, SSH, VMs) |
| `-l, --lang` | auto | Force parser language (`go`, `ts`) |
| `--exclude` | builtin | Extra globs to skip (defaults already ignore `node_modules`, `.git`, `vendor`, `*_test.go`, `*.spec.ts`) |
| `-h, --help`, `--version` | | Built-in |

## AI enrichment (optional)

The AI path is opt-in and runs as a second pass after metrics are computed:

- Disabled by default. Detector reads `GEMINI_API_KEY` from env (or `--api-key`).
- Never blocks the initial UI render — a skeleton shows first, insights stream in.
- Only metric payloads are sent to the model. **No source code or API keys leave the machine.**

## Why this exists

Most "architecture" tooling dumps raw numbers or a static diagram. archi's central view is
the main-sequence scatter plot (abstraction vs. instability with the balanced line drawn
through it), so the "zone of pain" and "zone of uselessness" are immediately visible. Pick a
module in the side panel and you see its coupling, connascence, complexity, and — when AI is
enabled — a plain-language explanation of what's rigid and what to do about it.

## Documentation

Full design specs (in Portuguese) live in [`docs/`](./docs):

- [`docs/func_requirements.md`](./docs/func_requirements.md) — RFD-001..012 feature list
- [`docs/non_func_requirements.md`](./docs/non_func_requirements.md) — RNF-001..009 hard constraints
- [`docs/cli_config.md`](./docs/cli_config.md) — CLI surface and flags
- [`docs/stack.md`](./docs/stack.md) — mandated libraries
- [`docs/layout.md`](./docs/layout.md) — web UI layout / design principles

[`CHANGELOG.md`](./CHANGELOG.md) records every released version in Keep-a-Changelog format.

## Releasing

Releases are tag-driven. CI runs on every push/PR; a `v*.*.*` tag (or manual dispatch with a
`version` input) triggers the release job, which compiles native binaries for every target,
slices the matching `## [<version>]` section out of `CHANGELOG.md` into the GitHub Release
body, and fans out to Homebrew, AUR, and nfpm-packaged DEB/RPM.

```sh
git tag v0.1.0 && git push origin v0.1.0
gh run watch
```

See [`.github/workflows/CI.yml`](./.github/workflows/CI.yml) for the full pipeline.

## License

MIT — see [`LICENSE`](./LICENSE).