# Marco 05: loop em tempo real, API HTTP e fila em memória

**Status**: concluída
**Roadmap**: `docs/ROADMAP.md`, marco 5
**Documentos de referência**: `docs/ARQUITETURA.md` (seções 2 a 4); `docs/BOTS.md` só para `BOT-01` a `BOT-03`, já usados no marco 4

## Objetivo

Levar a partida para o servidor. Cada partida roda num loop próprio, em tempo real, com uma fase de planejamento (os bots enviam planos para uma fila) e uma fase de execução (as etapas são liberadas uma a uma). Uma API HTTP permite criar partidas, consultar o estado atual com o cronômetro sincronizado, rever o histórico e listar os bots. É a base que o frontend (marco 6) vai consultar e que o Kafka (marco 8) e o ranking (marco 9) vão estender.

### Termos usados nesta spec

- **Relógio de teste**: relógio controlado pelo teste, que só avança quando o teste manda. Os critérios de tempo usam esse relógio, e não o relógio real.
- **Resposta de estado**: o JSON devolvido por `GET /partidas/{nome}/estado` (decisão 5).
- **Registro do turno**: o que fica gravado de cada turno executado: estado do início, as três versões das ações de cada jogador, infrações, falhas de bot e relatórios de etapa (decisão 6).
- `t0`: instante em que a fase de planejamento do turno começa. `p` = `prazo_planejamento_ms`. `d` = `duracao_etapa_ms`. `n` = quantidade de relatórios de etapa do turno.

## Escopo

**Dentro**
- **Loop da partida** (`internal/partida`): fases de planejamento e execução, com tempo de verdade, e fim da partida.
  - No planejamento, os bots rodam em paralelo com a chamada protegida do marco 4 (cópia do estado, prazo, `panic`).
  - Na execução, entram `Validar` e `ResolverTurno`, e as etapas são liberadas a cada `duracao_etapa_ms`.
- **Registro em memória** de cada partida: os turnos com as três versões das ações e o desfecho.
- **Fila** (`internal/fila`): interface e implementação em memória. Tem os tópicos `planos-enviados`, `turno-resolvido` e `partida-finalizada`.
- **API HTTP** (`internal/api`):
  - `POST /partidas`, `GET /partidas/{nome}/estado`, `GET /partidas/{nome}/historico` e `GET /bots`;
  - `GET /partidas` (API-10, decisão 9);
  - horário do servidor em toda resposta e formato de erro único.
- **Servidor** (`cmd/servidor`): sobe a API numa porta, lê os mapas de um diretório e encerra os loops ao receber sinal de término.
- IDs de regra para as seções 2 a 4 de `docs/ARQUITETURA.md` (decisão 1).

**Fora** (fica para outro marco)
- `POST /mapas` (editor): marco 7.
- `POST /partidas/{nome}/turnos/{n}/plano` (humanos): versão 2.
- `GET /ranking` e o consumo de `partida-finalizada`: marco 9.
- Kafka: marco 8. A interface da fila é feita para ele, mas só a implementação em memória entra agora.
- Frontend: marco 6.
- Gravar partidas em disco ou banco: as partidas vivem enquanto o processo vive (decisão 7).
- Autenticação, limites de requisição, HTTPS.

## Regras cobertas

IDs de `PAR`, `FILA` e `API` acrescentados em `docs/ARQUITETURA.md` (decisão 1).

- `PAR-01` a `PAR-05`: loop por partida, fases, jogador sem plano, registro.
- `FILA-01`, `FILA-02`: interface da fila e tópicos.
- `API-01`, `API-02`, `API-04`, `API-05`, `API-07`, `API-09`, `API-10`: consulta periódica, horário do servidor, endpoints deste marco.
- `BOT-01` a `BOT-03`: cópia do estado, prazo e `panic`, saída bruta gravada.
- `VAL-01` a `VAL-08`, `RES-01` a `RES-04`: validação e resolução de cada turno.
- `FIM-02` a `FIM-04`, `DEC-06`, `DEC-07`, `DEC-09`: fim da partida e ações executadas.
- `EST-01`, `EST-02`: prazo de planejamento e duração da etapa.
- `MAP-03`, `MAP-05`, `MAP-06`: estado inicial a partir do mapa e das versões.

## Critérios de aceitação

Cada critério é verificável por um teste automatizado. Salvo menção contrária, os testes usam o relógio de teste, bots de teste registrados num catálogo de teste e mapas pequenos.

### Loop da partida

- **CA-01** (PAR-01, PAR-02, EST-01): **Dada** uma partida criada no instante `t0`, **quando** o estado é consultado antes de `t0 + p`, **então** a fase é `PLANEJAMENTO`, o turno é 1, não há etapas liberadas e o fim da fase é `t0 + p`.
- **CA-02** (PAR-03, EST-02): **Dado** o relógio em `t1 = t0 + p`, **quando** o estado é consultado em `t1 + (k−1)·d` para k = 1…n, **então** a fase é `EXECUCAO`, a etapa atual é k, os relatórios liberados são exatamente os das etapas 1…k, e o fim da fase é `t1 + n·d`.
- **CA-03** (PAR-03): **Dado** o relógio em `t1 + n·d` numa partida não terminada, **quando** o estado é consultado, **então** a fase é `PLANEJAMENTO` do turno seguinte, o estado é o novo estado devolvido por `ResolverTurno`, e o fim da fase é `t1 + n·d + p`.
- **CA-04** (PAR-04, BOT-02): **Dado** um bot que só devolve o plano depois do fim da fase de planejamento, **quando** o turno é executado, **então** o jogador executa `ESPERAR` em todas as etapas, e o registro do turno marca a falha `PRAZO_ESTOURADO`.
- **CA-05** (BOT-02): **Dado** um bot que entra em `panic`, **quando** o turno é executado, **então** o servidor não cai, o jogador executa `ESPERAR`, o registro marca a falha `PANICO` com o valor do `panic`, e a partida continua.
- **CA-06** (BOT-01): **Dado** um bot que altera o estado recebido, **quando** a partida roda, **então** o histórico é idêntico ao da mesma partida com um bot que só espera.
- **CA-07** (PAR-05, BOT-03, VAL-05, DEC-09): **Dado** um bot cuja 2ª ação sai do tabuleiro, **quando** o turno é executado, **então** o registro do turno traz, para esse jogador:
  - as ações planejadas, iguais à saída bruta do bot;
  - as ações validadas, com `ESPERAR` a partir da 2ª;
  - as ações executadas, com o resultado de cada etapa;
  - a infração `VAL-02`.
- **CA-08** (DEC-06, FIM-02): **Dada** uma partida que termina na etapa k de um turno, **quando** o relógio passa de `t1 + (k−1)·d`, **então** a fase continua `EXECUCAO` com as etapas 1…k liberadas até `t1 + k·d`. Depois disso a fase é `ENCERRADA`, com o desfecho (vencedor) e sem fim de fase, e nenhum turno novo começa.
- **CA-09** (FIM-04, DEC-07): **Dada** uma partida com `limite_turnos` 2 e bots que só esperam, **quando** o relógio passa do fim do turno 2, **então** a fase é `ENCERRADA`, com empate entre todos, e o histórico tem exatamente 2 turnos.
- **CA-10** (PAR-01): **Dadas** duas partidas criadas em instantes diferentes, **quando** o relógio avança, **então** cada uma segue as próprias fases e os próprios turnos, sem interferir na outra.
- **CA-11** (RES-03): **Dada** uma partida com 4 bots `aleatorio-v1` e semente 1 no mapa de exemplo, **quando** ela roda até o fim, **então** cada turno do histórico (estado inicial, planos validados e relatórios) é igual ao de uma simulação síncrona com os mesmos bots e semente (`Planejar` → `Validar` → `ResolverTurno`).

### Fila

- **CA-12** (FILA-01): **Dada** a fila em memória, **quando** mensagens são publicadas em tópicos diferentes, **então** quem consome um tópico recebe só as mensagens dele, na ordem de publicação. Este teste de contrato fica reutilizável pela implementação Kafka.
- **CA-13** (FILA-02, PAR-02): **Dado** um turno em planejamento, **quando** os bots respondem, **então** cada plano é publicado em `planos-enviados` com partida, turno e jogador, e o turno é resolvido com os planos consumidos dessa fila.
- **CA-14** (PAR-04): **Dado** um plano publicado em `planos-enviados` depois do fim do planejamento do turno, ou com turno diferente, **quando** o turno é executado, **então** esse plano não é usado.
- **CA-15** (RES-01): **Dados** dois planos do mesmo jogador para o mesmo turno, **quando** o turno é executado, **então** vale o primeiro publicado.
- **CA-16** (FILA-02): **Dada** uma partida que roda até o fim, **quando** os tópicos são lidos, **então** há uma mensagem em `turno-resolvido` por turno (com o registro do turno) e exatamente uma em `partida-finalizada` (com o nome e o desfecho).

### API

- **CA-17** (API-09, API-02): **Quando** se faz `GET /bots`, **então** a resposta é 200 com as versões do catálogo em ordem alfabética e o horário do servidor.
- **CA-18** (API-04, MAP-03): **Dado** um corpo válido (nome, mapa, bots e, se quiser, semente), **quando** se faz `POST /partidas`, **então** a resposta é 201 com a resposta de estado do turno 1 em `PLANEJAMENTO`, e o loop está rodando.
- **CA-19** (API-04, MAP-05, MAP-06): **Dado** um corpo inválido, **quando** se faz `POST /partidas`, **então** a resposta é a da decisão 4, com o motivo no campo `erro`, e nenhuma partida é criada. Os casos testados são:
  - JSON malformado;
  - nome vazio ou com caracteres fora de `[a-z0-9-]`;
  - mapa inexistente ou inválido;
  - versão de bot desconhecida;
  - quantidade de bots errada;
  - nome já usado (409).
- **CA-20** (API-05, API-02): **Dada** uma partida existente, **quando** se faz `GET /partidas/{nome}/estado`, **então** a resposta é 200 com os campos da decisão 5. Para um nome desconhecido, é 404 com `erro`.
- **CA-21** (API-05): **Durante** a fase de execução, **quando** o estado é consultado, **então** a resposta nunca traz relatórios de etapas ainda não liberadas (CA-02).
- **CA-22** (API-07, PAR-05): **Dada** uma partida em andamento ou encerrada, **quando** se faz `GET /partidas/{nome}/historico`, **então** a resposta traz:
  - o mapa, os bots e a semente;
  - só os turnos já totalmente executados, cada um com o registro do turno;
  - o desfecho, se a partida terminou.

  Para um nome desconhecido, é 404.
- **CA-23** (API-02): **Dada** qualquer resposta da API, inclusive as de erro, **então** ela tem `horario_servidor`. As respostas de partida também têm `fase` e `fim_da_fase`.
- **CA-24** (API-01): **Dados** os endpoints fora deste marco (`POST /mapas`, `POST /partidas/{nome}/turnos/{n}/plano`, `GET /ranking`), **quando** chamados, **então** a resposta é 501 com `erro`. **Dado** um método não suportado num endpoint deste marco, a resposta é 405.

### Servidor

- **CA-25**: **Dado** o servidor iniciado numa porta livre com um diretório de mapas, **quando** se faz `GET /bots` e `POST /partidas` com o mapa `exemplo`, **então** as respostas são 200 e 201. Ao receber o pedido de término, o servidor para de aceitar requisições e encerra os loops das partidas.

## Decisões

Questões levantadas no rascunho e aprovadas pelo usuário em 2026-10-01, com as propostas como estavam. As decisões 1 e 9 foram registradas em `docs/ARQUITETURA.md`, e a 7 virou o marco 11 em `docs/ROADMAP.md`.

1. **IDs de regra para `docs/ARQUITETURA.md`, seções 2 a 4.** Hoje essas seções não têm IDs. acrescentar, sem mudar o conteúdo:
   - **PAR-01** cada partida tem um loop próprio, em tempo real, e quem a consulta vê o estado atual.
   - **PAR-02** o planejamento dura `prazo_planejamento_ms`; os planos vão para a fila.
   - **PAR-03** ao fim do prazo, o servidor consome a fila, passa os planos por `Validar` e `ResolverTurno` e libera uma etapa a cada `duracao_etapa_ms`.
   - **PAR-04** jogador sem plano dentro do prazo executa `ESPERAR` em todas as etapas.
   - **PAR-05** tudo é registrado: as três versões das ações e o resultado de cada etapa.
   - **FILA-01** fila atrás de interface; primeiro em memória, depois Kafka, sem mudar quem usa.
   - **FILA-02** tópicos `planos-enviados`, `turno-resolvido` e `partida-finalizada`.
   - **API-01** sem WebSockets; o frontend consulta periodicamente.
   - **API-02** toda resposta inclui o horário do servidor; as de partida também a fase e o fim da fase.
   - **API-03** a **API-09**, um por endpoint, na ordem da lista do documento.
2. **Planejamento termina mais cedo quando todos enviaram?** O documento diz "ao fim do prazo". Decisão: manter o prazo fixo. O ritmo fica previsível para quem assiste, e o cronômetro do frontend fica simples. Com o mapa de exemplo, uma partida leva até 50 × (1 s + 7 × 1 s) ≈ 7 min.
3. **Quando cada etapa é liberada.**
   - a etapa k fica visível de `t1 + (k−1)·d` até `t1 + k·d`;
   - a fase de execução dura `n·d`;
   - o turno é resolvido inteiro no início da execução, e só a liberação é gradual.

   No fim da partida (DEC-06), a última etapa também fica visível por `d` antes de `ENCERRADA`.
4. **Erros da API.**
   - corpo `{"erro": "...", "horario_servidor": "..."}`;
   - 400 para corpo inválido, mapa inexistente ou inválido, bot desconhecido e quantidade errada;
   - 404 para partida desconhecida;
   - 409 para nome repetido;
   - 405 para método não suportado;
   - 501 para endpoints de outros marcos.
5. **Resposta de estado.**

   ```json
   {
     "nome": "final-1",
     "fase": "EXECUCAO",
     "turno": 5,
     "etapa": 3,
     "horario_servidor": "2026-10-01T12:00:03.250Z",
     "fim_da_fase": "2026-10-01T12:00:08.000Z",
     "estado": { ... },
     "etapas": [ { ... }, { ... }, { ... } ],
     "desfecho": { ... }
   }
   ```

   - `fase`: `PLANEJAMENTO`, `EXECUCAO` ou `ENCERRADA`.
   - `etapa`: 0 no planejamento.
   - `estado`: sempre o `Estado` do início do turno.
   - `etapas`: os `RelatorioEtapa` já liberados no turno. Em `ENCERRADA`, são todos os do último turno.
   - `fim_da_fase`: ausente em `ENCERRADA`.
   - `desfecho`: o de `VerificarFim`.
   - Horários em RFC 3339, em UTC, com milissegundos.
6. **Corpo de `POST /partidas` e formato do histórico.**
   - **Corpo**: `{"nome": "final-1", "mapa": "exemplo", "bots": ["aleatorio-v1", ...], "semente": 1}`.
     - `mapa` é o nome de um arquivo `<mapa>.json` do diretório de mapas.
     - `bots` aceita uma só versão para todas as posições, como no terminal.
     - `semente` é opcional, com padrão 1.
   - **Histórico**: `{"nome", "mapa", "bots", "semente", "turnos": [...], "desfecho"}`. Cada turno tem `turno`, `estado` (início do turno), `jogadores` e `etapas` (os relatórios).
     - Cada item de `jogadores` tem `jogador_id`, `planejadas`, `validadas`, `executadas`, `infracoes` e `falha`.
     - Cada item de `executadas` tem `etapa`, `acao` e `resultado`, tirados dos relatórios (DEC-09).
7. **Persistência.** O glossário diz que a partida "fica salva". Neste marco as partidas ficam em memória e somem quando o servidor reinicia. Gravar em disco vira um item novo no `docs/ROADMAP.md`, depois do ranking, porque o ranking também vai precisar de dados salvos.
8. **Bot que não termina.** O mesmo problema do marco 4, agora num servidor que fica de pé. Decisão: aceitar a goroutine perdida por enquanto e registrar o risco. A solução de verdade (bots em processo separado) só faz sentido com os bots de outras IAs, no marco 10.
9. **`GET /partidas`.** O frontend do marco 6 precisa saber quais partidas existem, e o documento não tem esse endpoint. Decisão: acrescentar como **API-10**, que lista nome, fase, turno e desfecho de cada partida, ordenadas por criação.
10. **CORS.** O frontend vai rodar em outra porta durante o desenvolvimento. Decisão: não abrir CORS no servidor. O marco 6 usa o proxy do servidor de desenvolvimento do Svelte/Vite.
11. **Servidor.** Flags `-porta` (padrão 8080) e `-mapas` (padrão `mapas/`, procurando também em `../mapas/`, como no terminal). Encerramento limpo ao receber SIGINT ou SIGTERM.
