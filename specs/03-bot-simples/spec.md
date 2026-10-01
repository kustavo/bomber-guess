# Marco 03: bot simples

**Status**: concluída
**Roadmap**: `docs/ROADMAP.md`, marco 3
**Documentos de referência**: `docs/BOTS.md`, `docs/REGRAS.md`

## Objetivo

Entregar o primeiro bot, `aleatorio-v1`: anda aleatoriamente (com uma semente explícita), planta bombas de vez em quando e foge das explosões que consegue prever. Não pretende jogar bem; serve para haver algo jogando nos marcos seguintes (terminal, servidor, frontend) e como referência de bot que cumpre o contrato de `docs/BOTS.md`.

### Termos usados nesta spec

- **Simulação do turno**: resolver o turno (as regras de `docs/REGRAS.md`, seções 4 a 7) a partir do estado recebido, com o plano candidato do bot e todos os outros jogadores executando `ESPERAR`. O bot só prevê o que já está no tabuleiro e o que ele mesmo planta; não tenta adivinhar os adversários.
- **Plano seguro**: plano em que, na simulação do turno, o bot termina `VIVO` e a sua casa final não é alcançada pela explosão de nenhuma bomba que sobra no tabuleiro ao fim do turno (considerando pilhas, reações em cadeia e blocos como ficam ao fim do turno).
- **Plano sobrevivente**: plano em que, na simulação do turno, o bot termina `VIVO` (sem exigir nada da casa final).

## Escopo

**Dentro**
- Subpacote `backend/internal/bots/aleatorio`, versão `aleatorio-v1`, implementando a interface `Bot`.
- Aleatoriedade controlada por uma semente explícita, escolhida ao criar o bot.
- Previsão de explosões pela simulação do turno: bombas de turnos anteriores, bombas que o próprio bot planta, pilhas, reações em cadeia e blocos destrutíveis que param o fogo.
- Escolha do plano por prioridade: seguro, senão sobrevivente, senão o que sobrevive mais etapas.
- Plantar bombas só em plano seguro.
- Testes do bot exigidos por `docs/BOTS.md` (planos válidos no mapa de exemplo, sem pânico com estado vazio ou jogador morto) e testes de fuga em tabuleiros pequenos.

**Fora** (fica para outro marco)
- Catálogo de bots (registro de versões, `GET /bots`): marco 4 cria o que o terminal precisar; a API é do marco 5.
- Chamar o bot com limite de tempo e tratar `panic` (BOT-02) e gravar a saída bruta (BOT-03): marcos 4 e 5.
- Prever as ações dos adversários, caçar adversários, estratégia de vitória: bots do marco 10.
- Histórico de turnos (`Historico`): versão 2.

## Regras cobertas

- `BOT-01`: o bot trabalha sobre a cópia recebida e não altera o estado.
- `BOT-02`: o bot responde bem dentro do `prazo_planejamento_ms` (o limite em si é aplicado por quem chama, em outro marco).
- `BOT-04`, `VAL-01` a `VAL-06`, `ACA-01`, `ACA-02`, `BOM-02`, `MOV-02`: o bot nunca gera infração.
- `MOV-03`, `MOV-05`: o bot não planeja entrar em bloco destrutível (não aposta), então nunca tem movimento bloqueado.
- `BOM-04` a `BOM-12`, `ORD-01` a `ORD-07`: usadas na previsão das explosões.
- `EST-03`, `EST-04`, `EST-08`, `FIM-01`: limites do plano e jogador morto.

## Critérios de aceitação

Cada critério é verificável por um teste automatizado. Salvo menção contrária, "simular" é a simulação do turno definida acima, e os tabuleiros são pequenos, montados à mão.

### Contrato

- **CA-01** (BOT-01): **Dado** o bot criado com qualquer semente, **quando** `Versao()` é chamada, **então** devolve `"aleatorio-v1"`.
- **CA-02** (BOT-01): **Dado** um estado qualquer, **quando** `Planejar` é chamado, **então** o estado recebido continua igual (comparado com uma cópia feita antes da chamada), inclusive listas e jogadores.
- **CA-03** (ACA-02, VAL-06, EST-03): **Dado** um jogador vivo com `acoes_por_turno` n, **quando** `Planejar` é chamado, **então** o plano tem exatamente n ações, com `etapa` 1, 2, …, n, e `direcao` preenchida só em `MOVER`.
- **CA-04** (EST-08, FIM-01): **Dado** um estado vazio (`Estado{}`), ou um `jogadorID` inexistente, ou um jogador `MORTO`, **quando** `Planejar` é chamado, **então** não há `panic` e o plano é vazio.
- **CA-05** (BOT-02): **Dado** o mapa de exemplo, **quando** `Planejar` é chamado para cada jogador, **então** cada chamada termina em menos de 10% do `prazo_planejamento_ms` do mapa.

### Validade

- **CA-06** (BOT-04, VAL-01 a VAL-06, MOV-02, BOM-02): **Dado** o mapa de exemplo e as sementes 1 a 50, **quando** uma partida só com bots `aleatorio-v1` é jogada por até `limite_turnos` turnos (chamando `Planejar`, `Validar` e `ResolverTurno`), **então** `Validar` nunca devolve infração.
- **CA-07** (MOV-03, MOV-05): **Na** mesma partida do CA-06, **então** nenhum relatório de etapa tem ação `BLOQUEADA` ou `ABORTADA` de um bot `aleatorio-v1`.
- **CA-08** (MOV-03): **Dado** um jogador cercado por blocos destrutíveis e fixos, sem bombas, **quando** `Planejar` é chamado com várias sementes, **então** nenhum plano tem `MOVER` para uma casa com bloco destrutível.

### Aleatoriedade e determinismo

- **CA-09**: **Dados** o mesmo estado, o mesmo `jogadorID` e a mesma semente, **quando** `Planejar` é chamado duas vezes no mesmo bot ou em dois bots distintos, **então** os planos são idênticos. (O plano depende só da semente, do estado e do `jogadorID`, não de chamadas anteriores.)
- **CA-10**: **Dado** um jogador em campo aberto, sem bombas, **quando** `Planejar` é chamado com as sementes 1 a 20, **então** surgem pelo menos 2 planos diferentes.
- **CA-11**: **Dados** dois jogadores em posições simétricas em campo aberto, **quando** `Planejar` é chamado para cada um com a mesma semente, **então**, entre as sementes 1 a 20, há pelo menos uma em que os dois planos não são espelhados; o mesmo vale para o mesmo jogador em dois turnos diferentes (planos diferentes em pelo menos uma semente).
- **CA-12** (BOM-01, BOM-02): **Na** partida do CA-06, **então** em pelo menos uma das sementes algum plano tem `PLANTAR`.

### Fuga

Em todos os critérios desta seção vale para as sementes 1 a 50.

- **CA-13a** (BOM-05, BOM-06): **Dado** o jogador sobre uma bomba de um turno anterior com `pavio_restante` 2 e potência 1, com uma única casa vizinha livre, **quando** o plano é simulado, **então** ele sobrevive (sai pela casa livre na etapa 1 e na etapa 2 já está fora do alcance).
- **CA-13b** (ORD-07): **Dado** o jogador a 1 casa de uma bomba com `pavio_restante` 1 e potência 1, com a única casa livre do lado oposto à bomba, **quando** o plano é simulado, **então** ele sobrevive (move-se na própria etapa da explosão; o movimento é resolvido antes).
- **CA-14** (BOM-04): **Dada** uma pilha de duas bombas de potência 1 (alcance 2) com pavio 2, e o jogador a 2 casas dela em linha reta, **quando** o plano é simulado, **então** ele sobrevive.
- **CA-15** (BOM-09): **Dada** uma bomba A com pavio 1 que alcança a bomba B (pavio 6), e o jogador fora do alcance de A mas dentro do alcance de B, **quando** o plano é simulado, **então** ele sobrevive (o bot prevê a reação em cadeia na etapa 1).
- **CA-16** (BOM-07): **Dado** o jogador em uma casa protegida por um bloco destrutível entre ele e uma bomba com pavio 3, e o resto do tabuleiro fechado por blocos fixos, **quando** o plano é simulado, **então** ele sobrevive (a previsão respeita o bloco que para o fogo).
- **CA-17** (BOM-11): **Dado** um estado com uma bomba que sobra para o turno seguinte (pavio maior que as etapas do turno) e uma casa fora do seu alcance acessível ao jogador, **quando** o plano é simulado, **então** o bot termina o turno fora do alcance dela (plano seguro).
- **CA-18**: **Dado** um tabuleiro em que existe plano seguro só com `MOVER` e `ESPERAR`, **quando** `Planejar` é chamado, **então** o plano devolvido é seguro (a fuga é garantida, não depende da sorte da semente).
- **CA-19**: **Dado** um tabuleiro sem plano seguro mas com plano sobrevivente, **quando** `Planejar` é chamado, **então** o plano devolvido é sobrevivente.
- **CA-20**: **Dado** um jogador sem nenhuma saída (sobre uma bomba com `pavio_restante` 1, cercado por blocos fixos), **quando** `Planejar` é chamado, **então** não há `panic` e o plano é válido (sem infração).

### Bombas do próprio bot

- **CA-21** (BOM-01, BOM-05, BOM-11): **Dado** qualquer estado do CA-06 em que o plano tem `PLANTAR`, **quando** o plano é simulado, **então** o plano é seguro: o bot sobrevive às próprias bombas, inclusive às que sobram para o turno seguinte.
- **CA-22**: **Dado** um jogador cuja única casa livre é a atual (cercado por blocos fixos), **quando** `Planejar` é chamado com as sementes 1 a 50, **então** nenhum plano tem `PLANTAR`.

## Decisões

Questões levantadas no rascunho e aprovadas pelo usuário em 2026-10-01, com as propostas como estavam. São decisões locais deste bot; nenhuma regra de `docs/` muda.

1. **De onde vem a semente?** A interface `Planejar(estado, jogadorID)` não recebe semente. A semente é passada ao criar o bot (construtor `aleatorio.Novo(semente)`), e a cada chamada o gerador é derivado de `semente + turno + jogadorID`. Assim o plano só depende de (semente, estado, jogador) (CA-09), dois bots aleatórios na mesma partida podem usar a mesma semente sem jogar igual, e um replay é reproduzível.
2. **O que o bot sabe dos adversários?** Nada além do estado. Na simulação os outros ficam parados e não plantam. Bombas plantadas por eles neste turno podem matá-lo; isso é aceitável para um bot simples.
3. **Definição de "seguro" para o fim do turno.** Fora do alcance de **todas** as bombas que sobram, independente do pavio (conservador, mais fácil de testar).
4. **Garantia de fuga.** Se existe um plano seguro só com movimentos e esperas, o bot o encontra sempre (CA-18). Os planos com bombas podem ser buscados por sorteio, sem garantia; na dúvida o bot não planta.
5. **Com que frequência plantar?** Tentar plantar em cerca de metade dos turnos, em etapa sorteada, desde que exista plano seguro com a bomba; no máximo 1 bomba por turno (simples e sempre dentro de `bombas_por_turno`). Testado só como "planta em algum momento" (CA-12), sem fixar a taxa.
6. **Andar "aleatoriamente" entre os planos seguros.** Entre os planos seguros, escolher por sorteio (sem preferir ir para perto de blocos ou adversários). Ficar parado é aceitável quando sorteado.
7. **Fim antecipado na simulação (DEC-06).** Se a simulação termina antes da última etapa (sobra no máximo um vivo) com o bot vivo, conta como plano seguro, pois a partida acaba ali.
8. **Catálogo de bots** (`docs/BOTS.md`, "Onde colocar"). Fica para o marco 4, que precisa escolher bots por nome no terminal; aqui só se entrega o pacote do bot.
9. **Limite de tempo do CA-05.** 10% de `prazo_planejamento_ms` (100 ms no mapa de exemplo), folgado o bastante para não ser um teste instável.

Correção durante a implementação, aprovada pelo usuário em 2026-10-01: o CA-13 original (jogador sobre bomba com pavio 1 e potência 1) não tinha plano sobrevivente, porque a casa vizinha ainda está no alcance (BOM-06). Foi dividido em CA-13a (pavio 2) e CA-13b (ORD-07, jogador a 1 casa).
