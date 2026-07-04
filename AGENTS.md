# AGENTS.md

Repo is active (Go + web dashboard). Keep docs and implementation aligned; docs remain the normative high-level reference.

## Project

`archi` — static-analysis CLI that parses a codebase, computes architectural metrics (coupling $C_a/C_e$, instability $I$, abstraction $A$, main-sequence distance $D$, cyclomatic complexity, connascence), and serves an embedded web UI locally for visual diagnosis. Optional AI (Gemini) enrichment of insights.

Reference specs before implementing or reviewing:
- `docs/func_requirements.md` — RFD-001..016, the authoritative feature list
- `docs/non_func_requirements.md` — RNF-001..010, hard constraints (see below)
- `docs/cli_config.md` — exact CLI surface and flags
- `docs/stack.md` — mandated libraries
- `docs/layout.md` — web UI layout / design principles

## Mandated stack (from `docs/stack.md`)

- Go. Single static binary, no runtime deps (RNF-004).
- CLI: `github.com/spf13/cobra`; styled help via `github.com/charmbracelet/lipgloss`; spinner implemented in `internal/ui`.
- Parsing: `github.com/tree-sitter/go-tree-sitter` with per-language grammar packages. For Go targets only, stdlib `go/ast` + `go/parser` is preferred over tree-sitter.
- HTTP server: `github.com/go-chi/chi/v5` (or `net/http` Go 1.22+); frontend embedded via stdlib `embed`.
- AI: `google.golang.org/genai`, streaming, model `gemini-2.5-flash`.
- Frontend built with Vite + TypeScript; graphs via D3.js or Cytoscape.js. Built assets get `go:embed`-ded into the binary.

## Hard constraints (RNFs that change implementation choices)

- Parser must be resilient: never abort on a syntax-broken file — log, skip, continue (RNF-003).
- 100% offline for core analysis; AI must not block initial UI render — show skeleton and stream in (RNF-007, RNF-008).
- **No source code or API keys may leave the machine.** Metrics payload sent to LLM must omit source; keys read only from env (`GEMINI_API_KEY`) or `--api-key`, never persisted to disk (RNF-009).
- Perf budgets: <3s analysis for ≤500 files; <200MB RAM (RNF-001/002).

## CLI contract (`docs/cli_config.md`)

Default: `archi [path]` — path defaults to `.`; runs scan then opens browser at `http://localhost:8080`.

Flags: `-a/--ai`, `--api-key <string>`, `-p/--port <int>` (default 8080), `--no-browser`, `-l/--lang <string>`, `--exclude <strings>`.

Subcommands: `archi export --format json|markdown` (no server), `archi check --max-distance <float>` (exits non-zero on violation, for CI).

Default excludes already include `node_modules`, `.git`, `vendor`, `dist`, `build`, `out`, `venv`, `.venv`, `__pycache__`, `*_test.go`, `*.spec.ts`, `*.spec.js`, `*.spec.tsx`, `*.pyc`.

## Conventions taught by the docs

- UI copy should be human/action-oriented ("Onde o código está rígido"), not raw metric names — see `docs/layout.md` before writing frontend strings.
- Default theme is dark; analytical report + comparative KPIs are central, with scatter plot as secondary support.

## Build & dev (verified)

- Go toolchain on this machine lives at `/usr/local/go/bin` (not on `$PATH` by default): prefix commands with `export PATH=/usr/local/go/bin:$PATH`.
- Build CLI: `go build -o archi ./cmd/archi`. Verify: `go vet ./...` (no test suite yet).
- Frontend lives in `web/` (Vite + TS + d3). Its compiled output in `web/dist/` is `go:embed`-ded into the binary at build time via `web/embed.go` — so you **must** run `npm install && npm run build` inside `web/` before `go build`, otherwise the embedded UI is stale/empty.
- Frontend dev server proxies `/api` to `http://127.0.0.1:8080` (see `web/vite.config.ts`); run the Go server (`archi --no-browser`) and `npm run dev` together for hot-reload.
- Module path is `github.com/mpraes/archi` (matches git remote). Sub-packages live under `internal/{cli,parser,analyzer,server,ai,ui,model}`; `web/` holds embed + frontend. `cmd/archi/main.go` is the entrypoint.
- Parser dispatch is in `internal/parser/parser.go`; Go uses stdlib `go/ast`, JS/TS and Python use tree-sitter bindings.
- Run on this repo: `./archi export --format json` and `./archi check --max-distance 0.7` (exits 1 on violations, for CI).