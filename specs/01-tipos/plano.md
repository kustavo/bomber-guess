# Marco 01: plano técnico

**Spec**: `spec.md` (aprovada)

## Visão geral

Um módulo Go em `backend/` com um único pacote nesta etapa, `internal/jogo`. Os tipos espelham o JSON de `docs/REGRAS.md` e `docs/EDITOR.md` campo a campo. Os testes de serialização usam como fixture os próprios exemplos JSON dos docs: se o doc mudar sem o código acompanhar, o teste quebra. O mapa de exemplo fica em `mapas/` na raiz e também é verificado por teste.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/go.mod` | Módulo `github.com/kustavo/bomber-guess/backend`, `go 1.27`, sem dependências. |
| `backend/internal/jogo/doc.go` | Comentário do pacote: puro, sem I/O, linguagem ubíqua de `docs/REGRAS.md`. |
| `backend/internal/jogo/enums.go` | `TipoAcao`, `Direcao`, `Status` e suas constantes. |
| `backend/internal/jogo/estado.go` | `Posicao`, `Config`, `Atributos`, `Morte`, `Jogador`, `Bomba`, `Estado`, `Copiar`, `CalcularEtapas`. |
| `backend/internal/jogo/plano.go` | `Acao`, `Plano`. |
| `backend/internal/jogo/posicao.go` | `Posicao.Vizinha`, `Config.NoTabuleiro`. |
| `backend/internal/jogo/bot.go` | Interface `Bot`. |
| `backend/internal/jogo/mapa.go` | `Mapa`, `ErrMapaInvalido`, `LerMapa`, `VerificarMapa`, `EstadoInicial`. |
| `backend/internal/jogo/*_test.go` | Testes de tabela por arquivo, mais `docs_test.go` (helper que extrai JSON dos docs) e `pureza_test.go`. |
| `mapas/exemplo.json` | Mapa 15×13 no layout clássico. |

## Tipos e assinaturas

```go
// enums.go
type TipoAcao string // MOVER, PLANTAR, ESPERAR
type Direcao string  // CIMA, BAIXO, ESQUERDA, DIREITA
type Status string   // VIVO, MORTO

// estado.go
type Posicao struct { X int `json:"x"`; Y int `json:"y"` }

type Config struct {
    Largura             int `json:"largura"`
    Altura              int `json:"altura"`
    LimiteTurnos        int `json:"limite_turnos"`
    PrazoPlanejamentoMs int `json:"prazo_planejamento_ms"`
    DuracaoEtapaMs      int `json:"duracao_etapa_ms"`
}

// Atributos que valem para as bombas e o plano de um jogador.
// É o "jogador_padrao" do mapa e vai embutido em Jogador.
type Atributos struct {
    BombasPorTurno int `json:"bombas_por_turno"`
    Potencia       int `json:"potencia"`
    PavioPadrao    int `json:"pavio_padrao"`
    AcoesPorTurno  int `json:"acoes_por_turno"`
}

type Morte struct { Turno int `json:"turno"`; Etapa int `json:"etapa"` }

type Jogador struct {
    ID        string  `json:"id"`
    Posicao   Posicao `json:"posicao"`
    Status    Status  `json:"status"`
    Morte     *Morte  `json:"morte,omitempty"`
    Atributos         // campos achatados no JSON
    BotVersao string  `json:"bot_versao"`
}

type Bomba struct {
    Posicao       Posicao `json:"posicao"`
    JogadorID     string  `json:"jogador_id"`
    Potencia      int     `json:"potencia"`
    PavioRestante int     `json:"pavio_restante"`
}

type Estado struct {
    Turno              int       `json:"turno"`
    Config             Config    `json:"config"`
    EtapasNesteTurno   int       `json:"etapas_neste_turno"`
    BlocosFixos        []Posicao `json:"blocos_fixos"`
    BlocosDestrutiveis []Posicao `json:"blocos_destrutiveis"`
    Bombas             []Bomba   `json:"bombas"`
    Jogadores          []Jogador `json:"jogadores"`
}

func (e Estado) Copiar() Estado
func CalcularEtapas(jogadores []Jogador) int // maior AcoesPorTurno entre os VIVO (EST-03)

// plano.go
type Acao struct {
    Etapa   int      `json:"etapa"`
    Tipo    TipoAcao `json:"tipo"`
    Direcao Direcao  `json:"direcao,omitempty"`
}

type Plano struct {
    JogadorID string `json:"jogador_id"`
    Turno     int    `json:"turno"`
    Acoes     []Acao `json:"acoes"`
}

// posicao.go
func (p Posicao) Vizinha(d Direcao) Posicao // direção desconhecida: devolve p
func (c Config) NoTabuleiro(p Posicao) bool

// bot.go
type Bot interface {
    Versao() string
    Planejar(estado Estado, jogadorID string) []Acao
}

// mapa.go
type Mapa struct {
    Nome               string    `json:"nome"`
    Config             Config    `json:"config"`
    JogadorPadrao      Atributos `json:"jogador_padrao"`
    BlocosFixos        []Posicao `json:"blocos_fixos"`
    BlocosDestrutiveis []Posicao `json:"blocos_destrutiveis"`
    PosicoesIniciais   []Posicao `json:"posicoes_iniciais"`
}

var ErrMapaInvalido = errors.New("mapa inválido")

func LerMapa(dados []byte) (Mapa, error)  // decodifica e chama VerificarMapa
func VerificarMapa(m Mapa) error          // MAP-05; erros embrulham ErrMapaInvalido
func EstadoInicial(m Mapa, versoes []string) (Estado, error) // MAP-03, MAP-04, MAP-06
```

## Decisões

- **D1** `Atributos` embutido em `Jogador`. **Motivo**: o mesmo conjunto aparece em `jogador_padrao` do mapa e achatado no jogador. Com o embed, `encoding/json` promove os campos e a cópia em `EstadoInicial` vira uma atribuição.
- **D2** Fixtures dos testes de serialização são extraídas de `docs/REGRAS.md` (primeiro bloco ```json das seções 2 e 3) e `docs/EDITOR.md`. **Motivo**: doc e código não podem divergir sem quebrar o teste. Ler arquivo em teste não fere a pureza do pacote, que vale para código de produção.
- **D3** Igualdade "semântica" de JSON: decodificar os dois lados em `any` e comparar com `reflect.DeepEqual`.
- **D4** Slices vazios saem como `[]`, não `null`. `EstadoInicial` e `Copiar` sempre criam slices não nulos. **Motivo**: bots de outras linguagens e o frontend leem o JSON e não deveriam tratar `null`.
- **D5** Nomes no singular: o tipo é `Jogador`, mas a lista no `Estado` se chama `Jogadores`. Verbos no infinitivo para operações (`Copiar`, `Planejar`, `Verificar`). Para o mapa usamos `VerificarMapa`, e não `Validar`, porque `Validar` fica reservado para planos (`VAL`).
- **D6** Erros de mapa embrulham `ErrMapaInvalido` (`fmt.Errorf("%w: ...")`). Os testes checam com `errors.Is`, e a mensagem diz qual casa ou campo falhou.
- **D7** A pureza (CA-12) é verificada por um teste que analisa os imports dos arquivos não-teste do pacote com `go/parser`, e não por `go list`. **Motivo**: roda dentro de `go test`, sem depender de subprocesso.
- **D8** O mapa de exemplo é gerado uma vez por um script descartável e commitado como JSON comum. Os blocos destrutíveis seguem um padrão fixo e simétrico. O teste do CA-11 garante as propriedades, não o desenho exato.
- **D9** `go 1.27` no `go.mod`, a versão instalada. **Motivo**: acompanhar a versão atual do Go. As outras IAs devem usar a mesma versão, ou uma mais nova, para compilar seus bots.

## Riscos

- **Extração de JSON do Markdown é frágil.** Se alguém reorganizar o doc, o teste falha com uma mensagem clara ("bloco json da seção 2 não encontrado"), e não com um falso positivo.
- **O embed achata os campos no JSON.** Um campo novo em `Atributos` aparece automaticamente no `Jogador`. Isso é desejado, mas precisa ser lembrado ao adicionar power-ups (versão 2).
