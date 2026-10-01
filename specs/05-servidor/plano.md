# Marco 05: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

O núcleo é uma `Partida` (`internal/partida`) que funciona como uma **máquina de estados guiada pelo tempo**. Ela não dorme nem tem timer próprio.

- **`Avancar(agora)`** leva a partida até o instante dado. Atravessa quantas fases forem necessárias e devolve o próximo instante em que algo muda.
- **`Visao(agora)`** e **`Historico()`** leem o estado. `Visao` chama `Avancar(agora)` antes de responder, então o que ela mostra está sempre coerente com o horário, mesmo que ninguém tenha avançado a partida antes.

Os testes chamam `Avancar` com instantes escolhidos à mão; esse é o "relógio de teste" da spec. No servidor, um laço por partida chama `Avancar` com o relógio real e espera até o próximo instante.

Cada turno passa por estas etapas:

1. **Início do planejamento** (`t0`): a partida dispara uma goroutine por jogador vivo. Cada uma chama o bot com `bots.Chamar(..., prazo 0)`, que entrega uma cópia do estado e captura `panic`. Depois publica um `PlanoEnviado` no tópico `planos-enviados`, inclusive quando houve `panic`; nesse caso a mensagem traz a falha e nenhuma ação. Quem impõe o prazo (BOT-02, PAR-04) é o fim da fase, e não um timer na chamada.
2. **Fim do planejamento** (`t1 = t0 + p`): a partida lê o seu consumidor de `planos-enviados`. Fica com as mensagens dela, do turno atual e só a primeira de cada jogador (CA-14, CA-15). Quem não tem mensagem recebe a falha `PRAZO_ESTOURADO` e um plano vazio. Em seguida, `Validar` e `ResolverTurno` resolvem o turno inteiro de uma vez (decisão 3).
3. **Execução**: `Visao` mostra os relatórios das etapas 1…k, com k = ⌊(agora − t1)/d⌋ + 1.
4. **Fim da execução** (`t1 + n·d`): o registro do turno entra no histórico e é publicado em `turno-resolvido`. Se a partida terminou, a fase passa a `ENCERRADA` e o desfecho é publicado em `partida-finalizada`. Senão, começa o planejamento do turno seguinte.

O **Gerenciador** (`internal/partida`) cria partidas a partir de um pedido: valida o nome, carrega o mapa do diretório e cria os bots pelo catálogo. Ele guarda as partidas por nome, na ordem de criação, e liga ou não o laço de cada uma, conforme o modo automático.

A **API** (`internal/api`) é uma camada fina sobre o Gerenciador. Ela usa `net/http` com os padrões de rota do Go 1.22 e cuida dos códigos de status, do JSON de erro e do `horario_servidor` em todas as respostas.

A **fila** (`internal/fila`) é uma interface com `Publicar` e `Consumir`; o consumidor tem uma leitura que não bloqueia. A implementação em memória é síncrona e ordenada.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/internal/fila/fila.go` | `Mensagem`, `Fila`, `Consumidor`, constantes dos tópicos. |
| `backend/internal/fila/memoria.go` | `Memoria`, `NovaMemoria`. |
| `backend/internal/fila/filateste/contrato.go` | `TestarContrato(t, nova)`: teste de contrato reutilizável (CA-12). Fica fora de `_test.go` para o pacote do Kafka (marco 8) poder importá-lo. |
| `backend/internal/fila/memoria_test.go` | Aplica o contrato à `Memoria`. |
| `backend/internal/partida/registro.go` | `Fase`, `PlanoEnviado`, `AcaoExecutada`, `RegistroJogador`, `RegistroTurno`, `Historico`, `Visao`, `Finalizada`. |
| `backend/internal/partida/partida.go` | `Partida`, `Nova`, `Avancar`, `Visao`, `Historico`; disparo dos bots e fechamento do turno. |
| `backend/internal/partida/gerenciador.go` | `Relogio`, `RelogioReal`, `Pedido`, erros, `ConfigGerenciador`, `Gerenciador` e o laço em tempo real (`rodar`). |
| `backend/internal/partida/apoio_test.go` | Bots de teste (que esperam, bloqueiam até serem liberados, entram em `panic`, alteram o estado, saem do tabuleiro, matam), mapas em ASCII e `esperarPlanos`. |
| `backend/internal/partida/fases_test.go` | CA-01 a CA-03, CA-13. |
| `backend/internal/partida/registro_test.go` | CA-04 a CA-07. |
| `backend/internal/partida/fila_test.go` | CA-14, CA-15. |
| `backend/internal/partida/fim_test.go` | CA-08 a CA-11, CA-16. |
| `backend/internal/partida/gerenciador_test.go` | Pedidos válidos e inválidos (base de CA-18 e CA-19), lista na ordem de criação, modo automático com relógio real. |
| `backend/internal/api/api.go` | `Novo` (o `http.Handler`), rotas, `responder`, `responderErro`, horário em RFC 3339 com ms. |
| `backend/internal/api/respostas.go` | Corpos JSON de cada endpoint. |
| `backend/internal/api/api_test.go` | CA-17 a CA-24, com `httptest` e um relógio de teste. |
| `backend/cmd/servidor/main.go` | `main`, `rodar(ctx, args, erros, pronto)`: flags, diretório de mapas, `http.Server` e encerramento. |
| `backend/cmd/servidor/main_test.go` | CA-25. |

## Tipos e assinaturas

```go
package fila

const (
    PlanosEnviados    = "planos-enviados"
    TurnoResolvido    = "turno-resolvido"
    PartidaFinalizada = "partida-finalizada"
)

// Mensagem é um item publicado num tópico. Chave é o nome da partida.
type Mensagem struct {
    Chave string
    Valor []byte // JSON
}

// Fila publica mensagens em tópicos (FILA-01). Implementações: Memoria; Kafka no marco 8.
type Fila interface {
    Publicar(topico string, m Mensagem) error
    // Consumir cria um consumidor que recebe as mensagens publicadas no
    // tópico a partir de agora, na ordem de publicação.
    Consumir(topico string) (Consumidor, error)
}

// Consumidor lê as mensagens de um tópico.
type Consumidor interface {
    // Ler devolve, sem bloquear, as mensagens que chegaram desde a última leitura.
    Ler() ([]Mensagem, error)
    Fechar() error
}

// Memoria é a fila em memória: Publicar entrega a mensagem na hora a todos
// os consumidores do tópico. Segura para uso concorrente.
type Memoria struct { /* ... */ }

func NovaMemoria() *Memoria
```

```go
package partida

// Fase é a fase atual de uma partida.
type Fase string

const (
    Planejamento Fase = "PLANEJAMENTO"
    Execucao     Fase = "EXECUCAO"
    Encerrada    Fase = "ENCERRADA"
)

// PlanoEnviado é a mensagem de planos-enviados (FILA-02, BOT-03).
type PlanoEnviado struct {
    Partida   string      `json:"partida"`
    Turno     int         `json:"turno"`
    JogadorID string      `json:"jogador_id"`
    Acoes     []jogo.Acao `json:"acoes"`             // saída bruta do bot
    Falha     bots.Falha  `json:"falha,omitempty"`   // PANICO; PRAZO_ESTOURADO nunca é publicado
    Detalhe   string      `json:"detalhe,omitempty"`
}

// AcaoExecutada é a ação de uma etapa e o que aconteceu com ela (DEC-09).
type AcaoExecutada struct {
    Etapa     int                `json:"etapa"`
    Acao      jogo.Acao          `json:"acao"`
    Resultado jogo.ResultadoAcao `json:"resultado"`
}

// RegistroJogador são as três versões das ações de um jogador num turno (PAR-05).
type RegistroJogador struct {
    JogadorID  string          `json:"jogador_id"`
    Planejadas []jogo.Acao     `json:"planejadas"`
    Validadas  []jogo.Acao     `json:"validadas"`
    Executadas []AcaoExecutada `json:"executadas"`
    Infracoes  []jogo.Infracao `json:"infracoes"`
    Falha      bots.Falha      `json:"falha,omitempty"`
    Detalhe    string          `json:"detalhe,omitempty"`
}

// RegistroTurno é o que fica gravado de cada turno executado; também é a
// mensagem de turno-resolvido.
type RegistroTurno struct {
    Partida   string                `json:"partida"`
    Turno     int                   `json:"turno"`
    Estado    jogo.Estado           `json:"estado"` // início do turno
    Jogadores []RegistroJogador     `json:"jogadores"` // vivos no início do turno, na ordem do estado
    Etapas    []jogo.RelatorioEtapa `json:"etapas"`
}

// Historico é o registro completo de uma partida (API-07, decisão 6).
type Historico struct {
    Nome     string          `json:"nome"`
    Mapa     string          `json:"mapa"`
    Bots     []string        `json:"bots"`
    Semente  uint64          `json:"semente"`
    Turnos   []RegistroTurno `json:"turnos"`
    Desfecho *jogo.Desfecho  `json:"desfecho,omitempty"`
}

// Finalizada é a mensagem de partida-finalizada.
type Finalizada struct {
    Partida  string        `json:"partida"`
    Desfecho jogo.Desfecho `json:"desfecho"`
}

// Visao é o que se vê da partida num instante (decisão 5, sem o horário do servidor).
type Visao struct {
    Nome      string
    Fase      Fase
    Turno     int
    Etapa     int        // 0 no planejamento
    FimDaFase time.Time  // zero em ENCERRADA
    Estado    jogo.Estado
    Etapas    []jogo.RelatorioEtapa
    Desfecho  jogo.Desfecho
}

// Config é o que identifica uma partida.
type Config struct {
    Nome    string
    Mapa    string
    Bots    []string
    Semente uint64
}

// Partida é uma partida em tempo real. Segura para uso concorrente.
type Partida struct { /* mutex, config, estado, fase, inicioDaFase, turno resolvido, histórico, fila, consumidor */ }

// Nova cria a partida e começa o planejamento do turno 1 em agora. jogadores[i]
// joga com estado.Jogadores[i]. Assina planos-enviados antes de disparar os bots.
func Nova(cfg Config, estado jogo.Estado, jogadores []jogo.Bot, f fila.Fila, agora time.Time) (*Partida, error)

// Avancar leva a partida até agora e devolve o próximo instante em que algo
// muda (fim da fase) e se ela está encerrada.
func (p *Partida) Avancar(agora time.Time) (proximo time.Time, encerrada bool)

// Visao avança até agora e devolve o que se vê da partida.
func (p *Partida) Visao(agora time.Time) Visao

// Historico devolve uma cópia do histórico (só turnos totalmente executados).
func (p *Partida) Historico() Historico

// Relogio dá a hora atual. RelogioReal usa time.Now.
type Relogio interface{ Agora() time.Time }
type RelogioReal struct{}

// Pedido é o corpo de POST /partidas (decisão 6).
type Pedido struct {
    Nome    string   `json:"nome"`
    Mapa    string   `json:"mapa"`
    Bots    []string `json:"bots"`
    Semente *uint64  `json:"semente,omitempty"`
}

var (
    ErrPedidoInvalido      = errors.New("pedido inválido")      // 400
    ErrNomeRepetido        = errors.New("nome já usado")        // 409
    ErrPartidaDesconhecida = errors.New("partida desconhecida") // 404
)

type ConfigGerenciador struct {
    Catalogo   bots.Catalogo
    DirMapas   string
    Fila       fila.Fila
    Relogio    Relogio
    Automatico bool // liga o laço em tempo real de cada partida; false nos testes
}

type Gerenciador struct { /* ... */ }

func NovoGerenciador(cfg ConfigGerenciador) *Gerenciador
func (g *Gerenciador) Criar(p Pedido) (*Partida, error) // erros embrulham ErrPedidoInvalido ou ErrNomeRepetido
func (g *Gerenciador) Obter(nome string) (*Partida, error)
func (g *Gerenciador) Listar() []*Partida      // ordem de criação
func (g *Gerenciador) Versoes() []string       // do catálogo
func (g *Gerenciador) Relogio() Relogio
func (g *Gerenciador) Encerrar()               // para os laços e espera que terminem
```

```go
package api

// Novo devolve o handler da API sobre o gerenciador (API-01 a API-10).
func Novo(g *partida.Gerenciador) http.Handler
```

```go
package main // cmd/servidor

// rodar sobe o servidor até ctx ser cancelado; pronto recebe o endereço
// em que ele escuta (útil com -porta 0).
func rodar(ctx context.Context, args []string, erros io.Writer, pronto func(endereco string)) int
```

### Respostas da API

| Rota | Status | Corpo (sempre com `horario_servidor`) |
|---|---|---|
| `GET /bots` | 200 | `{"bots": [...]}` |
| `GET /partidas` | 200 | `{"partidas": [{"nome", "fase", "turno", "fim_da_fase", "desfecho"}]}` |
| `POST /partidas` | 201 | resposta de estado |
| `GET /partidas/{nome}/estado` | 200 | resposta de estado (decisão 5) |
| `GET /partidas/{nome}/historico` | 200 | `Historico` + `fase` + `fim_da_fase` |
| `POST /mapas`, `POST /partidas/{nome}/turnos/{n}/plano`, `GET /ranking` | 501 | `{"erro"}` |
| método errado numa rota conhecida | 405 | `{"erro"}` |
| qualquer outra rota | 404 | `{"erro"}` |

## Decisões

- **D1**: A `Partida` não dorme: `Avancar(agora)` e `Visao(agora)` recebem o instante, e um laço à parte (`rodar`) chama `Avancar` com o relógio real. **Motivo**: os critérios de tempo ficam determinísticos e rápidos, sem relógio falso com timers. Como `Visao` também avança, o estado mostrado nunca fica atrasado em relação ao horário, mesmo que o laço atrase.
- **D2**: Os bots rodam com `bots.Chamar(..., prazo 0)` e publicam o resultado na fila. O prazo é o fim da fase. **Motivo**: o prazo de `PAR-02` e `PAR-04` é um só, o da fase, então não existe um segundo timer que possa discordar dele. A cópia do estado e o `recover` continuam vindo do marco 4.
- **D3**: Um `panic` também é publicado em `planos-enviados`, com a falha e sem ações. **Motivo**: o teste espera N mensagens antes de fechar o planejamento; sem isso, um `panic` ainda não registrado viraria `PRAZO_ESTOURADO` por acaso. De quebra, a fila guarda a saída bruta de todo bot (BOT-03).
- **D4**: Cada partida tem o próprio consumidor de `planos-enviados`, criado em `Nova` antes de disparar os bots. No fim do planejamento, ela lê tudo e filtra por partida e turno. **Motivo**: não precisa de despachante entre partidas, e funciona igual com Kafka (um consumidor por partida). Mensagens atrasadas são lidas no turno seguinte e descartadas pelo número do turno (CA-14).
- **D5**: `Consumidor.Ler` não bloqueia. **Motivo**: o fim da fase é um instante; a partida lê o que chegou até ali e segue. Na memória é exato; no Kafka (marco 8) vira uma leitura com tempo curto.
- **D6**: O turno é resolvido inteiro no fim do planejamento, mas o registro do turno só entra no histórico e em `turno-resolvido` no fim da execução. **Motivo**: decisão 3 e CA-21/CA-22, que proíbem mostrar etapas futuras, nem pelo estado nem pelo histórico.
- **D7**: Se `Avancar` recebe um instante várias fases à frente (servidor atrasado), ele atravessa as fases em ordem. Um planejamento atravessado assim pode fechar antes de os bots responderem, e eles ficam com `PRAZO_ESTOURADO`. **Motivo**: o tempo da partida manda; só acontece se o processo travar.
- **D8**: Nomes de partida e de mapa seguem `^[a-z0-9-]{1,64}$`, e o mapa é lido de `<DirMapas>/<mapa>.json`. **Motivo**: a decisão 6 pede o formato; aplicá-lo também ao mapa evita que um pedido leia arquivos fora do diretório (`../`).
- **D9**: As rotas são registradas pelo caminho, sem método, e cada handler despacha o método e responde 405 em JSON. Uma rota `/` pega o resto e responde 404 em JSON. **Motivo**: os 405 e 404 automáticos do `ServeMux` são texto puro, sem `horario_servidor` (CA-23).
- **D10**: Os horários vão no JSON como `2006-01-02T15:04:05.000Z`, em UTC, por um tipo próprio da API. **Motivo**: decisão 5. A `Visao` fica com `time.Time` e não sabe nada de JSON.
- **D11**: O laço em tempo real (`Automatico`) é testado uma vez com o relógio real, numa partida pequena com prazo e etapas de poucos milissegundos, que precisa chegar a `ENCERRADA`. Os demais testes usam `Automatico: false` e chamam `Avancar`. **Motivo**: cobre o laço sem deixar a suíte lenta ou instável.

- **D12** (acrescentada na implementação): ao encerrar, o servidor espera as requisições em andamento por 2 s e então fecha à força o que sobrou, com código de saída 0. **Motivo**: o `net/http` só considera ociosa uma conexão que nunca enviou nada depois de 5 s, e isso fazia o encerramento falhar de vez em quando (achado ao repetir o teste do CA-25).

## Riscos

- **Bots de teste que bloqueiam**: ficam presos até o teste liberá-los. Todo bot assim é liberado em `t.Cleanup`.
- **Espera por planos nos testes**: `esperarPlanos` consulta um consumidor próprio do teste até ver N mensagens, com limite de 2 s. É a única espera com tempo real nos testes da partida.
- **Bot que nunca termina no servidor** (decisão 8): a goroutine fica perdida e segura uma cópia do estado. Fica aceito até o marco 10.
- **Memória do servidor**: históricos e consumidores ficam em memória enquanto o processo vive. O consumidor de uma partida encerrada é fechado; o histórico fica, de propósito (decisão 7).
