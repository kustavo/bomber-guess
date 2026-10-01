# Editor de mapas e partidas

- Ao criar uma partida, o usuário vê um mapa vazio e um editor no estilo "maker" com os itens: bloco fixo, bloco destrutível e posição inicial.
- Para cada posição inicial, o usuário escolhe qual versão de bot vai jogar ali. A lista vem do catálogo de bots (`GET /bots`).
- O usuário dá um nome à partida e a inicia (`POST /mapas` e `POST /partidas`).
- Pelo nome é possível assistir à partida a qualquer momento.

O mapa salvo segue o mesmo formato dos arquivos em `mapas/`. Coordenadas conforme `docs/REGRAS.md`, seção 1.

## Formato do mapa

```json
{
  "nome": "exemplo",
  "config": {
    "largura": 15,
    "altura": 13,
    "limite_turnos": 50,
    "turno_fechamento": 30,
    "prazo_planejamento_ms": 1000,
    "duracao_etapa_ms": 5
  },
  "jogador_padrao": {
    "bombas_por_turno": 2,
    "potencia": 2,
    "pavio_padrao": 3,
    "acoes_por_turno": 7
  },
  "blocos_fixos": [{"x": 1, "y": 1}],
  "blocos_destrutiveis": [{"x": 2, "y": 4}],
  "posicoes_iniciais": [{"x": 0, "y": 0}, {"x": 14, "y": 12}]
}
```

- **MAP-01** `config` tem o mesmo formato do `config` do `Estado` e é copiado para ele.
- **MAP-02** `jogador_padrao` define os atributos iniciais de todos os jogadores.
- **MAP-03** O mapa não sabe quais bots jogam. O estado inicial é criado com `EstadoInicial(mapa, versoes)`: a versão `i` joga na posição inicial `i`, com id `jogador_<i+1>`.
- **MAP-04** O estado inicial começa no turno 1, sem bombas, com todos os jogadores `VIVO`.
- **MAP-05** Um mapa é inválido se: `largura` ou `altura` ≤ 0; algum bloco ou posição inicial está fora do tabuleiro; uma posição inicial está sobre um bloco; a mesma casa tem bloco fixo e destrutível; há menos de 2 posições iniciais; `limite_turnos` < 1; `turno_fechamento` < 0, ou ≥ 1 e ≥ `limite_turnos` (FEC-01); algum atributo de `jogador_padrao` (`bombas_por_turno`, `potencia`, `pavio_padrao`, `acoes_por_turno`) é < 1.
- **MAP-06** `EstadoInicial` também falha se a quantidade de versões for diferente da quantidade de posições iniciais.
