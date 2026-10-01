package bots

import (
	"reflect"
	"testing"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// botDeTeste executa planejar e devolve o que ela devolver.
type botDeTeste struct {
	planejar func(estado jogo.Estado, jogadorID string) []jogo.Acao
}

func (b botDeTeste) Versao() string { return "teste-v1" }
func (b botDeTeste) Planejar(e jogo.Estado, id string) []jogo.Acao {
	return b.planejar(e, id)
}

func estadoDeTeste() jogo.Estado {
	return jogo.Estado{
		Turno:              1,
		Config:             jogo.Config{Largura: 3, Altura: 1, LimiteTurnos: 5, PrazoPlanejamentoMs: 50},
		BlocosFixos:        []jogo.Posicao{},
		BlocosDestrutiveis: []jogo.Posicao{{X: 1, Y: 0}},
		Bombas:             []jogo.Bomba{{Posicao: jogo.Posicao{X: 0, Y: 0}, JogadorID: "jogador_1", Potencia: 1, PavioRestante: 3}},
		Jogadores: []jogo.Jogador{
			{ID: "jogador_1", Status: jogo.Vivo, Atributos: jogo.Atributos{AcoesPorTurno: 3}},
			{ID: "jogador_2", Posicao: jogo.Posicao{X: 2, Y: 0}, Status: jogo.Vivo, Atributos: jogo.Atributos{AcoesPorTurno: 3}},
		},
	}
}

var umPlano = []jogo.Acao{{Etapa: 1, Tipo: jogo.Mover, Direcao: jogo.Direita}}

func TestChamar(t *testing.T) {
	casos := []struct {
		nome     string
		planejar func(jogo.Estado, string) []jogo.Acao
		prazo    time.Duration
		esperado Resposta
	}{
		{
			nome:     "BOT-02 bot dentro do prazo devolve a saída bruta",
			planejar: func(jogo.Estado, string) []jogo.Acao { return umPlano },
			prazo:    time.Second,
			esperado: Resposta{Acoes: umPlano},
		},
		{
			nome:     "BOT-02 CA-05 prazo estourado",
			planejar: func(jogo.Estado, string) []jogo.Acao { time.Sleep(2 * time.Second); return umPlano },
			prazo:    20 * time.Millisecond,
			esperado: Resposta{Acoes: []jogo.Acao{}, Falha: PrazoEstourado},
		},
		{
			nome:     "BOT-02 CA-06 panic",
			planejar: func(jogo.Estado, string) []jogo.Acao { panic("bum") },
			prazo:    time.Second,
			esperado: Resposta{Acoes: []jogo.Acao{}, Falha: Panico, Detalhe: "bum"},
		},
		{
			nome:     "BOT-02 prazo zero é sem limite",
			planejar: func(jogo.Estado, string) []jogo.Acao { time.Sleep(30 * time.Millisecond); return umPlano },
			prazo:    0,
			esperado: Resposta{Acoes: umPlano},
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			inicio := time.Now()
			r := Chamar(botDeTeste{c.planejar}, estadoDeTeste(), "jogador_1", c.prazo)
			if !reflect.DeepEqual(r, c.esperado) {
				t.Errorf("obtido %+v, esperado %+v", r, c.esperado)
			}
			if c.prazo > 0 && time.Since(inicio) > c.prazo+time.Second {
				t.Errorf("esperou %v, prazo %v", time.Since(inicio), c.prazo)
			}
		})
	}
}

func TestChamarEntregaCopia(t *testing.T) {
	e := estadoDeTeste()
	antes := e.Copiar()
	var recebidoID string
	vandalo := botDeTeste{func(c jogo.Estado, id string) []jogo.Acao {
		recebidoID = id
		c.Bombas[0].PavioRestante = 99
		c.BlocosDestrutiveis[0] = jogo.Posicao{X: 9, Y: 9}
		c.Jogadores[1].Status = jogo.Morto
		c.Jogadores[1].Morte = &jogo.Morte{Turno: 1, Etapa: 1}
		c.Turno = 42
		return nil
	}}
	Chamar(vandalo, e, "jogador_1", time.Second)
	if !reflect.DeepEqual(e, antes) {
		t.Errorf("BOT-01 CA-04 estado alterado pelo bot:\n antes:  %+v\n depois: %+v", antes, e)
	}
	if recebidoID != "jogador_1" {
		t.Errorf("bot recebeu o id %q", recebidoID)
	}
}
