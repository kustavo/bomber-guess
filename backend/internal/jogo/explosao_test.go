package jogo

import (
	"reflect"
	"slices"
	"testing"
)

// p é um atalho para Posicao nas tabelas.
func p(x, y int) Posicao { return Posicao{X: x, Y: y} }

func TestExplosao(t *testing.T) {
	// Sem planos: as bombas têm pavio 1 e explodem na etapa 1.
	casos := []struct {
		nome       string
		estado     Estado
		chamas     []Posicao // ordenadas por y e depois x
		destruidos []Posicao
		mortes     []string
		explosoes  int
	}{
		{
			nome: "BOM-06 potência 2 em campo aberto alcança 9 casas",
			estado: montar(t, `
				1....
				.....
				.....
				.....
				....2
			`, comBomba(2, 2, "jogador_1", 2, 1)),
			chamas:     []Posicao{p(2, 0), p(2, 1), p(0, 2), p(1, 2), p(2, 2), p(3, 2), p(4, 2), p(2, 3), p(2, 4)},
			destruidos: []Posicao{},
			mortes:     []string{},
			explosoes:  1,
		},
		{
			nome: "BOM-06 TAB-01 a explosão não passa da borda",
			estado: montar(t, `
				....1
				.....
				.....
				.....
				....2
			`, comBomba(1, 2, "jogador_1", 2, 1)),
			chamas:     []Posicao{p(1, 0), p(1, 1), p(0, 2), p(1, 2), p(2, 2), p(3, 2), p(1, 3), p(1, 4)},
			destruidos: []Posicao{},
			mortes:     []string{},
			explosoes:  1,
		},
		{
			// Bomba de potência 3 em (3,3): fixo a 1 casa acima, destrutível a 1 casa
			// à esquerda (com outro atrás), a 2 à direita e a 3 abaixo.
			nome: "BOM-07 bloco fixo e blocos destrutíveis param o fogo",
			estado: montar(t, `
				1......
				.......
				...#...
				.++..+.
				.......
				.......
				...+..2
			`, comBomba(3, 3, "jogador_1", 3, 1)),
			chamas:     []Posicao{p(2, 3), p(3, 3), p(4, 3), p(5, 3), p(3, 4), p(3, 5), p(3, 6)},
			destruidos: []Posicao{p(2, 3), p(5, 3), p(3, 6)},
			mortes:     []string{},
			explosoes:  1,
		},
		{
			// A (potência 3) em (1,1) atinge B (potência 1) em (2,1).
			nome: "BOM-07 BOM-09 o fogo para na bomba atingida, que explode com a própria potência",
			estado: montar(t, `
				......1
				.......
				......2
			`, comBomba(1, 1, "jogador_1", 3, 1), comBomba(2, 1, "jogador_2", 1, 9)),
			chamas:     []Posicao{p(1, 0), p(2, 0), p(0, 1), p(1, 1), p(2, 1), p(3, 1), p(1, 2), p(2, 2)},
			destruidos: []Posicao{},
			mortes:     []string{},
			explosoes:  2,
		},
		{
			nome: "BOM-09 cadeia A, B, C na mesma etapa",
			estado: montar(t, `
				......1
				.......
				......2
			`, comBomba(0, 1, "jogador_1", 1, 1), comBomba(1, 1, "jogador_1", 1, 9), comBomba(2, 1, "jogador_2", 1, 9)),
			chamas:     []Posicao{p(0, 0), p(1, 0), p(2, 0), p(0, 1), p(1, 1), p(2, 1), p(3, 1), p(0, 2), p(1, 2), p(2, 2)},
			destruidos: []Posicao{},
			mortes:     []string{},
			explosoes:  3,
		},
		{
			nome: "BOM-08 ORD-05 a explosão atravessa jogadores e mata quem está na casa da bomba",
			estado: montar(t, `
				......4
				312....
				.......
			`, comBomba(0, 1, "jogador_4", 3, 1)),
			chamas:     []Posicao{p(0, 0), p(0, 1), p(1, 1), p(2, 1), p(3, 1), p(0, 2)},
			destruidos: []Posicao{},
			mortes:     []string{"jogador_1", "jogador_2", "jogador_3"},
			explosoes:  1,
		},
		{
			nome: "BOM-12 bomba de jogador morto explode normalmente",
			estado: montar(t, `
				1...3
				.....
				....2
			`, comMorto("jogador_3", 1, 1), comBomba(2, 1, "jogador_3", 1, 1)),
			chamas:     []Posicao{p(2, 0), p(1, 1), p(2, 1), p(3, 1), p(2, 2)},
			destruidos: []Posicao{},
			mortes:     []string{},
			explosoes:  1,
		},
		{
			// A em (1,1) e B em (3,1), com um bloco destrutível entre elas.
			nome: "DEC-01 ORD-06 bloco atingido por duas explosões conta uma vez e para as duas",
			estado: montar(t, `
				......1
				..+....
				......2
			`, comBomba(1, 1, "jogador_1", 2, 1), comBomba(3, 1, "jogador_2", 2, 1)),
			chamas:     []Posicao{p(1, 0), p(3, 0), p(0, 1), p(1, 1), p(2, 1), p(3, 1), p(4, 1), p(5, 1), p(1, 2), p(3, 2)},
			destruidos: []Posicao{p(2, 1)},
			mortes:     []string{},
			explosoes:  2,
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			novo, relatorios := ResolverTurno(c.estado, nil)
			r := relatorioDaEtapa(t, relatorios, 1)
			if !reflect.DeepEqual(r.Chamas, c.chamas) {
				t.Errorf("chamas = %v\n esperado %v", r.Chamas, c.chamas)
			}
			if !reflect.DeepEqual(r.BlocosDestruidos, c.destruidos) {
				t.Errorf("blocos destruídos = %v, esperado %v", r.BlocosDestruidos, c.destruidos)
			}
			if !reflect.DeepEqual(r.Mortes, c.mortes) {
				t.Errorf("mortes = %v, esperado %v", r.Mortes, c.mortes)
			}
			if len(r.Explosoes) != c.explosoes {
				t.Errorf("%d explosões, esperado %d", len(r.Explosoes), c.explosoes)
			}
			if len(r.Bombas) != 0 {
				t.Errorf("BOM-09 sobraram bombas: %+v", r.Bombas)
			}
			for _, b := range c.destruidos {
				if slices.Contains(novo.BlocosDestrutiveis, b) {
					t.Errorf("ORD-06 bloco %s continua no novo estado", b)
				}
			}
		})
	}
}

func TestExplosaoPorOrigem(t *testing.T) {
	casos := []struct {
		nome   string
		estado Estado
		origem Posicao
		chamas []Posicao
	}{
		{
			nome: "BOM-09 o fogo da primeira explosão não continua além da bomba atingida",
			estado: montar(t, `
				......1
				.......
				......2
			`, comBomba(1, 1, "jogador_1", 3, 1), comBomba(2, 1, "jogador_2", 1, 9)),
			origem: p(1, 1),
			chamas: []Posicao{p(1, 0), p(0, 1), p(1, 1), p(2, 1), p(1, 2)},
		},
		{
			nome: "DEC-01 o fogo de cada explosão para no bloco destrutível",
			estado: montar(t, `
				......1
				..+....
				......2
			`, comBomba(1, 1, "jogador_1", 2, 1), comBomba(3, 1, "jogador_2", 2, 1)),
			origem: p(1, 1),
			chamas: []Posicao{p(1, 0), p(0, 1), p(1, 1), p(2, 1), p(1, 2)},
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			_, relatorios := ResolverTurno(c.estado, nil)
			for _, e := range relatorioDaEtapa(t, relatorios, 1).Explosoes {
				if e.Origem == c.origem {
					if !reflect.DeepEqual(e.Chamas, c.chamas) {
						t.Errorf("chamas = %v, esperado %v", e.Chamas, c.chamas)
					}
					return
				}
			}
			t.Errorf("nenhuma explosão com origem %s", c.origem)
		})
	}
}

func TestBombaPlantadaNaEtapaDaExplosao(t *testing.T) {
	// A explosão de (0,1) atinge (2,1), onde jogador_1 planta na mesma etapa.
	e := montar(t, `
		.....2
		..1...
		.....3
	`, comBomba(0, 1, "jogador_2", 2, 1))
	_, relatorios := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", "P")})
	t.Run("BOM-09 BOM-05 bomba recém-plantada atingida explode em cadeia", func(t *testing.T) {
		r := relatorioDaEtapa(t, relatorios, 1)
		if len(r.Explosoes) != 2 || r.Explosoes[1].Origem != p(2, 1) {
			t.Errorf("explosões = %+v, esperado (0,1) e (2,1)", r.Explosoes)
		}
		if !slices.Contains(r.Chamas, p(4, 1)) {
			t.Errorf("chamas %v não incluem o alcance da bomba nova", r.Chamas)
		}
	})
}

func TestExplosoesSimultaneasIndependemDaOrdem(t *testing.T) {
	// A (potência 3) em (0,1) e B (potência 1) em (1,1) explodem juntas. Se B
	// fosse retirada antes, o fogo de A chegaria a (3,1), onde está jogador_3.
	desenho := `
		......1
		...3...
		......2
	`
	a := comBomba(0, 1, "jogador_1", 3, 1)
	b := comBomba(1, 1, "jogador_2", 1, 1)
	_, r1 := ResolverTurno(montar(t, desenho, a, b), nil)
	_, r2 := ResolverTurno(montar(t, desenho, b, a), nil)
	t.Run("DEC-05 bomba que explode na mesma etapa ainda para o fogo", func(t *testing.T) {
		r := relatorioDaEtapa(t, r1, 1)
		if slices.Contains(r.Chamas, p(3, 1)) || len(r.Mortes) != 0 {
			t.Errorf("o fogo de A passou de B: chamas %v, mortes %v", r.Chamas, r.Mortes)
		}
	})
	t.Run("DEC-05 RES-03 a ordem das bombas no estado não muda o resultado", func(t *testing.T) {
		x, y := relatorioDaEtapa(t, r1, 1), relatorioDaEtapa(t, r2, 1)
		if !reflect.DeepEqual(x.Chamas, y.Chamas) || !reflect.DeepEqual(x.BlocosDestruidos, y.BlocosDestruidos) ||
			!reflect.DeepEqual(x.Mortes, y.Mortes) || !reflect.DeepEqual(x.Explosoes, y.Explosoes) {
			t.Errorf("resultados diferentes:\n %+v\n %+v", x, y)
		}
	})
}

func TestFogoDuraUmaEtapa(t *testing.T) {
	e := montar(t, `
		.....2
		..1...
		.....3
	`, comBomba(0, 1, "jogador_2", 1, 1))
	novo, _ := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", ". E")})
	t.Run("BOM-10 entrar na etapa 2 em casa atingida na etapa 1 não mata", func(t *testing.T) {
		j, _ := buscarJogador(novo, "jogador_1")
		if j.Status != Vivo || j.Posicao != p(1, 1) {
			t.Errorf("jogador_1 = %+v, esperado vivo em (1,1)", j)
		}
	})
}
