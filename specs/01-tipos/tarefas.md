# Marco 01: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: `backend/go.mod`, `doc.go` e `pureza_test.go` (só imports da biblioteca padrão). (CA-12, parte 1)
- [x] **T2**: `enums.go`, `estado.go` (só os tipos), `plano.go` e `docs_test.go` (helper que extrai JSON dos docs); testes de ida e volta e de omissão de chaves. (CA-01, CA-02, CA-03, CA-04)
- [x] **T3**: `posicao.go`: `Vizinha` e `NoTabuleiro`. (CA-05, CA-06)
- [x] **T4**: `Estado.Copiar()` e teste de isolamento de cada slice e de `Morte`. (CA-13)
- [x] **T5**: `bot.go`: a interface `Bot` e um bot de teste em `package jogo_test`, que garante que a interface pode ser implementada por um pacote externo. (CA-07)
- [x] **T6**: `mapa.go`: `LerMapa`, `VerificarMapa`, `EstadoInicial` e `CalcularEtapas`, com testes de tabela para cada caso inválido. (CA-08, CA-09, CA-10, CA-12, parte 2)
- [x] **T7**: `mapas/exemplo.json` e o teste das propriedades do layout. (CA-11)
- [x] **T8**: verificação final: todo CA com teste; `gofmt -l`, `go vet ./...` e `go test ./...` limpos; status da spec → `concluída`.
