package jogo

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// mapaValido é um mapa 5×5 pequeno e válido, base dos casos inválidos.
func mapaValido() Mapa {
	return Mapa{
		Nome:               "teste",
		Config:             Config{Largura: 5, Altura: 5, LimiteTurnos: 10, PrazoPlanejamentoMs: 100, DuracaoEtapaMs: 5},
		JogadorPadrao:      Atributos{BombasPorTurno: 1, Potencia: 2, PavioPadrao: 3, AcoesPorTurno: 4},
		BlocosFixos:        []Posicao{{X: 1, Y: 1}},
		BlocosDestrutiveis: []Posicao{{X: 2, Y: 0}},
		PosicoesIniciais:   []Posicao{{X: 0, Y: 0}, {X: 4, Y: 4}, {X: 4, Y: 0}},
	}
}

func TestMapaIdaEVolta(t *testing.T) {
	idaEVolta[Mapa](t, blocoJSON(t, "EDITOR.md", "## Formato do mapa"))
}

func TestLerMapaExemploDoDoc(t *testing.T) {
	if _, err := LerMapa(blocoJSON(t, "EDITOR.md", "## Formato do mapa")); err != nil {
		t.Errorf("MAP-05 exemplo de docs/EDITOR.md deveria ser válido: %v", err)
	}
}

func TestMapaInvalido(t *testing.T) {
	casos := []struct {
		nome    string
		alterar func(m *Mapa)
		trecho  string // parte esperada da mensagem de erro
	}{
		{"MAP-05 largura zero", func(m *Mapa) { m.Config.Largura = 0 }, "largura"},
		{"MAP-05 largura negativa", func(m *Mapa) { m.Config.Largura = -3 }, "largura"},
		{"MAP-05 altura zero", func(m *Mapa) { m.Config.Altura = 0 }, "altura"},
		{"MAP-05 bloco fixo fora do tabuleiro", func(m *Mapa) { m.BlocosFixos = append(m.BlocosFixos, Posicao{X: 5, Y: 0}) }, "(5,0)"},
		{"MAP-05 bloco destrutível fora do tabuleiro", func(m *Mapa) { m.BlocosDestrutiveis = append(m.BlocosDestrutiveis, Posicao{X: 0, Y: -1}) }, "(0,-1)"},
		{"MAP-05 posição inicial fora do tabuleiro", func(m *Mapa) { m.PosicoesIniciais[2] = Posicao{X: 0, Y: 5} }, "(0,5)"},
		{"MAP-05 posição inicial sobre bloco fixo", func(m *Mapa) { m.PosicoesIniciais[2] = Posicao{X: 1, Y: 1} }, "(1,1)"},
		{"MAP-05 posição inicial sobre bloco destrutível", func(m *Mapa) { m.PosicoesIniciais[2] = Posicao{X: 2, Y: 0} }, "(2,0)"},
		{"MAP-05 casa com bloco fixo e destrutível", func(m *Mapa) { m.BlocosDestrutiveis = append(m.BlocosDestrutiveis, Posicao{X: 1, Y: 1}) }, "(1,1)"},
		{"MAP-05 uma posição inicial", func(m *Mapa) { m.PosicoesIniciais = m.PosicoesIniciais[:1] }, "posições iniciais"},
		{"MAP-05 nenhuma posição inicial", func(m *Mapa) { m.PosicoesIniciais = nil }, "posições iniciais"},
		{"MAP-05 limite_turnos zero", func(m *Mapa) { m.Config.LimiteTurnos = 0 }, "limite_turnos"},
		{"MAP-05 bombas_por_turno zero", func(m *Mapa) { m.JogadorPadrao.BombasPorTurno = 0 }, "bombas_por_turno"},
		{"MAP-05 potencia zero", func(m *Mapa) { m.JogadorPadrao.Potencia = 0 }, "potencia"},
		{"MAP-05 pavio_padrao zero", func(m *Mapa) { m.JogadorPadrao.PavioPadrao = 0 }, "pavio_padrao"},
		{"MAP-05 acoes_por_turno negativo", func(m *Mapa) { m.JogadorPadrao.AcoesPorTurno = -1 }, "acoes_por_turno"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			m := mapaValido()
			c.alterar(&m)
			for nomeFuncao, err := range map[string]error{
				"VerificarMapa": VerificarMapa(m),
				"EstadoInicial": func() error {
					e, err := EstadoInicial(m, []string{"a", "b", "c"}[:min(3, len(m.PosicoesIniciais))])
					if err != nil && !reflect.DeepEqual(e, Estado{}) {
						t.Errorf("EstadoInicial devolveu estado junto com o erro: %+v", e)
					}
					return err
				}(),
			} {
				if !errors.Is(err, ErrMapaInvalido) {
					t.Fatalf("%s: erro %v, esperado ErrMapaInvalido", nomeFuncao, err)
				}
				if !strings.Contains(err.Error(), c.trecho) {
					t.Errorf("%s: mensagem %q não cita %q", nomeFuncao, err, c.trecho)
				}
			}
		})
	}
}

func TestLerMapaInvalido(t *testing.T) {
	casos := []struct {
		nome  string
		dados string
	}{
		{"MAP-05 JSON malformado", `{"nome": "x",`},
		{"MAP-05 JSON que não é objeto", `[1, 2]`},
		{"MAP-05 JSON válido com mapa inválido", `{"config": {"largura": 0, "altura": 5}}`},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			m, err := LerMapa([]byte(c.dados))
			if !errors.Is(err, ErrMapaInvalido) {
				t.Errorf("erro %v, esperado ErrMapaInvalido", err)
			}
			if !reflect.DeepEqual(m, Mapa{}) {
				t.Errorf("LerMapa devolveu mapa junto com o erro: %+v", m)
			}
		})
	}
}

func TestEstadoInicialQuantidadeDeVersoes(t *testing.T) {
	casos := []struct {
		nome    string
		versoes []string
	}{
		{"MAP-06 menos versões que posições", []string{"a", "b"}},
		{"MAP-06 mais versões que posições", []string{"a", "b", "c", "d"}},
		{"MAP-06 nenhuma versão", nil},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e, err := EstadoInicial(mapaValido(), c.versoes)
			if !errors.Is(err, ErrMapaInvalido) || !strings.Contains(err.Error(), "versões") {
				t.Errorf("erro %v, esperado ErrMapaInvalido citando versões", err)
			}
			if !reflect.DeepEqual(e, Estado{}) {
				t.Errorf("devolveu estado junto com o erro: %+v", e)
			}
		})
	}
}

func TestEstadoInicial(t *testing.T) {
	m := mapaValido()
	versoes := []string{"claude-v1", "aleatorio-v1", "gpt-v1"}
	e, err := EstadoInicial(m, versoes)
	if err != nil {
		t.Fatal(err)
	}
	casos := []struct {
		nome     string
		obtido   any
		esperado any
	}{
		{"MAP-04 turno 1", e.Turno, 1},
		{"MAP-01 config do mapa", e.Config, m.Config},
		{"MAP-04 sem bombas", e.Bombas, []Bomba{}},
		{"MAP-01 blocos fixos do mapa", e.BlocosFixos, m.BlocosFixos},
		{"MAP-01 blocos destrutíveis do mapa", e.BlocosDestrutiveis, m.BlocosDestrutiveis},
		{"EST-03 etapas = maior acoes_por_turno dos vivos", e.EtapasNesteTurno, 4},
		{"MAP-02 MAP-03 MAP-04 jogadores", e.Jogadores, []Jogador{
			{ID: "jogador_1", Posicao: Posicao{X: 0, Y: 0}, Status: Vivo, Atributos: m.JogadorPadrao, BotVersao: "claude-v1"},
			{ID: "jogador_2", Posicao: Posicao{X: 4, Y: 4}, Status: Vivo, Atributos: m.JogadorPadrao, BotVersao: "aleatorio-v1"},
			{ID: "jogador_3", Posicao: Posicao{X: 4, Y: 0}, Status: Vivo, Atributos: m.JogadorPadrao, BotVersao: "gpt-v1"},
		}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if !reflect.DeepEqual(c.obtido, c.esperado) {
				t.Errorf("obtido %+v, esperado %+v", c.obtido, c.esperado)
			}
		})
	}
}

func TestEstadoInicialSlicesNaoNulos(t *testing.T) {
	m := mapaValido()
	m.BlocosFixos, m.BlocosDestrutiveis = nil, nil
	e, err := EstadoInicial(m, []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if e.BlocosFixos == nil || e.BlocosDestrutiveis == nil || e.Bombas == nil {
		t.Errorf("MAP-04 (D4) estado inicial com slice nulo: %+v", e)
	}
}

func TestEstadoInicialNaoAlteraMapa(t *testing.T) {
	m := mapaValido()
	e, err := EstadoInicial(m, []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	e.BlocosFixos[0].X = 9
	e.BlocosDestrutiveis[0].X = 9
	e.Jogadores[0].Posicao.X = 9
	e.Config.Largura = 9
	if !reflect.DeepEqual(m, mapaValido()) {
		t.Errorf("MAP-03 (CA-12) alterar o estado inicial mudou o mapa: %+v", m)
	}
}

func TestCalcularEtapas(t *testing.T) {
	jogador := func(acoes int, status Status) Jogador {
		return Jogador{Status: status, Atributos: Atributos{AcoesPorTurno: acoes}}
	}
	casos := []struct {
		nome      string
		jogadores []Jogador
		esperado  int
	}{
		{"EST-03 maior entre os vivos", []Jogador{jogador(3, Vivo), jogador(7, Vivo), jogador(5, Vivo)}, 7},
		{"EST-03 mortos não contam", []Jogador{jogador(3, Vivo), jogador(9, Morto)}, 3},
		{"EST-03 nenhum vivo dá zero", []Jogador{jogador(9, Morto)}, 0},
		{"EST-03 sem jogadores dá zero", nil, 0},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if obtido := CalcularEtapas(c.jogadores); obtido != c.esperado {
				t.Errorf("obtido %d, esperado %d", obtido, c.esperado)
			}
		})
	}
}
