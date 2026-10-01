package jogo

import (
	"reflect"
	"testing"
)

func TestPavio(t *testing.T) {
	// jogador_1 (potência 2, pavio 3) planta na etapa 2 em (1,0) e foge para a direita.
	e := montar(t, `
		.1.....
		.......
		......2
	`, comAcoes(5))
	_, relatorios := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", ". P D D D")})
	bomba := func(pavio int) []Bomba {
		return []Bomba{{Posicao: Posicao{1, 0}, JogadorID: "jogador_1", Potencia: 2, PavioRestante: pavio}}
	}
	casos := []struct {
		nome     string
		etapa    int
		bombas   []Bomba
		explodiu bool
	}{
		{"BOM-05 ORD-03 nada antes de plantar", 1, []Bomba{}, false},
		{"BOM-01 BOM-05 ORD-03 plantada na etapa 2 com pavio 3, na casa do jogador", 2, bomba(3), false},
		{"BOM-05 ORD-03 pavio 2 na etapa 3", 3, bomba(2), false},
		{"BOM-05 ORD-03 pavio 1 na etapa 4", 4, bomba(1), false},
		{"BOM-05 pavio 3 na etapa 2 explode na etapa 5", 5, []Bomba{}, true},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			r := relatorioDaEtapa(t, relatorios, c.etapa)
			if !reflect.DeepEqual(r.Bombas, c.bombas) {
				t.Errorf("bombas = %+v, esperado %+v", r.Bombas, c.bombas)
			}
			if explodiu := len(r.Explosoes) == 1 && r.Explosoes[0].Origem == (Posicao{1, 0}) && r.Explosoes[0].Potencia == 2; explodiu != c.explodiu {
				t.Errorf("explosões = %+v, esperado explodir = %v", r.Explosoes, c.explodiu)
			}
		})
	}
}

func TestBombaAtravessaTurnos(t *testing.T) {
	// Turno de 7 etapas; jogador_1 planta na etapa 6 e fica parado.
	e := montar(t, `
		1....
		.....
		....2
	`)
	novo, _ := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", ". . . . . P")})
	t.Run("BOM-11 BOM-05 bomba plantada na etapa 6 de 7 chega ao turno seguinte com pavio 2", func(t *testing.T) {
		esperado := []Bomba{{Posicao: Posicao{0, 0}, JogadorID: "jogador_1", Potencia: 2, PavioRestante: 2}}
		if !reflect.DeepEqual(novo.Bombas, esperado) {
			t.Errorf("bombas = %+v, esperado %+v", novo.Bombas, esperado)
		}
	})
	t.Run("BOM-11 explode na etapa 2 do turno seguinte", func(t *testing.T) {
		_, relatorios := ResolverTurno(novo, nil)
		if len(relatorioDaEtapa(t, relatorios, 1).Explosoes) != 0 {
			t.Error("explodiu na etapa 1")
		}
		if len(relatorioDaEtapa(t, relatorios, 2).Explosoes) != 1 {
			t.Error("não explodiu na etapa 2")
		}
	})
}

func TestPilha(t *testing.T) {
	// Pilha em (5,1) de um tabuleiro 11×3, com pavio 1: explode na etapa 1.
	desenho := `
		1..........
		...........
		..........2
	`
	casos := []struct {
		nome      string
		potencias []int
		alcance   int
	}{
		{"BOM-04 bomba sozinha de potência 1", []int{1}, 1},
		{"BOM-03 BOM-04 potências 2 e 3 geram alcance 4", []int{2, 3}, 4},
		{"BOM-03 BOM-04 três bombas de potência 2 geram alcance 4", []int{2, 2, 2}, 4},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			var opcoes []opcao
			for _, p := range c.potencias {
				opcoes = append(opcoes, comBomba(5, 1, "jogador_1", p, 1))
			}
			_, relatorios := ResolverTurno(montar(t, desenho, opcoes...), nil)
			r := relatorioDaEtapa(t, relatorios, 1)
			if len(r.Explosoes) != 1 || r.Explosoes[0].Potencia != c.alcance {
				t.Fatalf("explosões = %+v, esperado uma de potência %d", r.Explosoes, c.alcance)
			}
			chamas := conjunto(r.Chamas)
			if !chamas[Posicao{5 + c.alcance, 1}] || !chamas[Posicao{5 - c.alcance, 1}] {
				t.Errorf("chamas %v não alcançam %d casas", r.Chamas, c.alcance)
			}
			if chamas[Posicao{6 + c.alcance, 1}] || chamas[Posicao{4 - c.alcance, 1}] {
				t.Errorf("chamas %v passam de %d casas", r.Chamas, c.alcance)
			}
		})
	}
}

func TestPilhaExplodeJunto(t *testing.T) {
	e := montar(t, `
		1....
		.....
		....2
	`, comBomba(2, 1, "jogador_1", 1, 1), comBomba(2, 1, "jogador_2", 1, 3))
	novo, relatorios := ResolverTurno(e, nil)
	t.Run("BOM-04 bomba de pavio 1 leva junto a de pavio 3, como uma única explosão", func(t *testing.T) {
		r := relatorioDaEtapa(t, relatorios, 1)
		if len(r.Explosoes) != 1 || r.Explosoes[0].Potencia != 2 {
			t.Errorf("explosões = %+v, esperado uma de potência 2", r.Explosoes)
		}
		if len(r.Bombas) != 0 || len(novo.Bombas) != 0 {
			t.Errorf("sobraram bombas: relatório %+v, estado %+v", r.Bombas, novo.Bombas)
		}
	})
}

func TestPlantarDuasNaMesmaCasa(t *testing.T) {
	e := montar(t, `
		1....
		....2
	`)
	_, relatorios := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", "P P")})
	t.Run("BOM-03 VAL-03 duas bombas do mesmo jogador formam uma pilha", func(t *testing.T) {
		r := relatorioDaEtapa(t, relatorios, 2)
		esperado := []Bomba{
			{Posicao: Posicao{0, 0}, JogadorID: "jogador_1", Potencia: 2, PavioRestante: 2},
			{Posicao: Posicao{0, 0}, JogadorID: "jogador_1", Potencia: 2, PavioRestante: 3},
		}
		if !reflect.DeepEqual(r.Bombas, esperado) {
			t.Errorf("bombas = %+v, esperado %+v", r.Bombas, esperado)
		}
	})
}
