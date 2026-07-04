# Stack Técnica (estado atual)

Este documento descreve a stack efetivamente usada hoje no projeto.

## 1. CLI e experiência de terminal

- Linguagem principal: **Go** (`go 1.26.x`).
- Framework de CLI: [`github.com/spf13/cobra`](https://github.com/spf13/cobra).
- Help estilizado: [`github.com/charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss).
- Spinner: implementação interna em Go (`internal/ui`), sem runtime externo.

## 2. Parsing e análise estática

- **Go**: parsing com stdlib `go/parser` + `go/ast`.
- **JavaScript/TypeScript**: Tree-sitter com:
  - [`github.com/tree-sitter/go-tree-sitter`](https://github.com/tree-sitter/go-tree-sitter)
  - [`github.com/tree-sitter/tree-sitter-javascript`](https://github.com/tree-sitter/tree-sitter-javascript)
  - [`github.com/tree-sitter/tree-sitter-typescript`](https://github.com/tree-sitter/tree-sitter-typescript)
- **Python**: Tree-sitter com:
  - [`github.com/tree-sitter/tree-sitter-python`](https://github.com/tree-sitter/tree-sitter-python)

## 3. Backend HTTP local

- Roteador HTTP: [`github.com/go-chi/chi/v5`](https://github.com/go-chi/chi/v5).
- API local para métricas, detalhes de módulo e streaming SSE de IA.
- Assets frontend embutidos no binário com `embed`.

## 4. IA opcional

- SDK: `google.golang.org/genai`.
- Modelo padrão: `gemini-2.5-flash`.
- Streaming de resposta para frontend via SSE (`/api/ai/insights`).

## 5. Frontend embutido

- Build tool: **Vite**.
- Linguagem: **TypeScript**.
- Visualização: **D3.js**.
- Saída compilada em `web/dist`, embutida no binário por `go:embed` (`web/embed.go`).

## 6. Fluxo de execução

1. CLI recebe argumentos e flags.
2. Parser varre arquivos com filtros de exclusão padrão + extras.
3. Analyzer calcula métricas arquiteturais e hotspots.
4. Servidor local expõe API + dashboard embutido.
5. Browser é aberto automaticamente (salvo `--no-browser`).
6. IA (se habilitada) entra em segundo plano por streaming.