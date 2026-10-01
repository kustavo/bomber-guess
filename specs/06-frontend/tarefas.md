# Marco 06: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: Scaffold de `frontend/`: `package.json`, Vite, Svelte 5, TypeScript, Vitest + jsdom + Testing Library, proxy para a API, `.gitignore`. Um teste trivial de fumaça passando em `npm test`, e `npm run check` limpo. (base para todos)
- [x] **T2**: `tipos.ts` e `testes/fabricas.ts` (construtores de `Estado`, `RelatorioEtapa`, `RespostaEstado` e API falsa). (base para CA-01 a CA-15)
- [x] **T3**: `tabuleiro.ts`: `calcularTabuleiro` com testes de tabela: planejamento, etapa k, blocos destruídos acumulados, chamas só da etapa k, jogador morto fora do tabuleiro, `bloqueado` a partir de `movimentos_bloqueados`. (CA-01, CA-02, CA-03, CA-04)
- [x] **T4**: `cronometro.ts` e `ritmo.ts`: defasagem, tempo restante nunca negativo, relógios com ±1 h, `null` em `ENCERRADA`, intervalos de consulta e entre etapas. (CA-05, CA-06, CA-07, CA-10)
- [x] **T5**: `api.ts`: `criarCliente` com `fetch` injetado; sucesso, 404 com `erro`, 5xx e falha de rede viram `ErroApi`. (CA-14, CA-15)
- [x] **T6**: `acompanhamento.ts`: consulta periódica com `Relogio` de teste, uma requisição por vez, fila de etapas (em dia e atrasada), troca de turno, parada em `ENCERRADA` e em 404, aviso e retomada em falha, primeira resposta já sincroniza. (CA-08, CA-09, CA-10, CA-11, CA-12, CA-14, CA-15)
- [x] **T7**: `Tabuleiro.svelte` e `PainelJogadores.svelte`: SVG com `data-*`, cores por jogador, pavio nas bombas, chamas, marca de bloqueio; painel com morte (turno e etapa). Testes de componente. (CA-01, CA-03, CA-04)
- [x] **T8**: `Cronometro.svelte`, `textos.ts` e `TelaPartida.svelte`: cabeçalho com turno, etapa e fase; cronômetro ou desfecho; mensagens de partida não encontrada e sem conexão. Testes de componente com cliente falso. (CA-07, CA-12, CA-14, CA-15)
- [x] **T9**: `rota.ts`, `ListaPartidas.svelte` e `App.svelte`: lista na ordem recebida com fase, turno e desfecho, link `#/partidas/<nome>`, abertura direta pelo endereço. Testes. (CA-13)
- [x] **T10**: Verificação manual de ponta a ponta: servidor Go + `npm run dev`, `POST /partidas` com o mapa `exemplo` e 4 `aleatorio-v1`; o tabuleiro aparece, as etapas animam, o turno avança, o cronômetro bate com o servidor e o desfecho aparece no fim. Registrar o resultado aqui. (CA-16)
  - **Resultado (2026-10-01)**: servidor (`go -C backend run ./cmd/servidor`) e `npm run dev` pelo `.claude/launch.json`; partida `teste-1` no mapa `exemplo` com 4 `aleatorio-v1`, semente 1.
    - O proxy do Vite chega à API (`/partidas`, `/partidas/{nome}/estado`).
    - A lista mostra a partida; o link abre a tela; o tabuleiro 15×13 aparece com blocos, jogadores, bombas e chamas, e as etapas avançam uma a uma.
    - Cronômetro comparado com `fim_da_fase` do servidor em 5 amostras: diferença de cerca de 20 ms (tique de 100 ms); a etapa da tela é a do servidor, com atraso de até uma consulta (250 ms) na troca de fase.
    - A partida terminou no turno 50 (`limite_turnos`); a tela mostrou "Encerrada" e "Empate entre jogador_1, jogador_2, jogador_3, jogador_4", sem cronômetro, e a lista também.
    - `#/partidas/nao-existe` mostra "Partida não encontrada."; com o servidor Go parado, a lista mostra o aviso de conexão e mantém a última lista.
- [x] **T11**: Verificação final: todo CA com teste (`grep -rn "CA-" frontend/src`); `npm run check` e `npm test` limpos; `gofmt -l`, `go vet ./...` e `go test ./...` continuam limpos; propor a linha de verificação do frontend no `AGENTS.md`; status da spec → `concluída`.
- [x] **T12** (marco 12): `tipos.ts` com `turno_fechamento?` e `blocos_fechados`; `fabricas.ts` com `blocos_fechados: []`; `calcularTabuleiro` acumula `blocos_fechados` até a etapa k (bloco fixo, sem destrutível); testes de tabela para etapa antes e depois do fechamento e para morte por fechamento no painel. (CA-17)
- [x] **T13**: Verificação final de novo: `npm run check` e `npm test` limpos; status da spec → `concluída`.
  - **Resultado (2026-10-01)**: `npm run check` sem erros e `npm test` com 79 testes passando. Verificação manual com uma cópia temporária do mapa `exemplo` (`turno_fechamento` 2, prazos curtos) e 4 `aleatorio-v1`, semente 1. Na etapa 7 do turno 2, a tela mostrou 94 blocos fixos (42 do mapa e 52 da borda). Os jogadores 1, 2 e 4 aparecem como "Morto no turno 2, etapa 7", e o desfecho é a vitória de jogador_3.
