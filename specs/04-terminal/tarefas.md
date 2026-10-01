# Marco 04: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: Catálogo em `internal/bots`: `Fabrica`, `Catalogo`, `Padrao` (com `aleatorio-v1`), `Versoes` e `Criar`. (CA-01, CA-02)
- [x] **T2**: `Chamar` em `internal/bots`: cópia do estado, goroutine com `recover`, prazo, e prazo ≤ 0 sem limite. Testes de unidade com bots de teste que alteram o estado, dormem e entram em `panic`. (CA-04, CA-05, CA-06)
- [x] **T3**: `quadro` em `cmd/terminal/desenho.go`: legenda e prioridade, `&` para dois jogadores, mortos fora, eventos, blocos removidos depois da etapa. (CA-14, CA-15)
- [x] **T4**: Helpers de teste (bots de teste, catálogo de teste, mapa em `t.TempDir()`) e o laço `jogar`: chamada protegida, validação, avisos, resolução, desenho, atraso e linha do desfecho. (CA-04, CA-05, CA-06, CA-07, CA-09, CA-10, CA-11, CA-12)
- [x] **T5**: `rodar` e `main`: argumentos, mapa padrão, uma versão para todas as posições, `-listar`, limite de 9 jogadores e erros de uso com código 2. (CA-03, CA-17, CA-18, CA-19)
- [x] **T6**: De ponta a ponta no mapa de exemplo: desfecho, turnos dentro do limite, saída determinística, dimensões e quantidade de tabuleiros, atraso por argumento e padrão do mapa. (CA-08, CA-13, CA-16, CA-20)
- [x] **T7**: Verificação final: todo CA com teste; `gofmt -l`, `go vet ./...` e `go test ./...` limpos; status da spec → `concluída`.
