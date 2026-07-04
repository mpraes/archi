# Requisitos Não Funcionais (RNF)

## 1. Desempenho, Eficiência e Escala

- **RNF-001: Tempo de Resposta (Métrica No-Wait)**  
  O tempo total de análise estática local (parsing e cálculo de métricas) para um projeto de médio porte (até 500 arquivos) deve ser inferior a 3 segundos.

- **RNF-002: Pegada de Memória Limitada**  
  O consumo de memória RAM do binário durante o processamento do código não deve exceder 200MB, garantindo que a ferramenta rode de forma leve em qualquer máquina ou pipeline de CI.

- **RNF-003: Resiliência a Erros de Sintaxe**  
  O parser não deve interromper a execução caso encontre um arquivo de código corrompido ou com erro de sintaxe. O sistema deve registrar o aviso, ignorar o arquivo problemático e continuar a análise do restante do projeto.

## 2. Distribuição e Experiência do Usuário (Zero Atrito)

- **RNF-004: Binário Único e Auto-Contido**  
  A aplicação CLI deve ser compilada em um binário estático único para as principais plataformas (Linux, macOS, Windows). Não deve ser exigida nenhuma dependência pré-instalada (como Node.js, Python ou Docker).

- **RNF-005: Interface Web "Standalone" Embutida**  
  O servidor web local iniciado pela CLI deve servir a interface gráfica a partir de assets HTML/JS embutidos (embedded) diretamente no binário compilado, funcionando sem necessidade de download de dependências da internet.

- **RNF-006: Feedback Visual no Terminal**  
  A interface de linha de comando deve fornecer feedback em tempo real através de spinners ou barras de progresso elegantes enquanto realiza o parsing e o cálculo das métricas.

## 3. Modos de Operação, Segurança e Privacidade

- **RNF-007: Modo Offline Autônomo (Default)**  
  O sistema deve ser 100% operacional sem conexão com a internet. Todos os gráficos, cálculos e mapeamentos de conascência devem carregar normalmente na interface do browser de forma local.

- **RNF-008: Injeção de IA Assíncrona e Não-Bloqueante**  
  Quando a IA estiver ativa, a chamada de rede para a API externa não deve travar a renderização inicial da interface web. O painel deve abrir instantaneamente com os gráficos locais, exibindo um estado de carregamento (*skeleton screen*) apenas no componente de insights da IA até que a resposta chegue.

- **RNF-009: Privacidade Total de Código e Chaves de API**  
  Nenhum trecho de código-fonte ou chave de API do usuário pode ser enviado ou armazenado por servidores de terceiros gerenciados pelo projeto. O app deve ser estritamente _stateless_ e as chaves de API devem ser lidas exclusivamente de variáveis de ambiente (`GEMINI_API_KEY`, etc.), nunca salvas em arquivos locais de configuração.