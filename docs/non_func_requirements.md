# Requisitos Não Funcionais (RNF)

Este documento registra as restrições não funcionais aplicadas no projeto.

## 1. Desempenho e robustez

- **RNF-001: Meta de tempo para análise local**  
  Para projetos de porte médio (ordem de até 500 arquivos), o scan deve permanecer em faixa de execução rápida para uso interativo.

- **RNF-002: Meta de memória controlada**  
  O processamento deve manter footprint previsível e adequado para uso em máquina de desenvolvimento e CI.

- **RNF-003: Tolerância a erro de sintaxe**  
  Erros de parsing por arquivo não podem abortar o scan completo; o sistema registra warning e continua.

## 2. Distribuição e UX

- **RNF-004: Binário único**  
  Execução principal via binário único Go (sem dependências de runtime externas para o usuário final).

- **RNF-005: Frontend embutido**  
  A interface web é servida a partir de assets embutidos (`go:embed`) no binário.

- **RNF-006: Feedback de execução em CLI**  
  O usuário deve receber feedback visual durante o scan via spinner de terminal.

## 3. Operação offline e IA

- **RNF-007: Operação local por padrão**  
  Métricas, dashboard e API funcionam localmente sem dependência de serviços externos.

- **RNF-008: IA não bloqueante**  
  Quando habilitada, IA não bloqueia o carregamento inicial; os insights entram por streaming assíncrono.

- **RNF-009: Segurança de dados**  
  Nenhum código-fonte é enviado ao modelo. Somente métricas estruturais são serializadas para IA.

- **RNF-010: Gestão de chave de API**  
  Chave Gemini é lida apenas de `--api-key` ou `GEMINI_API_KEY`, sem persistência em disco pelo projeto.