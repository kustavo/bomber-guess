# Marco 07: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

### Backend

- [x] **T1**: `Gerenciador.SalvarMapa` com gravação exclusiva e atômica (D1 a D3). Testes de tabela: mapa válido gravado e relido igual por `jogo.LerMapa`; nome inválido (`""`, `../x`, `a/b`, `A B`); um caso por condição de MAP-05; nome já existente sem alterar o arquivo; nenhum arquivo, nem temporário, sobra após erro. (CA-01, CA-03, CA-04, CA-05)
- [x] **T2**: `POST /mapas` na API (`salvarMapa`, `respostaMapa`, D4); tirar o 501. Testes `httptest`: 201 com `nome` e `horario_servidor`; corpo não JSON e campo desconhecido (400); MAP-05 (400); repetido (409); `GET /mapas` (405); salvar e depois `POST /partidas` com esse mapa (201). Atualizar a tabela do CA-24 do marco 5. (CA-01 a CA-06)

### Frontend: lógica pura

- [x] **T3**: Tipos (`Atributos`, `Mapa`, `RespostaBots`, `RespostaMapa`, `PedidoPartida`) e cliente da API (`buscarBots`, `salvarMapa`, `criarPartida`, com `pedir` genérico). Testes com `fetch` falso: método, caminho, corpo JSON e `ErroApi` nos erros. `clienteFalso` em `fabricas.ts` com os métodos novos. Proxy de `/mapas` no Vite. (CA-12, CA-14, CA-15)
- [x] **T4**: `editor.ts`: `novoMapaEmEdicao`, `itemEm`, `aplicarFerramenta`, `redimensionar`, `paraMapa`, `botsDasPosicoes` (D5 a D7). Testes. (CA-07, CA-08, CA-09, CA-10, CA-11)
- [x] **T5**: `editor.ts`: `pendencias` e `sugerirNomeMapa` (D10, D11). Testes. (CA-13)
- [x] **T6**: `criacao.ts`: `chaveMapa` e `iniciarPartida` (D8). Testes com cliente falso: sucesso nas duas chamadas e ordem delas; erro no mapa sem `POST /partidas`; erro na partida devolvendo `mapaSalvo`; nova tentativa sem `POST /mapas`; mapa alterado grava de novo; servidor fora do ar ("servidor indisponível"). (CA-14, CA-15, CA-16)
- [x] **T7**: Rota `#/criar` em `rota.ts`. Testes. (CA-17)

### Frontend: telas

- [x] **T8**: `TabuleiroEditor.svelte` (D13, D14): casas clicáveis, itens e números das posições iniciais. Coberto pelos testes da T9. (CA-08, CA-09)
- [x] **T9**: `TelaCriar.svelte`: formulário com padrões, paleta, tabuleiro, seletores de bot, aviso de pendências, botão Iniciar, erros e navegação (D9 a D12). Testes com Testing Library: seletores com o catálogo e primeira versão por padrão; botão desabilitado com aviso e sem requisições; clique em casas muda o mapa e a numeração; Iniciar com sucesso chama as duas rotas com os corpos certos e muda o hash; erro exibido com o mapa intacto; nova tentativa após 409 da partida sem novo `POST /mapas`. (CA-08, CA-09, CA-12, CA-13, CA-14, CA-15, CA-16)
- [x] **T10**: `App.svelte` despacha `criar`; link "Criar partida" na `ListaPartidas`. Testes: clicar no link leva ao editor; abrir `#/criar` direto mostra o editor vazio. (CA-17)

### Reabertura (2026-10-01)

- [x] **T13**: `normalizarNome` e as pendências de nome fora do padrão (D15); campos de nome da `TelaCriar` convertem para minúsculas. Testes de `editor.ts` e da tela. (CA-20)
- [x] **T14**: `outline: none` nas casas do `TabuleiroEditor` (D16). Teste da folha de estilos do componente; captura no Chrome headless clicando e movendo o mouse. (CA-21)
  - **Resultado (2026-10-01)**: reproduzido no Chrome headless (por CDP): depois de clicar numa casa e mover o mouse, a casa focada tinha `outline: auto 5px`, desenhado como um anel preto e branco de várias casas de raio. Com `outline: none`, a mesma sequência mostra só o realce da casa sob o mouse.
- [x] **T15**: `npm run check` e `npm test` limpos; status da spec → `concluída`.

### Verificação

- [x] **T11**: Verificação manual de ponta a ponta: servidor Go + `npm run dev`; desenhar um mapa, escolher bots, Iniciar, ver a partida avançar. Registrar o resultado aqui e apagar de `mapas/` o mapa criado no teste. (CA-19)
  - **Resultado (2026-10-01)**: servidor Go (`-C backend run ./cmd/servidor`) e `npm run dev` pelo navegador embutido. Mapa 15 × 13 com 2 blocos fixos, 2 destrutíveis e 3 posições iniciais (bot `aleatorio-v2` na 2ª); prazo 300 ms e etapa 200 ms. Iniciar levou a `#/partidas/teste-editor`, com o tabuleiro desenhado, os três jogadores e os bots certos no painel; o turno avançou de 2 para 5 em 4 s. Repetir com o mesmo nome mostrou `nome já usado: mapa "teste-editor"` e manteve o mapa no editor. O arquivo gravado tinha permissão 0644 e nenhum temporário sobrou. `mapas/teste-editor.json` apagado.
- [x] **T12**: Verificação final: todo CA com teste; `gofmt -l`, `go vet ./...` e `go test ./...` limpos; em `frontend/`, `npm run check` e `npm test` limpos; status da spec → `concluída`. (CA-18)
