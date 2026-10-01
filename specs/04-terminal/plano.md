# Marco 04: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

São três peças, cada uma testável sozinha:

1. **`internal/bots`** (pacote novo, ao lado dos subpacotes de cada bot), com duas responsabilidades:
   - o **catálogo**: um mapa de versão para fábrica (`func(semente) jogo.Bot`), com listagem ordenada e criação por versão;
   - a **chamada protegida** `Chamar`: copia o estado (BOT-01), roda `Planejar` numa goroutine com `recover` e espera o resultado até o prazo (BOT-02).

   O marco 5 reaproveita as duas.
2. **Laço da partida**, em `cmd/terminal`: síncrono e sem relógio. Para cada turno, ele chama os bots vivos com `Chamar`, valida os planos com `jogo.Validar`, imprime os avisos, resolve com `jogo.ResolverTurno` e desenha cada relatório de etapa. Repete até `jogo.VerificarFim` indicar o fim e termina com a linha do desfecho.
3. **Desenho**, em `cmd/terminal`: um `quadro` guarda o tamanho do tabuleiro e os blocos e desenha o estado inicial e cada `RelatorioEtapa`. Depois de desenhar uma etapa, ele remove os blocos destruídos nela (CA-15, ORD-06).

O `main` só chama `rodar(args, saida, erros, catalogo) int`. Os testes chamam `rodar` direto, com um catálogo de bots de teste e mapas JSON gravados em `t.TempDir()`. Tudo o que pode dar erro de uso é verificado antes da primeira linha da saída padrão (CA-17, CA-18).

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/internal/bots/catalogo.go` | `Fabrica`, `Catalogo`, `Padrao`, `Versoes`, `Criar`. |
| `backend/internal/bots/chamar.go` | `Falha`, `Resposta`, `Chamar` (BOT-01, BOT-02). |
| `backend/internal/bots/catalogo_test.go` | CA-01, CA-02. |
| `backend/internal/bots/chamar_test.go` | Testes de unidade de `Chamar`: cópia do estado, prazo, `panic`, bot normal (base de CA-04 a CA-06). |
| `backend/cmd/terminal/main.go` | `main`, `rodar`, leitura dos argumentos e erros de uso (decisões 1, 6). |
| `backend/cmd/terminal/partida.go` | Laço da partida (`jogar`), avisos por turno e linha do desfecho. |
| `backend/cmd/terminal/desenho.go` | `quadro`: desenho do início e de cada etapa, com os eventos (decisões 3, 4). |
| `backend/cmd/terminal/apoio_test.go` | Bots de teste, catálogo de teste, helper que grava um mapa JSON e helper que roda `rodar` e captura as saídas. |
| `backend/cmd/terminal/desenho_test.go` | CA-14, CA-15. |
| `backend/cmd/terminal/partida_test.go` | CA-04 a CA-07, CA-09 a CA-12. |
| `backend/cmd/terminal/main_test.go` | CA-03, CA-17 a CA-19. |
| `backend/cmd/terminal/ponta_test.go` | CA-08, CA-13, CA-16, CA-20 e o mapa padrão (D8), de ponta a ponta no mapa de exemplo. |

## Tipos e assinaturas

```go
package bots // internal/bots

// Fabrica cria um bot com a semente dada.
type Fabrica func(semente uint64) jogo.Bot

// Catalogo associa cada versão de bot à sua fábrica.
type Catalogo map[string]Fabrica

// Padrao devolve o catálogo com todos os bots do projeto. Um bot novo entra
// aqui com uma linha.
func Padrao() Catalogo

// Versoes devolve as versões em ordem alfabética (CA-01).
func (c Catalogo) Versoes() []string

// Criar cria o bot da versão dada; erro que cita a versão se ela não existe (CA-02).
func (c Catalogo) Criar(versao string, semente uint64) (jogo.Bot, error)

// Falha diz por que a chamada ao bot não produziu um plano (BOT-02).
type Falha string

const (
    SemFalha       Falha = ""
    PrazoEstourado Falha = "PRAZO_ESTOURADO"
    Panico         Falha = "PANICO"
)

// Resposta é o resultado de uma chamada protegida.
type Resposta struct {
    Acoes   []jogo.Acao // saída bruta do bot; vazia em caso de falha
    Falha   Falha
    Detalhe string      // valor do panic, em texto; vazio nos outros casos
}

// Chamar entrega ao bot uma cópia do estado (BOT-01) e espera o plano até o
// prazo. Prazo estourado ou panic devolvem Acoes vazias e a falha (BOT-02).
// prazo <= 0 significa sem limite. Ao estourar o prazo, não espera a
// goroutine do bot terminar (decisão 5).
func Chamar(bot jogo.Bot, estado jogo.Estado, jogadorID string, prazo time.Duration) Resposta
```

```go
package main // cmd/terminal

// rodar executa o terminal com os argumentos dados e devolve o código de
// saída: 0 para partida terminada ou listagem, 2 para erro de uso.
func rodar(args []string, saida, erros io.Writer, catalogo bots.Catalogo) int

// resumo é o que jogar devolve para os testes conferirem a saída.
type resumo struct {
    desfecho   jogo.Desfecho
    relatorios int // total de relatórios de etapa (CA-16)
}

// jogar roda a partida do estado inicial até o fim, imprimindo em saida.
func jogar(estado jogo.Estado, jogadores []jogo.Bot, atraso time.Duration, saida io.Writer) resumo

// quadro desenha o tabuleiro e acompanha os blocos destrutíveis ao longo da partida.
type quadro struct {
    config       jogo.Config
    fixos        map[jogo.Posicao]bool
    destrutiveis map[jogo.Posicao]bool
}

func novoQuadro(estado jogo.Estado) *quadro
func (q *quadro) inicio(w io.Writer, estado jogo.Estado)    // "Turno 1, início" + tabuleiro
func (q *quadro) etapa(w io.Writer, r jogo.RelatorioEtapa) // cabeçalho, tabuleiro, eventos; depois remove os blocos destruídos
```

### Formato da saída (decisão 4)

```
Turno 1, início
1.+.2
.#.#.

infração: jogador_1, turno 1, etapa 2, VAL-02: <motivo>
aviso: jogador_2 estourou o prazo de 50ms no turno 1 (BOT-02)
aviso: jogador_2 entrou em pânico no turno 1 (BOT-02): <valor do panic>

Turno 1, etapa 1
1.+.2
.#.#.

Turno 1, etapa 2
o1+.2
.#.#.

Turno 1, etapa 4
***.2
*#.#.
mortes: jogador_1
blocos destruídos: (2,0)

Fim: vitória de jogador_2 no turno 1, etapa 4
```

As outras linhas de fim são `Fim: empate, todos morreram no turno T, etapa E` e `Fim: empate por limite de turnos entre jogador_1, jogador_3`.

## Decisões

- **D1**: O catálogo e `Chamar` ficam em `internal/bots`, e o laço da partida fica em `cmd/terminal`. **Motivo**: o marco 5 precisa do catálogo e da chamada protegida, mas terá um laço próprio, com relógio e fila, em `internal/partida`. O laço síncrono do terminal é curto, e movê-lo agora anteciparia o desenho do marco 5.
- **D2**: Em caso de falha, `Chamar` devolve ações vazias, e o laço manda ao validador um plano vazio. **Motivo**: um plano vazio já é executado como `ESPERAR` em todas as etapas (ACA-03, RES-01), sem regra nova e sem infração.
- **D3**: `Chamar` faz `estado.Copiar()` antes de iniciar a goroutine, e o laço chama `Validar` e `ResolverTurno` com o estado original. **Motivo**: BOT-01. Um bot que altera a cópia, mesmo depois do prazo, não alcança o estado da partida (CA-04).
- **D4**: Os bots de cada turno são chamados um de cada vez, na ordem de `jogadores`. **Motivo**: a saída fica determinística (CA-13), e o pior caso por turno é jogadores × prazo, aceitável no terminal. O marco 5 pode paralelizar.
- **D5**: `-atraso` é um `time.Duration` com padrão −1, que significa "usar `duracao_etapa_ms` do mapa". O laço dorme `atraso` antes de desenhar cada etapa, mas não antes do tabuleiro inicial. **Motivo**: distingue "sem argumento" de `-atraso 0`, e dá exatamente (tabuleiros − 1) esperas (CA-20).
- **D6**: Os jogadores são numerados pela posição em `jogadores` + 1 e desenhados a partir de `RelatorioEtapa.Jogadores`, que traz todos na ordem do `Estado`; os mortos são pulados. **Motivo**: decisão 3 e EST-08, sem depender do formato do id.
- **D7**: O desfecho é impresso com o turno e a etapa do último relatório. **Motivo**: `Desfecho` não guarda quando a partida acabou, e o último relatório é exatamente a etapa do fim (DEC-06).
- **D8**: O mapa padrão é o primeiro arquivo que existir entre `mapas/exemplo.json` e `../mapas/exemplo.json`. Um `-mapa` explícito é usado como foi dado. **Motivo**: decisão 1.
- **D9**: Os argumentos são lidos com `flag.NewFlagSet(..., flag.ContinueOnError)`, com a saída do `flag` apontada para `erros`. **Motivo**: `rodar` é testável sem `os.Exit`, e os erros de argumento seguem a decisão 6.

## Riscos

- **Testes de tempo (CA-05, CA-20)**: CA-05 usa um bot que dorme 2 s com prazo de 50 ms, e o teste só exige que a partida inteira termine em menos de 1 s. CA-20 usa atraso de 5 ms e só confere o limite inferior. As margens são largas para não falhar por instabilidade.
- **Goroutines perdidas nos testes**: o bot que dorme continua vivo depois do prazo (decisão 5). Ele dorme e termina sozinho, sem efeito em outros testes.
- **`duracao_etapa_ms` do mapa de exemplo** está como 1000 na alteração local ainda sem commit. O terminal fica lento sem `-atraso 0`, o que é esperado. Os testes sempre passam o atraso.
