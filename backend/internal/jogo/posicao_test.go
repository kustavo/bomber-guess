package jogo

import "testing"

func TestVizinha(t *testing.T) {
	centro := Posicao{X: 5, Y: 5}
	casos := []struct {
		nome     string
		direcao  Direcao
		esperado Posicao
	}{
		{"TAB-03 TAB-04 CIMA diminui y", Cima, Posicao{X: 5, Y: 4}},
		{"TAB-03 TAB-04 BAIXO aumenta y", Baixo, Posicao{X: 5, Y: 6}},
		{"TAB-03 TAB-04 ESQUERDA diminui x", Esquerda, Posicao{X: 4, Y: 5}},
		{"TAB-03 TAB-04 DIREITA aumenta x", Direita, Posicao{X: 6, Y: 5}},
		{"TAB-04 direção desconhecida devolve a própria posição", Direcao("DIAGONAL"), centro},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if obtido := centro.Vizinha(c.direcao); obtido != c.esperado {
				t.Errorf("obtido %+v, esperado %+v", obtido, c.esperado)
			}
		})
	}
}

func TestNoTabuleiro(t *testing.T) {
	config := Config{Largura: 15, Altura: 13}
	casos := []struct {
		nome     string
		posicao  Posicao
		esperado bool
	}{
		{"TAB-01 origem está dentro", Posicao{X: 0, Y: 0}, true},
		{"TAB-01 canto oposto está dentro", Posicao{X: 14, Y: 12}, true},
		{"TAB-01 x negativo está fora", Posicao{X: -1, Y: 0}, false},
		{"TAB-01 y negativo está fora", Posicao{X: 0, Y: -1}, false},
		{"TAB-01 x igual à largura está fora", Posicao{X: 15, Y: 0}, false},
		{"TAB-01 y igual à altura está fora", Posicao{X: 0, Y: 13}, false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if obtido := config.NoTabuleiro(c.posicao); obtido != c.esperado {
				t.Errorf("obtido %v, esperado %v", obtido, c.esperado)
			}
		})
	}
}
