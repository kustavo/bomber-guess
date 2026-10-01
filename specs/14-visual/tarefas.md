# Marco 14: tarefas

Ordem de execução. Cada tarefa é pequena, termina com testes passando e cita os critérios que atende.

### Lógica pura

- [x] **T1**: `sprites.ts`: paleta, os 12 sprites desenhados em texto, `NOMES_SPRITES`, `retangulos`, `idSprite` (D2, D3). Testes: cada nome existe; 16 linhas de 16 letras; só letras da paleta; `retangulos` junta pixels vizinhos de mesma cor, pula transparentes e cobre exatamente os pixels pintados. (CA-01)
- [x] **T2**: `chamas.ts`: `formasDasChamas` (D6). Testes de tabela: origem; braços horizontal e vertical; pontas nas quatro direções; braço de uma casa; explosão sem braços; borda do tabuleiro (só um lado); duas explosões que se cruzam; casa fora de explosão. (CA-03)
- [x] **T3**: `TabuleiroExibido.explosoes` (D7) e `explosao()` em `fabricas.ts`. Testes: início do turno sem explosões; etapa k com as explosões só da etapa k. (CA-09)

### Componentes

- [x] **T4**: `SpritesSvg.svelte` e `Sprite.svelte` (D1, D10); `App` monta `SpritesSvg`. Testes: um `<symbol id="sprite-<nome>">` por nome; retângulos com as cores do sprite; nenhum `<image>` nem `href` externo. (CA-01, CA-02)
- [x] **T5**: `Tabuleiro.svelte` com sprites: piso, blocos, bombas com o pavio, chamas pela forma, jogadores com cor e número, marca de bloqueio (D4, D5). Testes novos para os sprites (`href` de cada `<use>`, cor do jogador, forma de cada chama, casa fechada como `bloco-fixo`). Os testes do marco 6 passam sem mudança. (CA-02, CA-04, CA-05, CA-06, CA-07, CA-08, CA-09)
- [x] **T6**: `TabuleiroEditor.svelte` com sprites e camadas (D9). Testes: blocos e posições iniciais com o `href` certo, cor e número das posições. Os testes do marco 7 passam sem mudança. (CA-10)
- [x] **T7**: Paleta da `TelaCriar` com ícones de sprite (a borracha mostra o piso). Teste do `href` de cada botão. (CA-11)

### Animação

- [x] **T8**: `sprites.css` (D8, D10): classes de animação, `crispEdges`, variável `--transicao` do jogador e `prefers-reduced-motion`; `Tabuleiro` aplica as classes. Testes: bombas e chamas com a classe; o CSS tem a media query que desliga as animações e a transição. (CA-12)

### Verificação

- [x] **T9**: `npm run check` e `npm test` limpos em `frontend/`, com os testes dos marcos 6 e 7 sem alteração. (CA-13)
- [x] **T10**: Verificação manual: servidor Go e `npm run dev`; criar uma partida pelo editor, assistir, ampliar, alternar tema claro e escuro; capturas de tela. Retocar os desenhos se preciso. Registrar o resultado aqui e apagar de `mapas/` o mapa criado. (CA-14)
  - **Resultado (2026-10-01)**: servidor Go e `npm run dev`; o editor foi conferido vazio (piso, grade e ícones da paleta), e uma partida foi criada com `POST /partidas` no mapa `exemplo` (4 bots, semente 3), sem gravar mapa novo. Como o painel do navegador embutido estava escondido e devolvia capturas congeladas, as capturas foram feitas com o Chrome headless: tema claro, tema escuro (`--force-dark-mode`) e ampliado 3× (`--force-device-scale-factor=3`). Sprites nítidos, cruz da explosão com centro, braços e pontas, jogadores com cor e número, bomba com o pavio. Ajuste feito: frestas finas entre as casas de grama, corrigidas com um retângulo de fundo na cor do piso (`Tabuleiro.svelte`). Movimento reduzido conferido pelo teste da folha de estilos (`estilos.test.ts`), não no navegador.
- [x] **T11**: Verificação final: todo CA com teste ou verificação registrada; status da spec → `concluída`.

### Reabertura (2026-10-01)

- [x] **T12**: Pulso sutil e centrado (D11): prop `centrado` no `Sprite`; CSS com `view-box` e `scale(1.04)`. Testes: `transform-origin` da bomba no centro da casa; folha de estilos sem escala acima de 1,04. (CA-15)
- [x] **T13**: `Grade.svelte` nas duas telas (D12); contorno escuro das casas do editor removido. Testes: linhas internas certas, `pointer-events="none"`, posição da grade entre as camadas. (CA-16)
- [x] **T14**: `bombas_por_turno` e `bombas_usadas` no `JogadorExibido`; `textoBombasUsadas` (D13). Testes de tabela: planejamento, plantar executado, plantar descartado ou ignorado, etapas posteriores à exibida. (CA-17)
- [x] **T15**: Cartão do jogador no `Tabuleiro` (D14). Testes: aparece com mouse e com foco; conteúdo; some ao sair; acompanha a etapa; só para jogador vivo. (CA-17)
- [x] **T16**: `npm run check` e `npm test` limpos; captura no Chrome headless (grade, bomba e cartão); status da spec → `concluída`. (CA-13, CA-14)
  - **Resultado (2026-10-01)**: com o servidor Go e o `npm run dev` já abertos pelo usuário, foi criada a partida `teste-visual` (mapa `exemplo`, 4 bots, só em memória) e capturada no Chrome headless por CDP, com o mouse sobre o `jogador_1` na etapa 4: o cartão mostrou `jogador_1`, `aleatorio-v1`, `Bombas restantes: 2` e `usou 0 de 2 (0/2)`, à direita do jogador. A grade aparecia, mas fraca sobre a grama; a opacidade subiu de 0,25 para 0,35. O pulso da bomba (4 %, centrado na casa) foi conferido pelos testes; uma captura parada não mostra animação.

