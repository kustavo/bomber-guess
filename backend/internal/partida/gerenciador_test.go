package partida

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
