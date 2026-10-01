# Regras do jogo

## Como ler este documento

Cada regra tem um **ID estável** (ex.: `BOM-05`). Specs, testes e commits citam as regras por esse ID; o nome de cada caso de teste começa com o ID que ele cobre.

- IDs nunca são renumerados nem reaproveitados. Regra removida fica como `~~BOM-05~~ removida: <motivo>`.
- Regra nova recebe o próximo número livre da seção, mesmo que fique fora de ordem.
- Mudou uma regra? Atualize aqui **antes** do código (ver `specs/README.md`).

## Glossário

- **Mapa**: layout do tabuleiro (tamanho, blocos e posições iniciais).
- **Partida**: sequência de turnos até haver um vencedor ou empate. Tem um nome único, é jogada em um mapa com um conjunto de bots, fica salva e pode ser assistida a qualquer momento.
- **Turno**: ciclo completo de planejamento (cada jogador envia seu plano) e execução (ações são resolvidas).
- **Etapa**: subdivisão de tempo dentro de um turno. Cada jogador executa no máximo uma ação por etapa.
- **Ação**: o que um jogador faz em uma etapa (`MOVER`, `PLANTAR` bomba ou `ESPERAR`).
- **Bloco fixo**: casa bloqueada permanentemente. Não pode ser atravessada nem destruída.
- **Bloco destrutível**: casa bloqueada que é destruída quando atingida por uma explosão.
- **Potência**: quantas casas a explosão alcança em cada direção.
- **Pavio**: quantas etapas faltam para a bomba explodir.
- **Pilha**: conjunto de bombas na mesma casa; explode como uma única explosão.
- **Chama**: casa atingida por uma explosão na etapa em que ela ocorre.
- **Reação em cadeia**: bomba atingida por uma chama explode na mesma etapa.
- **Estado** (`Estado`): foto completa da partida no início de um turno (seção 2).
- **Plano** (`Plano`): lista de ações que um jogador envia para um turno (seção 3).
- **Movimento bloqueado**: movimento para uma casa que ainda tem bloco destrutível na hora da execução; aborta o restante do plano.
- **Infração**: ação rejeitada pelo validador; conta contra o bot no ranking.
- **Posição inicial**: casa onde um jogador começa a partida, definida no mapa.
- **Bot**: jogador controlado por uma implementação da interface `Bot`.
- **Catálogo de bots**: lista das versões de bot disponíveis (`GET /bots`).
- **Relatório de etapa** (`RelatorioEtapa`): o que aconteceu em uma etapa: posições, ações executadas, bombas, explosões, mortes e blocos destruídos. Usado no replay e no registro.
- **Fechamento**: a partir de `turno_fechamento`, ao fim de cada turno o anel mais externo ainda aberto do tabuleiro vira bloco fixo, matando quem estiver nele (FEC-01 a FEC-07). Existe para evitar empates.
- **Anel**: conjunto das casas a uma mesma distância da borda; o anel 0 é a borda (FEC-02).
- **Desfecho** (`Desfecho`): situação da partida deduzida do estado: em andamento, vitória de um jogador ou empate (DEC-06).

## 1. Tabuleiro e coordenadas

- **TAB-01** O tabuleiro é uma grade retangular de casas quadradas.
- **TAB-02** Posições são representadas como `{"x": 12, "y": 8}`.
- **TAB-03** A origem `(0, 0)` fica no canto superior esquerdo; `x` cresce para a direita e `y` cresce para baixo.
- **TAB-04** Casas vizinhas são apenas as ortogonais (cima, baixo, esquerda, direita).

## 2. Estado do jogo (Estado)

No início de cada turno, o estado completo é representado assim:

```json
{
  "turno": 5,
  "config": {
      "largura": 15,
      "altura": 13,
      "limite_turnos": 50,
      "turno_fechamento": 30,
      "prazo_planejamento_ms": 1000,
      "duracao_etapa_ms": 5
  },
  "etapas_neste_turno": 7,
  "blocos_fixos": [{"x": 1, "y": 1}, {"x": 3, "y": 1}],
  "blocos_destrutiveis": [{"x": 2, "y": 4}],
  "bombas": [
    {
      "posicao": {"x": 5, "y": 6},
      "jogador_id": "jogador_2",
      "potencia": 2,
      "pavio_restante": 3
    }
  ],
  "jogadores": [
    {
      "id": "jogador_1",
      "posicao": {"x": 0, "y": 0},
      "status": "VIVO",
      "bombas_por_turno": 2,
      "potencia": 2,
      "pavio_padrao": 3,
      "acoes_por_turno": 7,
      "bot_versao": "claude-v1"
    },
    {
      "id": "jogador_2",
      "posicao": {"x": 5, "y": 6},
      "status": "MORTO",
      "morte": {"turno": 4, "etapa": 6},
      "bombas_por_turno": 2,
      "potencia": 2,
      "pavio_padrao": 3,
      "acoes_por_turno": 7,
      "bot_versao": "aleatorio-v1"
    }
  ]
}
```

Detalhes dos campos:

- **EST-01** `prazo_planejamento_ms`: tempo que os jogadores têm para enviar as ações no início de cada turno.
- **EST-02** `duracao_etapa_ms`: intervalo entre uma etapa e a próxima durante a execução (usado pelo loop do servidor e pelo frontend na animação). Não afeta as regras.
- **EST-03** `etapas_neste_turno`: igual ao maior `acoes_por_turno` entre os jogadores vivos. Jogadores com menos ações ficam parados nas etapas restantes.
- **EST-04** `bombas_por_turno`: quantas bombas o jogador pode plantar em cada turno. O estoque é renovado no início de todo turno. Ele pode plantar várias bombas antes da primeira explodir, desde que respeite esse limite.
- **EST-05** `potencia` e `pavio_padrao`: valores aplicados às bombas que o jogador plantar. O pavio é medido em **etapas**.
- **EST-06** `status`: `VIVO` ou `MORTO`.
- **EST-07** `morte`: turno e etapa em que o jogador morreu. Ausente enquanto ele está vivo.
- **EST-08** `posicao` de um jogador morto é a casa onde ele morreu; ele não ocupa mais o tabuleiro (FIM-01).
- **EST-09** `turno_fechamento`: turno a partir do qual o tabuleiro começa a fechar (FEC-01). Opcional; ausente ou `0` desliga o fechamento.

## 3. Ações

| Tipo | Efeito |
|---|---|
| `MOVER` | Move para uma casa vizinha (`CIMA`, `BAIXO`, `ESQUERDA`, `DIREITA`) |
| `PLANTAR` | Planta uma bomba na casa onde o jogador está |
| `ESPERAR` | Não faz nada |

Formato do **plano** de um jogador para um turno:

```json
{
  "jogador_id": "jogador_1",
  "turno": 5,
  "acoes": [
    {"etapa": 1, "tipo": "MOVER", "direcao": "BAIXO"},
    {"etapa": 2, "tipo": "MOVER", "direcao": "BAIXO"},
    {"etapa": 3, "tipo": "PLANTAR"},
    {"etapa": 4, "tipo": "MOVER", "direcao": "ESQUERDA"},
    {"etapa": 5, "tipo": "ESPERAR"}
  ]
}
```

- **ACA-01** Cada ação é de um dos tipos da tabela acima; `MOVER` exige `direcao`.
- **ACA-02** O plano pode ter no máximo `acoes_por_turno` ações.
- **ACA-03** Se o plano tiver menos ações que `etapas_neste_turno`, as etapas que faltam são preenchidas com `ESPERAR`.

## 4. Regras de movimento

- **MOV-01** Só é permitido mover para uma casa vizinha ortogonal.
- **MOV-02** Não é permitido sair do tabuleiro nem entrar em bloco fixo.
- **MOV-03** Entrar em bloco destrutível só funciona se ele já tiver sido destruído em uma etapa anterior (inclusive neste mesmo turno). Um bot pode planejar passar por um bloco apostando que alguém vai destruí-lo antes.
- **MOV-04** É permitido entrar em uma casa ocupada por outro jogador ou por uma bomba.
- **MOV-05** **Movimento bloqueado**: se na hora da execução a casa de destino ainda tiver um bloco destrutível, o jogador fica parado naquela etapa e **o restante do seu plano é abortado** (viram `ESPERAR`).

## 5. Regras de bombas e explosões

- **BOM-01** A bomba é plantada na casa atual do jogador, com a `potencia` e o `pavio_padrao` que ele possui naquele momento.
- **BOM-02** O jogador não pode plantar mais bombas por turno do que seu `bombas_por_turno`.
- **BOM-03** Pode haver várias bombas na mesma casa, inclusive do mesmo jogador.
- **BOM-04** **Bombas empilhadas**: a potência da explosão é a maior potência entre as bombas da casa, mais 1 casa para cada bomba adicional. Ex.: bombas de potência 2 e 3 na mesma casa geram uma explosão de potência 4; três bombas de potência 2 geram potência 4.
   - Quando qualquer bomba da pilha explode, a pilha inteira explode junto, como uma única explosão.
- **BOM-05** O pavio **não** diminui na etapa em que a bomba foi plantada. A partir da etapa seguinte, diminui 1 por etapa. Quando chega a 0, a bomba explode. Ex.: pavio 3 plantado na etapa 2 explode na etapa 5.
- **BOM-06** A explosão se espalha em cruz, alcançando até `potencia` casas em cada direção, além da própria casa da bomba.
- **BOM-07** Em cada direção, a explosão para ao encontrar:
   - um **bloco fixo**: não o atinge;
   - um **bloco destrutível**: destrói esse bloco e para ali;
   - uma **bomba**: atinge a casa da bomba e para ali.
- **BOM-08** A explosão atravessa jogadores (pode matar vários na mesma linha).
- **BOM-09** Reação em cadeia: uma bomba atingida por uma explosão explode na mesma etapa, com a **sua própria potência** (ou a potência da pilha, conforme BOM-04), a partir da sua casa. O fogo da primeira explosão não continua além dela.
- **BOM-10** O fogo da explosão dura apenas a etapa em que ocorre.
- **BOM-11** Bombas que não explodiram continuam no tabuleiro no turno seguinte, com o pavio restante preservado.
- **BOM-12** Bombas de um jogador morto continuam ativas e explodem normalmente.

## 6. Ordem de resolução de cada etapa

`ResolverTurno` deve seguir esta ordem rigorosamente:

- **ORD-01** **Movimentos**: todos são aplicados ao mesmo tempo. Movimentos para uma casa com bloco destrutível são bloqueados, e o restante do plano daquele jogador é abortado.
- **ORD-02** **Bombas**: todas as bombas plantadas na etapa são colocadas.
- **ORD-03** **Pavio**: diminui em 1 em todas as bombas, **exceto** as plantadas nesta etapa.
- **ORD-04** **Explosões**: bombas com pavio 0 explodem, considerando pilhas e reações em cadeia.
- **ORD-05** **Mortes**: jogadores em casas atingidas morrem.
- **ORD-06** **Blocos**: blocos destrutíveis atingidos são removidos e ficam transitáveis a partir da próxima etapa.
- **ORD-07** Consequência prática: um jogador que sai do alcance da explosão na mesma etapa em que ela acontece sobrevive, porque o movimento é resolvido antes.
- **ORD-08** **Fechamento**: depois da última etapa do turno (após ORD-06), aplica-se o fechamento do tabuleiro, se houver (FEC-03).

## 7. Morte, vitória e empate

- **FIM-01** Um jogador atingido tem `status` alterado para `MORTO`, com o turno e a etapa da morte registrados em `morte` (EST-07). Suas ações seguintes são descartadas e ele sai do tabuleiro.
- **FIM-02** Vence o último jogador vivo.
- **FIM-03** Se todos os jogadores restantes morrerem na mesma etapa, a partida termina em empate.
- **FIM-04** Ao atingir `limite_turnos`, a partida termina em empate entre os sobreviventes.

### Fechamento do tabuleiro

Para evitar empates, o tabuleiro fecha de fora para dentro a partir de um turno configurado.

- **FEC-01** O fechamento é ligado por `turno_fechamento` (EST-09) com valor ≥ 1. Ausente ou `0`, nada desta seção se aplica.
- **FEC-02** **Anel** `n` é o conjunto das casas `(x, y)` com `min(x, y, largura − 1 − x, altura − 1 − y) = n`. O anel 0 é a borda do tabuleiro; o anel 1 é a borda do que sobra dentro dela; e assim por diante.
- **FEC-03** Ao fim de cada turno `T ≥ turno_fechamento`, depois da última etapa executada (ORD-08), o anel `T − turno_fechamento` vira bloco fixo: toda casa desse anel passa a estar em `blocos_fixos`. Ex.: com `turno_fechamento` 30, a borda fecha ao fim do turno 30, o anel 1 ao fim do turno 31. Quando o anel não tem mais casas (o tabuleiro já fechou por completo), nada acontece.
- **FEC-04** Jogador vivo em uma casa que fecha morre. A morte é registrada no turno `T` e na última etapa executada desse turno (EST-07). Se com isso todos os restantes morrerem, é empate (FIM-03).
- **FEC-05** Bloco destrutível em uma casa que fecha deixa de existir e vira bloco fixo; não conta como bloco destruído.
- **FEC-06** Bombas em uma casa que fecha são removidas sem explodir.
- **FEC-07** Se a partida terminou no meio do turno (DEC-06), não há fechamento nesse turno.

## 8. Decisões

Casos que a especificação original não definia. Cada item começa como **[PROPOSTA]** e vira **[DECIDIDO]** quando confirmado. DEC-01 a DEC-09 foram decididas na spec do marco 2 (`specs/02-resolucao/spec.md`).

- **DEC-01** [DECIDIDO] **Bloco destrutível atingido por duas explosões na mesma etapa**: conta como um único bloco destruído (importa para estatísticas).
- **DEC-02** [DECIDIDO] **Morte registrada em qual turno/etapa quando o jogador morre por bomba de turno anterior**: o turno e a etapa atuais.
- **DEC-03** [DECIDIDO] **Jogador morto**: deixa de contar para `etapas_neste_turno` já no turno seguinte.
- **DEC-04** [DECIDIDO] **Bombas restantes após o fim da partida**: são ignoradas.
- **DEC-05** [DECIDIDO] **Bombas que explodem na mesma etapa**: toda bomba presente no início da fase de explosões (ORD-04) para o fogo das outras (BOM-07), mesmo que ela também exploda nessa etapa. O resultado não depende da ordem em que as bombas são processadas.
- **DEC-06** [DECIDIDO] **Fim no meio do turno**: quando, ao fim de uma etapa, restar no máximo um jogador vivo, a partida termina ali e as etapas seguintes não são executadas. O fim é deduzido do estado (jogadores vivos e `turno`), sem campo próprio no JSON. Resolver o turno de uma partida terminada não faz nada.
- **DEC-07** [DECIDIDO] **Limite de turnos**: o turno de número `limite_turnos` é jogado; se ao fim dele restarem 2 ou mais jogadores vivos, a partida termina em empate entre eles (FIM-04).
- **DEC-08** [DECIDIDO] **Direção em ação que não é `MOVER`**: é ignorada e removida pelo validador, sem infração.
- **DEC-09** [DECIDIDO] **Ações executadas**: o relatório de cada etapa traz, para cada jogador, a ação de fato executada, distinguindo movimento bloqueado (MOV-05), ação abortada e ação descartada por morte (FIM-01). É dele que sai a terceira versão das ações (`docs/ARQUITETURA.md`, seção 1.3).
