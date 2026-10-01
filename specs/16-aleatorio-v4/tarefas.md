# Marco 16: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

- [x] **T1**: `VersaoV4`, `NovoV4`, campo `preveAmeaca`, `Versao()`; registro no catálogo. Testes de versão e catálogo. (CA-01, CA-03)
- [x] **T2**: `cruz` e `calcularAmeaca` (D1). Testes de tabela: corredor com as etapas exatas; blocos fixo e destrutível; adversário morto e o próprio bot; pavio longo para a ameaça final. (CA-06, CA-07, CA-08, CA-09)
- [x] **T3**: Parâmetro `a *ameaca` em `preverLinha` e nas funções de planejamento (D2, D3); corpo do `Planejar` vira `planejarCom`. v1 a v3 passam `nil`. Todos os testes existentes passam sem mudança (guarda de que nada mudou nas versões antigas).
- [x] **T4**: `protegido` e as duas passadas do v4 (D4 a D7). Testes: igual ao v3 sem ameaça; protegido quando existe; seguro quando não existe; plano com bomba só protegido; janela sem trocar anel alvo por proteção. (CA-04, CA-10, CA-11, CA-12)
- [x] **T5**: v4 na tabela `versoesDeTeste` (testes herdados, tempo e partida só com v4). (CA-02, CA-05)
- [x] **T6a** (R1): peso na ameaça, `nivel(w)` e os níveis no `planejarV4` (D9, D10). Testes: pesos do corredor; plano protegido no menor nível possível. (CA-16, CA-17)
- [x] **T6b** (R2): `mira` (D11). Testes de tabela: fora do alcance, beco sem saída, 1 de 4, bomba do turno seguinte. (CA-18)
- [x] **T6c** (R2): `melhorPlanoComMira` e o uso no v4 (D12, D13). Testes: encurralar quando possível; a mira nunca vence a proteção; testes herdados e tempo continuam passando. (CA-19, CA-20, CA-05)
  - Ajustes na implementação, dentro do D12: a busca com mira usa um gerador próprio (`geradorMira`), para que, sem mira, o resto do plano saia igual ao do v3 (CA-04); em cada nível, se o plano com mira não passar na conferência de proteção, o v4 tenta o mesmo nível sem mira antes de subir. O teste do CA-17 dá ao bot pavio 9, porque com a mira ele passou a matar o adversário parado na simulação, o que apaga a ameaça da conta. Exceção R3 aplicada no CA-21 herdado do marco 3. `TestTempoDePlanejamento` com o v4: 0,15 s no total.
- [x] **T6**: `causaDaMorte` e comparação v4 × v3 no mapa de exemplo, com as metas e os registros (D8). Se as metas não forem atingidas, parar e voltar com os números (decisões 4 e 6). (CA-13, CA-14, CA-15)
  - **Resultado (sementes 1 a 200, mapa de exemplo)**:
    - 4 × v3: 349 mortes por bomba de adversário do mesmo turno, 22 por bomba já no tabuleiro, 103 por fechamento, 20 pela própria bomba, 45 pela própria e de adversário; 76 empates.
    - 4 × v4: 176 por bomba de adversário do mesmo turno (50 % do v3; meta ≤ 70 %), 2 por bomba já no tabuleiro, 134 por fechamento, 75 pela própria bomba, 62 pela própria e de adversário; 155 empates.
    - 2 × v3 contra 2 × v4: v4 venceu 54, v3 venceu 18 (3 ×; meta ≥ 1,5 ×), 128 empates.
    - Sem a mira (só R1), tinha sido 199 mortes (57 %) e 35 × 26 vitórias; só com a zona de ameaça (primeira versão), 251 (72 %) e 40 × 43.
- [x] **T7**: Verificação final: todo CA com teste; `gofmt -l`, `go vet ./...` e `go test ./...` limpos; status da spec → `concluída`.
