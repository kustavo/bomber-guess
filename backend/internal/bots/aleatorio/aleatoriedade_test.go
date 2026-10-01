package aleatorio

import (
	"reflect"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// campoAberto tem jogador_1 e jogador_2 em posições simétricas pelo eixo
// vertical do meio.
const campoAberto = `
	.......
	.......
	1.....2
	.......
	.......
`

func TestMesmaEntradaMesmoPlano(t *testing.T) {
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		casos := []struct {
			nome   string
			opcoes []opcao
		}{
			{"CA-09 campo aberto", nil},
			{"CA-09 com bombas", []opcao{comBomba(1, 2, "jogador_2", 2, 2), comBomba(5, 1, "jogador_2", 1, 9)}},
			{"CA-09 turno 7", []opcao{comTurno(7)}},
		}
		for _, c := range casos {
			t.Run(c.nome, func(t *testing.T) {
				e := montar(t, campoAberto, c.opcoes...)
				for semente := uint64(1); semente <= 20; semente++ {
					bot := novo(semente)
					primeiro := bot.Planejar(e.Copiar(), "jogador_1")
					bot.Planejar(e.Copiar(), "jogador_2") // outra chamada no meio não muda nada
					mesmoBot := bot.Planejar(e.Copiar(), "jogador_1")
					outroBot := novo(semente).Planejar(e.Copiar(), "jogador_1")
					if !reflect.DeepEqual(primeiro, mesmoBot) || !reflect.DeepEqual(primeiro, outroBot) {
						t.Fatalf("semente %d: planos diferentes:\n %+v\n %+v\n %+v", semente, primeiro, mesmoBot, outroBot)
					}
				}
			})
		}
	})
}

func TestSementesDiferentesPlanosDiferentes(t *testing.T) {
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		e := montar(t, campoAberto)
		distintos := map[string]bool{}
		for semente := uint64(1); semente <= 20; semente++ {
			distintos[formatar(novo(semente).Planejar(e.Copiar(), "jogador_1"))] = true
		}
		if len(distintos) < 2 {
			t.Errorf("CA-10 só %d plano distinto em 20 sementes", len(distintos))
		}
	})
}

func TestJogadoresETurnosDiferentes(t *testing.T) {
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		e := montar(t, campoAberto)
		outroTurno := montar(t, campoAberto, comTurno(2))
		espelhoDiferente, turnoDiferente := false, false
		for semente := uint64(1); semente <= 20; semente++ {
			bot := novo(semente)
			p1 := bot.Planejar(e.Copiar(), "jogador_1")
			p2 := bot.Planejar(e.Copiar(), "jogador_2")
			if !reflect.DeepEqual(espelhar(p1), p2) {
				espelhoDiferente = true
			}
			if !reflect.DeepEqual(p1, bot.Planejar(outroTurno.Copiar(), "jogador_1")) {
				turnoDiferente = true
			}
		}
		if !espelhoDiferente {
			t.Error("CA-11 jogadores simétricos com planos sempre espelhados")
		}
		if !turnoDiferente {
			t.Error("CA-11 o mesmo jogador tem o mesmo plano em turnos diferentes")
		}
	})
}

// espelhar troca ESQUERDA e DIREITA.
func espelhar(plano []jogo.Acao) []jogo.Acao {
	r := append([]jogo.Acao{}, plano...)
	for i, a := range r {
		switch a.Direcao {
		case jogo.Esquerda:
			r[i].Direcao = jogo.Direita
		case jogo.Direita:
			r[i].Direcao = jogo.Esquerda
		}
	}
	return r
}

// formatar devolve o plano como texto, para comparar e contar planos.
func formatar(plano []jogo.Acao) string {
	s := ""
	for _, a := range plano {
		s += string(a.Tipo) + ":" + string(a.Direcao) + " "
	}
	return s
}
