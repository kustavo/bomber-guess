package jogo

import (
	"reflect"
	"testing"
)

func estadoDeTeste() Estado {
	return Estado{
		Turno:              5,
		Config:             Config{Largura: 15, Altura: 13},
		EtapasNesteTurno:   7,
		BlocosFixos:        []Posicao{{X: 1, Y: 1}},
		BlocosDestrutiveis: []Posicao{{X: 2, Y: 4}},
		Bombas:             []Bomba{{Posicao: Posicao{X: 5, Y: 6}, JogadorID: "jogador_2", Potencia: 2, PavioRestante: 3}},
		Jogadores: []Jogador{
			{ID: "jogador_1", Status: Vivo, Atributos: Atributos{AcoesPorTurno: 7}},
			{ID: "jogador_2", Status: Morto, Morte: &Morte{Turno: 4, Etapa: 6}},
		},
	}
}

func TestCopiar(t *testing.T) {
	casos := []struct {
		nome    string
		alterar func(e *Estado)
	}{
		{"BOT-01 campo simples", func(e *Estado) { e.Turno = 99; e.Config.Largura = 1 }},
		{"BOT-01 blocos fixos", func(e *Estado) { e.BlocosFixos[0].X = 9 }},
		{"BOT-01 blocos destrutíveis", func(e *Estado) { e.BlocosDestrutiveis[0].Y = 9 }},
		{"BOT-01 bombas", func(e *Estado) { e.Bombas[0].PavioRestante = 0 }},
		{"BOT-01 jogadores", func(e *Estado) { e.Jogadores[0].Posicao.X = 9; e.Jogadores[0].AcoesPorTurno = 1 }},
		{"BOT-01 EST-07 morte", func(e *Estado) { e.Jogadores[1].Morte.Etapa = 1 }},
		{"BOT-01 append nos slices", func(e *Estado) {
			e.BlocosFixos = append(e.BlocosFixos[:0], Posicao{X: 7})
			e.Bombas = append(e.Bombas[:0], Bomba{JogadorID: "x"})
			e.Jogadores = append(e.Jogadores[:0], Jogador{ID: "x"})
		}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			original := estadoDeTeste()
			copia := original.Copiar()
			if !reflect.DeepEqual(original, copia) {
				t.Fatalf("cópia diferente do original\n original: %+v\n cópia:    %+v", original, copia)
			}
			c.alterar(&copia)
			if !reflect.DeepEqual(original, estadoDeTeste()) {
				t.Errorf("alterar a cópia mudou o original: %+v", original)
			}
		})
	}
}

func TestCopiarSlicesNaoNulos(t *testing.T) {
	copia := Estado{}.Copiar()
	if copia.BlocosFixos == nil || copia.BlocosDestrutiveis == nil || copia.Bombas == nil || copia.Jogadores == nil {
		t.Errorf("BOT-01 (D4) cópia de estado vazio tem slice nulo: %+v", copia)
	}
}
