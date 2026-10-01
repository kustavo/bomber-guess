package aleatorio

import (
	"strconv"
	"strings"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// atributosDeTeste são os atributos padrão dos jogadores montados por montar.
var atributosDeTeste = jogo.Atributos{BombasPorTurno: 2, Potencia: 2, PavioPadrao: 3, AcoesPorTurno: 7}

// opcao ajusta o estado montado por montar.
type opcao func(e *jogo.Estado)

// montar monta um Estado no turno 1 a partir de um desenho em ASCII (D9,
// cópia do helper de internal/jogo): '.' casa livre, '#' bloco fixo, '+'
// bloco destrutível e '1'–'9' o jogador jogador_N, vivo e com
// atributosDeTeste. etapas_neste_turno é calculado depois das opções.
func montar(t *testing.T, desenho string, opcoes ...opcao) jogo.Estado {
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
	e := jogo.Estado{
		Turno:              1,
		Config:             jogo.Config{Largura: len(linhas[0]), Altura: len(linhas), LimiteTurnos: 50, PrazoPlanejamentoMs: 1000, DuracaoEtapaMs: 5},
		BlocosFixos:        []jogo.Posicao{},
		BlocosDestrutiveis: []jogo.Posicao{},
		Bombas:             []jogo.Bomba{},
		Jogadores:          []jogo.Jogador{},
	}
	var jogadores [10]*jogo.Jogador
	for y, l := range linhas {
		if len(l) != e.Config.Largura {
			t.Fatalf("montar: linha %d tem largura %d, esperado %d", y, len(l), e.Config.Largura)
		}
		for x, c := range l {
			p := jogo.Posicao{X: x, Y: y}
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
				jogadores[n] = &jogo.Jogador{ID: "jogador_" + strconv.Itoa(n), Posicao: p, Status: jogo.Vivo, Atributos: atributosDeTeste, BotVersao: Versao}
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
	e.EtapasNesteTurno = jogo.CalcularEtapas(e.Jogadores)
	return e
}

// comBomba acrescenta uma bomba ao estado.
func comBomba(x, y int, jogadorID string, potencia, pavio int) opcao {
	return func(e *jogo.Estado) {
		e.Bombas = append(e.Bombas, jogo.Bomba{Posicao: jogo.Posicao{X: x, Y: y}, JogadorID: jogadorID, Potencia: potencia, PavioRestante: pavio})
	}
}

// comAtributos altera os atributos de um jogador.
func comAtributos(jogadorID string, alterar func(a *jogo.Atributos)) opcao {
	return func(e *jogo.Estado) {
		for i := range e.Jogadores {
			if e.Jogadores[i].ID == jogadorID {
				alterar(&e.Jogadores[i].Atributos)
			}
		}
	}
}

// comMorto marca um jogador como morto no turno 1, etapa 1.
func comMorto(jogadorID string) opcao {
	return func(e *jogo.Estado) {
		for i := range e.Jogadores {
			if e.Jogadores[i].ID == jogadorID {
				e.Jogadores[i].Status = jogo.Morto
				e.Jogadores[i].Morte = &jogo.Morte{Turno: 1, Etapa: 1}
			}
		}
	}
}

// comTurno define o turno do estado.
func comTurno(turno int) opcao {
	return func(e *jogo.Estado) { e.Turno = turno }
}

// conferirContrato falha se o plano não tiver exatamente acoes_por_turno
// ações numeradas 1, 2, 3… com direção só em MOVER (CA-03).
func conferirContrato(t *testing.T, e jogo.Estado, jogadorID string, plano []jogo.Acao) {
	t.Helper()
	eu, _ := encontrar(e, jogadorID)
	if len(plano) != eu.AcoesPorTurno {
		t.Fatalf("plano com %d ações, esperado %d: %+v", len(plano), eu.AcoesPorTurno, plano)
	}
	for i, a := range plano {
		if a.Etapa != i+1 {
			t.Errorf("ação %d com etapa %d", i+1, a.Etapa)
		}
		if (a.Tipo == jogo.Mover) != (a.Direcao != "") {
			t.Errorf("ação %d: tipo %s com direção %q", i+1, a.Tipo, a.Direcao)
		}
	}
}

// categoria classifica um plano pela definição da spec (seção "Termos").
type categoria int

const (
	morre categoria = iota
	sobrevivente
	seguro
)

func (c categoria) String() string {
	return [...]string{"morre", "sobrevivente", "seguro"}[c]
}

// simular resolve o turno com o plano do jogador e os outros esperando.
func simular(e jogo.Estado, jogadorID string, plano []jogo.Acao) (jogo.Estado, []jogo.RelatorioEtapa) {
	return jogo.ResolverTurno(e, []jogo.Plano{{JogadorID: jogadorID, Turno: e.Turno, Acoes: plano}})
}

// classificar diz se o plano é seguro, só sobrevivente ou se o jogador morre
// na simulação do turno. Independente da implementação: a zona de perigo
// final é verificada explodindo de uma vez, no estado real, todas as bombas
// que sobram. Fim antecipado com o jogador vivo conta como seguro (decisão 7).
func classificar(e jogo.Estado, jogadorID string, plano []jogo.Acao) categoria {
	final, relatorios := simular(e, jogadorID, plano)
	if eu, _ := encontrar(final, jogadorID); eu.Status != jogo.Vivo {
		return morre
	}
	if len(relatorios) < e.EtapasNesteTurno || jogo.VerificarFim(final).Terminada {
		return seguro
	}
	resto := final.Copiar()
	resto.Config.TurnoFechamento = 0 // só as bombas: o fechamento do turno seguinte não conta
	for i := range resto.Bombas {
		resto.Bombas[i].PavioRestante = 1
	}
	depois, _ := jogo.ResolverTurno(resto, nil)
	if eu, _ := encontrar(depois, jogadorID); eu.Status != jogo.Vivo {
		return sobrevivente
	}
	return seguro
}

// versaoDeTeste é uma versão do bot coberta pelos testes (D10 do marco 13).
type versaoDeTeste struct {
	nome string
	novo func(semente uint64) *Bot
}

var versoesDeTeste = []versaoDeTeste{{Versao, Novo}, {VersaoV2, NovoV2}, {VersaoV3, NovoV3}, {VersaoV4, NovoV4}}

// paraCadaVersao roda f num subteste por versão do bot (CA-02 do marco 13).
func paraCadaVersao(t *testing.T, f func(t *testing.T, novo func(uint64) *Bot)) {
	t.Helper()
	for _, v := range versoesDeTeste {
		t.Run(v.nome, func(t *testing.T) { f(t, v.novo) })
	}
}
