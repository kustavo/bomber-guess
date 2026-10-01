# Marco 07: editor de mapas e criação de partidas

**Status**: concluída
**Roadmap**: `docs/ROADMAP.md`, marco 7
**Documentos de referência**: `docs/EDITOR.md`; contrato da API já definido em `specs/05-servidor/spec.md` (decisões 4 a 6) e telas do marco 6 (`specs/06-frontend/spec.md`)

## Objetivo

Permitir criar partidas pela tela, sem editar JSON à mão. O usuário desenha um mapa num editor no estilo "maker" (bloco fixo, bloco destrutível, posição inicial), escolhe o bot de cada posição inicial, dá nome à partida e a inicia; em seguida passa a assisti-la na tela do marco 6. Para isso o servidor ganha o `POST /mapas`, que hoje responde 501.

### Termos usados nesta spec

- **Item**: o que o editor coloca numa casa: bloco fixo, bloco destrutível ou posição inicial. Uma casa tem no máximo um item.
- **Ferramenta**: o item selecionado na paleta, ou a borracha (remove o item da casa).
- **Mapa em edição**: o mapa que a tela monta, no formato de `docs/EDITOR.md`.
- **API falsa** e **diretório temporário**: nos testes do frontend, as respostas HTTP são controladas pelo teste; nos do backend, os mapas são gravados num diretório temporário.

## Escopo

**Dentro**
- Backend: `POST /mapas` (API-03) grava o mapa no diretório de mapas do servidor, como `<nome>.json`, depois de validá-lo (MAP-05).
- Frontend, nova tela **Criar partida** (rota por hash, com link na lista de partidas):
  - campos de `config` e `jogador_padrao`, preenchidos com valores padrão (decisão 3);
  - tabuleiro vazio de `largura` × `altura`, editável com a paleta: bloco fixo, bloco destrutível, posição inicial e borracha;
  - posições iniciais numeradas na ordem em que foram colocadas (MAP-03);
  - para cada posição inicial, um seletor de versão de bot com o catálogo de `GET /bots` (API-09);
  - nome da partida, nome do mapa e semente;
  - botão **Iniciar**: `POST /mapas` e depois `POST /partidas` (API-04); com sucesso, abre a tela da partida (`#/partidas/<nome>`);
  - exibição dos erros devolvidos pela API (campo `erro`).
- Testes automatizados do backend (Go) e do frontend (Vitest).

**Fora** (fica para outro marco)
- Abrir, listar ou editar mapas já salvos (`GET /mapas`, `GET /mapas/{nome}`) e criar partida a partir de um mapa existente pela tela (decisão 6).
- Apagar ou sobrescrever mapas (decisão 2).
- Ferramentas de conveniência: espelhamento/simetria, desfazer, pintar arrastando, gerar mapa aleatório.
- Mudanças nas regras do jogo ou no formato do mapa.
- Mudanças em `POST /partidas`: o corpo da decisão 6 do marco 5 já aceita uma versão por posição.

## Regras cobertas

- `API-03`: `POST /mapas` salva um mapa criado no editor.
- `API-04`: `POST /partidas` cria a partida com os bots de cada posição (consumido, sem mudanças).
- `API-09`: `GET /bots` fornece as versões para os seletores (consumido).
- `API-02`: toda resposta inclui `horario_servidor` (vale para o `POST /mapas`).
- `MAP-01`, `MAP-02`: o mapa em edição tem `config` e `jogador_padrao` no formato de `docs/EDITOR.md`.
- `MAP-03`: a versão `i` joga na posição inicial `i`; a ordem das posições no mapa define os ids `jogador_<i+1>`.
- `MAP-05`: validação do mapa (no servidor; o editor impede por construção parte dos casos).
- `MAP-06`: quantidade de versões igual à de posições iniciais.

## Critérios de aceitação

Backend: testes Go com `httptest` e diretório temporário. Frontend: Vitest + Testing Library, com API falsa.

### `POST /mapas` (backend)

- **CA-01** (API-03, API-02, MAP-01, MAP-02): **Dado** um mapa válido com nome novo, **quando** se faz `POST /mapas`, **então** a resposta é 201 com `{"nome", "horario_servidor"}`, o arquivo `<nome>.json` existe no diretório de mapas e, carregado com a mesma função que carrega `mapas/exemplo.json`, é igual ao mapa enviado.
- **CA-02** (API-03, API-04): **Dado** um mapa salvo pelo `POST /mapas`, **quando** se faz `POST /partidas` com esse nome de mapa e um bot por posição inicial, **então** a resposta é 201 (o servidor encontra o mapa sem reiniciar).
- **CA-03** (MAP-05): **Dado** um mapa inválido, **quando** se faz `POST /mapas`, **então** a resposta é 400 com o motivo em `erro` e nenhum arquivo é criado. Casos testados: um por condição de MAP-05 (dimensão ≤ 0, item fora do tabuleiro, posição inicial sobre bloco, casa com bloco fixo e destrutível, menos de 2 posições iniciais, `limite_turnos` < 1, `turno_fechamento` inválido, `area_minima` inválida, atributo de `jogador_padrao` < 1).
- **CA-04** (API-03): **Dado** um corpo que não é JSON de mapa, ou um `nome` vazio ou fora do padrão de nome de mapa (decisão 1, ex.: `../x`, `a/b`, `A B`), **quando** se faz `POST /mapas`, **então** a resposta é 400 com `erro` e nada é gravado fora nem dentro do diretório de mapas.
- **CA-05** (API-03): **Dado** que já existe `<nome>.json` no diretório de mapas, **quando** se faz `POST /mapas` com esse nome, **então** a resposta é 409 com `erro` e o arquivo existente não muda (decisão 2).
- **CA-06** (API-03): **Dado** um método diferente de `POST` em `/mapas`, **quando** chamado, **então** a resposta é 405 (e `POST /mapas` não responde mais 501).

### Mapa em edição (lógica pura, frontend)

- **CA-07** (MAP-01, MAP-02): **Dado** um editor recém-aberto, **quando** o mapa em edição é montado, **então** ele não tem blocos nem posições iniciais, e `config` e `jogador_padrao` têm os valores padrão da decisão 3.
- **CA-08**: **Dada** uma ferramenta de item selecionada, **quando** se clica numa casa, **então** a casa passa a ter esse item, substituindo o que havia nela; **quando** se clica com a borracha, a casa fica vazia. Nenhuma casa fica com dois itens.
- **CA-09** (MAP-03): **Dadas** três posições iniciais colocadas na ordem A, B, C, **quando** o mapa em edição é montado, **então** `posicoes_iniciais` é `[A, B, C]` e a tela as numera 1, 2 e 3; **quando** B é removida (borracha ou outro item por cima), **então** fica `[A, C]`, numeradas 1 e 2, e o bot escolhido para C continua com C.
- **CA-10** (MAP-05): **Dado** um mapa com itens, **quando** `largura` ou `altura` diminui, **então** os itens que ficariam fora do tabuleiro são removidos e o mapa em edição continua só com casas dentro do tabuleiro.
- **CA-11** (MAP-01, MAP-02, MAP-03): **Dado** um mapa desenhado no editor, **quando** ele é convertido para o corpo do `POST /mapas`, **então** o JSON tem exatamente os campos e tags de `docs/EDITOR.md` (`nome`, `config`, `jogador_padrao`, `blocos_fixos`, `blocos_destrutiveis`, `posicoes_iniciais`).

### Tela Criar partida (componentes, frontend)

- **CA-12** (API-09): **Dada** a tela aberta, **quando** `GET /bots` responde, **então** cada posição inicial tem um seletor com exatamente as versões do catálogo, na ordem recebida; uma posição nova já vem com a primeira versão selecionada (decisão 4).
- **CA-13** (MAP-05, MAP-06): **Dado** um mapa com menos de 2 posições iniciais, ou sem nome de partida ou de mapa, **quando** a tela é exibida, **então** o botão Iniciar está desabilitado e há um aviso dizendo o que falta; nenhuma requisição é feita.
- **CA-14** (API-03, API-04, MAP-03, MAP-06): **Dado** um mapa válido com N posições iniciais e um bot escolhido para cada, **quando** se clica em Iniciar, **então** a tela faz `POST /mapas` com o mapa (CA-11) e, após 201, `POST /partidas` com `{"nome", "mapa", "bots", "semente"}`, em que `bots[i]` é a versão escolhida para a posição `i`; após 201, a rota passa a `#/partidas/<nome>`.
- **CA-15** (API-03, API-04): **Dada** uma resposta de erro (400, 409 ou servidor fora do ar) em `POST /mapas` ou em `POST /partidas`, **quando** se clica em Iniciar, **então** a tela mostra o `erro` recebido (ou "servidor indisponível"), continua no editor com o mapa intacto e, se o erro veio do `POST /mapas`, não faz `POST /partidas`.
- **CA-16** (API-03, API-04): **Dado** que o `POST /mapas` deu 201 e o `POST /partidas` falhou (ex.: 409 por nome de partida repetido), **quando** o usuário corrige o nome da partida sem mexer no mapa e clica de novo em Iniciar, **então** a tela não repete o `POST /mapas` e faz só o `POST /partidas`; se o mapa foi alterado nesse meio-tempo, o `POST /mapas` volta a ser feito (decisão 5).
- **CA-17**: **Dada** a lista de partidas, **quando** se clica em "Criar partida", **então** a rota passa à tela do editor; e abrir essa rota diretamente (recarregar a página) mostra o editor vazio.

### Verificação

- **CA-18**: `gofmt`, `go vet` e `go test ./...` limpos; em `frontend/`, `npm run check` sem erros e `npm test` passando.
- **CA-19**: **Dado** o servidor Go rodando e o frontend no servidor de desenvolvimento, **quando** se desenha um mapa, escolhem-se os bots e clica-se em Iniciar, **então** a partida aparece na tela do marco 6 e o turno avança. Verificação manual registrada no `tarefas.md` (como o CA-16 do marco 6).

## Decisões

1. **Nome do mapa.** Vira nome de arquivo, então precisa ser seguro. Decisão: o mesmo padrão que o `POST /partidas` já exige para o nome do mapa (marco 5, D8): `^[a-z0-9-]{1,64}$`. Qualquer outro valor dá 400 (CA-04). Com um padrão só, todo mapa salvo pode ser usado numa partida. (Ajustada no planejamento: a versão aprovada propunha `^[a-z0-9][a-z0-9_-]{0,63}$`, que aceitaria mapas impossíveis de usar.) O editor sugere como nome do mapa o nome da partida, quando este já obedece ao padrão.
2. **Mapa com nome já existente.** Decisão: 409, sem sobrescrever (CA-05). Sobrescrever ou apagar mapas fica para quando houver edição de mapas salvos. Isso também protege `exemplo.json`.
3. **Valores padrão e limites do formulário.** `docs/EDITOR.md` diz "mapa vazio", mas não diz o tamanho nem a `config`. Decisão: padrões iguais aos de `mapas/exemplo.json` (15 × 13, `limite_turnos` 50, `turno_fechamento` 30, `area_minima` 5 × 5, `prazo_planejamento_ms` 1000, `duracao_etapa_ms` 1000, `jogador_padrao` 2/2/3/7). O formulário limita `largura` e `altura` a 1–30 só na tela, para o tabuleiro caber; não se cria regra nova de domínio (o servidor continua aplicando apenas MAP-05).
4. **Bot padrão de uma posição nova.** Decisão: a primeira versão do catálogo (ordem alfabética, marco 5 CA-17). Se `GET /bots` falhar, a tela mostra o erro e o Iniciar fica desabilitado.
5. **`POST /mapas` deu certo e `POST /partidas` falhou.** O mapa já está gravado; repetir o `POST /mapas` daria 409. Decisão: a tela lembra que aquele mapa (mesmo nome e mesmo conteúdo) já foi salvo e pula o `POST /mapas` na nova tentativa (CA-16). Alternativa descartada: tornar o `POST /mapas` idempotente para conteúdo igual, que complica o servidor.
6. **Partida a partir de mapa existente.** Hoje a única forma de usar `exemplo.json` pela tela seria redesenhá-lo. Decisão: fora deste marco, como `docs/EDITOR.md` descreve (o usuário começa com um mapa vazio). Um marco futuro pode trazer `GET /mapas` e "abrir mapa".
7. **Semente.** `docs/EDITOR.md` não fala dela. Decisão: campo opcional na tela, preenchido com 1 (o padrão do servidor), enviado sempre em `POST /partidas`.
8. **Validação no frontend.** Decisão: o editor só impede o que é barato e óbvio (menos de 2 posições, nomes vazios, CA-13; dois itens na mesma casa e itens fora do tabuleiro são impossíveis por construção, CA-08 e CA-10). O restante de MAP-05 (`turno_fechamento`, `area_minima`, atributos < 1) fica com o servidor, cujo `erro` é exibido (CA-15). Assim a regra tem uma só implementação.
