# Ranking e estatísticas

- Cada partida finalizada (evento `partida-finalizada`) atualiza o ranking das versões de bot.
- **[SUPOSIÇÃO]** Pontuação: vitória vale 3 pontos, empate 1 e derrota 0. Como critério de desempate, conta quantos turnos o bot sobreviveu.
- Estatísticas extras por bot, calculadas a partir do registro de ações (planejadas, validadas, executadas):
  - taxa de infrações;
  - movimentos bloqueados;
  - bombas plantadas;
  - oponentes eliminados.
- Recomendação: para comparar IAs de forma justa, rodar várias partidas no mesmo mapa trocando os bots de posição inicial, já que posição dá vantagem e bots determinísticos repetem a mesma partida.

Endpoints: `GET /ranking` e `GET /bots` (ver `docs/ARQUITETURA.md`).
