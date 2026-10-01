package jogo

import "testing"

func TestPlanoIdaEVolta(t *testing.T) {
	idaEVolta[Plano](t, blocoJSON(t, "REGRAS.md", "## 3."))
}

func TestAcaoOmiteDirecao(t *testing.T) {
	casos := []struct {
		nome       string
		acao       Acao
		temDirecao bool
	}{
		{"ACA-01 MOVER tem direcao", Acao{Etapa: 1, Tipo: Mover, Direcao: Baixo}, true},
		{"ACA-01 PLANTAR não tem direcao", Acao{Etapa: 1, Tipo: Plantar}, false},
		{"ACA-01 ESPERAR não tem direcao", Acao{Etapa: 1, Tipo: Esperar}, false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			_, tem := chaves(t, c.acao)["direcao"]
			if tem != c.temDirecao {
				t.Errorf("chave direcao presente = %v, esperado %v", tem, c.temDirecao)
			}
		})
	}
}
