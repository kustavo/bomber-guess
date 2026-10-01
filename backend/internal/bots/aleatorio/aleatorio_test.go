package aleatorio

import (
	"reflect"
	"testing"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

func TestVersao(t *testing.T) {
	for _, semente := range []uint64{0, 1, 42} {
		if v := Novo(semente).Versao(); v != "aleatorio-v1" {
			t.Errorf("BOT-01 CA-01 semente %d: Versao() = %q", semente, v)
		}
		if v := NovoV2(semente).Versao(); v != "aleatorio-v2" {
			t.Errorf("BOT-01 CA-01 v2 semente %d: Versao() = %q", semente, v)
		}
	}
}

func TestPlanoComAcoesPorTurno(t *testing.T) {
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		casos := []struct {
			nome  string
			acoes int
		}{
			{"ACA-02 CA-03 uma ação", 1},
			{"ACA-02 CA-03 três ações", 3},
			{"ACA-02 CA-03 sete ações", 7},
		}
		for _, c := range casos {
			t.Run(c.nome, func(t *testing.T) {
				e := montar(t, `
					1....
					.....
					....2
				`, comAtributos("jogador_1", func(a *jogo.Atributos) { a.AcoesPorTurno = c.acoes }))
				for semente := range uint64(20) {
					conferirContrato(t, e, "jogador_1", novo(semente).Planejar(e.Copiar(), "jogador_1"))
				}
			})
		}
	})
}

func TestPlanoVazio(t *testing.T) {
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		casos := []struct {
			nome      string
			estado    func(t *testing.T) jogo.Estado
			jogadorID string
		}{
			{"EST-08 CA-04 estado vazio", func(*testing.T) jogo.Estado { return jogo.Estado{} }, "jogador_1"},
			{"EST-08 CA-04 jogador inexistente", func(t *testing.T) jogo.Estado { return montar(t, "1.2") }, "jogador_9"},
			{"FIM-01 CA-04 jogador morto", func(t *testing.T) jogo.Estado { return montar(t, "1.2", comMorto("jogador_1")) }, "jogador_1"},
		}
		for _, c := range casos {
			t.Run(c.nome, func(t *testing.T) {
				if plano := novo(1).Planejar(c.estado(t), c.jogadorID); len(plano) != 0 {
					t.Errorf("plano %+v, esperado vazio", plano)
				}
			})
		}
	})
}

func TestNaoAlteraOEstado(t *testing.T) {
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		inicial, err := jogo.EstadoInicial(lerMapaExemplo(t), []string{Versao, Versao, Versao, Versao})
		if err != nil {
			t.Fatal(err)
		}
		casos := []struct {
			nome   string
			estado jogo.Estado
		}{
			{"BOT-01 CA-02 mapa de exemplo", inicial},
			{"BOT-01 CA-02 com bombas e jogador morto", montar(t, campoAberto,
				comBomba(1, 2, "jogador_2", 2, 1), comBomba(1, 2, "jogador_2", 1, 9), comBomba(4, 4, "jogador_1", 1, 3),
				comMorto("jogador_2"))},
			{"BOT-01 CA-02 cercado sobre bomba", montar(t, "#1#2", comBomba(1, 0, "jogador_2", 1, 1))},
		}
		for _, c := range casos {
			t.Run(c.nome, func(t *testing.T) {
				antes := c.estado.Copiar()
				for semente := uint64(1); semente <= 20; semente++ {
					for _, j := range c.estado.Jogadores {
						novo(semente).Planejar(c.estado, j.ID)
					}
				}
				if !reflect.DeepEqual(c.estado, antes) {
					t.Errorf("estado alterado:\n antes:  %+v\n depois: %+v", antes, c.estado)
				}
			})
		}
	})
}

func TestTempoDePlanejamento(t *testing.T) {
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		m := lerMapaExemplo(t)
		inicial, err := jogo.EstadoInicial(m, []string{Versao, Versao, Versao, Versao})
		if err != nil {
			t.Fatal(err)
		}
		// Um estado do meio da partida, com bombas e blocos destruídos.
		meio := inicial
		bot := novo(7)
		for range 10 {
			var planos []jogo.Plano
			for _, j := range meio.Jogadores {
				planos = append(planos, jogo.Plano{JogadorID: j.ID, Turno: meio.Turno, Acoes: bot.Planejar(meio.Copiar(), j.ID)})
			}
			meio, _ = jogo.ResolverTurno(meio, planos)
		}
		// Um estado dentro da janela com jogador preso atrás de blocos
		// destrutíveis, tirado de uma partida com o v2 (CA-05 do marco 13).
		janela, ok := estadoComPreso(t, m)
		if !ok {
			t.Fatal("BOT-02 CA-05 nenhuma partida chegou à janela com jogador preso")
		}
		limite := time.Duration(m.Config.PrazoPlanejamentoMs) * time.Millisecond / 10
		for nome, e := range map[string]jogo.Estado{"inicial": inicial, "meio": meio, "janela": janela} {
			const chamadas = 20
			inicio := time.Now()
			for semente := range uint64(chamadas) {
				for _, j := range e.Jogadores {
					novo(semente).Planejar(e.Copiar(), j.ID)
				}
			}
			media := time.Since(inicio) / time.Duration(chamadas*len(e.Jogadores))
			if media >= limite {
				t.Errorf("BOT-02 CA-05 estado %s: média de %v por chamada, limite %v", nome, media, limite)
			}
		}
	})
}

// estadoComPreso joga partidas com 4 aleatorio-v2 no mapa e devolve o
// primeiro estado dentro da janela de antecipação com um jogador vivo preso.
func estadoComPreso(t *testing.T, m jogo.Mapa) (jogo.Estado, bool) {
	t.Helper()
	for semente := uint64(1); semente <= 20; semente++ {
		e, err := jogo.EstadoInicial(m, []string{VersaoV2, VersaoV2, VersaoV2, VersaoV2})
		if err != nil {
			t.Fatal(err)
		}
		bot := NovoV2(semente)
		for !jogo.VerificarFim(e).Terminada {
			if naJanela(e.Config, e.Turno) {
				for _, j := range e.Jogadores {
					if _, preso := blocoDeAbertura(e, j.Posicao, anelAlvo(e.Config, e.Turno)); preso && j.Status == jogo.Vivo {
						return e, true
					}
				}
			}
			var planos []jogo.Plano
			for _, j := range e.Jogadores {
				if j.Status == jogo.Vivo {
					planos = append(planos, jogo.Plano{JogadorID: j.ID, Turno: e.Turno, Acoes: bot.Planejar(e.Copiar(), j.ID)})
				}
			}
			e, _ = jogo.ResolverTurno(e, planos)
		}
	}
	return jogo.Estado{}, false
}
