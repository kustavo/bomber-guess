# Marco 12: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: Docs: EST-09, ORD-08, FEC-01 a FEC-07 e glossário em `REGRAS.md`; MAP-05 e exemplo em `EDITOR.md`; `blocos_fechados` em `ARQUITETURA.md` 1.2. (CA-10)
- [x] **T2**: `Config.TurnoFechamento`, `RelatorioEtapa.BlocosFechados` e validação no mapa; casos novos em `mapa_test.go`; relatório completo em `resolver_test.go`. (CA-09, CA-10)
- [x] **T3**: `fechamento.go`: `Anel`, `CasasQueFecham`, `fechar` e chamada em `ResolverTurno`; `fechamento_test.go`. (CA-01 a CA-08)
- [x] **T4**: Terminal: evento `fechamento: N casas` e blocos fixos novos no quadro; teste. (CA-11)
- [x] **T5**: `aleatorio-v1`: perigo nas casas que fecham, auxiliar sem fechamento, oráculo de teste ajustado; caso novo em `fuga_test.go`, conferido falhando sem a mudança. (CA-12)
- [x] **T6**: `mapas/exemplo.json` com `turno_fechamento` 30; partidas do bot com as sementes de 1 a 50 continuam passando; medição de empates registrada na spec. (CA-13)
- [x] **T7**: Verificação final: todo CA com teste; `gofmt -l`, `go vet ./...` e `go test ./...` limpos; status da spec → `concluída`.
- [x] **T8** (adendo): `Area`, `Config.AreaMinima`, `AreaMinimaEfetiva`, `AnelFecha`; `CasasQueFecham` respeita FEC-08; `VerificarMapa` valida `area_minima`; docs (`REGRAS.md` EST-10 e FEC-08, `EDITOR.md`, `ROADMAP.md`) e `mapas/exemplo.json` com 5×5; testes. (CA-14 a CA-18)
