package main

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/bots/aleatorio"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// botDeTeste é um bot que só existe nos testes.
type botDeTeste struct {
	versao   string
	planejar func(e jogo.Estado, id string) []jogo.Acao
}

func (b botDeTeste) Versao() string { return b.versao }
func (b botDeTeste) Planejar(e jogo.Estado, id string) []jogo.Acao {
	return b.planejar(e, id)
}

// acoes converte uma sequência compacta em ações numeradas: C, B, E, D movem
// para CIMA, BAIXO, ESQUERDA, DIREITA; P planta; '.' espera.
func acoes(seq string) []jogo.Acao {
	direcoes := map[string]jogo.Direcao{"C": jogo.Cima, "B": jogo.Baixo, "E": jogo.Esquerda, "D": jogo.Direita}
	r := []jogo.Acao{}
	for i, s := range strings.Fields(seq) {
		a := jogo.Acao{Etapa: i + 1, Tipo: jogo.Esperar}
		if d, ok := direcoes[s]; ok {
			a.Tipo, a.Direcao = jogo.Mover, d
		} else if s == "P" {
			a.Tipo = jogo.Plantar
		}
		r = append(r, a)
	}
	return r
}

// noTurno1 joga a sequência dada no turno 1 e espera nos demais.
func noTurno1(versao, seq string) jogo.Bot {
	return botDeTeste{versao, func(e jogo.Estado, _ string) []jogo.Acao {
		if e.Turno == 1 {
			return acoes(seq)
		}
		return nil
	}}
}

// catalogoDeTeste tem o aleatorio-v1 e os bots de teste.
func catalogoDeTeste() bots.Catalogo {
	fixo := func(b jogo.Bot) bots.Fabrica { return func(uint64) jogo.Bot { return b } }
	return bots.Catalogo{
		aleatorio.Versao: func(s uint64) jogo.Bot { return aleatorio.Novo(s) },
		"espera":         fixo(botDeTeste{"espera", func(jogo.Estado, string) []jogo.Acao { return nil }}),
		"bombista":       fixo(noTurno1("bombista", "P B D")), // planta e foge em L
		"suicida":        fixo(noTurno1("suicida", "P")),      // planta e fica
		"fora":           fixo(noTurno1("fora", "D C D")),     // 2ª ação sai do tabuleiro a partir de (0,0)
		"dorminhoco": fixo(botDeTeste{"dorminhoco", func(e jogo.Estado, _ string) []jogo.Acao {
			if e.Turno == 1 {
				time.Sleep(2 * time.Second)
				return acoes("D")
			}
			return nil
		}}),
		"panico": fixo(botDeTeste{"panico", func(e jogo.Estado, _ string) []jogo.Acao {
			if e.Turno == 1 {
				panic("bum")
			}
			return nil
		}}),
		"vandalo": fixo(botDeTeste{"vandalo", func(e jogo.Estado, id string) []jogo.Acao {
			e.Bombas = e.Bombas[:0]
			for i := range e.BlocosDestrutiveis {
				e.BlocosDestrutiveis[i] = jogo.Posicao{X: -1, Y: -1}
			}
			for i := range e.Jogadores {
				if e.Jogadores[i].ID != id {
					e.Jogadores[i].Status = jogo.Morto
				}
			}
			return nil
		}}),
	}
}

// mapaDe monta um mapa a partir de um desenho em ASCII: '.' livre, '#' bloco
// fixo, '+' bloco destrutível e '1'–'9' a posição inicial N.
func mapaDe(t *testing.T, desenho string, ajustes ...func(*jogo.Mapa)) jogo.Mapa {
	t.Helper()
	var linhas []string
	for _, l := range strings.Split(desenho, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			linhas = append(linhas, l)
		}
	}
	m := jogo.Mapa{
		Nome:               "teste",
		Config:             jogo.Config{Largura: len(linhas[0]), Altura: len(linhas), LimiteTurnos: 10, PrazoPlanejamentoMs: 500},
		JogadorPadrao:      jogo.Atributos{BombasPorTurno: 1, Potencia: 2, PavioPadrao: 3, AcoesPorTurno: 7},
		BlocosFixos:        []jogo.Posicao{},
		BlocosDestrutiveis: []jogo.Posicao{},
	}
	iniciais := map[int]jogo.Posicao{}
	for y, l := range linhas {
		for x, c := range l {
			p := jogo.Posicao{X: x, Y: y}
			switch {
			case c == '#':
				m.BlocosFixos = append(m.BlocosFixos, p)
			case c == '+':
				m.BlocosDestrutiveis = append(m.BlocosDestrutiveis, p)
			case c >= '1' && c <= '9':
				iniciais[int(c-'0')] = p
			}
		}
	}
	for n := 1; n <= len(iniciais); n++ {
		m.PosicoesIniciais = append(m.PosicoesIniciais, iniciais[n])
	}
	for _, a := range ajustes {
		a(&m)
	}
	if err := jogo.VerificarMapa(m); err != nil {
		t.Fatal(err)
	}
	return m
}

func comLimite(n int) func(*jogo.Mapa) { return func(m *jogo.Mapa) { m.Config.LimiteTurnos = n } }
func comPrazo(ms int) func(*jogo.Mapa) {
	return func(m *jogo.Mapa) { m.Config.PrazoPlanejamentoMs = ms }
}

// partida joga o mapa com as versões dadas do catálogo de teste, sem atraso.
func partida(t *testing.T, m jogo.Mapa, versoes ...string) (string, resumo) {
	t.Helper()
	e, err := jogo.EstadoInicial(m, versoes)
	if err != nil {
		t.Fatal(err)
	}
	jogadores := make([]jogo.Bot, len(versoes))
	for i, v := range versoes {
		if jogadores[i], err = catalogoDeTeste().Criar(v, 1); err != nil {
			t.Fatal(err)
		}
	}
	var saida strings.Builder
	r := jogar(e, jogadores, 0, &saida)
	return saida.String(), r
}

// tabuleiroDa devolve as linhas do tabuleiro impresso logo depois do
// cabeçalho dado (ex.: "Turno 1, etapa 3"), ou nil se ele não aparece.
func tabuleiroDa(saida, cabecalho string, altura int) []string {
	linhas := strings.Split(saida, "\n")
	for i, l := range linhas {
		if l == cabecalho && i+altura < len(linhas) {
			return linhas[i+1 : i+1+altura]
		}
	}
	return nil
}

// ultimaLinha devolve a última linha não vazia.
func ultimaLinha(saida string) string {
	linhas := strings.Split(strings.TrimRight(saida, "\n"), "\n")
	return linhas[len(linhas)-1]
}

// cabecalhos conta as linhas "Turno T, etapa E" e devolve o maior turno visto.
func cabecalhos(saida string) (etapas, maiorTurno int) {
	for _, l := range strings.Split(saida, "\n") {
		var turno, etapa int
		if strings.HasPrefix(l, "Turno ") && strings.Contains(l, ", etapa ") {
			partes := strings.Fields(strings.NewReplacer(",", "").Replace(l))
			turno, _ = strconv.Atoi(partes[1])
			etapa, _ = strconv.Atoi(partes[3])
			if etapa > 0 {
				etapas++
				maiorTurno = max(maiorTurno, turno)
			}
		}
	}
	return etapas, maiorTurno
}
