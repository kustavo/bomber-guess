# Marco 02: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: `mapa.go`: `VerificarMapa` rejeita `limite_turnos` < 1 e atributos de `jogador_padrao` < 1; casos novos em `mapa_test.go`. (CA-49)
- [x] **T2**: `tabuleiro_test.go`: helper `montar` (desenho ASCII, opções de bombas e atributos) com teste próprio do helper.
- [x] **T3**: `validar.go`: `Infracao` e `Validar` com VAL-01 a VAL-03, VAL-05 e VAL-06 a VAL-08, DEC-08; `validar_test.go`. (CA-01 a CA-15)
- [x] **T4**: `relatorio.go` (tipos) e o esqueleto de `resolver.go`: estrutura `mesa`, preenchimento com `ESPERAR`, indexação dos planos, laço de etapas, relatório por etapa e novo estado (turno + 1, `etapas_neste_turno`). Movimentos (ORD-01, MOV-01 a MOV-05) com resultado `BLOQUEADA`/`ABORTADA`; `movimento_test.go`. (CA-16, CA-17, CA-20, CA-46)
- [x] **T5**: Bombas e pavio (ORD-02, ORD-03, BOM-01 a BOM-05, BOM-11) em `resolver.go`; explosão simples de uma pilha em `explosao.go`, sem cadeia; `bomba_test.go`. (CA-21 a CA-25)
- [x] **T6**: `explosao.go` completo: paradas em blocos e bombas, reação em cadeia, DEC-05, DEC-01; remoção de blocos (ORD-06); `explosao_test.go` e os casos de bloco destruído em `movimento_test.go`. (CA-18, CA-19, CA-26 a CA-35)
- [x] **T7**: Mortes (ORD-05, FIM-01) e ordem da etapa; `etapa_test.go`. (CA-36 a CA-39)
- [x] **T8**: `fim.go`: `Desfecho`, `VerificarFim`; parada no meio do turno e partida já terminada em `ResolverTurno`; `fim_test.go`. (CA-40 a CA-44)
- [x] **T9**: Relatório completo, robustez (resultado `IGNORADA`, planos duplicados ou de jogador inexistente) e pureza/determinismo (chamada repetida, planos embaralhados, entradas intactas); ida e volta dos exemplos JSON de `docs/ARQUITETURA.md` (1.1 `Infracao`, 1.2 `RelatorioEtapa`); `resolver_test.go`. (CA-45, CA-47, CA-48)
- [x] **T10**: P1, P2 e P3 aplicadas em `docs/` na fase de planejamento.
- [x] **T11**: verificação final: todo CA com teste (`grep` de cada ID de regra nos testes); `gofmt -l`, `go vet ./...` e `go test ./...` limpos; status da spec → `concluída`.
