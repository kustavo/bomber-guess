# Marco 14: visual do tabuleiro (sprites)

**Status**: concluída (reaberta e concluída em 2026-10-01 com o CA-15 a CA-17)
**Roadmap**: `docs/ROADMAP.md`, marco 14
**Documentos de referência**: `specs/06-frontend/spec.md` (tabuleiro exibido, decisão 3) e `specs/07-editor/spec.md` (tabuleiro do editor)

## Objetivo

Hoje o tabuleiro é desenhado com quadrados e círculos lisos. Este marco troca esses desenhos por **sprites em pixel art** com cara de jogo, inspirados em Bomberman, na tela de assistir e no editor, com animação básica. Nada muda no que é exibido nem quando: muda só a aparência.

### Termos usados nesta spec

- **Sprite**: o desenho de um tipo de elemento do tabuleiro, numa grade de pixels (decisão 1).
- **Elemento**: o que ocupa ou marca uma casa no tabuleiro exibido (marco 6) ou no mapa em edição (marco 7): piso, bloco fixo, bloco destrutível, bomba, chama, jogador ou posição inicial.
- **Forma da chama**: o desenho de uma casa em chamas conforme a posição dela na explosão: centro, braço horizontal, braço vertical ou ponta (decisão 3).

## Escopo

**Dentro**
- Sprites originais, feitos para o projeto, de: piso, bloco fixo, bloco destrutível, bomba, chama (centro, braço horizontal, braço vertical e ponta em quatro direções) e jogador.
- O jogador usa a cor de cada um (a de hoje, por ordem em `estado.jogadores`) e continua com o número.
- Tela de assistir (marco 6): blocos, bombas, chamas e jogadores com os sprites. A bomba continua mostrando o `pavio_restante`, e o bloqueio continua com uma marca visível.
- Editor (marco 7): blocos e posições iniciais com os mesmos sprites; os botões da paleta mostram o sprite do item.
- Animação básica em CSS: bomba pulsando, chama tremulando e o deslize do jogador que já existe. Nada de animação por quadros.
- Testes automatizados (Vitest) da lógica e dos componentes.

**Fora** (fica para outro marco)
- Sprites de terceiros, imagens baixadas da internet ou cópia de sprites de jogos comerciais.
- Animação por quadros: caminhada por direção, explosão em vários quadros (decisão 4).
- Jogador virado para a direção em que andou (decisão 4).
- Sons.
- Sprite próprio para casa fechada pelo fechamento: continua aparecendo como bloco fixo (marco 6, CA-17; decisão 5).
- Mudanças na API, no backend ou nas regras.

## Regras cobertas

Nenhuma regra nova. O marco precisa respeitar o que as telas já mostram:

- `EST-06` a `EST-08`, `FIM-01`: jogador morto não aparece no tabuleiro (marco 6, CA-03).
- `DEC-09`: marca do movimento `BLOQUEADA` (marco 6, CA-04).
- `FEC-03` a `FEC-05`: casas fechadas aparecem como bloco fixo (marco 6, CA-17).
- `MAP-03`: posições iniciais numeradas na ordem (marco 7, CA-09).

## Critérios de aceitação

Vitest + Testing Library com jsdom, como nos marcos 6 e 7, salvo menção contrária.

### Sprites

- **CA-01**: **Dado** o conjunto de sprites, **quando** ele é carregado, **então** existe um sprite para cada um destes nomes: `piso`, `bloco-fixo`, `bloco-destrutivel`, `bomba`, `jogador`, `chama-centro`, `chama-horizontal`, `chama-vertical`, `chama-ponta-cima`, `chama-ponta-baixo`, `chama-ponta-esquerda`, `chama-ponta-direita`. Cada um é uma grade de pixels do tamanho da decisão 1 e usa só cores da paleta do conjunto.
- **CA-02**: **Dado** o tabuleiro de assistir ou do editor, **quando** ele é desenhado, **então** não há nenhuma requisição de rede por imagem: os sprites vêm com o próprio app (decisão 2).

### Forma da chama (lógica pura)

- **CA-03**: **Dadas** as `explosoes` de um relatório de etapa (cada uma com `origem` e `chamas`), **quando** a forma de cada casa em chamas é calculada, **então**, dentro de uma explosão:
  - a `origem` é `centro`;
  - uma casa na mesma linha da origem é `horizontal`, e na mesma coluna, `vertical`;
  - a casa mais distante da origem em cada direção é a `ponta` dessa direção (a mais à esquerda é `ponta-esquerda` etc.);
  - se duas ou mais explosões dão formas diferentes à mesma casa, ela é `centro`;
  - uma casa em `chamas` que não aparece em nenhuma explosão é `centro`.

### Tela de assistir

- **CA-04** (API-05): **Dada** uma resposta com blocos fixos, blocos destrutíveis, bombas, chamas e jogadores vivos, **quando** o tabuleiro é desenhado, **então** cada elemento usa o sprite do seu tipo, nas mesmas casas de hoje. Os atributos que os testes dos marcos 6 e 7 já usam (`data-tipo`, `data-x`, `data-y`, `data-jogador`, `data-pavio`, `data-bloqueado`) continuam iguais, e esses testes passam sem mudança.
- **CA-05**: **Dados** dois ou mais jogadores, **quando** o tabuleiro é desenhado, **então** cada jogador usa o sprite `jogador` com a sua cor de hoje (`corJogador`, pela ordem em `estado.jogadores`) e mostra o seu número.
- **CA-06**: **Dada** uma bomba com `pavio_restante` = n, **quando** é desenhada, **então** usa o sprite `bomba` e mostra n.
- **CA-07** (DEC-09): **Dado** um jogador em `movimentos_bloqueados`, **quando** a etapa é desenhada, **então** a marca de bloqueio continua visível sobre o sprite do jogador.
- **CA-08** (FEC-03, FEC-05): **Dada** uma casa fechada pelo fechamento, **quando** a etapa é desenhada, **então** ela usa o sprite `bloco-fixo`.
- **CA-09**: **Dadas** as chamas de uma etapa, **quando** são desenhadas, **então** cada casa usa o sprite da sua forma (CA-03).

### Editor

- **CA-10** (MAP-03): **Dado** um mapa em edição com blocos e posições iniciais, **quando** o tabuleiro do editor é desenhado, **então** blocos usam os sprites dos blocos, e cada posição inicial usa o sprite `jogador` com a cor e o número da posição.
- **CA-11**: **Dada** a paleta do editor, **quando** é exibida, **então** os botões de bloco fixo, bloco destrutível e posição inicial mostram o sprite do item, e a borracha mostra o `piso`.

### Animação

- **CA-12**: **Dados** bombas e chamas no tabuleiro, **quando** é desenhado, **então** bombas e chamas têm a marca de animação (classe CSS) correspondente. Com `prefers-reduced-motion: reduce`, as animações ficam desligadas (decisão 6; verificado por teste da folha de estilos ou, se o jsdom não permitir, na verificação manual do CA-14).

### Reabertura (2026-10-01, depois do uso real)

- **CA-15** (decisão 8): **Dada** uma bomba no tabuleiro, **quando** ela é animada, **então** só aumenta e diminui de tamanho, de leve (no máximo 4 %), sempre em torno do centro da casa: não sai do lugar nem balança. Com movimento reduzido, continua parada (CA-12).
- **CA-16** (decisão 9): **Dado** o tabuleiro de assistir ou do editor, **quando** é desenhado, **então** há uma linha clara e fina entre todas as casas vizinhas, por cima do piso e dos blocos e por baixo de bombas, chamas e jogadores, e essa linha não recebe cliques.
- **CA-17** (EST-04, BOM-02, DEC-09; decisão 10): **Dado** um jogador vivo no tabuleiro de assistir, **quando** o mouse passa sobre ele, **então** aparece um cartão com:
  - o id do jogador (ex.: `jogador_1`);
  - a versão do bot (`bot_versao`);
  - as bombas que ainda pode plantar neste turno, `bombas_por_turno` − usadas;
  - as bombas usadas no turno no formato "usou u de n (u/n)", em que u conta as ações `PLANTAR` com resultado `EXECUTADA` nos relatórios das etapas 1…k já exibidas e n é `bombas_por_turno`.

  No planejamento (etapa 0), u é 0. O cartão some quando o mouse sai do jogador, e os números acompanham a etapa exibida enquanto ele está aberto.

### Verificação

- **CA-13**: Em `frontend/`, `npm run check` sem erros e `npm test` passando, incluindo todos os testes dos marcos 6 e 7.
- **CA-14**: **Dado** o servidor Go e o frontend rodando, **quando** se cria uma partida pelo editor e se assiste a ela, **então** os sprites aparecem nítidos (sem borrão ao ampliar), em tema claro e escuro, e as animações funcionam. Verificação manual com captura de tela, registrada no `tarefas.md`.

## Decisões

1. **Tamanho e formato dos sprites.** Decisão: grade de 16 × 16 pixels por casa, desenhada com retângulos de 1 pixel em SVG, com `shape-rendering: crispEdges`. Fica nítida em qualquer tamanho de tabuleiro, é fácil de testar pelo DOM e não precisa de ferramenta de imagem.
2. **De onde vêm os sprites.** Decisão: escritos à mão no próprio código do frontend, como dados (linhas de texto com uma letra por cor), e convertidos em SVG. Nada de arquivo de imagem nem rede (CA-02), e quem mexer no projeto edita o desenho no próprio texto.
3. **Forma da chama.** O relatório de etapa traz `explosoes`, cada uma com `origem` e `chamas`. Decisão: calcular a forma por explosão, a partir da origem (CA-03), para dar o efeito de "cruz" do Bomberman sem mudar a API. Onde duas explosões se cruzam, a casa vira `centro`.
4. **Direção do jogador e animação por quadros.** Decisão: fora deste marco. O jogador tem um sprite só, visto de frente. Animação por quadros pede estado de animação por jogador e fica para depois, se valer a pena.
5. **Casa fechada.** Decisão: continua com o sprite de bloco fixo, como diz o CA-17 do marco 6. Um sprite próprio (ex.: parede de metal) seria bonito, mas mudaria esse critério; pode entrar depois com uma reabertura pequena.
6. **Movimento reduzido.** Decisão: respeitar `prefers-reduced-motion`: sem pulsar e sem tremular, mas com o deslize do jogador (que ajuda a seguir a partida) encurtado para um corte seco.
7. **Piso.** Decisão: o fundo do tabuleiro passa a ser o sprite `piso` repetido em cada casa (grama, como no Bomberman clássico), no lugar do retângulo verde liso. No editor, cada casa clicável usa esse piso.
8. **Pulso da bomba (reabertura).** O pulso de 9 % ficou exagerado e, como a origem da escala era a caixa do desenho (e não a casa), a bomba parecia sair do lugar. Decisão: escala de no máximo 4 %, com a origem no centro da casa, num ritmo mais lento (cerca de 1,2 s por ciclo).
9. **Linhas entre as casas (reabertura).** Decisão: uma grade de linhas claras e finas (branco com pouca opacidade) desenhada uma vez sobre todo o tabuleiro, nas duas telas. No editor, ela substitui o contorno escuro das casas clicáveis, para as duas telas ficarem iguais.
10. **Cartão do jogador (reabertura).** O `<title>` nativo do SVG demora a aparecer e não dá para formatar. Decisão: um cartão HTML próprio, posicionado junto do jogador, que aparece ao passar o mouse e também ao focar o jogador pelo teclado. "Bombas que ainda tem" segue o EST-04: o estoque é por turno e se renova no início de cada turno; bombas de turnos anteriores ainda no tabuleiro não contam. O `<title>` atual continua (ele carrega o `BLOQUEADA` do CA-07). No editor não há cartão: lá ainda não existem jogadores, só posições iniciais.
