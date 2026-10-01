package jogo

import "testing"

func TestEnums(t *testing.T) {
	casos := []struct {
		nome     string
		obtido   string
		esperado string
	}{
		{"ACA-01 tipo MOVER", string(Mover), "MOVER"},
		{"ACA-01 tipo PLANTAR", string(Plantar), "PLANTAR"},
		{"ACA-01 tipo ESPERAR", string(Esperar), "ESPERAR"},
		{"TAB-04 direção CIMA", string(Cima), "CIMA"},
		{"TAB-04 direção BAIXO", string(Baixo), "BAIXO"},
		{"TAB-04 direção ESQUERDA", string(Esquerda), "ESQUERDA"},
		{"TAB-04 direção DIREITA", string(Direita), "DIREITA"},
		{"EST-06 status VIVO", string(Vivo), "VIVO"},
		{"EST-06 status MORTO", string(Morto), "MORTO"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if c.obtido != c.esperado {
				t.Errorf("obtido %q, esperado %q", c.obtido, c.esperado)
			}
		})
	}
}
