package jogo_test

import (
	"reflect"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// botVandalo altera tudo o que recebe, para provar que o estado de quem
// chamou fica intacto.
type botVandalo struct{}

func (botVandalo) Versao() string { return "vandalo-v1" }

func (botVandalo) Planejar(estado jogo.Estado, jogadorID string) []jogo.Acao {
	estado.Turno = -1
	estado.BlocosFixos[0].X = -1
	estado.BlocosDestrutiveis[0].X = -1
	estado.Bombas[0].PavioRestante = -1
	estado.Jogadores[0].Posicao.X = -1
	estado.Jogadores[1].Morte.Turno = -1
	return []jogo.Acao{{Etapa: 1, Tipo: jogo.Esperar}}
}

// Verificação em tempo de compilação: um pacote externo implementa jogo.Bot.
var _ jogo.Bot = botVandalo{}

func novoEstado() jogo.Estado {
	return jogo.Estado{
		Turno:              3,
		BlocosFixos:        []jogo.Posicao{{X: 1, Y: 1}},
		BlocosDestrutiveis: []jogo.Posicao{{X: 2, Y: 2}},
		Bombas:             []jogo.Bomba{{JogadorID: "jogador_1", PavioRestante: 2}},
		Jogadores: []jogo.Jogador{
			{ID: "jogador_1", Status: jogo.Vivo},
			{ID: "jogador_2", Status: jogo.Morto, Morte: &jogo.Morte{Turno: 2, Etapa: 1}},
		},
	}
}

func TestBot(t *testing.T) {
	casos := []struct {
		nome string
		bot  jogo.Bot
	}{
		{"BOT-01 bot que altera o estado recebido não altera o original", botVandalo{}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			estado := novoEstado()
			acoes := c.bot.Planejar(estado.Copiar(), "jogador_1")
			if len(acoes) != 1 || c.bot.Versao() != "vandalo-v1" {
				t.Fatalf("bot devolveu %v, versão %q", acoes, c.bot.Versao())
			}
			if !reflect.DeepEqual(estado, novoEstado()) {
				t.Errorf("estado de quem chamou foi alterado: %+v", estado)
			}
		})
	}
}
