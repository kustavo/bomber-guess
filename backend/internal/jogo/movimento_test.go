package jogo

import (
	"reflect"
	"testing"
)

// relatorioDaEtapa devolve o relatório da etapa n (1, 2, 3…) ou falha o teste.
func relatorioDaEtapa(t *testing.T, relatorios []RelatorioEtapa, n int) RelatorioEtapa {
	t.Helper()
	if n < 1 || n > len(relatorios) {
		t.Fatalf("etapa %d pedida, mas há %d relatórios", n, len(relatorios))
	}
	return relatorios[n-1]
}

// situacao devolve a situação do jogador no relatório ou falha o teste.
func situacao(t *testing.T, r RelatorioEtapa, id string) JogadorEtapa {
	t.Helper()
	for _, j := range r.Jogadores {
		if j.ID == id {
			return j
		}
	}
	t.Fatalf("jogador %s ausente do relatório da etapa %d", id, r.Etapa)
	return JogadorEtapa{}
}

// esperado é o que se espera de um jogador ao fim de uma etapa.
type esperado struct {
	etapa     int
	id        string
	posicao   Posicao
	resultado ResultadoAcao
}

func conferir(t *testing.T, relatorios []RelatorioEtapa, esperados []esperado) {
	t.Helper()
	for _, e := range esperados {
		j := situacao(t, relatorioDaEtapa(t, relatorios, e.etapa), e.id)
		if j.Posicao != e.posicao || j.Resultado != e.resultado {
			t.Errorf("etapa %d, %s: posição %s resultado %s; esperado %s %s",
				e.etapa, e.id, j.Posicao, j.Resultado, e.posicao, e.resultado)
		}
	}
}

func TestMovimento(t *testing.T) {
	casos := []struct {
		nome      string
		estado    Estado
		planos    func(e Estado) []Plano
		esperados []esperado
	}{
		{
			nome:   "MOV-04 MOV-01 entrar na casa de outro jogador",
			estado: montar(t, "12.", comAcoes(1)),
			planos: func(e Estado) []Plano { return []Plano{plano(t, e, "jogador_1", "D")} },
			esperados: []esperado{
				{1, "jogador_1", Posicao{1, 0}, Executada},
				{1, "jogador_2", Posicao{1, 0}, Executada},
			},
		},
		{
			nome:   "MOV-04 dois jogadores trocam de casa",
			estado: montar(t, "12.", comAcoes(1)),
			planos: func(e Estado) []Plano {
				return []Plano{plano(t, e, "jogador_1", "D"), plano(t, e, "jogador_2", "E")}
			},
			esperados: []esperado{
				{1, "jogador_1", Posicao{1, 0}, Executada},
				{1, "jogador_2", Posicao{0, 0}, Executada},
			},
		},
		{
			nome:      "MOV-04 entrar em casa com bomba",
			estado:    montar(t, "1..2", comAcoes(1), comBomba(1, 0, "jogador_2", 1, 5)),
			planos:    func(e Estado) []Plano { return []Plano{plano(t, e, "jogador_1", "D")} },
			esperados: []esperado{{1, "jogador_1", Posicao{1, 0}, Executada}},
		},
		{
			nome: "MOV-05 ORD-01 movimento bloqueado aborta o resto do plano",
			estado: montar(t, `
				1+.
				...
				..2
			`, comAcoes(3)),
			planos: func(e Estado) []Plano { return []Plano{plano(t, e, "jogador_1", "D P B")} },
			esperados: []esperado{
				{1, "jogador_1", Posicao{0, 0}, Bloqueada},
				{2, "jogador_1", Posicao{0, 0}, Abortada},
				{3, "jogador_1", Posicao{0, 0}, Abortada},
			},
		},
		{
			// A bomba em (2,0) explode na etapa 3 e destrói o bloco em (1,0).
			nome: "MOV-03 ORD-06 bloco destruído na etapa 3 é atravessado na etapa 4",
			estado: montar(t, `
				1+..
				....
				...2
			`, comAcoes(4), comBomba(2, 0, "jogador_2", 1, 3)),
			planos: func(e Estado) []Plano { return []Plano{plano(t, e, "jogador_1", ". . . D")} },
			esperados: []esperado{
				{3, "jogador_1", Posicao{0, 0}, Executada},
				{4, "jogador_1", Posicao{1, 0}, Executada},
			},
		},
		{
			nome: "MOV-03 ORD-01 ORD-06 bloco destruído na etapa 3 ainda bloqueia na etapa 3",
			estado: montar(t, `
				1+..
				....
				...2
			`, comAcoes(4), comBomba(2, 0, "jogador_2", 1, 3)),
			planos: func(e Estado) []Plano { return []Plano{plano(t, e, "jogador_1", ". . D D")} },
			esperados: []esperado{
				{3, "jogador_1", Posicao{0, 0}, Bloqueada},
				{4, "jogador_1", Posicao{0, 0}, Abortada},
			},
		},
		{
			nome: "ACA-03 EST-03 jogador com menos ações espera nas etapas restantes",
			estado: montar(t, `
				1....
				....2
			`, comAtributos("jogador_1", func(a *Atributos) { a.AcoesPorTurno = 5 }),
				comAtributos("jogador_2", func(a *Atributos) { a.AcoesPorTurno = 3 })),
			planos: func(e Estado) []Plano {
				return []Plano{plano(t, e, "jogador_1", "D D D D"), plano(t, e, "jogador_2", "E E E")}
			},
			esperados: []esperado{
				{3, "jogador_2", Posicao{1, 1}, Executada},
				{4, "jogador_2", Posicao{1, 1}, Executada},
				{5, "jogador_2", Posicao{1, 1}, Executada},
				{4, "jogador_1", Posicao{4, 0}, Executada},
				{5, "jogador_1", Posicao{4, 0}, Executada},
			},
		},
		{
			nome:   "RES-01 jogador sem plano espera em todas as etapas",
			estado: montar(t, "1..2", comAcoes(3)),
			planos: func(e Estado) []Plano { return []Plano{plano(t, e, "jogador_1", "D")} },
			esperados: []esperado{
				{1, "jogador_2", Posicao{3, 0}, Executada},
				{2, "jogador_2", Posicao{3, 0}, Executada},
				{3, "jogador_2", Posicao{3, 0}, Executada},
			},
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			_, relatorios := ResolverTurno(c.estado, c.planos(c.estado))
			if len(relatorios) != c.estado.EtapasNesteTurno {
				t.Errorf("%d relatórios, esperado %d", len(relatorios), c.estado.EtapasNesteTurno)
			}
			conferir(t, relatorios, c.esperados)
		})
	}
}

func TestMovimentoBloqueadoNoRelatorio(t *testing.T) {
	e := montar(t, `
		1+.
		..2
	`, comAcoes(3))
	_, relatorios := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", "D P D")})
	casos := []struct {
		nome       string
		etapa      int
		bloqueados []string
		acao       Acao
	}{
		{"MOV-05 etapa 1 registra o movimento bloqueado", 1, []string{"jogador_1"}, Acao{1, Mover, Direita}},
		{"MOV-05 etapa 2 traz a ação abortada do plano", 2, []string{}, Acao{2, Plantar, ""}},
		{"MOV-05 etapa 3 traz a ação abortada do plano", 3, []string{}, Acao{3, Mover, Direita}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			r := relatorioDaEtapa(t, relatorios, c.etapa)
			if !reflect.DeepEqual(r.MovimentosBloqueados, c.bloqueados) {
				t.Errorf("movimentos bloqueados = %v, esperado %v", r.MovimentosBloqueados, c.bloqueados)
			}
			if j := situacao(t, r, "jogador_1"); j.Acao != c.acao {
				t.Errorf("ação = %+v, esperado %+v", j.Acao, c.acao)
			}
			if len(r.Bombas) != 0 {
				t.Errorf("MOV-05 PLANTAR abortado não deveria plantar: %+v", r.Bombas)
			}
		})
	}
}
