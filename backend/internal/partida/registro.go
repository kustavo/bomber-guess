// Package partida roda partidas em tempo real no servidor (docs/ARQUITETURA.md,
// seção 2): fases de planejamento e execução, registro dos turnos e o
// gerenciador que cria e guarda as partidas.
package partida

import (
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// Fase é a fase atual de uma partida.
type Fase string

const (
	Planejamento Fase = "PLANEJAMENTO" // PAR-02
	Execucao     Fase = "EXECUCAO"     // PAR-03
	Encerrada    Fase = "ENCERRADA"
)

// PlanoEnviado é a mensagem de planos-enviados (FILA-02, BOT-03). Um panic
// também é publicado, com a falha e sem ações (D3); PRAZO_ESTOURADO nunca é.
type PlanoEnviado struct {
	Partida   string      `json:"partida"`
	Turno     int         `json:"turno"`
	JogadorID string      `json:"jogador_id"`
	Acoes     []jogo.Acao `json:"acoes"` // saída bruta do bot
	Falha     bots.Falha  `json:"falha,omitempty"`
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
	Estado    jogo.Estado           `json:"estado"`    // início do turno
	Jogadores []RegistroJogador     `json:"jogadores"` // vivos no início do turno, na ordem do estado
	Etapas    []jogo.RelatorioEtapa `json:"etapas"`
}

// Historico é o registro completo de uma partida (API-07).
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

// Visao é o que se vê da partida num instante (API-05).
type Visao struct {
	Nome      string
	Fase      Fase
	Turno     int
	Etapa     int       // 0 no planejamento
	FimDaFase time.Time // zero em ENCERRADA
	Estado    jogo.Estado
	Etapas    []jogo.RelatorioEtapa // só as já liberadas
	Desfecho  jogo.Desfecho
}

// Config é o que identifica uma partida.
type Config struct {
	Nome    string
	Mapa    string
	Bots    []string
	Semente uint64
}
