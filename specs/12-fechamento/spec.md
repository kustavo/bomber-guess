# Marco 12: fechamento do tabuleiro

**Status**: concluída (com o adendo da área mínima, 2026-10-01)
**Roadmap**: `docs/ROADMAP.md`, marco 12
**Documentos de referência**: `docs/REGRAS.md` (seções 2, 6 e 7), `docs/ARQUITETURA.md` (seção 1.2), `docs/EDITOR.md` (formato do mapa)

> Spec escrita depois da implementação (commit `1480eba`), para registrar a mudança de regra no processo. As regras foram decididas com o usuário em 2026-10-01 antes do código (ver "Decisões").

## Objetivo

Reduzir os empates. Com bots cautelosos, quase toda partida chegava a `limite_turnos` com vários sobreviventes (194 de 200 partidas do `aleatorio-v1` no mapa de exemplo). A partir de um turno configurado, o tabuleiro passa a fechar de fora para dentro, um anel por turno, e quem fica no anel que fecha morre. Assim os jogadores são empurrados para o centro até sobrar um.

## Escopo

**Dentro**
- Campo opcional `turno_fechamento` no `config` do estado e do mapa, com validação no mapa.
- Fechamento ao fim do turno dentro de `ResolverTurno`: casas viram bloco fixo, jogadores morrem, blocos destrutíveis viram fixos e bombas somem.
- Campo `blocos_fechados` no relatório de etapa.
- Função pública para prever as casas que fecham no turno (usada por bots).
- Desenho do fechamento no terminal (marco 4).
- O bot `aleatorio-v1` (marco 3) evita terminar o turno em casa que fecha.
- `mapas/exemplo.json` com `turno_fechamento` 30.
- **Adendo**: área mínima que nunca fecha (`area_minima`, 5×5 por padrão), com validação no mapa.

**Fora**
- Mostrar o fechamento no frontend: fica para o marco 6 (CA-17 de `specs/06-frontend/spec.md`).
- Velocidade de fechamento configurável (mais de um turno por anel) ou formatos diferentes de anel.
- Pontuação específica para morte por fechamento no ranking (marco 9).

## Regras cobertas

- `EST-09`: `turno_fechamento` no `config`; ausente ou `0` desliga.
- `FEC-01`: o fechamento só vale com `turno_fechamento` ≥ 1.
- `FEC-02`: definição de anel; a borda é o anel 0.
- `FEC-03`: ao fim do turno `T ≥ turno_fechamento`, o anel `T − turno_fechamento` vira bloco fixo.
- `FEC-04`: quem está numa casa que fecha morre, na última etapa executada do turno.
- `FEC-05`: bloco destrutível que fecha vira fixo e não conta como destruído.
- `FEC-06`: bombas em casa que fecha somem sem explodir.
- `FEC-07`: partida terminada no meio do turno não fecha.
- `ORD-08`: o fechamento vem depois da última etapa.
- `MAP-05`: `turno_fechamento` < 0, ou ≥ 1 e ≥ `limite_turnos`, torna o mapa inválido; `area_minima` negativa ou maior que o tabuleiro também.
- `EST-10`: `area_minima` no `config`; ausente ou zerada vale 5×5 (adendo).
- `FEC-08`: o anel só fecha se o retângulo que sobra tiver pelo menos a área mínima (adendo).
- Respeitadas sem mudança: `FIM-02`, `FIM-03`, `FIM-04`, `DEC-06`, `DEC-07`.

## Critérios de aceitação

Testes em `backend/internal/jogo/fechamento_test.go`, salvo indicação.

- **CA-01** (FEC-02): **Dado** um tabuleiro 5×4, **quando** se pede o anel de uma casa, **então** cantos e bordas dão 0 e casas internas dão 1.
- **CA-02** (FEC-01, EST-09): **Dado** um estado sem `turno_fechamento`, **quando** se calculam as casas que fecham, **então** a lista é vazia em qualquer turno.
- **CA-03** (FEC-03): **Dado** `turno_fechamento` 3, **quando** se calculam as casas que fecham, **então** no turno 2 nenhuma fecha; no turno 3, toda a borda; no turno 4, o anel 1 sem as casas que já são bloco fixo; depois que o tabuleiro fecha por completo, nenhuma.
- **CA-04** (FEC-03, FEC-04, ORD-08): **Dado** um turno de fechamento com 3 etapas e dois jogadores na borda, **quando** o turno é resolvido, **então** só o relatório da última etapa tem `blocos_fechados` e as mortes; os dois morrem com `morte` = turno e etapa 3; o novo estado tem os blocos fixos novos; o jogador de dentro vence.
- **CA-05** (FEC-04, FIM-03): **Dados** todos os jogadores vivos na borda que fecha, **quando** o turno é resolvido, **então** a partida termina em empate.
- **CA-06** (FEC-05, FEC-06): **Dados** um bloco destrutível e uma bomba na borda que fecha, **quando** o turno é resolvido, **então** o bloco vira fixo sem aparecer em `blocos_destruidos`, a bomba some sem explodir e não aparece nem no relatório nem no novo estado.
- **CA-07** (FEC-07, DEC-06): **Dada** uma partida que termina na etapa 1 de um turno de fechamento, **quando** o turno é resolvido, **então** nada fecha e o sobrevivente na borda continua vivo.
- **CA-08** (FEC-03): **Dado** um estado cujo slice `blocos_fixos` tem capacidade sobrando, **quando** o turno de fechamento é resolvido, **então** o estado recebido continua igual, inclusive na capacidade do slice.
- **CA-09** (MAP-05, FEC-01): **Dado** um mapa com `turno_fechamento` −1 ou igual a `limite_turnos`, **quando** é verificado, **então** é inválido e a mensagem cita `turno_fechamento` (`mapa_test.go`).
- **CA-10** (EST-09, DEC-09): **Dados** os exemplos JSON de `docs/REGRAS.md` (seção 2, com `turno_fechamento`) e `docs/ARQUITETURA.md` (1.2, com `blocos_fechados`), **quando** fazem ida e volta, **então** o JSON é o mesmo; em turnos sem fechamento, `blocos_fechados` sai como `[]` (`docs_test.go`, `resolver_test.go`).
- **CA-11** (FEC-03, FEC-05): **Dado** um relatório com `blocos_fechados`, **quando** o terminal desenha a etapa, **então** imprime `fechamento: N casas` e, a partir da etapa seguinte, desenha essas casas como bloco fixo (`backend/cmd/terminal/desenho_test.go`).
- **CA-12** (FEC-04): **Dado** o `aleatorio-v1` num canto de um tabuleiro aberto no turno de fechamento, **quando** planeja, **então** o plano é seguro: ele termina fora da borda, para todas as sementes de 1 a 50 (`backend/internal/bots/aleatorio/fuga_test.go`).
- **CA-13** (FEC-01, MAP-05): **Dado** `mapas/exemplo.json` com `turno_fechamento` 30, **quando** partidas só com `aleatorio-v1` são jogadas com as sementes de 1 a 50, **então** continuam valendo os critérios do marco 3 (sem infrações nem movimentos bloqueados, bombas planejadas com fuga segura) (`partida_test.go` do bot).

### Adendo: área mínima

Pedido pelo usuário em 2026-10-01, depois da entrega: o fechamento para antes de sobrar menos que uma área configurável, para que a partida termine por confronto e não por todos morrerem juntos no centro. Testes em `backend/internal/jogo/fechamento_test.go` e `mapa_test.go`.

- **CA-14** (EST-10, FEC-08): **Dado** um tabuleiro 15×13 sem `area_minima`, **quando** se pergunta se cada anel pode fechar, **então** os anéis 0 a 3 fecham e o anel 4 não (sobrariam 5×3, menos que 5×5).
- **CA-15** (FEC-08): **Dada** `area_minima` 1×1 no 15×13, **quando** se pergunta, **então** o anel 5 fecha (sobram 3×1) e o 6 não; largura e altura são comparadas sem girar (9×7 cabe em 9×7, 7×9 não).
- **CA-16** (FEC-03, FEC-08): **Dado** um tabuleiro 7×7 com área mínima padrão e fechamento a partir do turno 1, **quando** três turnos são resolvidos, **então** só a borda fecha, os jogadores do anel 1 continuam vivos e a partida segue.
- **CA-17** (MAP-05, FEC-08): **Dado** um mapa com `area_minima` de largura negativa ou maior que o tabuleiro, **quando** é verificado, **então** é inválido e a mensagem cita `area_minima`.
- **CA-18** (EST-10): **Dado** o exemplo JSON de `docs/REGRAS.md` (seção 2, com `area_minima`), **quando** faz ida e volta, **então** o JSON é o mesmo (`docs_test.go`).

## Decisões

Aprovadas pelo usuário em 2026-10-01, antes da implementação.

1. **Momento**: ao fim do turno, depois da última etapa. O bot recebe o estado seguinte com o anel já fechado e pode prever o próximo com `turno` e `config`.
2. **Bombas em casa que fecha**: somem sem explodir.
3. **Configuração**: `turno_fechamento` opcional; ausente ou `0` desliga. Mapas antigos continuam válidos.
4. **`limite_turnos`**: mantido como rede de segurança; o mapa é inválido se `turno_fechamento` ≥ `limite_turnos`.
5. **Velocidade**: um anel por turno, como pedido.
6. **Área mínima (adendo)**: configurável em `area_minima` (largura e altura); ausente vale 5×5. O fechamento para no primeiro anel que deixaria menos que isso. Quem sobra dentro dela joga até haver vencedor ou até `limite_turnos`.

## Resultado medido

200 partidas de `aleatorio-v1` no mapa de exemplo (sementes 1 a 200):

| `turno_fechamento` | Empates | Duração média |
|---|---|---|
| desligado | 194 | 49,8 turnos |
| 30 | 85 | 32,7 turnos |

Os empates que sobram vêm sobretudo de bots iguais que morrem juntos quando o centro fecha.

Com o adendo (área mínima 5×5, sobra o retângulo 7×5), as mesmas 200 partidas com 4 `aleatorio-v1` dão 69 empates.
