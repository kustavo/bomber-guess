# Marco 15: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: `VersaoV3`, `NovoV3`, campo `variasBombas` e `Versao()`; registro no catálogo. Testes: versão; catálogo lista as três e cria cada uma. (CA-01, CA-03)
- [x] **T2**: `sortearEtapas` (D4). Testes de tabela: n etapas distintas, em ordem, dentro de 1..acoes; n = acoes devolve todas. (CA-07)
- [x] **T3**: `tentarPlantarVarias` (D4). Testes: plano com exatamente n `PLANTAR`, seguro pela simulação (vivo e fora da zona de perigo final, com pilha e bomba que atravessa o turno); sem plano possível devolve false. (CA-09)
- [x] **T4**: `planejarComBombas` e o despacho em `Planejar` (D1 a D3). Testes: nunca mais `PLANTAR` que `bombas_por_turno`, sem infração no `Validar` (bombas_por_turno 1 a 3); tabuleiro aberto com ao menos um plano de 2 bombas nas sementes 1 a 50; queda para 1 bomba quando 2 não cabem. (CA-07, CA-08, CA-11)
- [x] **T5**: v3 igual ao v2 com `bombas_por_turno` 1 e dentro da janela. (CA-04, CA-05)
- [x] **T6**: Testes herdados do v2 (e do v1) rodando também com o v3 (D6), incluindo tempo de planejamento no mapa de exemplo e a partida sem infrações só com v3. (CA-02, CA-06)
  - O v3 entrou na tabela `versoesDeTeste`; todos os testes que rodavam com o v1 e o v2 (incluindo `TestTempoDePlanejamento` e `TestPartidaNoMapaExemplo`, com os 4 jogadores em v3) passaram sem mudança, porque nenhum deles fixava "no máximo uma bomba". Os nomes desses casos citam os CAs do marco 3 e do marco 13, de onde vêm.
- [x] **T7**: Jogo no tabuleiro aberto 7 × 7 (D7): o v3 nunca morre nas sementes 1 a 50, e em pelo menos uma semente sobra bomba dele para o turno seguinte. (CA-10)
  - **Resultado**: nas sementes 1 a 50, o v3 nunca morreu; 134 turnos terminaram com bomba dele no tabuleiro.
- [x] **T8**: Comparação no mapa de exemplo, só registrada (D8). (CA-12)
  - **Resultado (sementes 1 a 200, mapa de exemplo)**:
    - 4 × v3: jogadores-turno com 0 bombas 17 634, com 1 bomba 9 864, com 2 bombas 4 005; 103 mortes por fechamento, 439 por bomba, 76 empates.
    - 2 × v2 contra 2 × v3: v2 venceu 65, v3 venceu 58, 77 empates. O v2 nunca plantou 2 bombas (0: 9 033, 1: 7 136); o v3 plantou 2 em 2 001 jogadores-turno (0: 8 933, 1: 4 964). 102 mortes por fechamento, 424 por bomba.
    - Leitura: o v3 usa a segunda bomba em cerca de 13 % dos turnos, mas isso não o fez ganhar do v2. Sem meta (decisão 5).
- [x] **T9**: Verificação final: todo CA com teste; `gofmt -l`, `go vet ./...` e `go test ./...` limpos; status da spec → `concluída`.
