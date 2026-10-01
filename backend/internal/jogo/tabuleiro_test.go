package jogo

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// atributosDeTeste são os atributos padrão dos jogadores montados por montar.
var atributosDeTeste = Atributos{BombasPorTurno: 2, Potencia: 2, PavioPadrao: 3, AcoesPorTurno: 7}

// opcao ajusta o estado montado por montar.
type opcao func(e *Estado)

// montar monta um Estado no turno 1 a partir de um desenho em ASCII (D11):
// '.' casa livre, '#' bloco fixo, '+' bloco destrutível e '1'–'9' o jogador
// jogador_N, vivo e com atributosDeTeste. Espaços nas pontas das linhas e
// linhas vazias são ignorados. etapas_neste_turno é calculado depois das opções.
func montar(t *testing.T, desenho string, opcoes ...opcao) Estado {
	t.Helper()
	var linhas []string
	for _, l := range strings.Split(desenho, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			linhas = append(linhas, l)
		}
	}
	if len(linhas) == 0 {
		t.Fatal("montar: desenho vazio")
	}
	e := Estado{
		Turno:              1,
		Config:             Config{Largura: len(linhas[0]), Altura: len(linhas), LimiteTurnos: 50, PrazoPlanejamentoMs: 100, DuracaoEtapaMs: 5},
		BlocosFixos:        []Posicao{},
		BlocosDestrutiveis: []Posicao{},
		Bombas:             []Bomba{},
		Jogadores:          []Jogador{},
	}
	var jogadores [10]*Jogador
	for y, l := range linhas {
		if len(l) != e.Config.Largura {
			t.Fatalf("montar: linha %d tem largura %d, esperado %d", y, len(l), e.Config.Largura)
		}
		for x, c := range l {
			p := Posicao{X: x, Y: y}
			switch {
			case c == '.':
			case c == '#':
				e.BlocosFixos = append(e.BlocosFixos, p)
			case c == '+':
				e.BlocosDestrutiveis = append(e.BlocosDestrutiveis, p)
			case c >= '1' && c <= '9':
				n := int(c - '0')
				if jogadores[n] != nil {
					t.Fatalf("montar: jogador %d repetido", n)
				}
				jogadores[n] = &Jogador{ID: "jogador_" + strconv.Itoa(n), Posicao: p, Status: Vivo, Atributos: atributosDeTeste, BotVersao: "teste"}
			default:
				t.Fatalf("montar: caractere %q desconhecido em (%d,%d)", c, x, y)
			}
		}
	}
	for _, j := range jogadores {
		if j != nil {
			e.Jogadores = append(e.Jogadores, *j)
		}
	}
	for _, o := range opcoes {
		o(&e)
	}
	e.EtapasNesteTurno = CalcularEtapas(e.Jogadores)
	return e
}

// comBomba acrescenta uma bomba ao estado.
func comBomba(x, y int, jogadorID string, potencia, pavio int) opcao {
	return func(e *Estado) {
		e.Bombas = append(e.Bombas, Bomba{Posicao: Posicao{X: x, Y: y}, JogadorID: jogadorID, Potencia: potencia, PavioRestante: pavio})
	}
}

// comAtributos altera os atributos de um jogador.
func comAtributos(jogadorID string, alterar func(a *Atributos)) opcao {
	return func(e *Estado) {
		for i := range e.Jogadores {
			if e.Jogadores[i].ID == jogadorID {
				alterar(&e.Jogadores[i].Atributos)
			}
		}
	}
}

// comAcoes define acoes_por_turno de todos os jogadores.
func comAcoes(n int) opcao {
	return func(e *Estado) {
		for i := range e.Jogadores {
			e.Jogadores[i].AcoesPorTurno = n
		}
	}
}

// comMorto marca um jogador como morto no turno e etapa dados.
func comMorto(jogadorID string, turno, etapa int) opcao {
	return func(e *Estado) {
		for i := range e.Jogadores {
			if e.Jogadores[i].ID == jogadorID {
				e.Jogadores[i].Status = Morto
				e.Jogadores[i].Morte = &Morte{Turno: turno, Etapa: etapa}
			}
		}
	}
}

// comTurno define o turno do estado.
func comTurno(turno int) opcao {
	return func(e *Estado) { e.Turno = turno }
}

// acoes converte uma sequência compacta em ações numeradas a partir da etapa 1:
// C, B, E, D movem para CIMA, BAIXO, ESQUERDA, DIREITA; P planta; '.' espera.
func acoes(t *testing.T, seq string) []Acao {
	t.Helper()
	direcoes := map[string]Direcao{"C": Cima, "B": Baixo, "E": Esquerda, "D": Direita}
	resultado := []Acao{}
	for i, s := range strings.Fields(seq) {
		a := Acao{Etapa: i + 1}
		switch {
		case direcoes[s] != "":
			a.Tipo, a.Direcao = Mover, direcoes[s]
		case s == "P":
			a.Tipo = Plantar
		case s == ".":
			a.Tipo = Esperar
		default:
			t.Fatalf("acoes: token %q desconhecido", s)
		}
		resultado = append(resultado, a)
	}
	return resultado
}

// plano monta o plano de um jogador para o turno do estado.
func plano(t *testing.T, e Estado, jogadorID, seq string) Plano {
	t.Helper()
	return Plano{JogadorID: jogadorID, Turno: e.Turno, Acoes: acoes(t, seq)}
}

func TestMontar(t *testing.T) {
	e := montar(t, `
		1.+
		.#2
	`, comBomba(1, 0, "jogador_1", 2, 3), comAtributos("jogador_2", func(a *Atributos) { a.AcoesPorTurno = 9 }))
	esperado := Estado{
		Turno:              1,
		Config:             Config{Largura: 3, Altura: 2, LimiteTurnos: 50, PrazoPlanejamentoMs: 100, DuracaoEtapaMs: 5},
		EtapasNesteTurno:   9,
		BlocosFixos:        []Posicao{{X: 1, Y: 1}},
		BlocosDestrutiveis: []Posicao{{X: 2, Y: 0}},
		Bombas:             []Bomba{{Posicao: Posicao{X: 1, Y: 0}, JogadorID: "jogador_1", Potencia: 2, PavioRestante: 3}},
		Jogadores: []Jogador{
			{ID: "jogador_1", Posicao: Posicao{X: 0, Y: 0}, Status: Vivo, Atributos: atributosDeTeste, BotVersao: "teste"},
			{ID: "jogador_2", Posicao: Posicao{X: 2, Y: 1}, Status: Vivo, Atributos: Atributos{BombasPorTurno: 2, Potencia: 2, PavioPadrao: 3, AcoesPorTurno: 9}, BotVersao: "teste"},
		},
	}
	if !reflect.DeepEqual(e, esperado) {
		t.Errorf("montar:\n obtido:   %+v\n esperado: %+v", e, esperado)
	}
	if obtido := acoes(t, "D P ."); !reflect.DeepEqual(obtido, []Acao{{1, Mover, Direita}, {2, Plantar, ""}, {3, Esperar, ""}}) {
		t.Errorf("acoes: %+v", obtido)
	}
}
