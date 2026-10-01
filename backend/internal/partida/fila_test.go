package partida

import (
	"encoding/json"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/fila"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// publicarPlano publica um plano à mão em planos-enviados.
func (m *mesa) publicarPlano(t *testing.T, pe PlanoEnviado) {
	t.Helper()
	valor, err := json.Marshal(pe)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.fila.Publicar(fila.PlanosEnviados, fila.Mensagem{Chave: pe.Partida, Valor: valor}); err != nil {
		t.Fatal(err)
	}
}

func TestPlanosIgnorados(t *testing.T) {
	casos := []struct {
		nome  string
		plano PlanoEnviado
	}{
		{"PAR-04 CA-14 plano de outro turno", PlanoEnviado{Partida: "teste", Turno: 2, JogadorID: "jogador_1", Acoes: acoes("D")}},
		{"PAR-04 CA-14 plano de outra partida", PlanoEnviado{Partida: "outra", Turno: 1, JogadorID: "jogador_1", Acoes: acoes("D")}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			lento, _ := bloqueado(t, "B")
			m := novaMesa(t, mapaDe(t, doisCantos), lento, espera)
			m.publicarPlano(t, c.plano)
			t1 := m.jogarTurno(t, t0, 1)
			v := m.Visao(t1.Add(7 * d))
			if pos := v.Estado.Jogadores[0].Posicao; pos != (jogo.Posicao{}) {
				t.Errorf("o plano ignorado foi usado: jogador_1 em %v", pos)
			}
		})
	}

	t.Run("PAR-04 CA-14 plano atrasado não vale no turno seguinte", func(t *testing.T) {
		lento, liberar := bloqueado(t, "B")
		m := novaMesa(t, mapaDe(t, doisCantos), lento, espera)
		t1 := m.jogarTurno(t, t0, 1)
		agora, _ := m.Avancar(t1) // início do turno 2
		liberar()                 // chegam o plano atrasado do turno 1 e o do turno 2
		t2 := m.jogarTurno(t, agora, 2)
		m.Avancar(t2.Add(7 * d))
		r := registroDe(t, m.Historico(), 2, "jogador_1")
		if r.Falha != "" || len(r.Planejadas) != 1 {
			t.Errorf("turno 2: %+v", r)
		}
		if n := len(m.Historico().Turnos); n != 2 {
			t.Errorf("%d turnos no histórico", n)
		}
	})
}

func TestPrimeiroPlanoVale(t *testing.T) {
	lento, liberar := bloqueado(t, "B")
	m := novaMesa(t, mapaDe(t, doisCantos), lento, espera)
	m.publicarPlano(t, PlanoEnviado{Partida: "teste", Turno: 1, JogadorID: "jogador_1", Acoes: acoes("D")})
	liberar() // o bot publica "B" depois
	t1 := m.jogarTurno(t, t0, 3)
	m.Avancar(t1.Add(7 * d))
	r := registroDe(t, m.Historico(), 1, "jogador_1")
	if len(r.Planejadas) != 1 || r.Planejadas[0].Direcao != jogo.Direita {
		t.Errorf("RES-01 CA-15 valeu %+v, esperado o primeiro (DIREITA)", r.Planejadas)
	}
}
