# Marco 02: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

Tudo continua no pacote puro `internal/jogo`, sem dependências novas. `Validar` percorre o plano uma vez, simulando a posição e contando bombas, e para na primeira infração. `ResolverTurno` copia o estado para uma estrutura de trabalho interna e executa as etapas, cada uma em seis fases que seguem ORD-01 a ORD-06 na ordem. Depois de cada etapa ela gera um `RelatorioEtapa` e verifica se a partida acabou (DEC-06). As explosões são calculadas em uma função separada: ela parte das casas com pavio 0 e propaga as reações em cadeia em largura. As bombas que servem de obstáculo são fixadas no início da fase de explosões (DEC-05). As listas de saída são ordenadas por posição, de modo que o resultado não depende da ordem das entradas (RES-03).

Os testes montam tabuleiros pequenos com um helper que lê desenhos em ASCII. Assim cada caso da tabela cabe em poucas linhas, e o desenho mostra a situação testada.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/internal/jogo/validar.go` | `Infracao`, `Validar`. |
| `backend/internal/jogo/relatorio.go` | `RelatorioEtapa`, `JogadorEtapa`, `ResultadoAcao` e constantes, `Explosao`. |
| `backend/internal/jogo/resolver.go` | `ResolverTurno` e a estrutura interna de trabalho (`mesa`), com uma função por fase (ORD-01 a ORD-06). |
| `backend/internal/jogo/explosao.go` | Cálculo interno de pilhas, alcance e reação em cadeia (BOM-04, BOM-06 a BOM-09, DEC-05). |
| `backend/internal/jogo/fim.go` | `Desfecho`, `VerificarFim` (FIM-02 a FIM-04, DEC-06, DEC-07). |
| `backend/internal/jogo/mapa.go` | `VerificarMapa` passa a exigir `limite_turnos` ≥ 1 e os atributos de `jogador_padrao` ≥ 1 (MAP-05). |
| `backend/internal/jogo/tabuleiro_test.go` | Helper de teste `montar(t, desenho, ...)`: monta um `Estado` a partir de um desenho em ASCII. |
| `backend/internal/jogo/validar_test.go` | CA-01 a CA-15. |
| `backend/internal/jogo/movimento_test.go` | CA-16 a CA-20. |
| `backend/internal/jogo/bomba_test.go` | CA-21 a CA-25. |
| `backend/internal/jogo/explosao_test.go` | CA-26 a CA-35. |
| `backend/internal/jogo/etapa_test.go` | CA-36 a CA-39 (ordem da etapa e mortes). |
| `backend/internal/jogo/fim_test.go` | CA-40 a CA-44. |
| `backend/internal/jogo/resolver_test.go` | CA-45 a CA-48 (relatório, robustez, pureza) e ida e volta dos exemplos JSON de `Infracao` e `RelatorioEtapa` de `docs/ARQUITETURA.md`. |
| `backend/internal/jogo/mapa_test.go` | Casos novos de CA-49. |

## Tipos e assinaturas

```go
// validar.go

// Infracao é uma ação rejeitada pelo validador (VAL-05). Cada plano gera
// no máximo uma.
type Infracao struct {
    JogadorID string `json:"jogador_id"`
    Etapa     int    `json:"etapa"`  // posição (1, 2, 3…) da ação rejeitada; 1 para VAL-07 e VAL-08
    Regra     string `json:"regra"`  // "VAL-01" … "VAL-08"
    Motivo    string `json:"motivo"` // texto para humanos
}

// Validar devolve o plano executável e as infrações (VAL-01 a VAL-08, DEC-08).
// Não altera o estado nem o plano recebidos.
func Validar(estado Estado, plano Plano) (Plano, []Infracao)

// relatorio.go

// ResultadoAcao diz o que aconteceu com a ação planejada de um jogador em uma etapa (DEC-09).
type ResultadoAcao string

const (
    Executada  ResultadoAcao = "EXECUTADA"  // a ação aconteceu como planejada
    Bloqueada  ResultadoAcao = "BLOQUEADA"  // MOV-05: ficou parado; o resto do plano foi abortado
    Abortada   ResultadoAcao = "ABORTADA"   // MOV-05: etapa depois de um bloqueio; executada como ESPERAR
    Descartada ResultadoAcao = "DESCARTADA" // FIM-01: jogador morto; nada executado
    Ignorada   ResultadoAcao = "IGNORADA"   // RES-02: ação impossível; executada como ESPERAR
)

// JogadorEtapa é a situação de um jogador ao fim de uma etapa.
type JogadorEtapa struct {
    ID        string        `json:"id"`
    Posicao   Posicao       `json:"posicao"`
    Status    Status        `json:"status"`
    Acao      Acao          `json:"acao"`      // a ação do plano para esta etapa (ESPERAR se não havia)
    Resultado ResultadoAcao `json:"resultado"`
}

// Explosao é a explosão de uma pilha (BOM-04, BOM-06).
type Explosao struct {
    Origem   Posicao   `json:"origem"`
    Potencia int       `json:"potencia"` // já considerando a pilha
    Chamas   []Posicao `json:"chamas"`   // inclui a origem
}

// RelatorioEtapa é o que aconteceu em uma etapa, usado no replay e no registro.
type RelatorioEtapa struct {
    Turno                int            `json:"turno"`
    Etapa                int            `json:"etapa"`
    Jogadores            []JogadorEtapa `json:"jogadores"`             // todos, na ordem do Estado
    Bombas               []Bomba        `json:"bombas"`                // no tabuleiro ao fim da etapa
    Explosoes            []Explosao     `json:"explosoes"`
    Chamas               []Posicao      `json:"chamas"`                // união das chamas, sem repetição
    Mortes               []string       `json:"mortes"`                // ids dos que morreram nesta etapa
    BlocosDestruidos     []Posicao      `json:"blocos_destruidos"`     // sem repetição (DEC-01)
    MovimentosBloqueados []string       `json:"movimentos_bloqueados"` // ids
}

// resolver.go

// ResolverTurno resolve o turno etapa a etapa (ORD-01 a ORD-06) e devolve o
// estado do turno seguinte e um relatório por etapa executada. Para no fim
// da etapa em que a partida termina (DEC-06). Para uma partida já terminada,
// devolve uma cópia do estado e nenhum relatório. Não altera as entradas.
func ResolverTurno(estado Estado, planos []Plano) (Estado, []RelatorioEtapa)

// fim.go

// Desfecho é a situação da partida deduzida de um Estado (DEC-06).
type Desfecho struct {
    Terminada     bool     `json:"terminada"`
    Empate        bool     `json:"empate"`
    Vencedor      string   `json:"vencedor,omitempty"` // id; vazio em empate ou partida em andamento
    Sobreviventes []string `json:"sobreviventes"`      // ids dos vivos
}

// VerificarFim deduz o desfecho: 1 vivo → vitória (FIM-02); 0 vivos → empate
// (FIM-03); 2+ vivos com turno > limite_turnos → empate (FIM-04, DEC-07).
func VerificarFim(estado Estado) Desfecho
```

Estrutura interna (não exportada) de `resolver.go`, para orientar:

```go
type mesa struct {
    turno, etapa       int
    config             Config
    fixos              map[Posicao]bool
    destrutiveis       map[Posicao]bool
    bombas             []bombaViva // Bomba + plantadaNaEtapa int
    jogadores          []jogadorVivo // Jogador + acoes []Acao, abortado bool, plantadas int
}
```

## Decisões

- **D1** `Validar` simula na seguinte ordem para cada ação i: limite de ações (VAL-01), `etapa` = i (VAL-06), tipo e direção (VAL-01), destino (VAL-02), bombas (VAL-03). Antes do laço vêm VAL-07 e VAL-08. **Motivo**: quando uma ação viola mais de uma regra, a regra citada fica determinística.
- **D2** No plano validado, as ações além de `acoes_por_turno` são **removidas**, e as demais ações inválidas viram `ESPERAR` com `etapa` renumerada. **Motivo**: é o que pedem CA-01 e CA-15; o preenchimento até `etapas_neste_turno` (ACA-03) é feito pelo `ResolverTurno`, que trata todos os jogadores do mesmo jeito.
- **D3** `ResolverTurno` ignora os campos `etapa` e `turno` das ações e dos planos e usa a posição da ação na lista; ações além de `acoes_por_turno` são descartadas. **Motivo**: RES-02. Quem garante a consistência é `Validar`.
- **D4** `ResolverTurno` recalcula `etapas_neste_turno` com `CalcularEtapas` em vez de confiar no campo recebido. **Motivo**: robustez (RES-02). Pela EST-03, o valor é o mesmo.
- **D5** O novo estado sempre tem `turno` + 1, mesmo quando a partida termina no meio do turno. **Motivo**: `VerificarFim` fica simples e a FIM-04 vira `turno > limite_turnos`. Para uma partida já terminada, nada muda (CA-43).
- **D6** Explosões: cada etapa fixa o conjunto de casas com bomba logo depois de ORD-03. Uma fila parte das casas com alguma bomba de pavio ≤ 0. Cada casa explode uma vez, com potência `max + (n − 1)` sobre **todas** as bombas da casa. Uma casa com bomba atingida pelo fogo entra na fila. Ao fim, todas as bombas das casas que explodiram são removidas. **Motivo**: BOM-04, BOM-09 e DEC-05. O resultado é um conjunto, o que garante CA-31.
- **D7** Ordem das listas de saída: `Jogadores` e `Mortes` na ordem de `Estado.Jogadores`; `Chamas`, `BlocosDestruidos`, `Explosoes` (pela origem) e `Bombas` (posição, depois `jogador_id`, depois pavio) ordenadas por (y, x). No novo `Estado`, os blocos mantêm a ordem original. **Motivo**: RES-03 e determinismo, com saída estável para os testes.
- **D8** Os planos são indexados por `jogador_id`; vale o primeiro de cada id, e ids inexistentes ou mortos são ignorados (RES-01). Como o processamento segue `Estado.Jogadores`, a ordem de `planos` só importa entre duplicatas. **Motivo**: RES-03.
- **D9** O relatório inclui os jogadores mortos em turnos anteriores (ação `ESPERAR`, resultado `DESCARTADA`). **Motivo**: o frontend desenha cada etapa só a partir do relatório.
- **D10** Uma ação inválida dentro do `ResolverTurno` (destino fora do tabuleiro ou bloco fixo, bomba além do limite, tipo ou direção desconhecidos) recebe resultado `IGNORADA` e não aborta o resto do plano. **Motivo**: RES-02 e CA-47.
- **D11** Helper de teste `montar(t, desenho string, opcoes...)`: `.` livre, `#` bloco fixo, `+` bloco destrutível, `1`–`9` jogador `jogador_N` vivo com atributos padrão (bombas 2, potência 2, pavio 3, ações 7). As bombas e os atributos diferentes entram por funções de opção. **Motivo**: deixar legíveis os cerca de 50 casos de tabela.

## Mudanças em `docs/` (aprovadas e aplicadas)

- **P1** Glossário de `docs/REGRAS.md`: **Relatório de etapa** e **Desfecho**.
- **P2** `docs/ARQUITETURA.md` 1.1 e 1.2: exemplos JSON de `Infracao` e `RelatorioEtapa` e a descrição de `VerificarFim`. Os testes fazem a ida e volta desses exemplos, como no marco 1.
- **P3** `docs/EDITOR.md` `MAP-05`: `limite_turnos` < 1 torna o mapa inválido.

## Riscos

- **Reação em cadeia com pilhas e bombas recém-plantadas**: é a parte mais sujeita a erro. Mitigação: casos de tabela dedicados (CA-28 a CA-31) e um teste que embaralha a ordem das bombas.
- **`limite_turnos` ≤ 0 em estado montado à mão**: o mapa agora proíbe (P3); um estado assim seria tratado como partida terminada por `VerificarFim`, o que é aceitável.
- **Pavio ≤ 0 vindo de um estado montado à mão**: pela D6, a bomba explode na primeira etapa. É um comportamento aceitável e não precisa de caso especial.
- **Tamanho de `resolver.go`**: dividido por fase em funções pequenas para não virar um bloco monolítico.
