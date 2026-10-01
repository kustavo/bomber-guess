package jogo

import (
	"reflect"
	"testing"
)

func TestValidar(t *testing.T) {
	// jogador_1 em (0,0); bloco fixo em (1,1); bloco destrutível em (2,1).
	base := `
		1....
		.#+..
		.....
	`
	type infracao struct {
		etapa int
		regra string
	}
	casos := []struct {
		nome     string
		estado   Estado
		plano    Plano
		esperado string    // ações validadas, no formato de acoes
		infracao *infracao // nil se não houver
	}{
		{
			nome:     "VAL-01 ACA-02 VAL-05 cinco ações com limite 3",
			estado:   montar(t, base, comAcoes(3)),
			plano:    plano(t, montar(t, base), "jogador_1", "D D E E D"),
			esperado: "D D E",
			infracao: &infracao{4, "VAL-01"},
		},
		{
			nome:     "VAL-01 ACA-01 MOVER sem direção",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_1", Turno: 1, Acoes: []Acao{{1, Mover, Direita}, {2, Mover, ""}, {3, Mover, Direita}}},
			esperado: "D . .",
			infracao: &infracao{2, "VAL-01"},
		},
		{
			nome:     "VAL-01 ACA-01 direção desconhecida",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_1", Turno: 1, Acoes: []Acao{{1, Mover, Direita}, {2, Mover, "NORTE"}, {3, Esperar, ""}}},
			esperado: "D . .",
			infracao: &infracao{2, "VAL-01"},
		},
		{
			nome:     "VAL-01 ACA-01 tipo desconhecido",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_1", Turno: 1, Acoes: []Acao{{1, Mover, Direita}, {2, "PULAR", ""}, {3, Plantar, ""}}},
			esperado: "D . .",
			infracao: &infracao{2, "VAL-01"},
		},
		{
			nome:     "VAL-02 MOV-02 VAL-05 sair do tabuleiro",
			estado:   montar(t, base),
			plano:    plano(t, montar(t, base), "jogador_1", "D C B"),
			esperado: "D . .",
			infracao: &infracao{2, "VAL-02"},
		},
		{
			nome: "VAL-02 MOV-02 entrar em bloco fixo na etapa 1",
			estado: montar(t, `
				...
				1#.
			`),
			plano:    Plano{JogadorID: "jogador_1", Turno: 1, Acoes: acoes(t, "D B")},
			esperado: ". .",
			infracao: &infracao{1, "VAL-02"},
		},
		{
			nome:     "VAL-02 posição simulada ação a ação",
			estado:   montar(t, base),
			plano:    plano(t, montar(t, base), "jogador_1", "D D E E E"),
			esperado: "D D E E .",
			infracao: &infracao{5, "VAL-02"},
		},
		{
			nome:     "VAL-04 bloco destrutível não é verificado",
			estado:   montar(t, base),
			plano:    plano(t, montar(t, base), "jogador_1", "D D B B"),
			esperado: "D D B B",
		},
		{
			nome:     "VAL-03 BOM-02 VAL-05 bomba além do limite",
			estado:   montar(t, base),
			plano:    plano(t, montar(t, base), "jogador_1", "P P . P D"),
			esperado: "P P . . .",
			infracao: &infracao{4, "VAL-03"},
		},
		{
			nome:     "VAL-03 EST-04 bombas de turnos anteriores não contam",
			estado:   montar(t, base, comBomba(4, 2, "jogador_1", 2, 2), comBomba(4, 1, "jogador_1", 2, 1)),
			plano:    plano(t, montar(t, base), "jogador_1", "P P"),
			esperado: "P P",
		},
		{
			nome:     "VAL-05 duas ações inválidas geram uma infração",
			estado:   montar(t, base),
			plano:    plano(t, montar(t, base), "jogador_1", "D C D C"),
			esperado: "D . . .",
			infracao: &infracao{2, "VAL-02"},
		},
		{
			nome:     "VAL-01 ACA-03 plano vazio",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_1", Turno: 1},
			esperado: "",
		},
		{
			nome:     "VAL-06 etapa repetida",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_1", Turno: 1, Acoes: []Acao{{1, Mover, Direita}, {1, Mover, Direita}, {3, Esperar, ""}}},
			esperado: "D . .",
			infracao: &infracao{2, "VAL-06"},
		},
		{
			nome:     "VAL-06 etapa pulada",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_1", Turno: 1, Acoes: []Acao{{1, Mover, Direita}, {3, Mover, Direita}, {4, Esperar, ""}}},
			esperado: "D . .",
			infracao: &infracao{2, "VAL-06"},
		},
		{
			nome:     "VAL-06 etapas fora de ordem",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_1", Turno: 1, Acoes: []Acao{{2, Mover, Direita}, {1, Mover, Direita}}},
			esperado: ". .",
			infracao: &infracao{1, "VAL-06"},
		},
		{
			nome:     "VAL-07 turno diferente do estado",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_1", Turno: 2, Acoes: acoes(t, "D D")},
			esperado: ". .",
			infracao: &infracao{1, "VAL-07"},
		},
		{
			nome:     "VAL-08 jogador inexistente",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_9", Turno: 1, Acoes: acoes(t, "D D")},
			esperado: "",
			infracao: &infracao{1, "VAL-08"},
		},
		{
			nome:     "VAL-08 jogador morto, sem infração",
			estado:   montar(t, base, comMorto("jogador_1", 1, 1)),
			plano:    plano(t, montar(t, base), "jogador_1", "D D"),
			esperado: "",
		},
		{
			nome:     "DEC-08 direção em PLANTAR e ESPERAR é removida",
			estado:   montar(t, base),
			plano:    Plano{JogadorID: "jogador_1", Turno: 1, Acoes: []Acao{{1, Plantar, Cima}, {2, Esperar, Baixo}}},
			esperado: "P .",
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			estadoAntes, planoAntes := c.estado.Copiar(), copiarPlano(c.plano)
			validado, infracoes := Validar(c.estado, c.plano)

			// CA-15: mesmo jogador e turno, ações numeradas, no máximo acoes_por_turno.
			if validado.JogadorID != c.plano.JogadorID || validado.Turno != c.plano.Turno {
				t.Errorf("jogador/turno = %s/%d, esperado %s/%d", validado.JogadorID, validado.Turno, c.plano.JogadorID, c.plano.Turno)
			}
			if esperado := acoes(t, c.esperado); !reflect.DeepEqual(validado.Acoes, esperado) {
				t.Errorf("ações validadas = %+v, esperado %+v", validado.Acoes, esperado)
			}
			if !reflect.DeepEqual(c.estado, estadoAntes) || !reflect.DeepEqual(c.plano, planoAntes) {
				t.Error("Validar alterou o estado ou o plano recebido")
			}

			switch {
			case c.infracao == nil && len(infracoes) != 0:
				t.Errorf("infrações = %+v, esperado nenhuma", infracoes)
			case c.infracao != nil && len(infracoes) != 1:
				t.Errorf("infrações = %+v, esperado exatamente uma", infracoes)
			case c.infracao != nil:
				i := infracoes[0]
				if i.JogadorID != c.plano.JogadorID || i.Etapa != c.infracao.etapa || i.Regra != c.infracao.regra || i.Motivo == "" {
					t.Errorf("infração = %+v, esperado etapa %d, regra %s, motivo preenchido", i, c.infracao.etapa, c.infracao.regra)
				}
			}
		})
	}
}

// copiarPlano devolve uma cópia profunda do plano.
func copiarPlano(p Plano) Plano {
	p.Acoes = append([]Acao(nil), p.Acoes...)
	return p
}
