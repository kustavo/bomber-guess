# Marco 05: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: `internal/fila`: interface, tópicos, `Memoria`, e `TestarContrato` aplicado à memória. (CA-12)
- [x] **T2**: `internal/partida`: tipos de registro, `Nova`, disparo dos bots via fila, `Avancar` e `Visao` com as fases e a liberação das etapas; helpers de teste. (CA-01, CA-02, CA-03, CA-13)
- [x] **T3**: Registro do turno: falhas (`PRAZO_ESTOURADO`, `PANICO`), três versões das ações, infrações, plano atrasado, de outro turno ou repetido. (CA-04, CA-05, CA-06, CA-07, CA-14, CA-15)
- [x] **T4**: Fim da partida e tópicos de saída: `ENCERRADA` depois da última etapa, limite de turnos, `turno-resolvido` e `partida-finalizada`, partidas independentes, igualdade com a simulação síncrona. (CA-08, CA-09, CA-10, CA-11, CA-16)
- [x] **T5**: `Gerenciador`: pedido (nome, mapa, bots, semente), erros, lista na ordem de criação, laço em tempo real e `Encerrar`. (CA-18, CA-19, CA-10)
- [x] **T6**: `internal/api`: rotas, respostas, erros em JSON, 405, 404, 501 e `horario_servidor`. (CA-17, CA-18, CA-19, CA-20, CA-21, CA-22, CA-23, CA-24)
- [x] **T7**: `cmd/servidor`: flags, diretório de mapas padrão, `http.Server`, encerramento por contexto ou sinal. (CA-25)
- [x] **T8**: Verificação final: todo CA com teste; `gofmt -l`, `go vet ./...` e `go test ./...` limpos, inclusive com `-race`; status da spec → `concluída`.
