# Configuração de CLI (`archi`)

Este documento descreve o comportamento **implementado hoje** no comando `archi`.

## Comando principal

Uso base:

```bash
archi [caminho_do_projeto] [flags]
```

- `caminho_do_projeto` é opcional; quando omitido, usa `.`.
- O fluxo padrão faz scan + sobe servidor local + abre navegador (exceto `--no-browser`).

Exemplos:

```bash
archi .
archi /home/usuario/projetos/meu-app
archi . --no-browser --port 3000
```

## Flags globais

### IA e autenticação

- `-a, --ai`  
  Força a rota de insights por IA a ficar ativa. Se não houver chave, a rota responde indisponível.

- `--api-key <string>`  
  Chave da API Gemini passada por flag (alternativa ao `GEMINI_API_KEY`).

Comportamento atual da IA:
- Se `--api-key` ou `GEMINI_API_KEY` existir, IA fica habilitada automaticamente.
- `--ai` é útil para forçar/validar o fluxo de IA mesmo sem env configurado.

### Servidor e browser

- `-p, --port <int>`  
  Porta do servidor local. Padrão: `8080`.

- `--no-browser`  
  Não abre navegador automaticamente.

### Escopo de parsing

- `-l, --lang <string>`  
  Força linguagem: `go`, `js`, `ts`, `py`, `all`.  
  Padrão: auto-detecção.

- `--exclude <strings>`  
  Padrões adicionais para ignorar no scan.

Excludes padrão já aplicados:
- diretórios: `node_modules`, `.git`, `vendor`, `dist`, `build`, `out`, `venv`, `.venv`, `__pycache__`
- arquivos: `*_test.go`, `*.spec.ts`, `*.spec.js`, `*.spec.tsx`, `*.pyc`

## Subcomandos

### `archi export`

Gera relatório sem iniciar servidor web.

```bash
archi export --format json
archi export --format markdown
```

- `--format`: `json` (padrão) ou `markdown`.
- Também aceita `md` como alias de markdown.

### `archi check`

Roda análise para uso em CI/CD e falha quando encontrar violação.

```bash
archi check --max-distance 0.8
```

- `--max-distance` padrão: `0.8`
- Se houver módulo com `D > limite`, o comando encerra com exit code não-zero.

## Contrato de help

- Root help customizado com `lipgloss`.
- Subcomandos usam help textual com flags herdadas + locais.
- Comandos disponíveis atualmente:
  - `export`
  - `check`

## Distribuição e atualização

As instruções de instalação, upgrade e reinstalação limpa por método de distribuição
(Homebrew, AUR, mise, binário direto, DEB e RPM) ficam centralizadas no `README.md`,
na seção **"Update to a new version"**.  
Este documento (`docs/cli_config.md`) descreve apenas o contrato de comandos e flags da CLI.