Para entregar essa sensação imediata de clareza ("uau, agora eu entendi meu código"), a interface precisa fugir completamente do visual "tabela cheia de números" do SonarQube ou de painéis de computação em nuvem carregados de informações.

O segredo aqui é o **minimalismo cognitivo**: remover o ruído visual, destacar apenas o que está quebrado e usar analogias visuais óbvias.

Abaixo está o conceito estrutural de como essa interface web embutida deve se parecer e se comportar para gerar esse impacto.

---

## O Layout: "Três Cliques para o Diagnóstico"

A tela é dividida em apenas três zonas visuais limpas, dispostas de forma fluida, usando um tema escuro confortável por padrão (estilo editor de código moderno).

### 1. O Cabeçalho (Status Geral)

Em vez de notas de A a F ou dezenas de gráficos de pizza, o topo da tela traz apenas uma frase direta e humana:

> 📊 **Status da Arquitetura:** Seu projeto possui **12 módulos**. O design está predominantemente **Saudável**, mas existem **2 Hotspots** críticos que estão travando sua evolução.

---

### 2. A Área Central: O Mapa de Calor da Sequência Principal

Aqui fica o coração visual da ferramenta. Um gráfico 2D minimalista feito com D3.js onde cada módulo do seu software é uma "bolinha" flutuante.

* **A Linha Ideal:** Uma linha diagonal suave corta o gráfico (A Sequência Principal). As bolinhas que representam seus módulos devem orbitar perto dela.
* **O Código de Cores Emocional:**
* Bolinhas **Verdes/Cinzas**: Módulos saudáveis, equilibrados.
* Bolinhas **Vermelho-Alerta**: Módulos na **Zona da Dor** (muito grandes, rígidos e acoplados).


* **O Efeito "Ficou Fácil":** O usuário não precisa saber o que é Acoplamento Aferente. Ele só precisa olhar para o gráfico e ver: *"Ok, por que o pacote `/payments` está tão grande, vermelho e isolado lá embaixo?"*

---

### 3. A Lateral Dinâmica: O Painel de Contexto (Onde a IA brilha)

Esta barra lateral só ganha conteúdo quando o usuário clica em uma das bolinhas do gráfico. Ela funciona como um raio-X instantâneo do módulo selecionado.

#### Sem IA (Modo Padrão)

Mostra uma lista curta e escaneável com o diagnóstico puramente matemático:

* **Quem ele puxa (Eferente):** Depende de 8 módulos.
* **Quem puxa ele (Aferente):** 0 módulos dependem dele.
* **Conascência Detectada:** A string `"STATUS_PENDING"` está duplicada rigidamente entre este arquivo e o arquivo `shipping.go`.

#### Com IA (Modo Enriquecido)

O componente de IA substitui os termos técnicos por uma explicação de um parágrafo que parece um colega sênior conversando com você:

> ✨ **Insight do Consultor Virtual:**
> *"Este módulo `/payments` virou um nó cego. Ele tenta resolver tudo sozinho (Complexidade Ciclomática 28) e está rigidamente amarrado ao `/shipping`. Se você mudar as regras de envio, seus pagamentos vão quebrar. Experimente quebrar a função `Process()` em duas."*

---

## Princípios de Design para causar o efeito "Ficou Fácil"

Para que o layout entregue essa experiência, o frontend precisa seguir três regras de ouro:

1. **Ocultar por Padrão (Progressive Disclosure):** Não mostre as funções, métodos e linhas de código logo de cara. Mostre o módulo. O usuário quer ver as funções? Ele clica e o módulo se "expande" em um micro-grafo na tela.
2. **Linguagem de Cores Consistente:** Se a conascência de significado é um elo perigoso, a linha que conecta os dois módulos no grafo deve piscar em vermelho quando o usuário passa o mouse por cima, mostrando visualmente o "efeito dominó" de uma alteração.
3. **Foco em Ação, Não em Estatística:** Em vez de uma seção chamada *"Métricas de Instabilidade"*, use títulos como *"Onde o código está rígido"* ou *"O que está pronto para ser isolado"*.

Esse design minimalista e focado em conexões humanas limpa a névoa mental de quem está lidando com um projeto legado gigante. Olhar para a tela vai parecer ver o mapa de uma cidade do alto, em vez de tentar se achar no meio de um engarrafamento de linhas de código.