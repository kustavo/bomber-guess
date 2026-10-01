# Marco 02: validador e resolução do turno

**Status**: concluída
**Roadmap**: `docs/ROADMAP.md`, marco 2
**Documentos de referência**: `docs/REGRAS.md`, `docs/ARQUITETURA.md` (seção 1); `docs/EDITOR.md` só para `MAP-05`

## Objetivo

Implementar as regras do jogo no pacote puro `internal/jogo`: o validador (`Validar`), que transforma a saída bruta de um bot em um plano executável e registra as infrações, e a resolução do turno (`ResolverTurno`), que executa os planos de todos os jogadores etapa por etapa e devolve o relatório de cada etapa e o estado do turno seguinte. É o coração do jogo: tudo o que vem depois (bots, terminal, servidor, ranking) só chama estas duas funções.

## Escopo

**Dentro**
- `Validar(estado, plano)`: plano validado e lista de infrações (VAL-01 a VAL-08).
- O tipo `Infracao`: jogador, etapa da ação, regra violada e motivo.
- `ResolverTurno(estado, planos)`: resolução etapa a etapa na ordem ORD-01 a ORD-06, com movimentos, bombas, pavio, explosões, pilhas, reações em cadeia, mortes e blocos.
- O tipo `RelatorioEtapa`: posições, ações executadas, bombas, chamas, mortes, blocos destruídos e movimentos bloqueados de cada etapa.
- O novo `Estado` do turno seguinte (turno, bombas, blocos, jogadores, `etapas_neste_turno`).
- Detecção do fim da partida (vitória, empate por morte simultânea, empate por limite de turnos), observável a partir do estado.
- Validação de mapa ampliada: `limite_turnos` < 1 ou atributos de `jogador_padrao` < 1 tornam o mapa inválido (MAP-05).

**Fora** (fica para outro marco)
- Chamar os bots, limite de tempo e `panic` de bot (BOT-02): marco 4/5.
- Gravação das três versões das ações (seção 1.3 de ARQUITETURA): aqui só se garante que o relatório contém o necessário para gravá-las; a gravação é do marco 5.
- Loop, fila, API, relógio: marco 5.
- Estatísticas e pontuação de infrações: marco 9.
- Power-ups (atributos que mudam durante a partida): versão 2.

## Regras cobertas

- `ACA-01` a `ACA-03`: tipos de ação, limite de ações, preenchimento com `ESPERAR`.
- `VAL-01` a `VAL-08`: validador.
- `MOV-01` a `MOV-05`: movimento, inclusive movimento bloqueado.
- `BOM-01` a `BOM-12`: bombas, pilhas, explosões, reação em cadeia.
- `ORD-01` a `ORD-07`: ordem de resolução da etapa.
- `FIM-01` a `FIM-04`: morte, vitória e empate.
- `EST-03`, `EST-04`, `EST-07`, `EST-08`: campos do estado que a resolução atualiza.
- `RES-01` a `RES-04`: jogador sem plano, robustez, independência da ordem, fim e relatório.
- `DEC-01` a `DEC-09`: decisões de `docs/REGRAS.md` seção 8.
- `MAP-05`: `limite_turnos` e atributos mínimos do mapa.

## Critérios de aceitação

Cada critério é verificável por um teste de tabela. Os estados dos testes são tabuleiros pequenos montados à mão; "pavio p" e "potência k" referem-se aos atributos do jogador que planta.

### Validador

- **CA-01** (VAL-01, ACA-02, VAL-05): **Dado** um jogador com `acoes_por_turno` 3, **quando** o plano tem 5 ações válidas, **então** o plano validado tem as 3 primeiras ações e há exatamente uma infração, na etapa 4, citando `VAL-01`.
- **CA-02** (VAL-01, ACA-01, VAL-05): **Dado** um plano cuja 2ª ação é `MOVER` sem direção, ou com direção desconhecida (`"NORTE"`), ou cuja ação tem tipo desconhecido (`"PULAR"`), **quando** validado, **então** a 1ª ação é mantida, a 2ª e todas as seguintes viram `ESPERAR`, e há uma única infração, na etapa 2, citando `VAL-01`.
- **CA-03** (VAL-02, MOV-02, VAL-05): **Dado** um jogador em `{0,0}`, **quando** o plano é `DIREITA, CIMA, BAIXO`, **então** a ação da etapa 2 (sairia do tabuleiro) e a da etapa 3 viram `ESPERAR` e há uma infração na etapa 2 citando `VAL-02`.
- **CA-04** (VAL-02, MOV-02): **Dado** um jogador ao lado de um bloco fixo, **quando** a 1ª ação move para o bloco, **então** todo o plano vira `ESPERAR` e há uma infração na etapa 1 citando `VAL-02`.
- **CA-05** (VAL-02): **Dado** um jogador em `{0,0}`, **quando** o plano é `DIREITA, DIREITA, ESQUERDA, ESQUERDA, ESQUERDA`, **então** a simulação ação a ação aceita as quatro primeiras e marca infração só na etapa 5 (a posição simulada é `{0,0}` e `ESQUERDA` sairia do tabuleiro).
- **CA-06** (VAL-04): **Dado** um jogador ao lado de um bloco destrutível, **quando** o plano move para o bloco e continua depois dele, **então** o plano validado é igual ao enviado e não há infrações.
- **CA-07** (VAL-03, BOM-02, VAL-05): **Dado** um jogador com `bombas_por_turno` 2, **quando** o plano tem `PLANTAR, PLANTAR, ESPERAR, PLANTAR, MOVER`, **então** as etapas 4 e 5 viram `ESPERAR` e há uma infração na etapa 4 citando `VAL-03`.
- **CA-08** (VAL-03, EST-04): **Dado** um estado com bombas do jogador ainda no tabuleiro, vindas de turnos anteriores, **quando** ele planta `bombas_por_turno` bombas neste turno, **então** não há infração (o estoque é por turno).
- **CA-09** (VAL-05): **Dado** um plano com duas ações inválidas (etapas 2 e 4), **quando** validado, **então** há uma única infração (a da etapa 2).
- **CA-10** (VAL-01, ACA-03): **Dado** um plano vazio (ou sem ações), **quando** validado, **então** o plano validado não tem ações e não há infrações.
- **CA-11** (VAL-06): **Dado** um plano cujas ações têm `etapa` fora da sequência 1, 2, 3… (repetida, pulada ou fora de ordem), **quando** validado, **então** a primeira ação cuja `etapa` difere da sua posição na lista e todas as seguintes viram `ESPERAR`, com infração citando `VAL-06`.
- **CA-12** (VAL-07): **Dado** um plano com `turno` diferente do turno do estado, **quando** validado, **então** todas as ações viram `ESPERAR` e há uma infração na etapa 1 citando `VAL-07`.
- **CA-13** (VAL-08): **Dado** um plano de um jogador morto ou inexistente no estado, **quando** validado, **então** o plano validado não tem ações; para o jogador morto não há infração, para o inexistente há uma infração citando `VAL-08`.
- **CA-14** (DEC-08): **Dada** uma ação `PLANTAR` ou `ESPERAR` com `direcao` preenchida, **quando** validada, **então** a direção é removida no plano validado e não há infração.
- **CA-15**: **Dado** qualquer plano, **quando** validado, **então** o plano validado tem o mesmo `jogador_id` e `turno`, as ações numeradas 1, 2, 3… e no máximo `acoes_por_turno` ações; o estado e o plano recebidos não são alterados.

### Movimento

- **CA-16** (MOV-01, MOV-04): **Dados** dois jogadores, **quando** um entra na casa do outro, ou os dois trocam de casa na mesma etapa, ou um entra em uma casa com bomba, **então** todos os movimentos acontecem.
- **CA-17** (MOV-05, ORD-01): **Dado** um jogador ao lado de um bloco destrutível que ninguém destrói, **quando** o plano é `MOVER` para o bloco, depois `PLANTAR` e `MOVER`, **então** ele fica parado na etapa 1, as etapas 2 e 3 viram `ESPERAR` (nenhuma bomba plantada), e o relatório da etapa 1 registra o movimento bloqueado.
- **CA-18** (MOV-03, ORD-06): **Dado** um bloco destrutível atingido por uma explosão na etapa 3, **quando** um jogador planeja entrar nele na etapa 4, **então** o movimento acontece.
- **CA-19** (MOV-03, ORD-01, ORD-06): **Dado** um bloco destrutível atingido por uma explosão na etapa 3, **quando** um jogador planeja entrar nele na etapa 3, **então** o movimento é bloqueado (o movimento é resolvido antes da explosão) e o resto do plano é abortado.
- **CA-20** (ACA-03, EST-03): **Dados** dois jogadores com `acoes_por_turno` 5 e 3, **quando** o turno é resolvido, **então** há 5 relatórios de etapa e o segundo jogador executa `ESPERAR` nas etapas 4 e 5; o mesmo vale para um plano com menos ações que o limite.

### Bombas e pavio

- **CA-21** (BOM-05, ORD-03, BOM-01): **Dado** um jogador com pavio 3 e potência 2, **quando** planta na etapa 2, **então** a bomba aparece no relatório da etapa 2 com pavio 3, na sua casa, com potência 2; tem pavio 2, 1 nas etapas 3, 4; e explode na etapa 5.
- **CA-22** (BOM-11, BOM-05): **Dado** um turno de 7 etapas e um jogador com pavio 3, **quando** planta na etapa 6, **então** o novo estado tem a bomba com `pavio_restante` 2, e no turno seguinte ela explode na etapa 2.
- **CA-23** (BOM-03, BOM-04): **Dadas** na mesma casa bombas de potência 2 e 3, **quando** a pilha explode, **então** o alcance é 4; **dadas** três bombas de potência 2, o alcance é 4.
- **CA-24** (BOM-04): **Dada** uma pilha com uma bomba de pavio 1 e outra de pavio 3 (inclusive de jogadores diferentes), **quando** a primeira chega a 0, **então** a pilha inteira explode como uma única explosão e nenhuma bomba sobra na casa.
- **CA-25** (BOM-03, VAL-03): **Dado** um jogador com `bombas_por_turno` 2, **quando** planta duas vezes na mesma casa, **então** as duas bombas formam uma pilha.

### Explosões

- **CA-26** (BOM-06, BOM-07): **Dada** uma bomba de potência 2 em campo aberto, **quando** explode, **então** as chamas são a casa da bomba mais 2 casas em cada direção (9 casas), sem passar das bordas do tabuleiro.
- **CA-27** (BOM-07): **Dada** uma bomba de potência 3 com um bloco fixo a 1 casa, um bloco destrutível a 2 casas e outro destrutível a 3 casas em direções diferentes, **quando** explode, **então** a casa do bloco fixo não é chama; o primeiro bloco destrutível de cada direção é chama, é destruído e o fogo para ali; um destrutível atrás de outro na mesma direção não é atingido.
- **CA-28** (BOM-07, BOM-09): **Dada** uma bomba A de potência 3 e uma bomba B a 1 casa dela, com potência 1, **quando** A explode, **então** B explode na mesma etapa com o próprio alcance (1) a partir da sua casa, e o fogo de A não continua além de B.
- **CA-29** (BOM-09): **Dada** uma cadeia A → B → C (cada uma alcança só a seguinte), **quando** A explode, **então** as três explodem na mesma etapa e são removidas.
- **CA-30** (BOM-09, BOM-05): **Dada** uma bomba plantada na etapa em que uma explosão atinge sua casa, **quando** a etapa é resolvida, **então** ela explode em cadeia na mesma etapa.
- **CA-31** (DEC-05): **Dadas** duas bombas com pavio 0 na mesma etapa, uma no caminho da outra, **quando** explodem, **então** o resultado (chamas, blocos, mortes) é o mesmo independentemente da ordem das bombas no estado: toda bomba presente no início da fase de explosões para o fogo das outras.
- **CA-32** (BOM-08, ORD-05): **Dados** dois jogadores na mesma linha dentro do alcance, **quando** a bomba explode, **então** os dois morrem; um jogador na casa da própria bomba também morre.
- **CA-33** (BOM-10): **Dada** uma explosão na etapa 3, **quando** um jogador entra em uma das casas atingidas na etapa 4, **então** ele sobrevive.
- **CA-34** (BOM-12): **Dada** uma bomba de um jogador já morto, **quando** o pavio chega a 0, **então** ela explode normalmente.
- **CA-35** (DEC-01, ORD-06): **Dado** um bloco destrutível atingido por duas explosões na mesma etapa, **então** ele aparece uma única vez nos blocos destruídos do relatório e o fogo das duas para nele.

### Ordem da etapa e mortes

- **CA-36** (ORD-07, ORD-01): **Dado** um jogador no alcance de uma bomba que explode na etapa 4, **quando** ele sai do alcance na etapa 4, **então** sobrevive; se sair na etapa 5, morre na etapa 4.
- **CA-37** (ORD-02, ORD-05): **Dado** um jogador que planta na etapa em que é atingido, **quando** a etapa é resolvida, **então** a bomba é colocada e o jogador morre.
- **CA-38** (FIM-01, EST-07, EST-08, DEC-02): **Dado** um jogador atingido na etapa 3 do turno 5 (inclusive por bomba de turno anterior), **quando** o turno é resolvido, **então** ele fica `MORTO`, com `morte` = turno 5, etapa 3, posição igual à casa onde morreu, e suas ações das etapas 4 em diante não são executadas.
- **CA-39** (FIM-01): **Dado** um jogador morto, **então** ele não é mais atingido por explosões (não aparece de novo nas mortes) e não planta nem se move.

### Fim da partida e novo estado

- **CA-40** (FIM-02, DEC-06): **Dados** dois jogadores, **quando** um morre na etapa 3 e o outro sobrevive, **então** a resolução para ao fim da etapa 3 (3 relatórios), o estado indica fim com vitória do sobrevivente, e as bombas restantes ficam como estão (DEC-04).
- **CA-41** (FIM-03, DEC-06): **Dados** os dois últimos jogadores vivos, **quando** ambos morrem na mesma etapa, **então** a resolução para naquela etapa e o estado indica empate sem vencedor.
- **CA-42** (FIM-04, DEC-07): **Dado** o turno igual a `limite_turnos` com 2 ou mais jogadores vivos ao fim, **quando** o turno é resolvido, **então** o estado indica empate entre os sobreviventes.
- **CA-43** (DEC-06): **Dado** um estado de partida já terminada, **quando** `ResolverTurno` é chamado, **então** devolve o mesmo estado e nenhum relatório.
- **CA-44** (EST-03, DEC-03, BOM-11): **Dado** um turno resolvido sem fim de partida, **então** o novo estado tem `turno` + 1, as bombas não explodidas com o pavio restante, os blocos destruídos removidos, os jogadores mortos com `status` e `morte` atualizados e `etapas_neste_turno` igual ao maior `acoes_por_turno` entre os vivos.

### Relatório, robustez e pureza

- **CA-45** (DEC-09, RES-04): **Dado** qualquer turno, **então** cada relatório de etapa traz o número da etapa, a posição e o status de cada jogador ao fim da etapa, a ação de fato executada por cada jogador (bloqueado/abortado/morto distinguíveis), as bombas no tabuleiro, as casas em chamas, as mortes, os blocos destruídos e os movimentos bloqueados.
- **CA-46** (RES-01): **Dado** um jogador vivo sem plano na lista, **quando** o turno é resolvido, **então** ele executa `ESPERAR` em todas as etapas.
- **CA-47** (RES-02, RES-01): **Dado** um plano não validado (com movimento para fora do tabuleiro ou bomba além do limite), **quando** passado direto para `ResolverTurno`, **então** a resolução não entra em pânico e a ação impossível é tratada como `ESPERAR`, sem abortar o resto; planos de jogador inexistente são ignorados e, havendo dois planos do mesmo jogador, vale o primeiro.
- **CA-48** (RES-03): **Dados** o mesmo estado e os mesmos planos, **quando** `ResolverTurno` é chamado duas vezes, ou com os planos em outra ordem, **então** os resultados são idênticos; o estado e os planos recebidos não são alterados.
- **CA-49** (MAP-05): **Dado** um mapa com `limite_turnos` < 1, ou com `bombas_por_turno`, `potencia`, `pavio_padrao` ou `acoes_por_turno` < 1 em `jogador_padrao`, **quando** convertido em estado inicial, **então** a conversão devolve erro descritivo.

## Decisões

Questões levantadas no rascunho e aprovadas pelo usuário em 2026-10-01, com as propostas como estavam. Registradas em `docs/`:

1. `DEC-01` a `DEC-04` confirmadas.
2. `etapa` das ações → `VAL-06`.
3. Plano com turno errado → `VAL-07`.
4. Plano de jogador morto ou inexistente → `VAL-08`.
5. Bombas que explodem na mesma etapa param o fogo umas das outras → `DEC-05`.
6. Fim no meio do turno, deduzido do estado → `DEC-06`.
7. Turno `limite_turnos` é jogado → `DEC-07`.
8. Direção em ação não-`MOVER` é removida sem infração → `DEC-08`.
9. Relatório traz as ações executadas → `DEC-09`.
10. `ResolverTurno` robusto e independente da ordem dos planos → `RES-01` a `RES-04` (`docs/ARQUITETURA.md` 1.2).
11. Atributos mínimos no mapa → `MAP-05` ampliada (CA-49).
12. Na fase de planejamento (2026-10-01): `limite_turnos` ≥ 1 também entra em `MAP-05` (CA-49); termos **Relatório de etapa** e **Desfecho** no glossário; exemplos JSON de `Infracao` e `RelatorioEtapa` em `docs/ARQUITETURA.md` 1.1 e 1.2.
