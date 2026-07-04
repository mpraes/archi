# Requisitos Funcionais de Domínio (RFD)

## 1. Mapeamento, Sintaxe e Granularidade Fina

- **RFD-001: Resolução de Escopo de Módulos**  
  O sistema deve varrer o diretório raiz informado e identificar automaticamente os limites de cada módulo ou pacote com base nos arquivos de manifesto da linguagem alvo (ex: `go.mod`, `package.json`).

- **RFD-002: Inventário de Estruturas Anatômicas (Blocos)**  
  O sistema deve realizar o parsing dos arquivos utilizando uma engine leve (Tree-sitter), catalogando individualmente cada Classe, Método e Função, registrando suas linhas de início/fim e escopo.

- **RFD-003: Mapeamento do Grafo de Chamadas Internas (Call Graph)**  
  O sistema deve identificar as relações de invocação de funções/métodos dentro do mesmo módulo para descobrir código órfão (nunca chamado) e funções centralizadoras (God Functions).

## 2. Métricas de Acoplamento e Sequência Principal

- **RFD-004: Construção do Grafo de Dependências**  
  O sistema deve extrair as diretivas de importação de cada arquivo para mapear as conexões externas e rastrear qual bloco de código específico (função/método) é o responsável por disparar aquele acoplamento eferente.

- **RFD-005: Cálculo de Acoplamento Aferente ($C_a$) e Eferente ($C_e$)**  
  O sistema deve computar o número de módulos externos que dependem de um módulo específico ($C_a$) e o número de módulos externos dos quais esse módulo depende ($C_e$).

- **RFD-006: Cálculo de Instabilidade ($I$) e Abstração ($A$)**  
  O sistema deve calcular o índice de instabilidade de cada módulo baseado na proporção de acoplamento, e o índice de abstração baseado na contagem de estruturas abstratas (interfaces ou assinaturas) em relação ao total de componentes concretos.

- **RFD-007: Cálculo da Distância da Sequência Principal ($D$)**  
  O sistema deve aplicar a fórmula $D = \left|A + I - 1\right|$ para cada módulo, classificando-os de forma determinística em zonas críticas (Zona da Dor ou Zona da Inutilidade).

- **RFD-008: Cálculo de Complexidade Ciclomática por Bloco**  
  O sistema deve contar o número de caminhos lineares independentes (estruturas de decisão como `if`, `for`, `switch`) em cada função ou método para enriquecer o diagnóstico de manutenibilidade.

## 3. Análise de Conascência Estática

- **RFD-009: Detecção de Conascência de Nome (CoN)**  
  O sistema deve sinalizar quando um componente depende explicitamente de assinaturas textuais rígidas (nomes exatos de métodos ou propriedades) de outro componente.

- **RFD-010: Detecção de Conascência de Significado / Valor (CoM)**  
  O sistema deve varrer o código em busca de valores literais idênticos (magic numbers ou strings hardcoded) compartilhados entre arquivos de módulos diferentes que possuam acoplamento lógico, inferindo dependência oculta de significado.

## 4. Integração Opcional com IA

- **RFD-011: Consolidação do Payload de Métricas**  
  O sistema deve serializar todos os resultados numéricos, grafos e alertas gerados em um esquema JSON compacto, omitindo completamente o código-fonte original.

- **RFD-012: Enriquecimento de Insights por IA (Execução Condicional)**  
  Caso a chave de API correspondente esteja presente no ambiente, o sistema deve enviar o JSON de métricas para o LLM e injetar na interface web recomendações arquiteturais textuais e acionáveis para os 3 pontos mais críticos detectados.