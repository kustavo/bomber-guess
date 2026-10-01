package jogo

import "testing"

func TestEstadoIdaEVolta(t *testing.T) {
	casos := []struct {
		nome  string
		dados []byte
	}{
		{"EST-01 a EST-08 e TAB-02 exemplo de docs/REGRAS.md seção 2", blocoJSON(t, "REGRAS.md", "## 2.")},
		{"EST-07 jogador vivo sem morte", []byte(`{"turno":1,"config":{"largura":3,"altura":3,"limite_turnos":1,"prazo_planejamento_ms":1,"duracao_etapa_ms":1},"etapas_neste_turno":1,"blocos_fixos":[],"blocos_destrutiveis":[],"bombas":[],"jogadores":[{"id":"jogador_1","posicao":{"x":0,"y":0},"status":"VIVO","bombas_por_turno":1,"potencia":1,"pavio_padrao":1,"acoes_por_turno":1,"bot_versao":"v"}]}`)},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			idaEVolta[Estado](t, c.dados)
		})
	}
}

func TestJogadorOmiteMorte(t *testing.T) {
	casos := []struct {
		nome     string
		jogador  Jogador
		temMorte bool
	}{
		{"EST-07 vivo não tem morte", Jogador{ID: "jogador_1", Status: Vivo}, false},
		{"EST-07 morto tem morte", Jogador{ID: "jogador_1", Status: Morto, Morte: &Morte{Turno: 4, Etapa: 6}}, true},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			_, tem := chaves(t, c.jogador)["morte"]
			if tem != c.temMorte {
				t.Errorf("chave morte presente = %v, esperado %v", tem, c.temMorte)
			}
		})
	}
}
