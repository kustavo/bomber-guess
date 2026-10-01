package partida

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/fila"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// relogioFixo é um relógio de teste que só muda quando o teste manda.
type relogioFixo struct{ agora time.Time }

func (r *relogioFixo) Agora() time.Time { return r.agora }

// dirMapas grava os mapas dados em JSON num diretório temporário.
func dirMapas(t *testing.T, mapas map[string]jogo.Mapa) string {
	t.Helper()
	dir := t.TempDir()
	for nome, m := range mapas {
		dados, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, nome+".json"), dados, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func catalogoDeTeste() bots.Catalogo {
	fixo := func(b jogo.Bot) bots.Fabrica { return func(uint64) jogo.Bot { return b } }
	return bots.Catalogo{"espera": fixo(espera), "bombista": fixo(noTurno1("P B D"))}
}

func novoGerenciadorDeTeste(t *testing.T, automatico bool, relogio Relogio) *Gerenciador {
	t.Helper()
	invalido := mapaDe(t, doisCantos)
	invalido.JogadorPadrao.AcoesPorTurno = 0
	rapido := mapaDe(t, emL)
	rapido.Config.PrazoPlanejamentoMs, rapido.Config.DuracaoEtapaMs = 20, 1
	g := NovoGerenciador(ConfigGerenciador{
		Catalogo:   catalogoDeTeste(),
		DirMapas:   dirMapas(t, map[string]jogo.Mapa{"pequeno": mapaDe(t, doisCantos), "invalido": invalido, "rapido": rapido}),
		Fila:       fila.NovaMemoria(),
		Relogio:    relogio,
		Automatico: automatico,
	})
	t.Cleanup(g.Encerrar)
	return g
}

func TestCriarPartida(t *testing.T) {
	relogio := &relogioFixo{t0}
	g := novoGerenciadorDeTeste(t, false, relogio)
	semente := uint64(7)
	p, err := g.Criar(Pedido{Nome: "final-1", Mapa: "pequeno", Bots: []string{"espera"}, Semente: &semente})
	if err != nil {
		t.Fatal(err)
	}
	v := p.Visao(t0)
	h := p.Historico()
	if v.Nome != "final-1" || v.Fase != Planejamento || v.Turno != 1 || !v.FimDaFase.Equal(t0.Add(p0(v))) {
		t.Errorf("API-04 CA-18 visão %+v", v)
	}
	if h.Mapa != "pequeno" || len(h.Bots) != 2 || h.Bots[1] != "espera" || h.Semente != 7 {
		t.Errorf("MAP-03 CA-18 histórico %+v", h)
	}
	if obtida, err := g.Obter("final-1"); err != nil || obtida != p {
		t.Errorf("Obter: %v", err)
	}
	if _, err := g.Obter("nenhuma"); !errors.Is(err, ErrPartidaDesconhecida) {
		t.Errorf("Obter desconhecida: %v", err)
	}
	p2, _ := g.Criar(Pedido{Nome: "b", Mapa: "pequeno", Bots: []string{"espera", "espera"}})
	if l := g.Listar(); len(l) != 2 || l[0] != p || l[1] != p2 {
		t.Errorf("API-10 lista fora da ordem de criação")
	}
	if p2.Historico().Semente != 1 {
		t.Errorf("semente padrão %d", p2.Historico().Semente)
	}
}

// p0 é o prazo de planejamento da visão.
func p0(v Visao) time.Duration {
	return time.Duration(v.Estado.Config.PrazoPlanejamentoMs) * time.Millisecond
}

func TestPedidoInvalido(t *testing.T) {
	g := novoGerenciadorDeTeste(t, false, &relogioFixo{t0})
	if _, err := g.Criar(Pedido{Nome: "usada", Mapa: "pequeno", Bots: []string{"espera"}}); err != nil {
		t.Fatal(err)
	}
	casos := []struct {
		nome   string
		pedido Pedido
		erro   error
	}{
		{"API-04 CA-19 nome vazio", Pedido{Mapa: "pequeno", Bots: []string{"espera"}}, ErrPedidoInvalido},
		{"API-04 CA-19 nome com maiúscula", Pedido{Nome: "Final", Mapa: "pequeno", Bots: []string{"espera"}}, ErrPedidoInvalido},
		{"API-04 CA-19 mapa fora do diretório", Pedido{Nome: "x", Mapa: "../pequeno", Bots: []string{"espera"}}, ErrPedidoInvalido},
		{"API-04 CA-19 mapa inexistente", Pedido{Nome: "x", Mapa: "nenhum", Bots: []string{"espera"}}, ErrPedidoInvalido},
		{"MAP-05 CA-19 mapa inválido", Pedido{Nome: "x", Mapa: "invalido", Bots: []string{"espera"}}, ErrPedidoInvalido},
		{"API-04 CA-19 bot desconhecido", Pedido{Nome: "x", Mapa: "pequeno", Bots: []string{"espera", "nenhum-v9"}}, ErrPedidoInvalido},
		{"MAP-06 CA-19 quantidade de bots errada", Pedido{Nome: "x", Mapa: "pequeno", Bots: []string{"espera", "espera", "espera"}}, ErrPedidoInvalido},
		{"MAP-06 CA-19 sem bots", Pedido{Nome: "x", Mapa: "pequeno"}, ErrPedidoInvalido},
		{"API-04 CA-19 nome repetido", Pedido{Nome: "usada", Mapa: "pequeno", Bots: []string{"espera"}}, ErrNomeRepetido},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if _, err := g.Criar(c.pedido); !errors.Is(err, c.erro) {
				t.Errorf("erro %v, esperado %v", err, c.erro)
			}
			if len(g.Listar()) != 1 {
				t.Errorf("a partida foi criada")
			}
		})
	}
}

func TestLacoEmTempoReal(t *testing.T) {
	g := novoGerenciadorDeTeste(t, true, nil)
	p, err := g.Criar(Pedido{Nome: "rapida", Mapa: "rapido", Bots: []string{"bombista", "espera"}})
	if err != nil {
		t.Fatal(err)
	}
	limite := time.Now().Add(2 * time.Second)
	for p.Historico().Desfecho == nil {
		if time.Now().After(limite) {
			t.Fatal("PAR-01 a partida não terminou sozinha")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if d := p.Historico().Desfecho; d.Vencedor != "jogador_1" {
		t.Errorf("PAR-01 desfecho %+v", d)
	}
	g.Encerrar()
}

func TestEncerrarParaOsLacos(t *testing.T) {
	g := novoGerenciadorDeTeste(t, true, nil)
	if _, err := g.Criar(Pedido{Nome: "longa", Mapa: "pequeno", Bots: []string{"espera"}}); err != nil {
		t.Fatal(err)
	}
	pronto := make(chan struct{})
	go func() { g.Encerrar(); close(pronto) }()
	select {
	case <-pronto:
	case <-time.After(time.Second):
		t.Fatal("Encerrar não parou o laço")
	}
}

// arquivosEm lista os nomes dos arquivos do diretório.
func arquivosEm(t *testing.T, dir string) []string {
	t.Helper()
	entradas, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var nomes []string
	for _, e := range entradas {
		nomes = append(nomes, e.Name())
	}
	return nomes
}

func TestSalvarMapa(t *testing.T) {
	g := novoGerenciadorDeTeste(t, false, &relogioFixo{t0})
	m := mapaDe(t, doisCantos)
	m.Nome = "novo-1"
	if err := g.SalvarMapa(m); err != nil {
		t.Fatal(err)
	}
	dados, err := os.ReadFile(filepath.Join(g.cfg.DirMapas, "novo-1.json"))
	if err != nil {
		t.Fatalf("API-03 CA-01 arquivo não gravado: %v", err)
	}
	lido, err := jogo.LerMapa(dados)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(lido, m) {
		t.Errorf("API-03 CA-01 relido %+v, esperado %+v", lido, m)
	}
	// CA-02: o mapa salvo já serve para criar uma partida.
	if _, err := g.Criar(Pedido{Nome: "final-1", Mapa: "novo-1", Bots: []string{"espera"}}); err != nil {
		t.Errorf("API-04 CA-02 criar no mapa salvo: %v", err)
	}
}

func TestSalvarMapaInvalido(t *testing.T) {
	valido := func(nome string) jogo.Mapa {
		m := mapaDe(t, doisCantos)
		m.Nome = nome
		return m
	}
	com := func(mudar func(m *jogo.Mapa)) jogo.Mapa {
		m := valido("novo")
		mudar(&m)
		return m
	}
	fora := jogo.Posicao{X: 99, Y: 0}
	casos := []struct {
		nome string
		mapa jogo.Mapa
		erro error
	}{
		{"API-03 CA-04 nome vazio", valido(""), ErrPedidoInvalido},
		{"API-03 CA-04 nome com ..", valido("../x"), ErrPedidoInvalido},
		{"API-03 CA-04 nome com barra", valido("a/b"), ErrPedidoInvalido},
		{"API-03 CA-04 nome com maiúscula e espaço", valido("A B"), ErrPedidoInvalido},
		{"API-03 CA-04 nome com sublinhado", valido("a_b"), ErrPedidoInvalido},
		{"MAP-05 CA-03 largura 0", com(func(m *jogo.Mapa) { m.Config.Largura = 0 }), ErrPedidoInvalido},
		{"MAP-05 CA-03 altura negativa", com(func(m *jogo.Mapa) { m.Config.Altura = -1 }), ErrPedidoInvalido},
		{"MAP-05 CA-03 bloco fixo fora", com(func(m *jogo.Mapa) { m.BlocosFixos = append(m.BlocosFixos, fora) }), ErrPedidoInvalido},
		{"MAP-05 CA-03 bloco destrutível fora", com(func(m *jogo.Mapa) { m.BlocosDestrutiveis = append(m.BlocosDestrutiveis, fora) }), ErrPedidoInvalido},
		{"MAP-05 CA-03 posição inicial fora", com(func(m *jogo.Mapa) { m.PosicoesIniciais[0] = fora }), ErrPedidoInvalido},
		{"MAP-05 CA-03 posição inicial sobre bloco", com(func(m *jogo.Mapa) { m.BlocosFixos = append(m.BlocosFixos, m.PosicoesIniciais[0]) }), ErrPedidoInvalido},
		{"MAP-05 CA-03 bloco fixo e destrutível na mesma casa", com(func(m *jogo.Mapa) {
			m.BlocosFixos = append(m.BlocosFixos, jogo.Posicao{X: 2, Y: 0})
			m.BlocosDestrutiveis = append(m.BlocosDestrutiveis, jogo.Posicao{X: 2, Y: 0})
		}), ErrPedidoInvalido},
		{"MAP-05 CA-03 uma posição inicial", com(func(m *jogo.Mapa) { m.PosicoesIniciais = m.PosicoesIniciais[:1] }), ErrPedidoInvalido},
		{"MAP-05 CA-03 limite_turnos 0", com(func(m *jogo.Mapa) { m.Config.LimiteTurnos = 0 }), ErrPedidoInvalido},
		{"MAP-05 FEC-01 CA-03 turno_fechamento igual ao limite", com(func(m *jogo.Mapa) { m.Config.TurnoFechamento = m.Config.LimiteTurnos }), ErrPedidoInvalido},
		{"MAP-05 FEC-08 CA-03 area_minima maior que o tabuleiro", com(func(m *jogo.Mapa) { m.Config.AreaMinima = jogo.Area{Largura: 99, Altura: 1} }), ErrPedidoInvalido},
		{"MAP-05 CA-03 potencia 0", com(func(m *jogo.Mapa) { m.JogadorPadrao.Potencia = 0 }), ErrPedidoInvalido},
		{"API-03 CA-05 nome já existente", valido("pequeno"), ErrNomeRepetido},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			g := novoGerenciadorDeTeste(t, false, &relogioFixo{t0})
			antes := arquivosEm(t, g.cfg.DirMapas)
			original, _ := os.ReadFile(filepath.Join(g.cfg.DirMapas, "pequeno.json"))
			err := g.SalvarMapa(c.mapa)
			if !errors.Is(err, c.erro) {
				t.Fatalf("erro %v, esperado %v", err, c.erro)
			}
			if depois := arquivosEm(t, g.cfg.DirMapas); !reflect.DeepEqual(depois, antes) {
				t.Errorf("arquivos %v, antes %v", depois, antes)
			}
			if atual, _ := os.ReadFile(filepath.Join(g.cfg.DirMapas, "pequeno.json")); !bytes.Equal(atual, original) {
				t.Error("pequeno.json mudou")
			}
		})
	}
}
