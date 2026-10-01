# Marco 13: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: `busca.go`: `modoBusca` vira objetivo na casa final (D3); todos os testes do marco 3 continuam passando sem mudança. (base para CA-02, CA-04)
- [x] **T2**: `VersaoV2`, `NovoV2`, campo `antecipa`, `Versao()`; v2 no catálogo; `paraCadaVersao` e testes existentes rodando para as duas versões. (CA-01, CA-02, CA-03)
- [x] **T3**: `fechamento.go`: `naJanela`, `anelAlvo`, `regiao`, `distancias`, `blocoDeAbertura`, `casasDePlantio`, com testes de tabela em `fechamento_test.go`. (base para CA-07 a CA-13)
- [x] **T4**: `antecipacao.go`: `planejarNaJanela` com os itens 1, 3, 4, 5 e 6 da decisão 3; desvio em `Planejar` só dentro da janela (D2). Testes de identidade com o v1 fora da janela e sem fechamento. (CA-04, CA-06, CA-07, CA-08, CA-09, CA-10)
- [x] **T5**: `tentarAbrir` e o item 2 da decisão 3. (CA-11, CA-12)
- [x] **T6**: Tabuleiro de vários turnos com bloco de abertura, v2 sobrevive e v1 morre em alguma semente. (CA-13)
- [x] **T7**: `jogarPartida` com um bot por posição, classificação de mortes e comparações no mapa de exemplo; CA-06 do marco 3 com só v2; CA-05 dentro da janela, com o bot preso. (CA-02, CA-05, CA-14, CA-15, CA-16)
- [x] **T8**: Verificação final: todo CA com teste (`grep -rn "CA-" backend/internal/bots`); `gofmt -l`, `go vet ./...` e `go test ./...` limpos; resultados do CA-16 registrados aqui; status da spec → `concluída`.
  - **Resultado (2026-10-01)**: `gofmt -l` vazio; `go vet ./...` e `go test ./...` limpos (pacote do bot em ~3,3 s, sem precisar de `-short`). CA-16, sementes 1 a 200 no mapa de exemplo com área mínima 5×5:
    - 4×v1: 587 mortes por fechamento (anel médio 2,34), 76 por bomba, 69 empates;
    - 4×v2: 91 mortes por fechamento (15,5% do v1; anel médio 2,63), 441 por bomba, 76 empates; nenhuma morte por fechamento com saída disponível;
    - 2×v1 + 2×v2: v2 venceu 82 e v1 venceu 18; 100 empates.
