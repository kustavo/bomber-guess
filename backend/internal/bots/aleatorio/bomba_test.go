package aleatorio

import (
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// planta informa se o plano tem PLANTAR.
func planta(plano []jogo.Acao) bool {
	for _, a := range plano {
		if a.Tipo == jogo.Plantar {
			return true
		}
	}
	return false
}

func TestCercadoNaoPlanta(t *testing.T) {
	e := montar(t, `
		###.
		#1#.
		###2
	`)
	for semente := uint64(1); semente <= 50; semente++ {
		if plano := planejarValido(t, e, semente); planta(plano) {
			t.Fatalf("BOM-01 CA-22 semente %d: plantou cercado: %+v", semente, plano)
		}
	}
}

func TestPlantaComPlanoSeguro(t *testing.T) {
	casos := []struct {
		nome    string
		desenho string
		opcoes  []opcao
	}{
		{"BOM-05 CA-21 campo aberto", campoAberto, nil},
		{"BOM-11 CA-21 campo aberto, turno curto", campoAberto, []opcao{comAtributos("jogador_1", func(a *jogo.Atributos) { a.AcoesPorTurno = 3 })}},
		{"BOM-09 CA-21 com bombas no tabuleiro", campoAberto, []opcao{comBomba(3, 2, "jogador_2", 2, 3), comBomba(3, 0, "jogador_2", 1, 8)}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e := montar(t, c.desenho, c.opcoes...)
			plantou := 0
			for semente := uint64(1); semente <= 50; semente++ {
				plano := planejarValido(t, e, semente)
				if !planta(plano) {
					continue
				}
				plantou++
				if obtido := classificar(e, "jogador_1", plano); obtido != seguro {
					t.Fatalf("semente %d: plano com bomba %s: %+v", semente, obtido, plano)
				}
			}
			if plantou == 0 {
				t.Error("nenhuma semente plantou")
			}
		})
	}
}
