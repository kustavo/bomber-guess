package jogo

import (
	"reflect"
	"testing"
)

func TestOrdemDaEtapa(t *testing.T) {
	// Bomba de potência 2 em (0,1) com pavio 4: explode na etapa 4.
	desenho := `
		.....2
		.1....
		.....3
	`
	bomba := comBomba(0, 1, "jogador_2", 2, 4)
	casos := []struct {
		nome   string
		seq    string
		status Status
		morte  *Morte
	}{
		{"ORD-07 ORD-01 sair do alcance na etapa da explosão salva", ". . . B", Vivo, nil},
		{"ORD-07 sair do alcance uma etapa depois não salva", ". . . . B", Morto, &Morte{Turno: 1, Etapa: 4}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e := montar(t, desenho, bomba)
			novo, _ := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", c.seq)})
			j, _ := buscarJogador(novo, "jogador_1")
			if j.Status != c.status || !reflect.DeepEqual(j.Morte, c.morte) {
				t.Errorf("jogador_1 = %s %+v, esperado %s %+v", j.Status, j.Morte, c.status, c.morte)
			}
		})
	}
}

func TestPlantarNaEtapaDaMorte(t *testing.T) {
	e := montar(t, `
		.....2
		.1....
		.....3
	`, comBomba(0, 1, "jogador_2", 2, 1))
	_, relatorios := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", "P")})
	t.Run("ORD-02 ORD-05 a bomba é colocada antes de o jogador morrer", func(t *testing.T) {
		r := relatorioDaEtapa(t, relatorios, 1)
		j := situacao(t, r, "jogador_1")
		if j.Status != Morto || j.Resultado != Executada {
			t.Errorf("jogador_1 = %+v, esperado morto com PLANTAR executado", j)
		}
		plantada := false
		for _, ex := range r.Explosoes {
			plantada = plantada || ex.Origem == p(1, 1) // a bomba nova explode em cadeia
		}
		if !plantada {
			t.Errorf("explosões = %+v, esperado uma com origem (1,1)", r.Explosoes)
		}
	})
}

func TestMorte(t *testing.T) {
	// Turno 5; bomba de turno anterior em (0,1) com pavio 3: explode na etapa 3.
	e := montar(t, `
		.....2
		.1....
		.....3
	`, comTurno(5), comBomba(0, 1, "jogador_2", 2, 3))
	novo, relatorios := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", ". . . D D")})
	j, _ := buscarJogador(novo, "jogador_1")
	casos := []struct {
		nome string
		ok   bool
	}{
		{"FIM-01 EST-07 DEC-02 morte no turno 5, etapa 3", j.Status == Morto && reflect.DeepEqual(j.Morte, &Morte{Turno: 5, Etapa: 3})},
		{"EST-08 posição é a casa onde morreu", j.Posicao == p(1, 1)},
		{"FIM-01 ações seguintes são descartadas", situacao(t, relatorioDaEtapa(t, relatorios, 4), "jogador_1").Resultado == Descartada &&
			situacao(t, relatorioDaEtapa(t, relatorios, 5), "jogador_1").Posicao == p(1, 1)},
		{"FIM-01 a morte aparece só na etapa 3", reflect.DeepEqual(relatorioDaEtapa(t, relatorios, 3).Mortes, []string{"jogador_1"})},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if !c.ok {
				t.Errorf("jogador_1 = %+v", j)
			}
		})
	}
}

func TestJogadorMortoSaiDoTabuleiro(t *testing.T) {
	// jogador_1 morreu no turno 1, em (1,1); uma explosão atinge essa casa no turno 2.
	e := montar(t, `
		.....2
		.1....
		.....3
	`, comTurno(2), comMorto("jogador_1", 1, 2), comBomba(0, 1, "jogador_2", 2, 1))
	novo, relatorios := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", "D P")})
	t.Run("FIM-01 jogador morto não é atingido de novo", func(t *testing.T) {
		if m := relatorioDaEtapa(t, relatorios, 1).Mortes; len(m) != 0 {
			t.Errorf("mortes = %v", m)
		}
		if j, _ := buscarJogador(novo, "jogador_1"); !reflect.DeepEqual(j.Morte, &Morte{Turno: 1, Etapa: 2}) {
			t.Errorf("morte alterada: %+v", j.Morte)
		}
	})
	t.Run("FIM-01 jogador morto não se move nem planta", func(t *testing.T) {
		for _, r := range relatorios {
			j := situacao(t, r, "jogador_1")
			if j.Posicao != p(1, 1) || j.Resultado != Descartada || j.Acao.Tipo != Esperar {
				t.Errorf("etapa %d: %+v", r.Etapa, j)
			}
		}
		if len(novo.Bombas) != 0 {
			t.Errorf("bombas = %+v", novo.Bombas)
		}
	})
}
