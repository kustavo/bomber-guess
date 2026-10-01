# Roadmap

Um marco por conversa. Em toda conversa vai o `AGENTS.md`; além dele, só os documentos listados no marco. ("Marco" aqui evita confusão com a **etapa** do jogo.)

Cada marco segue o fluxo de `specs/README.md` (`/especificar` → `/planejar` → `/implementar`) e tem sua spec em `specs/NN-nome/`. As descrições abaixo são o ponto de partida da spec, não a substituem.

| # | Marco | Documentos | Modelo sugerido |
|---|---|---|---|
| 1 | Tipos em Go + mapa de exemplo | REGRAS (seções 1 a 3), BOTS (interface), EDITOR (formato do mapa) | barato |
| 2 | Validador e resolução do turno com testes | REGRAS, ARQUITETURA (seção 1) | forte |
| 3 | Bot simples | BOTS, REGRAS | barato/médio |
| 4 | Partida no terminal | ARQUITETURA (seção 1) | barato |
| 5 | Loop em tempo real + API + fila em memória | ARQUITETURA (seções 2 a 4) | médio |
| 6 | Frontend Svelte para assistir | ARQUITETURA (seção 4) | barato |
| 7 | Editor de mapas | EDITOR | barato |
| 8 | Kafka | ARQUITETURA (seção 3) | médio |
| 9 | Ranking e estatísticas | RANKING | médio |
| 10 | Bots de outras IAs | BOTS, REGRAS | cada IA |
| 11 | Partidas salvas em disco | ARQUITETURA (seção 2) | médio |
| 12 | Fechamento do tabuleiro | REGRAS (seções 2, 6 e 7), ARQUITETURA (seção 1.2) | médio |

## 1. Tipos em Go

- Tipos `Estado`, `Plano`, `Acao`, `Bomba`, `Jogador` e a interface `Bot` em `backend/internal/jogo`.
- Serialização JSON idêntica aos exemplos de `docs/REGRAS.md` (teste de ida e volta).
- `mapas/exemplo.json` escrito à mão e função para carregá-lo como estado inicial.

## 2. Validador e resolução do turno

Pacote puro, com testes de tabela para cada regra. Casos mínimos:

- pavio (não diminui na etapa em que foi plantada; pavio 3 na etapa 2 explode na etapa 5) (BOM-05, ORD-03);
- pilha de bombas (potência combinada; a pilha explode junto) (BOM-03, BOM-04);
- reação em cadeia (o fogo para na bomba atingida, que explode com a própria potência) (BOM-07, BOM-09);
- bloco destrutível destruído no meio do turno e atravessado depois (MOV-03, ORD-06);
- movimento bloqueado que aborta o restante do plano (MOV-05);
- jogador que foge na mesma etapa da explosão e sobrevive (ORD-07);
- morte simultânea (empate) e vitória (FIM-02, FIM-03);
- bombas que atravessam turnos e bombas de jogador morto (BOM-11, BOM-12);
- validador (infrações): excesso de ações, direção inválida, saída do tabuleiro, bloco fixo, excesso de bombas (VAL-01 a VAL-05).

Antes de começar, confirmar as decisões `DEC-NN` marcadas como [PROPOSTA] em `docs/REGRAS.md` (seção 8).

## 3. Bot simples

Anda aleatoriamente (com semente) e foge de explosões previstas. Serve só para ter algo jogando.

## 4. Partida no terminal

`backend/cmd/terminal`: roda uma partida inteira entre bots e imprime o tabuleiro em ASCII a cada etapa. Aqui já se vê o jogo funcionando de verdade.

## 5. Loop em tempo real e API HTTP

Loop por partida com fases de planejamento e execução, fila atrás de uma interface (primeiro em memória) e endpoints de `docs/ARQUITETURA.md`.

## 6. Frontend Svelte (só assistir)

Consulta `GET /partidas/{nome}/estado`, desenha o tabuleiro, anima as etapas e mostra o cronômetro sincronizado com o servidor.

## 7. Editor de mapas

Conforme `docs/EDITOR.md`.

## 8. Kafka

Nova implementação da interface de fila; nada acima dela muda.

## 9. Ranking e estatísticas

Conforme `docs/RANKING.md`.

## 10. Bots criados por outras IAs

Cada IA recebe apenas `AGENTS.md`, `docs/BOTS.md` e `docs/REGRAS.md`. Ferramentas que leem `AGENTS.md` sozinhas (Codex, Cursor, Copilot etc.) já o recebem ao abrir o repositório; nas de chat, cole os três arquivos.

## 11. Partidas salvas em disco

As partidas do servidor (marco 5) vivem só em memória. Gravar o registro de cada partida para que sobreviva a um reinício e possa ser revista depois (`PAR-05`, glossário: a partida "fica salva").

## 12. Fechamento do tabuleiro

Para reduzir empates, a partir de `turno_fechamento` o tabuleiro fecha um anel por turno, de fora para dentro, e quem está no anel morre (FEC-01 a FEC-07). Feito fora da ordem, depois do marco 5, porque quase todas as partidas terminavam empatadas.

## Versão 2 (depois de tudo acima)

- `Bot.Planejar` recebe um terceiro parâmetro, `Historico`, para estudar padrões dos adversários.
- Suporte a jogadores humanos (aí o prazo por turno e o cronômetro voltam a fazer sentido).
- Power-ups: mais bombas por turno, mais potência, pavio diferente e mais ações por turno.
