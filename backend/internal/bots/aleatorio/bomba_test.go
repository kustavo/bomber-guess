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
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		e := montar(t, `
			###.
			#1#.
			###2
		`)
		for semente := uint64(1); semente <= 50; semente++ {
			if plano := planejarValido(t, novo, e, semente); planta(plano) {
				t.Fatalf("BOM-01 CA-22 semente %d: plantou cercado: %+v", semente, plano)
			}
		}
	})
}

func TestPlantaComPlanoSeguro(t *testing.T) {
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		casos := []struct {
			nome    string
			desenho string
			opcoes  []opcao
			// semBombaNoV4: no v4, todo plano com bomba fica num nível de
			// proteção pior que o melhor sem bomba, e ele não planta (R3 do
			// marco 16).
			semBombaNoV4 bool
		}{
			{"BOM-05 CA-21 campo aberto", campoAberto, nil, false},
			{"BOM-11 CA-21 campo aberto, turno curto", campoAberto, []opcao{comAtributos("jogador_1", func(a *jogo.Atributos) { a.AcoesPorTurno = 3 })}, true},
			{"BOM-09 CA-21 com bombas no tabuleiro", campoAberto, []opcao{comBomba(3, 2, "jogador_2", 2, 3), comBomba(3, 0, "jogador_2", 1, 8)}, false},
		}
		for _, c := range casos {
			t.Run(c.nome, func(t *testing.T) {
				e := montar(t, c.desenho, c.opcoes...)
				plantou := 0
				for semente := uint64(1); semente <= 50; semente++ {
					plano := planejarValido(t, novo, e, semente)
					if !planta(plano) {
						continue
					}
					plantou++
					if obtido := classificar(e, "jogador_1", plano); obtido != seguro {
						t.Fatalf("semente %d: plano com bomba %s: %+v", semente, obtido, plano)
					}
				}
				switch {
				case c.semBombaNoV4 && novo(1).Versao() == VersaoV4:
					if plantou > 0 {
						t.Errorf("CA-02 R3 (marco 16) o v4 plantou em %d sementes", plantou)
					}
				case plantou == 0:
					t.Error("nenhuma semente plantou")
				}
			})
		}
	})
}
