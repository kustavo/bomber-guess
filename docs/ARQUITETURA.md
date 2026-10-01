# Arquitetura

As regras do jogo estão em `docs/REGRAS.md`. Este arquivo trata de como o código se organiza em torno delas.

## 1. Núcleo do jogo (pacote `internal/jogo`)

Pacote puro: sem API, sem banco, sem fila, sem relógio. Recebe dados, devolve dados.

### 1.1 Validador (`Validar`)

`Validar(estado Estado, plano Plano) (Plano, []Infracao)`

Verifica apenas o que pode ser checado a partir do estado do início do turno:

- **VAL-01** O plano tem no máximo `acoes_por_turno` ações, com tipos e direções válidos (ACA-01, ACA-02).
- **VAL-02** Nenhum movimento sai do tabuleiro ou entra em bloco fixo (MOV-02). A posição é simulada ação a ação a partir da posição do jogador no início do turno.
- **VAL-03** A quantidade de ações `PLANTAR` não passa de `bombas_por_turno` (BOM-02).
- **VAL-04** Blocos destrutíveis **não** são verificados aqui, porque dependem das ações dos outros jogadores.
- **VAL-05** Tratamento de erro: a primeira ação inválida (infração) e todas as seguintes viram `ESPERAR`. Cada plano gera no máximo uma infração.
- **VAL-06** A i-ésima ação do plano tem `etapa` = i (1, 2, 3…). A primeira ação fora dessa sequência é uma infração.
- **VAL-07** O `turno` do plano é igual ao `turno` do estado. Caso contrário, o plano inteiro é descartado, com uma infração na etapa 1.
- **VAL-08** O plano de um jogador inexistente no estado é descartado, com uma infração. O plano de um jogador morto é descartado sem infração.

Cada infração (`Infracao`) registra o jogador, a posição da ação rejeitada no plano (`etapa`; 1 para VAL-07 e VAL-08), a regra violada e um motivo legível:

```json
{"jogador_id": "jogador_1", "etapa": 4, "regra": "VAL-03", "motivo": "3ª bomba no turno, limite 2"}
```

### 1.2 Resolução do turno (`ResolverTurno`)

`ResolverTurno(estado Estado, planos []Plano) (Estado, []RelatorioEtapa)`

Recebe os planos validados de todos os jogadores e o `Estado`, resolve cada etapa na ordem definida em `docs/REGRAS.md` (seção 6) e produz:

1. O **relatório de cada etapa** (`RelatorioEtapa`: posições, bombas, explosões, mortes, blocos destruídos, movimentos bloqueados), usado no replay.
2. O novo `Estado` para o turno seguinte.

- **RES-01** Jogador vivo sem plano executa `ESPERAR` em todas as etapas. Planos de jogador inexistente ou morto são ignorados; se houver mais de um plano para o mesmo jogador, vale o primeiro.
- **RES-02** `ResolverTurno` é robusto a planos não validados: uma ação impossível (movimento para fora do tabuleiro ou para bloco fixo, bomba além de `bombas_por_turno`, tipo inválido) é executada como `ESPERAR`, sem abortar o resto do plano e sem gerar infração.
- **RES-03** O resultado não depende da ordem dos planos recebidos.
- **RES-04** O fim da partida segue `DEC-06` e `DEC-07` de `docs/REGRAS.md`; o relatório de cada etapa segue `DEC-09`.

O fim é consultado com `VerificarFim(estado Estado) Desfecho`, que devolve `terminada`, `empate`, `vencedor` (id; ausente em empate ou partida em andamento) e `sobreviventes` (ids dos vivos).

#### Exemplo de relatório de etapa

Em cada jogador, `acao` é a ação do plano para a etapa e `resultado` diz o que aconteceu com ela (DEC-09): `EXECUTADA`, `BLOQUEADA` (MOV-05), `ABORTADA` (etapa depois de um bloqueio, executada como `ESPERAR`), `DESCARTADA` (jogador morto) ou `IGNORADA` (ação impossível, RES-02). `jogadores` traz todos, na ordem do `Estado`; as listas de posições vêm ordenadas por `y` e depois por `x`.

```json
{
  "turno": 5,
  "etapa": 5,
  "jogadores": [
    {"id": "jogador_1", "posicao": {"x": 6, "y": 5}, "status": "VIVO",
     "acao": {"etapa": 5, "tipo": "MOVER", "direcao": "CIMA"}, "resultado": "EXECUTADA"},
    {"id": "jogador_2", "posicao": {"x": 4, "y": 6}, "status": "MORTO",
     "acao": {"etapa": 5, "tipo": "ESPERAR"}, "resultado": "EXECUTADA"},
    {"id": "jogador_3", "posicao": {"x": 8, "y": 4}, "status": "VIVO",
     "acao": {"etapa": 5, "tipo": "MOVER", "direcao": "DIREITA"}, "resultado": "BLOQUEADA"},
    {"id": "jogador_4", "posicao": {"x": 12, "y": 10}, "status": "MORTO",
     "acao": {"etapa": 5, "tipo": "ESPERAR"}, "resultado": "DESCARTADA"}
  ],
  "bombas": [
    {"posicao": {"x": 2, "y": 0}, "jogador_id": "jogador_1", "potencia": 2, "pavio_restante": 2}
  ],
  "explosoes": [
    {"origem": {"x": 5, "y": 6}, "potencia": 2,
     "chamas": [{"x": 3, "y": 6}, {"x": 4, "y": 6}, {"x": 5, "y": 6}, {"x": 6, "y": 6}, {"x": 7, "y": 6}]}
  ],
  "chamas": [{"x": 3, "y": 6}, {"x": 4, "y": 6}, {"x": 5, "y": 6}, {"x": 6, "y": 6}, {"x": 7, "y": 6}],
  "mortes": ["jogador_2"],
  "blocos_destruidos": [{"x": 7, "y": 6}],
  "movimentos_bloqueados": ["jogador_3"]
}
```

### 1.3 Registro de ações

Para cada jogador e turno são gravadas três versões das ações:

1. **Planejadas**: saída bruta do bot.
2. **Validadas**: após o validador.
3. **Executadas**: o que de fato aconteceu após a resolução do turno.

Isso permite medir, por exemplo, quantas infrações cada IA comete e quantas vezes ela errou uma previsão.

## 2. Loop da partida (pacote `internal/partida`)

A partida roda **em tempo real** no servidor. Cada partida tem seu próprio loop de jogo, e quem abre a partida assiste ao estado atual, não a um replay.

Cada turno tem duas fases:

1. **Planejamento**: o servidor abre o turno com prazo de `prazo_planejamento_ms`. Os jogadores enviam seus planos, que vão para a fila. Os bots rodam nessa fase como qualquer outro jogador; na versão 2, humanos enviam via API.
2. **Execução**: ao fim do prazo, o servidor consome a fila, passa as ações por `Validar` e `ResolverTurno` e libera as etapas uma a uma, a cada `duracao_etapa_ms`.

Jogador que não enviar o plano dentro do prazo fica com todas as etapas em `ESPERAR`.

**Registro**: tudo o que acontece é gravado (ações nas três versões e resultado de cada etapa), então é possível rever partidas antigas, mas isso é um recurso extra, não o modo principal.

## 3. Fila (pacote `internal/fila`)

A fila fica atrás de uma interface. Primeira implementação em memória; depois Kafka, sem mudar quem usa a interface.

Tópicos Kafka:

- `planos-enviados`: plano de cada jogador durante o planejamento (a fila do turno).
- `turno-resolvido`: resultado de cada turno.
- `partida-finalizada`: consumido pelo serviço de ranking e estatísticas.

## 4. API HTTP (pacote `internal/api`)

Sem WebSockets. O frontend consulta o estado periodicamente.

**Sincronização do cronômetro**: toda resposta da API inclui o horário do servidor, a fase atual e o horário de término dessa fase. O frontend calcula a diferença entre seu relógio e o do servidor e exibe o cronômetro a partir disso. Se a página for recarregada, basta consultar o endpoint de estado para se ressincronizar.

Endpoints:

- `POST /mapas`: salva um mapa criado no editor.
- `POST /partidas`: cria uma partida (nome, mapa, bots em cada posição) e inicia o loop.
- `GET /partidas/{nome}/estado`: estado atual, fase, etapa em execução, horário do servidor e fim da fase.
- `POST /partidas/{nome}/turnos/{n}/plano`: envia o plano de um jogador (usado pelos humanos na v2).
- `GET /partidas/{nome}/historico`: registro completo, para rever partidas encerradas.
- `GET /ranking` e `GET /bots`.
