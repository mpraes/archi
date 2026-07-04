# Requisitos Funcionais de Domínio (RFD)

Este documento reflete o **estado funcional implementado**.

## 1. Descoberta e parsing

- **RFD-001: Descoberta de módulos por diretório/language path**  
  O sistema varre a raiz informada e organiza módulos por diretório (com resolução por `go.mod` para nomes Go quando disponível).

- **RFD-002: Parsing resiliente multi-linguagem**  
  O parser suporta Go, JavaScript/TypeScript e Python, com auto-detecção ou força por `--lang`.

- **RFD-003: Inventário anatômico por arquivo**  
  O sistema extrai funções, métodos e tipos/classes, com linhas de início/fim, complexidade e chamadas detectadas.

## 2. Métricas arquiteturais

- **RFD-004: Grafo de dependências internas**  
  O sistema resolve imports internos por módulo para construir acoplamento eferente.

- **RFD-005: Acoplamento aferente/eferente (`Ca`/`Ce`)**  
  O sistema computa dependências de entrada e saída para cada módulo.

- **RFD-006: Instabilidade (`I`) e abstração (`A`)**  
  O sistema calcula instabilidade por proporção de acoplamento e abstração por razão de elementos abstratos/concretos.

- **RFD-007: Distância da sequência principal (`D`)**  
  O sistema aplica `D = |A + I - 1|` por módulo.

- **RFD-008: Complexidade ciclomática**  
  O sistema calcula complexidade por bloco e agrega `maxComplexity` e `totalComplexity` por módulo.

## 3. Diagnóstico de risco estrutural

- **RFD-009: Código órfão**  
  O sistema marca blocos não referenciados no call map coletado.

- **RFD-010: God blocks**  
  O sistema marca blocos com complexidade acima do limiar heurístico atual.

- **RFD-011: Conascência estática**  
  O sistema detecta:
  - conascência de nome (`kind: "name"`) por chamadas entre módulos;
  - conascência de significado (`kind: "meaning"`) por literais compartilhados entre módulos.

- **RFD-012: Hotspots**  
  O sistema destaca módulos em risco (zona de dor heurística e/ou presença de god blocks).

## 4. API local e interface

- **RFD-013: API de métricas local**  
  Endpoints locais:
  - `GET /api/metrics`
  - `GET /api/modules/{module}`
  - `GET /api/ai/enabled`
  - `GET /api/ai/insights` (SSE)

- **RFD-014: Dashboard com baseline local**  
  A interface compara o scan atual com o último snapshot salvo no `localStorage`, mostrando deltas de risco e regressões/evoluções.

## 5. IA opcional

- **RFD-015: Payload sem código-fonte**  
  O payload para IA contém apenas métricas e relações estruturais, sem conteúdo de código.

- **RFD-016: Streaming de insights**  
  Insights são enviados por SSE em chunks para não bloquear a renderização inicial do dashboard.