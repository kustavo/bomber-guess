# Marco 04: partida no terminal

**Status**: concluída
**Roadmap**: `docs/ROADMAP.md`, marco 4
**Documentos de referência**: `docs/ARQUITETURA.md` (seção 1); `docs/BOTS.md` só para `BOT-01`, `BOT-02` e o catálogo de bots, que o marco 3 deixou para este marco (`specs/03-bot-simples/spec.md`, decisão 8)

## Objetivo

Entregar o programa `backend/cmd/terminal`. Ele roda uma partida inteira entre bots, do estado inicial até o desfecho, e imprime o tabuleiro em ASCII ao fim de cada etapa. É a primeira vez que o jogo roda de ponta a ponta: mapa → bots → validador → resolução → fim. Também cria o catálogo de bots e a chamada protegida aos bots (cópia do estado, prazo e `panic`), que o servidor do marco 5 vai reaproveitar.

### Termos usados nesta spec

- **Rodar o terminal**: executar o programa com uma lista de argumentos, capturando a saída padrão, a saída de erro e o código de saída. Os testes fazem isso sem compilar um binário.
- **Bot de teste**: implementação de `Bot` que só existe nos testes (dorme além do prazo, entra em `panic`, altera o estado recebido, gera ações inválidas, planta sempre etc.) e é registrada num catálogo de teste.

## Escopo

**Dentro**
- **Catálogo de bots**: lista das versões disponíveis e criação de um bot a partir da versão e de uma semente. Começa só com `aleatorio-v1`.
- **Chamada protegida ao bot** (BOT-01, BOT-02): o bot recebe uma cópia do estado; se estourar `prazo_planejamento_ms` ou entrar em `panic`, todas as suas ações do turno viram `ESPERAR`.
- **Laço da partida, síncrono e sem relógio**: a cada turno, chamar os bots dos jogadores vivos, validar os planos (VAL-01 a VAL-08), resolver o turno e repetir até `VerificarFim` indicar o fim.
- **Desenho em ASCII** do tabuleiro ao fim de cada etapa, a partir do `RelatorioEtapa`, com os eventos da etapa (mortes, blocos destruídos, movimentos bloqueados).
- **Avisos por turno**: infrações (regra e motivo), prazo estourado e `panic`.
- **Linha final com o desfecho**: vitória, empate por morte simultânea ou empate por limite de turnos.
- **Argumentos de linha de comando**: mapa, versões dos bots, semente, atraso entre etapas e listar o catálogo.
- **Erros de uso**: mapa inexistente ou inválido, versão desconhecida, quantidade de bots errada.

**Fora** (fica para outro marco)
- Relógio de turnos com fases de planejamento e execução, fila, API HTTP, `GET /bots`: marco 5.
- Gravar as três versões das ações (ARQUITETURA 1.3) e a saída bruta dos bots (BOT-03): marco 5. Aqui as infrações só são impressas.
- Cores, animação, limpar a tela, interface interativa: o terminal só imprime texto em sequência.
- Pontuação de infrações e ranking: marco 9.
- Bots de outras IAs: marco 10. O catálogo só precisa permitir acrescentá-los com uma linha.

## Regras cobertas

- `BOT-01`: o bot recebe uma cópia do estado; alterá-la não afeta a partida.
- `BOT-02`: prazo de planejamento e `panic` viram `ESPERAR` em todas as ações do turno.
- `BOT-04`, `VAL-01` a `VAL-08`: os planos passam pelo validador antes da resolução; as infrações são mostradas.
- `RES-01` a `RES-04`, `ORD-01` a `ORD-07`: a resolução é a de `ResolverTurno`, sem regra nova.
- `EST-02`: `duracao_etapa_ms` como atraso padrão entre etapas.
- `EST-08`, `FIM-01`: jogador morto sai do tabuleiro desenhado.
- `FIM-02` a `FIM-04`, `DEC-06`, `DEC-07`: fim da partida, inclusive no meio do turno, e limite de turnos.
- `MAP-03`, `MAP-05`, `MAP-06`: estado inicial a partir do mapa e das versões; mapa inválido; quantidade de versões.

## Critérios de aceitação

Cada critério é verificável por um teste automatizado. Salvo menção contrária, o terminal roda com atraso 0 e os mapas de teste são pequenos.

### Catálogo

- **CA-01**: **Dado** o catálogo, **quando** se listam as versões, **então** a lista contém `aleatorio-v1`, está em ordem alfabética e não tem repetidos; criar um bot de cada versão listada devolve um bot cuja `Versao()` é essa versão.
- **CA-02**: **Dada** uma versão que não está no catálogo, **quando** se tenta criar o bot, **então** há erro que cita a versão.
- **CA-03**: **Dado** o argumento de listar o catálogo, **quando** o terminal roda, **então** imprime uma versão por linha, na ordem do CA-01, e sai com código 0 sem jogar.

### Chamada protegida ao bot

- **CA-04** (BOT-01): **Dado** um bot de teste que altera o estado recebido (remove bombas e blocos, mata adversários), **quando** a partida roda, **então** o resultado é idêntico ao da mesma partida com um bot que só espera.
- **CA-05** (BOT-02): **Dado** um bot de teste que dorme além de `prazo_planejamento_ms` em um turno, **quando** a partida roda, **então** naquele turno o jogador executa `ESPERAR` em todas as etapas, há um aviso de prazo estourado citando o jogador e o turno, a partida continua, e a espera pelo bot não passa muito do prazo.
- **CA-06** (BOT-02): **Dado** um bot de teste que entra em `panic` em um turno, **quando** a partida roda, **então** o programa não cai, o jogador executa `ESPERAR` naquele turno, há um aviso de `panic` citando o jogador e o turno, e a partida continua.

### Laço da partida

- **CA-07** (VAL-05, BOT-04): **Dado** um bot de teste cuja 2ª ação sai do tabuleiro, **quando** a partida roda, **então** há um aviso de infração com jogador, etapa, regra (`VAL-02`) e motivo, e o turno é resolvido com o plano validado (ações a partir da 2ª viram `ESPERAR`).
- **CA-08** (MAP-03, FIM-02 a FIM-04): **Dado** o mapa de exemplo com 4 bots `aleatorio-v1` e a semente 1, **quando** o terminal roda, **então** sai com código 0, a última linha é o desfecho, e nenhum turno impresso passa de `limite_turnos`.
- **CA-09** (DEC-06): **Dada** uma partida que termina no meio de um turno, **quando** o terminal roda, **então** o último tabuleiro impresso é o da etapa em que a partida terminou, e nenhuma etapa seguinte é impressa.
- **CA-10** (FIM-02): **Dado** um mapa pequeno em que um bot de teste mata o outro, **quando** a partida roda, **então** o desfecho é a vitória do sobrevivente, com turno e etapa do fim.
- **CA-11** (FIM-03): **Dados** dois bots de teste que morrem na mesma etapa, **quando** a partida roda, **então** o desfecho é empate por morte simultânea.
- **CA-12** (FIM-04, DEC-07): **Dado** um mapa com `limite_turnos` 2 e bots que só esperam, **quando** a partida roda, **então** são impressos exatamente os turnos 1 e 2, e o desfecho é empate por limite de turnos entre todos os jogadores.
- **CA-13**: **Dados** os mesmos argumentos (mapa, bots, semente), **quando** o terminal roda duas vezes, **então** a saída é idêntica; com outra semente, no mapa de exemplo, a saída é diferente.

### Desenho

- **CA-14** (EST-08, FIM-01): **Dado** um `Estado` do início do turno e um `RelatorioEtapa` montados à mão, **quando** a etapa é desenhada, **então** o texto é exatamente o esperado: cabeçalho com turno e etapa, uma linha por `y` com um caractere por casa segundo a legenda (decisão 3) e a prioridade definida, sem jogadores mortos, e depois os eventos da etapa.
- **CA-15** (ORD-06): **Dado** um bloco destrutível destruído na etapa 2, **quando** as etapas 2 e 3 são desenhadas, **então** a casa aparece como chama na etapa 2 e como livre na etapa 3.
- **CA-16**: **Dada** qualquer partida do CA-08, **quando** a saída é lida, **então** cada tabuleiro tem `altura` linhas de `largura` caracteres, e a quantidade de tabuleiros impressos é igual à quantidade de relatórios de etapa da partida (mais o do estado inicial, decisão 4).

### Argumentos e erros

- **CA-17** (MAP-05): **Dado** um caminho de mapa inexistente, ou um mapa inválido, **quando** o terminal roda, **então** sai com código diferente de 0, escreve o motivo na saída de erro e não escreve nada na saída padrão.
- **CA-18** (MAP-06): **Dada** uma versão de bot desconhecida, ou uma quantidade de versões diferente de 1 e da quantidade de posições iniciais, **quando** o terminal roda, **então** o comportamento é o do CA-17.
- **CA-19** (MAP-03): **Dada** uma única versão, **quando** o terminal roda, **então** ela joga em todas as posições iniciais; **dadas** N versões para N posições, a versão `i` joga na posição `i` (`jogador_<i+1>`).
- **CA-20** (EST-02): **Dado** um atraso de X ms (por argumento ou, sem argumento, `duracao_etapa_ms` do mapa), **quando** uma partida curta roda, **então** o tempo total é pelo menos (etapas impressas − 1) × X.

## Decisões

Questões levantadas no rascunho e aprovadas pelo usuário em 2026-10-01, com as propostas como estavam. Nenhuma regra de `docs/` muda.

1. **Argumentos.**
   - `-mapa <caminho>`, padrão `mapas/exemplo.json`, procurado também em `../mapas/exemplo.json` para funcionar com `go run ./cmd/terminal` dentro de `backend/`;
   - `-bots v1,v2,...`, padrão `aleatorio-v1`; uma versão sozinha joga em todas as posições;
   - `-semente N`, padrão 1;
   - `-atraso <duração>` (ex.: `200ms`, `0`), padrão `duracao_etapa_ms` do mapa;
   - `-listar`.
2. **Semente de cada bot.** Todos os bots são criados com a mesma `-semente`. No `aleatorio-v1` isso já dá jogos diferentes por jogador (marco 3, decisão 1).
3. **Legenda do desenho.** `.` casa livre, `#` bloco fixo, `+` bloco destrutível, `o` bomba (uma ou pilha), `*` chama, `1`–`9` jogador vivo (posição dele em `jogadores` + 1, que é o N de `jogador_N` no estado inicial). Quando há mais de uma coisa na casa, vale a prioridade jogador > chama > bomba > bloco. Dois jogadores na mesma casa aparecem como `&`. Sem borda e sem coordenadas. O terminal recusa partidas com mais de 9 jogadores (erro de uso).
4. **O que é impresso.** Em ordem:
   - o tabuleiro do estado inicial (cabeçalho `Turno 1, início`);
   - em cada turno, os avisos (infrações, prazo, `panic`), um por linha;
   - em cada etapa, o cabeçalho `Turno T, etapa E`, o tabuleiro e os eventos não vazios (`mortes: ...`, `blocos destruídos: (x,y) ...`, `bloqueados: ...`);
   - no fim, a linha `Fim: ...` (vitória de X no turno T, etapa E / empate: todos morreram no turno T, etapa E / empate por limite de turnos entre X, Y).

   Uma linha em branco separa as etapas.
5. **Bot preso em laço infinito.** Em Go não dá para interromper uma goroutine. Ao estourar o prazo, o terminal segue sem esperar o bot, e a goroutine fica perdida até o processo acabar. Aceitável para o terminal; o servidor do marco 5 vai precisar rever isso.
6. **Saída de erro e código de saída.** código 0 para partida terminada ou listagem, 2 para erro de uso (argumento, mapa, versão). Mensagens de erro só na saída de erro.
7. **Onde ficam o catálogo e a chamada protegida.** o catálogo em `backend/internal/bots` (pacote que importa os subpacotes de cada bot) e a chamada protegida num pacote reutilizável fora de `cmd/`, para o marco 5 usar. O local exato fica para o plano.
