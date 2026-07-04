Para o **archi**, o design dos argumentos (args) e flags no CLI deve ser extremamente intuitivo e seguir os padrões de ferramentas consolidadas no ecossistema Linux/Go. O objetivo é que o usuário consiga rodar o app de forma simples, mas com controle total através do terminal.

Usando a biblioteca `Cobra`, aqui está o design ideal de comandos, argumentos e flags para o **`archi`**:

---

## 1. O Comando Principal (Uso Padrão)

O uso mais comum não deve exigir comandos complexos. O argumento principal deve ser apenas o **caminho do diretório** a ser analisado. Se omitido, ele assume o diretório atual (`.`).

```bash
# Analisa o diretório atual e abre o browser
archi .

# Analisa um projeto em outro caminho
archi /home/usuario/projetos/meu-app

```

---

## 2. As Flags Principais (Opções Globais)

As flags modificam o comportamento do escaneamento ou ativam recursos específicos.

### 🔑 Autenticação e IA

* `-a, --ai`: Força ou ativa o modo de insights por IA. (Por padrão, o `archi` pode detectar automaticamente se a variável `GEMINI_API_KEY` existe no ambiente, mas essa flag é útil para forçar o comportamento ou validar a conexão).
* `--api-key <string>`: Permite passar a chave diretamente pelo comando, caso o usuário não queira (ou não saiba) configurar uma variável de ambiente.

### 🌐 Comportamento do Servidor/Browser

* `-p, --port <int>`: Define a porta onde o servidor web local vai rodar. *(Default: `8080`)*. Se a porta 8080 estiver ocupada, o `archi` pode procurar a próxima disponível automaticamente, mas a flag dá controle ao usuário.
* `--no-browser`: Roda o escaneamento e sobe o servidor, mas **não** abre o navegador automaticamente. Excelente para quem está rodando o app dentro de uma máquina virtual, WSL sem interface gráfica ou SSH.

### 🧹 Filtros de Escopo (Crucial para Performance)

* `-l, --lang <string>`: Força a linguagem do parser (ex: `go`, `ts`, `py`). Útil se o projeto for poliglota e o usuário quiser focar em apenas uma stack. *(Default: Auto-detectar)*.
* `--exclude <strings>`: Lista de pastas ou padrões para o `archi` ignorar no escaneamento. Por padrão, pastas como `node_modules`, `.git`, `vendor` e arquivos de teste (`*_test.go`, `*.spec.ts`) já devem ser ignoradas automaticamente para poupar memória.

---

## 3. Subcomandos (Para Casos de Uso Avançados)

À medida que o projeto amadurecer, você pode adicionar subcomandos específicos que rodam tarefas sem necessariamente abrir o navegador.

### `archi export`

Gera um relatório estático das métricas sem abrir o servidor local.

```bash
# Exporta o grafo de acoplamento e conascência para um arquivo JSON estruturado
archi export --format json > report.json

# Exporta em formato Markdown para colar em um Readme ou documentação interna
archi export --format markdown

```

### `archi check` (Ideal para CI/CD)

Roda a análise e retorna um código de saída (`exit code`) de erro caso alguma métrica arquitetural seja violada. Perfeito para colocar no GitHub Actions da empresa e impedir que códigos muito acoplados entrem em produção.

```bash
# Falha se algum módulo tiver uma Distância da Sequência Principal (D) maior que 0.8
archi check --max-distance 0.8

```

---

## Como isso se traduz no Help do Terminal (`archi --help`)

Ao digitar `archi` ou `archi --help`, o `Cobra` vai gerar uma tela limpa e minimalista usando o `Lipgloss` para formatar o texto com cores elegantes:

```text
Archi - Ferramenta de análise estática e diagnóstico visual de arquitetura de software.

Uso:
  archi [caminho_do_projeto] [flags]
  archi [comando]

Comandos Disponíveis:
  export      Exporta as métricas do projeto (JSON/Markdown)
  check       Valida limites arquiteturais no pipeline de CI

Flags:
  -a, --ai                 Ativa os insights do consultor virtual via IA
      --api-key string     Chave de API do Gemini (alternativa à variável de ambiente)
  -p, --port int           Porta para o servidor web local (default 8080)
      --no-browser         Não abre o navegador automaticamente após o escaneamento
  -l, --lang string        Força uma linguagem específica para o parsing
      --exclude strings    Pastas ou arquivos adicionais para ignorar
  -h, --help               Ajuda para o comando archi

Exemplo:
  archi . --ai --port 3000

```

Este design mantém o comando curto e direto para o dia a dia, mas dá superpoderes para o desenvolvedor avançado que quer automatizar a análise da arquitetura.