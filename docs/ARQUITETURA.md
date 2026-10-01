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
- **VAL-05** Tratamento de erro: a primeira ação inválida (infração) e todas as seguintes viram `ESPERAR`.

### 1.2 Resolução do turno (`ResolverTurno`)

`ResolverTurno(estado Estado, planos []Plano) (Estado, []RelatorioEtapa)`

Recebe os planos validados de todos os jogadores e o `Estado`, resolve cada etapa na ordem definida em `docs/REGRAS.md` (seção 6) e produz:

1. O **relatório de cada etapa** (`RelatorioEtapa`: posições, bombas, explosões, mortes, blocos destruídos, movimentos bloqueados), usado no replay.
2. O novo `Estado` para o turno seguinte.

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
