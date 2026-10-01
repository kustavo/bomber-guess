package partida

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/fila"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// t0 é o instante em que as partidas de teste são criadas.
var t0 = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

const (
	p = time.Second            // prazo_planejamento_ms dos mapas de teste
	d = 100 * time.Millisecond // duracao_etapa_ms dos mapas de teste
)

// botDeTeste é um bot que só existe nos testes.
type botDeTeste struct {
	planejar func(e jogo.Estado, id string) []jogo.Acao
}

func (b botDeTeste) Versao() string                                { return "teste" }
func (b botDeTeste) Planejar(e jogo.Estado, id string) []jogo.Acao { return b.planejar(e, id) }

// acoes converte uma sequência compacta em ações numeradas: C, B, E, D movem
// para CIMA, BAIXO, ESQUERDA, DIREITA; P planta; '.' espera.
func acoes(seq string) []jogo.Acao {
	direcoes := map[string]jogo.Direcao{"C": jogo.Cima, "B": jogo.Baixo, "E": jogo.Esquerda, "D": jogo.Direita}
	r := []jogo.Acao{}
	for i, s := range strings.Fields(seq) {
		a := jogo.Acao{Etapa: i + 1, Tipo: jogo.Esperar}
		if dir, ok := direcoes[s]; ok {
			a.Tipo, a.Direcao = jogo.Mover, dir
		} else if s == "P" {
			a.Tipo = jogo.Plantar
		}
		r = append(r, a)
	}
	return r
}

// espera só espera.
var espera = botDeTeste{func(jogo.Estado, string) []jogo.Acao { return nil }}

// sempre joga a mesma sequência em todo turno.
func sempre(seq string) botDeTeste {
	return botDeTeste{func(jogo.Estado, string) []jogo.Acao { return acoes(seq) }}
}

// noTurno1 joga a sequência no turno 1 e espera nos demais.
func noTurno1(seq string) botDeTeste {
	return botDeTeste{func(e jogo.Estado, _ string) []jogo.Acao {
		if e.Turno == 1 {
			return acoes(seq)
		}
		return nil
	}}
}

// panico entra em panic em todo turno.
var panico = botDeTeste{func(jogo.Estado, string) []jogo.Acao { panic("bum") }}

// vandalo altera o estado recebido e não joga.
var vandalo = botDeTeste{func(e jogo.Estado, id string) []jogo.Acao {
	e.Bombas = e.Bombas[:0]
	for i := range e.BlocosDestrutiveis {
		e.BlocosDestrutiveis[i] = jogo.Posicao{X: -1, Y: -1}
	}
	for i := range e.Jogadores {
		if e.Jogadores[i].ID != id {
			e.Jogadores[i].Status = jogo.Morto
		}
	}
	e.Turno = 99
	return nil
}}

// bloqueado só responde (com a sequência dada) depois que o teste o libera.
func bloqueado(t *testing.T, seq string) (botDeTeste, func()) {
	libera := make(chan struct{})
	var uma sync.Once
	liberar := func() { uma.Do(func() { close(libera) }) }
	t.Cleanup(liberar)
	return botDeTeste{func(jogo.Estado, string) []jogo.Acao {
		<-libera
		return acoes(seq)
	}}, liberar
}

// mapaDe monta um mapa a partir de um desenho em ASCII: '.' livre, '#' bloco
// fixo, '+' bloco destrutível e '1'–'9' a posição inicial N. Prazo p, etapa d.
func mapaDe(t *testing.T, desenho string) jogo.Mapa {
	t.Helper()
	var linhas []string
	for _, l := range strings.Split(desenho, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			linhas = append(linhas, l)
		}
	}
	m := jogo.Mapa{
		Nome: "teste",
		Config: jogo.Config{Largura: len(linhas[0]), Altura: len(linhas), LimiteTurnos: 10,
			PrazoPlanejamentoMs: int(p / time.Millisecond), DuracaoEtapaMs: int(d / time.Millisecond)},
		JogadorPadrao:      jogo.Atributos{BombasPorTurno: 1, Potencia: 2, PavioPadrao: 3, AcoesPorTurno: 7},
		BlocosFixos:        []jogo.Posicao{},
		BlocosDestrutiveis: []jogo.Posicao{},
	}
	iniciais := map[int]jogo.Posicao{}
	for y, l := range linhas {
		for x, c := range l {
			pos := jogo.Posicao{X: x, Y: y}
			switch {
			case c == '#':
				m.BlocosFixos = append(m.BlocosFixos, pos)
			case c == '+':
				m.BlocosDestrutiveis = append(m.BlocosDestrutiveis, pos)
			case c >= '1' && c <= '9':
				iniciais[int(c-'0')] = pos
			}
		}
	}
	for n := 1; n <= len(iniciais); n++ {
		m.PosicoesIniciais = append(m.PosicoesIniciais, iniciais[n])
	}
	if err := jogo.VerificarMapa(m); err != nil {
		t.Fatal(err)
	}
	return m
}

// mesa é uma partida de teste com a fila e um observador de planos-enviados.
type mesa struct {
	*Partida
	fila    *fila.Memoria
	planos  fila.Consumidor
	inicial jogo.Estado
}

// novaMesa cria a partida "teste" no mapa, em t0, com os bots dados.
func novaMesa(t *testing.T, m jogo.Mapa, jogadores ...jogo.Bot) *mesa {
	t.Helper()
	return novaMesaNa(t, fila.NovaMemoria(), "teste", m, t0, jogadores...)
}

func novaMesaNa(t *testing.T, f *fila.Memoria, nome string, m jogo.Mapa, inicio time.Time, jogadores ...jogo.Bot) *mesa {
	t.Helper()
	versoes := make([]string, len(jogadores))
	for i := range versoes {
		versoes[i] = "teste"
	}
	e, err := jogo.EstadoInicial(m, versoes)
	if err != nil {
		t.Fatal(err)
	}
	obs, err := f.Consumir(fila.PlanosEnviados)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { obs.Fechar() })
	pt, err := Nova(Config{Nome: nome, Mapa: m.Nome, Bots: versoes, Semente: 1}, e, jogadores, f, inicio)
	if err != nil {
		t.Fatal(err)
	}
	return &mesa{Partida: pt, fila: f, planos: obs, inicial: e}
}

// esperarPlanos espera até haver n planos desta partida no turno dado em
// planos-enviados, e os devolve.
func (m *mesa) esperarPlanos(t *testing.T, turno, n int) []PlanoEnviado {
	t.Helper()
	var vistos []PlanoEnviado
	limite := time.Now().Add(2 * time.Second)
	for {
		mensagens, err := m.planos.Ler()
		if err != nil {
			t.Fatal(err)
		}
		for _, msg := range mensagens {
			var pe PlanoEnviado
			if err := json.Unmarshal(msg.Valor, &pe); err != nil {
				t.Fatal(err)
			}
			if pe.Partida == m.cfg.Nome && pe.Turno == turno {
				vistos = append(vistos, pe)
			}
		}
		if len(vistos) >= n {
			return vistos
		}
		if time.Now().After(limite) {
			t.Fatalf("%d planos no turno %d, esperado %d", len(vistos), turno, n)
		}
		time.Sleep(time.Millisecond)
	}
}

// jogarTurno espera n planos do turno atual e avança até o fim do
// planejamento; devolve o instante t1 em que a execução começa.
func (m *mesa) jogarTurno(t *testing.T, agora time.Time, n int) time.Time {
	t.Helper()
	v := m.Visao(agora)
	if v.Fase != Planejamento {
		t.Fatalf("fase %s, esperado PLANEJAMENTO", v.Fase)
	}
	m.esperarPlanos(t, v.Turno, n)
	m.Avancar(v.FimDaFase)
	return v.FimDaFase
}

// jogarAteOFim joga turnos, esperando o plano de cada jogador vivo, até a
// partida acabar; devolve o instante do fim.
func (m *mesa) jogarAteOFim(t *testing.T) time.Time {
	t.Helper()
	agora := t0
	for range 200 {
		v := m.Visao(agora)
		if v.Fase == Encerrada {
			return agora
		}
		t1 := m.jogarTurno(t, agora, len(jogo.VerificarFim(v.Estado).Sobreviventes))
		agora, _ = m.Avancar(t1)
	}
	t.Fatal("a partida não terminou")
	return agora
}
