# Layout da Interface Web (estado atual)

O dashboard atual prioriza **diagnóstico comparativo** e não apenas visualização geométrica.

## Estrutura principal

### 1. Sidebar fixa

- Branding (`Archi / Painel Analítico`).
- Navegação lateral (resumo, módulos, conascência, risco, histórico).

### 2. Topbar + Hero

- Status curto com total de módulos e hotspots ativos.
- Badge de baseline indicando se há comparação com scan anterior.
- Hero com visão geral da arquitetura e indicadores-chave.

### 3. Grade de KPIs

Cards com:
- hotspots
- distância média
- instabilidade média
- complexidade máxima média
- complexidade total média
- total de conascências

Cada card mostra delta vs baseline quando disponível.

### 4. Área de trabalho principal

- **Centro (foco):** relatório analítico textual + tabela completa de módulos.
- **Direita:** painel de atividade comparativa e painel de detalhe do módulo selecionado.
- **Mapa visual (D3):** disponível em `<details>` como apoio secundário.

### 5. Painel de IA

- Exibido apenas quando IA está habilitada.
- Streaming incremental com skeleton inicial.

## Interações essenciais

1. Selecionar módulo na tabela (ou no mapa) abre contexto lateral:
   - acoplamento (`Ca`, `Ce`)
   - abstração
   - complexidade
   - god blocks
   - blocos órfãos
   - conascências relacionadas
2. Dashboard salva snapshot local e compara com a próxima execução.
3. Listas de regressões e evoluções destacam o que piorou/melhorou.

## Princípios de UX aplicados

- Linguagem orientada à ação (ex.: "O que está rígido").
- Progressive disclosure (detalhe sob demanda por módulo).
- Priorização de risco atual + tendência (baseline), não só estado estático.
- Mapa da sequência principal mantido como ferramenta de contexto visual, não como único centro de decisão.
