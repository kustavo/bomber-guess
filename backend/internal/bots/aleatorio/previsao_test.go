package aleatorio

import (
	"reflect"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// cruz devolve a casa (x,y) e as vizinhas até k casas em cada direção, dentro
// de um tabuleiro largura × altura, sem blocos.
func cruz(x, y, k, largura, altura int) []jogo.Posicao {
	c := []jogo.Posicao{{X: x, Y: y}}
	for i := 1; i <= k; i++ {
		for _, p := range []jogo.Posicao{{X: x - i, Y: y}, {X: x + i, Y: y}, {X: x, Y: y - i}, {X: x, Y: y + i}} {
			if p.X >= 0 && p.X < largura && p.Y >= 0 && p.Y < altura {
				c = append(c, p)
			}
		}
	}
	return c
}

func conjuntoDe(ps []jogo.Posicao) map[jogo.Posicao]bool {
	c := map[jogo.Posicao]bool{}
	for _, p := range ps {
		c[p] = true
	}
	return c
}

func TestPreverLinha(t *testing.T) {
	aberto := `
		.....
		.....
		.....
		.....
		1...2
	`
	casos := []struct {
		nome   string
		estado func(t *testing.T) jogo.Estado
		eu     string      // jogador que entra no estado auxiliar (D4); vazio para nenhum
		acoes  []jogo.Acao // ações dele
		chamas map[int][]jogo.Posicao
		perigo []jogo.Posicao
	}{
		{
			nome:   "BOM-05 bomba de pavio 2 explode na etapa 2",
			estado: func(t *testing.T) jogo.Estado { return montar(t, aberto, comBomba(2, 2, "jogador_2", 1, 2)) },
			chamas: map[int][]jogo.Posicao{2: cruz(2, 2, 1, 5, 5)},
		},
		{
			nome: "BOM-04 pilha de duas bombas de potência 1 alcança 2",
			estado: func(t *testing.T) jogo.Estado {
				return montar(t, aberto, comBomba(2, 2, "jogador_2", 1, 1), comBomba(2, 2, "jogador_2", 1, 4))
			},
			chamas: map[int][]jogo.Posicao{1: cruz(2, 2, 2, 5, 5)},
		},
		{
			nome: "BOM-09 reação em cadeia na etapa 1",
			estado: func(t *testing.T) jogo.Estado {
				return montar(t, "1....2", comBomba(0, 0, "jogador_2", 1, 1), comBomba(1, 0, "jogador_2", 2, 6))
			},
			chamas: map[int][]jogo.Posicao{1: {{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}}},
		},
		{
			nome:   "BOM-07 bloco destrutível para o fogo",
			estado: func(t *testing.T) jogo.Estado { return montar(t, ".1+..2", comBomba(0, 0, "jogador_2", 4, 1)) },
			chamas: map[int][]jogo.Posicao{1: {{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}},
		},
		{
			nome:   "BOM-11 bomba que sobra para o turno seguinte entra na zona de perigo final",
			estado: func(t *testing.T) jogo.Estado { return montar(t, aberto, comBomba(2, 2, "jogador_2", 1, 9)) },
			perigo: cruz(2, 2, 1, 5, 5),
		},
		{
			nome: "BOM-11 zona de perigo final respeita bloco destruído no turno",
			estado: func(t *testing.T) jogo.Estado {
				return montar(t, "1+...2", comBomba(0, 0, "jogador_2", 1, 2), comBomba(2, 0, "jogador_2", 2, 9))
			},
			chamas: map[int][]jogo.Posicao{2: {{X: 0, Y: 0}, {X: 1, Y: 0}}},
			perigo: []jogo.Posicao{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}, {X: 4, Y: 0}},
		},
		{
			nome:   "BOM-05 bomba do próprio jogador plantada na etapa 1 explode na etapa 4",
			estado: func(t *testing.T) jogo.Estado { return montar(t, "1....2") },
			eu:     "jogador_1",
			acoes:  []jogo.Acao{{Etapa: 1, Tipo: jogo.Plantar}, {Etapa: 2, Tipo: jogo.Mover, Direcao: jogo.Direita}},
			chamas: map[int][]jogo.Posicao{4: {{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}},
		},
		{
			nome:   "BOM-11 bomba do próprio jogador plantada na etapa 6 fica na zona de perigo final",
			estado: func(t *testing.T) jogo.Estado { return montar(t, "1....2") },
			eu:     "jogador_1",
			acoes: []jogo.Acao{
				{Etapa: 1, Tipo: jogo.Esperar}, {Etapa: 2, Tipo: jogo.Esperar}, {Etapa: 3, Tipo: jogo.Esperar},
				{Etapa: 4, Tipo: jogo.Esperar}, {Etapa: 5, Tipo: jogo.Esperar}, {Etapa: 6, Tipo: jogo.Plantar},
			},
			perigo: []jogo.Posicao{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e := c.estado(t)
			var eu *jogo.Jogador
			if c.eu != "" {
				j, _ := encontrar(e, c.eu)
				eu = &j
			}
			l := preverLinha(e, e.EtapasNesteTurno, eu, c.acoes, nil)
			if len(l.chamas) != e.EtapasNesteTurno+1 {
				t.Fatalf("%d etapas de chamas, esperado %d", len(l.chamas)-1, e.EtapasNesteTurno)
			}
			for etapa := 1; etapa <= e.EtapasNesteTurno; etapa++ {
				if esperado := conjuntoDe(c.chamas[etapa]); !reflect.DeepEqual(l.chamas[etapa], esperado) {
					t.Errorf("chamas da etapa %d: %v, esperado %v", etapa, l.chamas[etapa], esperado)
				}
			}
			if esperado := conjuntoDe(c.perigo); !reflect.DeepEqual(l.perigo, esperado) {
				t.Errorf("perigo: %v, esperado %v", l.perigo, esperado)
			}
		})
	}
}

func TestPreverLinhaIgnoraJogadores(t *testing.T) {
	// D2: o jogador atingido morre na simulação real e a partida acaba (DEC-06),
	// mas a linha do tempo continua até a última etapa.
	e := montar(t, "1...2", comBomba(0, 0, "jogador_2", 1, 1), comBomba(4, 0, "jogador_1", 1, 5))
	l := preverLinha(e, e.EtapasNesteTurno, nil, nil, nil)
	if !l.chamas[5][jogo.Posicao{X: 4, Y: 0}] {
		t.Errorf("DEC-06 chamas da etapa 5 ausentes: %v", l.chamas[5])
	}
}
