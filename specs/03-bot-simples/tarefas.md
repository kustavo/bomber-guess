# Marco 03: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: Esqueleto de `internal/bots/aleatorio`: `Versao`, `Novo`, `Bot.Versao` e um `Planejar` que devolve plano vazio para estado vazio, jogador inexistente ou morto, e `acoes_por_turno` ações `ESPERAR` numeradas nos demais casos. Helpers de teste `montar` e `comBomba` (D9). (CA-01, CA-03, CA-04)
- [x] **T2**: `preverLinha` com fantasmas (D2) e zona de perigo final (D3), com testes próprios: chamas da etapa certa, pilha, reação em cadeia, bloco que para o fogo, bomba que sobra para o turno seguinte. Helpers de teste `simular` e `classificar` pela definição da spec.
- [x] **T3**: `buscar` nos três modos (seguro, sobrevivente, mais longo), com ordem sorteada, sem entrar em bloco destrutível (D5) e parando depois de `acoes_por_turno` (D6). `Planejar` passa a usar a linha do tempo sem bomba e o gerador de D7. (CA-08, CA-13a, CA-13b, CA-14, CA-15, CA-16, CA-17, CA-18, CA-19, CA-20)
- [x] **T4**: Testes de aleatoriedade e determinismo: mesmo plano para a mesma entrada, variedade entre sementes, jogadores simétricos e turnos diferentes. (CA-09, CA-10, CA-11)
- [x] **T5**: Candidatos com bomba (D4, D8): sorteio de cerca de metade dos turnos, até 8 tentativas com etapa e prefixo sorteados, continuação exigindo plano seguro. Teste do jogador cercado que não planta. (CA-22)
- [x] **T6**: `jogarPartida(semente)` no mapa de exemplo com os 4 jogadores `aleatorio-v1` (D10): sem infração, sem movimento bloqueado ou abortado, alguma bomba plantada, todo plano com `PLANTAR` é seguro. (CA-06, CA-07, CA-12, CA-21)
- [x] **T7**: Pureza e tempo: o estado recebido não muda, inclusive nos estados com bombas e fantasmas; tempo médio por chamada no mapa de exemplo abaixo de 10% do prazo. (CA-02, CA-05)
- [x] **T8**: Verificação final: todo CA com teste (nome do caso começando pelo ID da regra ou do CA quando não há regra); `gofmt -l`, `go vet ./...` e `go test ./...` limpos; status da spec → `concluída`.
