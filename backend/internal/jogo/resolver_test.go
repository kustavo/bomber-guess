package jogo

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// estadoMovimentado tem um pouco de tudo em uma etapa: movimento bloqueado,
// movimento, plantio seguido de morte, reação em cadeia, bloco destruído e um
// jogador morto em turno anterior.
func estadoMovimentado(t *testing.T) (Estado, []Plano) {
	e := montar(t, `
		1+..+
		.....
		4.2.3
	`, comTurno(5), comMorto("jogador_4", 4, 6), comBomba(3, 2, "jogador_2", 1, 1))
	return e, []Plano{
		plano(t, e, "jogador_1", "D"),
		plano(t, e, "jogador_2", "C"),
		plano(t, e, "jogador_3", "P"),
		plano(t, e, "jogador_4", "D"),
	}
}

func TestRelatorioEtapa(t *testing.T) {
	e, planos := estadoMovimentado(t)
	_, relatorios := ResolverTurno(e, planos)
	esperado := RelatorioEtapa{
		Turno: 5,
		Etapa: 1,
		Jogadores: []JogadorEtapa{
			{ID: "jogador_1", Posicao: p(0, 0), Status: Vivo, Acao: Acao{1, Mover, Direita}, Resultado: Bloqueada},
			{ID: "jogador_2", Posicao: p(2, 1), Status: Vivo, Acao: Acao{1, Mover, Cima}, Resultado: Executada},
			{ID: "jogador_3", Posicao: p(4, 2), Status: Morto, Acao: Acao{1, Plantar, ""}, Resultado: Executada},
			{ID: "jogador_4", Posicao: p(0, 2), Status: Morto, Acao: Acao{1, Esperar, ""}, Resultado: Descartada},
		},
		Bombas: []Bomba{},
		Explosoes: []Explosao{
			{Origem: p(3, 2), Potencia: 1, Chamas: []Posicao{p(3, 1), p(2, 2), p(3, 2), p(4, 2)}},
			{Origem: p(4, 2), Potencia: 2, Chamas: []Posicao{p(4, 0), p(4, 1), p(3, 2), p(4, 2)}},
		},
		Chamas:               []Posicao{p(4, 0), p(3, 1), p(4, 1), p(2, 2), p(3, 2), p(4, 2)},
		Mortes:               []string{"jogador_3"},
		BlocosDestruidos:     []Posicao{p(4, 0)},
		MovimentosBloqueados: []string{"jogador_1"},
		BlocosFechados:       []Posicao{},
	}
	t.Run("DEC-09 RES-04 relatório completo da etapa", func(t *testing.T) {
		if r := relatorioDaEtapa(t, relatorios, 1); !reflect.DeepEqual(r, esperado) {
			t.Errorf("relatório:\n %+v\n esperado:\n %+v", r, esperado)
		}
	})
	t.Run("DEC-09 listas vazias saem como [] no JSON, nunca null", func(t *testing.T) {
		for _, r := range relatorios {
			dados, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(dados), "null") {
				t.Errorf("etapa %d: %s", r.Etapa, dados)
			}
		}
	})
}

func TestResolverRobusto(t *testing.T) {
	desenho := `
		1#..
		....
		...2
	`
	casos := []struct {
		nome       string
		planos     func(e Estado) []Plano
		resultados []ResultadoAcao // de jogador_1, etapa a etapa
		posicao    Posicao         // de jogador_1 ao fim do turno
	}{
		{
			nome: "RES-02 movimento para fora do tabuleiro ou bloco fixo é ignorado sem abortar",
			planos: func(e Estado) []Plano {
				return []Plano{plano(t, e, "jogador_1", "C D B")}
			},
			resultados: []ResultadoAcao{Ignorada, Ignorada, Executada},
			posicao:    p(0, 1),
		},
		{
			nome: "RES-02 BOM-02 bomba além do limite é ignorada",
			planos: func(e Estado) []Plano {
				return []Plano{plano(t, e, "jogador_1", "P P P")}
			},
			resultados: []ResultadoAcao{Executada, Executada, Ignorada},
			posicao:    p(0, 0),
		},
		{
			nome: "RES-02 tipo e direção desconhecidos são ignorados",
			planos: func(e Estado) []Plano {
				return []Plano{{JogadorID: "jogador_1", Turno: 1, Acoes: []Acao{{1, "PULAR", ""}, {2, Mover, "NORTE"}, {3, Mover, Baixo}}}}
			},
			resultados: []ResultadoAcao{Ignorada, Ignorada, Executada},
			posicao:    p(0, 1),
		},
		{
			nome: "RES-01 planos de jogador inexistente são ignorados e vale o primeiro de cada jogador",
			planos: func(e Estado) []Plano {
				return []Plano{plano(t, e, "jogador_9", "D"), plano(t, e, "jogador_1", "B"), plano(t, e, "jogador_1", "B B")}
			},
			resultados: []ResultadoAcao{Executada, Executada},
			posicao:    p(0, 1),
		},
		{
			nome: "RES-02 ACA-02 ações além de acoes_por_turno são descartadas",
			planos: func(e Estado) []Plano {
				return []Plano{plano(t, e, "jogador_1", "B B B B B B B B B")}
			},
			resultados: []ResultadoAcao{Executada, Executada, Ignorada, Ignorada, Ignorada, Ignorada, Ignorada},
			posicao:    p(0, 2),
		},
		{
			nome: "RES-02 etapa e turno das ações são ignorados",
			planos: func(e Estado) []Plano {
				return []Plano{{JogadorID: "jogador_1", Turno: 99, Acoes: []Acao{{7, Mover, Baixo}, {7, Mover, Direita}}}}
			},
			resultados: []ResultadoAcao{Executada, Executada},
			posicao:    p(1, 1),
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e := montar(t, desenho)
			novo, relatorios := ResolverTurno(e, c.planos(e))
			for i, esperado := range c.resultados {
				if j := situacao(t, relatorioDaEtapa(t, relatorios, i+1), "jogador_1"); j.Resultado != esperado {
					t.Errorf("etapa %d: resultado %s, esperado %s", i+1, j.Resultado, esperado)
				}
			}
			if j, _ := buscarJogador(novo, "jogador_1"); j.Posicao != c.posicao {
				t.Errorf("posição final %s, esperado %s", j.Posicao, c.posicao)
			}
		})
	}
}

func TestResolverPuroEDeterministico(t *testing.T) {
	e, planos := estadoMovimentado(t)
	estadoAntes := e.Copiar()
	planosAntes := make([]Plano, len(planos))
	for i, p := range planos {
		planosAntes[i] = copiarPlano(p)
	}
	novo1, rel1 := ResolverTurno(e, planos)
	novo2, rel2 := ResolverTurno(e, planos)
	invertidos := slices.Clone(planos)
	slices.Reverse(invertidos)
	novo3, rel3 := ResolverTurno(e, invertidos)

	casos := []struct {
		nome string
		ok   bool
	}{
		{"RES-03 mesma entrada gera a mesma saída", reflect.DeepEqual(novo1, novo2) && reflect.DeepEqual(rel1, rel2)},
		{"RES-03 a ordem dos planos não muda o resultado", reflect.DeepEqual(novo1, novo3) && reflect.DeepEqual(rel1, rel3)},
		{"RES-03 o estado e os planos recebidos não são alterados", reflect.DeepEqual(e, estadoAntes) && reflect.DeepEqual(planos, planosAntes)},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if !c.ok {
				t.Fail()
			}
		})
	}
	t.Run("RES-03 alterar o novo estado não altera a entrada", func(t *testing.T) {
		novo1.BlocosDestrutiveis[0] = p(9, 9)
		novo1.Jogadores[3].Morte.Turno = 99
		if !reflect.DeepEqual(e, estadoAntes) {
			t.Error("o novo estado compartilha memória com a entrada")
		}
	})
}

func TestExemplosDeArquitetura(t *testing.T) {
	t.Run("VAL-05 exemplo de Infracao de docs/ARQUITETURA.md 1.1", func(t *testing.T) {
		idaEVolta[Infracao](t, blocoJSON(t, "ARQUITETURA.md", "### 1.1"))
	})
	t.Run("DEC-09 exemplo de RelatorioEtapa de docs/ARQUITETURA.md 1.2", func(t *testing.T) {
		idaEVolta[RelatorioEtapa](t, blocoJSON(t, "ARQUITETURA.md", "### 1.2"))
	})
}
