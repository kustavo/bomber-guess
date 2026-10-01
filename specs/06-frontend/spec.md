# Marco 06: frontend Svelte para assistir às partidas

**Status**: concluída (reaberta e concluída em 2026-10-01 com o CA-17, por causa do marco 12)
**Roadmap**: `docs/ROADMAP.md`, marco 6
**Documentos de referência**: `docs/ARQUITETURA.md` (seção 4); formatos de resposta da API definidos em `specs/05-servidor/spec.md` (decisões 5, 6 e 9)

## Objetivo

Dar uma tela para assistir às partidas que o servidor (marco 5) roda em tempo real. O frontend lista as partidas, consulta periodicamente o estado de uma delas, desenha o tabuleiro, anima as etapas conforme são liberadas e mostra um cronômetro sincronizado com o relógio do servidor. Não envia nada ao jogo: só observa.

### Termos usados nesta spec

- **Resposta de estado**: o JSON de `GET /partidas/{nome}/estado` (marco 5, decisão 5): `nome`, `fase`, `turno`, `etapa`, `horario_servidor`, `fim_da_fase`, `estado`, `etapas`, `desfecho`.
- **Tabuleiro exibido**: o que a tela mostra num instante: blocos fixos, blocos destrutíveis, bombas, chamas e jogadores. É derivado do `estado` do início do turno e dos relatórios de etapa já liberados.
- **Defasagem**: diferença entre o relógio do servidor e o do navegador, estimada a cada resposta.
- **Relógio de teste** e **API falsa**: nos testes, o tempo e as respostas HTTP são controlados pelo teste; nenhum critério depende do relógio real nem de um servidor Go rodando (salvo CA-16).

## Escopo

**Dentro**
- App Svelte em `frontend/`, com servidor de desenvolvimento cujo proxy encaminha a API para o servidor Go (marco 5, decisão 10: sem CORS).
- **Lista de partidas** (`GET /partidas`, API-10): nome, fase, turno e desfecho; escolher uma abre a tela da partida.
- **Tela da partida** (`GET /partidas/{nome}/estado`, API-05), com consulta periódica (API-01):
  - tabuleiro desenhado a partir da resposta de estado;
  - etapas animadas na ordem em que são liberadas;
  - painel de jogadores: id, versão do bot, status e, se morto, turno e etapa da morte;
  - turno, etapa e fase atuais;
  - cronômetro até `fim_da_fase`, corrigido pela defasagem (API-02);
  - desfecho quando a partida termina.
- Tratamento de erros da API: partida inexistente (404), servidor fora do ar, resposta com `erro`.
- Testes automatizados do frontend (lógica pura e componentes).

**Fora** (fica para outro marco)
- Criar partidas pela tela (`POST /partidas`) e o editor de mapas: marco 7.
- Rever partidas encerradas etapa por etapa a partir de `GET /partidas/{nome}/historico`: fora (decisão 4).
- Ranking: marco 9.
- Jogadores humanos e envio de planos: versão 2.
- Servir o build de produção pelo servidor Go: fora (decisão 6).
- Mudanças na API do marco 5. Se faltar algo, a spec volta para revisão.

## Regras cobertas

- `API-01`: sem WebSockets; o frontend consulta periodicamente.
- `API-02`: cronômetro calculado a partir de `horario_servidor` e `fim_da_fase`; recarregar a página ressincroniza.
- `API-05`: consumo da resposta de estado.
- `API-10`: consumo da lista de partidas.
- `EST-02`: `duracao_etapa_ms` usado no ritmo da animação.
- `EST-06` a `EST-08`: status, morte e posição de jogador morto (que não ocupa mais o tabuleiro, `FIM-01`).
- `DEC-06`, `FIM-02` a `FIM-04`: exibição do desfecho (vitória ou empate).
- `DEC-09`: resultado de cada ação no relatório de etapa (`EXECUTADA`, `BLOQUEADA`...).
- `EST-09`, `FEC-03` a `FEC-05`: fechamento do tabuleiro (marco 12), exibido a partir de `blocos_fechados`.

## Critérios de aceitação

Cada critério é verificável por um teste automatizado (Vitest + Testing Library, decisão 1), com API falsa e relógio de teste, salvo menção contrária.

### Tabuleiro exibido (lógica pura)

- **CA-01** (API-05): **Dada** uma resposta em `PLANEJAMENTO` (etapa 0, sem relatórios), **quando** o tabuleiro exibido é calculado, **então** ele tem `largura` × `altura` casas, com os blocos fixos, os blocos destrutíveis, as bombas e os jogadores vivos exatamente nas posições do `estado`.
- **CA-02** (API-05, DEC-09): **Dada** uma resposta em `EXECUCAO` com os relatórios das etapas 1…k, **quando** o tabuleiro exibido é calculado, **então**:
  - posições e status dos jogadores e as bombas são os do relatório da etapa k;
  - as chamas são as `chamas` da etapa k (só da etapa k, não acumuladas);
  - os blocos destrutíveis são os do `estado` menos todos os `blocos_destruidos` das etapas 1…k.
- **CA-03** (EST-07, EST-08, FIM-01): **Dado** um jogador `MORTO`, **quando** o tabuleiro exibido é calculado, **então** ele não aparece como ocupante de casa; o painel de jogadores o mostra como morto, com turno e etapa da morte.
- **CA-04** (DEC-09): **Dado** um relatório de etapa com um jogador em `movimentos_bloqueados`, **quando** a etapa é exibida, **então** a tela indica o bloqueio desse jogador (marca visível com o resultado `BLOQUEADA`).

- **CA-17** (FEC-03, FEC-04, FEC-05): **Dada** uma resposta em `EXECUCAO` cuja etapa j tem `blocos_fechados`, **quando** o tabuleiro exibido é calculado para uma etapa k ≥ j, **então** essas casas aparecem como bloco fixo e não como bloco destrutível. Os jogadores mortos pelo fechamento saem do tabuleiro, e o painel mostra a morte na etapa j (como no CA-03). Para k < j, o tabuleiro ainda não mostra o fechamento.

### Cronômetro (lógica pura)

- **CA-05** (API-02): **Dada** uma resposta com `horario_servidor` = S, recebida quando o relógio do navegador marca L, **quando** a defasagem é calculada, **então** ela é S − L, e o tempo restante exibido no instante local L' é `fim_da_fase` − (L' + S − L), nunca negativo (mostra 0).
- **CA-06** (API-02): **Dados** os relógios do navegador e do servidor diferentes por ±1 h, **quando** o cronômetro é exibido, **então** o tempo restante é o mesmo que com relógios iguais.
- **CA-07** (API-02, DEC-06): **Dada** uma resposta em `ENCERRADA` (sem `fim_da_fase`), **quando** a tela é exibida, **então** não há cronômetro e o desfecho aparece no lugar.

### Consulta periódica e animação

- **CA-08** (API-01): **Dada** a tela de uma partida em andamento, **quando** o relógio de teste avança, **então** o estado é consultado a cada intervalo de consulta (decisão 2), sempre por `GET /partidas/{nome}/estado`, sem nenhuma conexão persistente.
- **CA-09** (API-01): **Dada** uma partida `ENCERRADA`, **quando** a resposta chega, **então** a consulta periódica para.
- **CA-10** (EST-02, API-05): **Dadas** duas respostas consecutivas em `EXECUCAO` em que a segunda libera as etapas k+1…m de uma vez (consulta atrasada), **quando** a tela as recebe, **então** cada etapa k+1…m é exibida em ordem, nenhuma é pulada, e o ritmo dessa exibição não é mais lento que `duracao_etapa_ms`.
- **CA-11** (API-05): **Dada** uma resposta do turno seguinte em `PLANEJAMENTO`, **quando** a tela a recebe, **então** o tabuleiro passa a ser o do novo `estado` (CA-01), sem resto de chamas ou marcas da etapa anterior.
- **CA-12** (API-02): **Dada** uma tela recém-aberta (ou recarregada) no meio de uma fase, **quando** a primeira resposta chega, **então** turno, etapa, fase e cronômetro já correspondem a ela, sem depender de respostas anteriores.

### Lista de partidas e erros

- **CA-13** (API-10): **Dada** uma resposta de `GET /partidas` com N partidas, **quando** a lista é exibida, **então** aparecem as N, na ordem recebida, com nome, fase, turno e, se encerrada, o desfecho (vencedor ou empate); escolher uma leva à tela dessa partida (endereço com o nome, que pode ser aberto diretamente).
- **CA-14** (API-05): **Dada** uma partida inexistente (404 com `erro`), **quando** a tela é aberta, **então** aparece a mensagem de partida não encontrada e a consulta periódica para.
- **CA-15** (API-01): **Dada** uma falha de rede ou resposta 5xx durante a consulta, **quando** ela ocorre, **então** a tela mantém o último tabuleiro, mostra um aviso de conexão e continua tentando; quando uma resposta válida chega, o aviso some.

### Integração

- **CA-16**: **Dado** o servidor Go do marco 5 rodando e o frontend no servidor de desenvolvimento com proxy, **quando** se cria uma partida por `POST /partidas` e se abre sua tela, **então** o tabuleiro aparece e o turno avança. Teste de ponta a ponta opcional (decisão 1); se não automatizado, vira verificação manual registrada no `tarefas.md`.

## Decisões

Questões levantadas no rascunho e aprovadas pelo usuário em 2026-10-01, com as propostas como estavam.

1. **Ferramentas.** Decisão: Svelte 5 + Vite (sem SvelteKit), TypeScript, Vitest + `@testing-library/svelte` com jsdom para CA-01 a CA-15. Roteamento simples por hash (`#/partidas/<nome>`), sem biblioteca. CA-16 como verificação manual, sem Playwright por enquanto (evita baixar navegadores).
2. **Intervalo de consulta.** Com `duracao_etapa_ms` de 1 s no mapa de exemplo, consultar a cada 1 s perderia o início das etapas. Decisão: consultar a cada `min(250 ms, duracao_etapa_ms / 2)`; na lista de partidas, a cada 2 s.
3. **Desenho e animação.** Decisão: tabuleiro em SVG (escala bem e é fácil de testar pelo DOM); movimento com transição curta (até metade de `duracao_etapa_ms`); chamas e bloqueios exibidos durante a etapa. Cada jogador com uma cor fixa pela ordem em `estado.jogadores`; bombas mostram o `pavio_restante`.
4. **Rever partidas encerradas.** `GET /partidas/{nome}/historico` permitiria um replay turno a turno. Decisão: fora deste marco ("só assistir" ao vivo); ao abrir uma encerrada, mostra-se o último turno e o desfecho (como já vem na resposta de estado). Replay entra junto com o marco 11 (partidas salvas).
5. **Etapas atrasadas (CA-10).** Decisão: se a consulta atrasar e várias etapas chegarem juntas, exibi-las em sequência acelerada (até alcançar a etapa atual do servidor), em vez de pular para a última.
6. **Build de produção.** Decisão: só o servidor de desenvolvimento com proxy neste marco; servir os arquivos estáticos pelo servidor Go fica para quando houver implantação.
7. **Idioma e textos da tela.** Decisão: tela em português, usando os termos do glossário (turno, etapa, planejamento, execução, encerrada, desfecho).
8. **Fechamento do tabuleiro (aprovada em 2026-10-01).** O marco 12 acrescentou `turno_fechamento` ao `config` e `blocos_fechados` ao relatório de etapa (`docs/ARQUITETURA.md`, 1.2). A spec foi reaberta para o CA-17. Decisão: o tabuleiro exibido trata `blocos_fechados` como `blocos_destruidos` ao contrário, acumulando até a etapa k e somando aos blocos fixos. Não há aviso de "fechamento no turno X" na tela neste marco.
