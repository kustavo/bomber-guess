package partida

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/fila"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// Relogio dá a hora atual.
type Relogio interface{ Agora() time.Time }

// RelogioReal usa time.Now.
type RelogioReal struct{}

func (RelogioReal) Agora() time.Time { return time.Now() }

// Pedido é o corpo de POST /partidas (decisão 6).
type Pedido struct {
	Nome    string   `json:"nome"`
	Mapa    string   `json:"mapa"`
	Bots    []string `json:"bots"`
	Semente *uint64  `json:"semente,omitempty"`
}

var (
	ErrPedidoInvalido      = errors.New("pedido inválido")
	ErrNomeRepetido        = errors.New("nome já usado")
	ErrPartidaDesconhecida = errors.New("partida desconhecida")
)

// nomeValido é o formato de nomes de partida e de mapa (D8).
var nomeValido = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

// ConfigGerenciador configura o Gerenciador.
type ConfigGerenciador struct {
	Catalogo   bots.Catalogo
	DirMapas   string
	Fila       fila.Fila
	Relogio    Relogio // nil: RelogioReal
	Automatico bool    // liga o laço em tempo real de cada partida; false nos testes
}

// Gerenciador cria e guarda as partidas do servidor. Seguro para uso concorrente.
type Gerenciador struct {
	cfg      ConfigGerenciador
	mu       sync.Mutex
	partidas map[string]*Partida
	ordem    []*Partida
	ctx      context.Context
	cancelar context.CancelFunc
	lacos    sync.WaitGroup
}

// NovoGerenciador cria um gerenciador sem partidas.
func NovoGerenciador(cfg ConfigGerenciador) *Gerenciador {
	if cfg.Relogio == nil {
		cfg.Relogio = RelogioReal{}
	}
	ctx, cancelar := context.WithCancel(context.Background())
	return &Gerenciador{cfg: cfg, partidas: map[string]*Partida{}, ctx: ctx, cancelar: cancelar}
}

// Criar valida o pedido, cria a partida e, no modo automático, liga o laço
// dela. Os erros embrulham ErrPedidoInvalido ou ErrNomeRepetido.
func (g *Gerenciador) Criar(p Pedido) (*Partida, error) {
	if !nomeValido.MatchString(p.Nome) {
		return nil, fmt.Errorf("%w: nome %q deve ter de 1 a 64 caracteres entre a-z, 0-9 e -", ErrPedidoInvalido, p.Nome)
	}
	if !nomeValido.MatchString(p.Mapa) {
		return nil, fmt.Errorf("%w: mapa %q deve ter de 1 a 64 caracteres entre a-z, 0-9 e -", ErrPedidoInvalido, p.Mapa)
	}
	dados, err := os.ReadFile(filepath.Join(g.cfg.DirMapas, p.Mapa+".json"))
	if err != nil {
		return nil, fmt.Errorf("%w: mapa %q não encontrado", ErrPedidoInvalido, p.Mapa)
	}
	m, err := jogo.LerMapa(dados)
	if err != nil {
		return nil, fmt.Errorf("%w: mapa %q: %v", ErrPedidoInvalido, p.Mapa, err)
	}
	versoes := p.Bots
	if len(versoes) == 1 {
		versoes = slices.Repeat(versoes, len(m.PosicoesIniciais))
	}
	estado, err := jogo.EstadoInicial(m, versoes) // MAP-06
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPedidoInvalido, err)
	}
	semente := uint64(1)
	if p.Semente != nil {
		semente = *p.Semente
	}
	jogadores := make([]jogo.Bot, len(versoes))
	for i, v := range versoes {
		if jogadores[i], err = g.cfg.Catalogo.Criar(v, semente); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPedidoInvalido, err)
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if _, existe := g.partidas[p.Nome]; existe {
		return nil, fmt.Errorf("%w: %q", ErrNomeRepetido, p.Nome)
	}
	cfg := Config{Nome: p.Nome, Mapa: p.Mapa, Bots: versoes, Semente: semente}
	partida, err := Nova(cfg, estado, jogadores, g.cfg.Fila, g.cfg.Relogio.Agora())
	if err != nil {
		return nil, err
	}
	g.partidas[p.Nome] = partida
	g.ordem = append(g.ordem, partida)
	if g.cfg.Automatico {
		g.lacos.Add(1)
		go g.rodar(partida)
	}
	return partida, nil
}

// SalvarMapa grava m em <DirMapas>/<m.Nome>.json (API-03). Os erros embrulham
// ErrPedidoInvalido (nome fora de nomeValido, MAP-05 via jogo.VerificarMapa)
// ou ErrNomeRepetido (o arquivo já existe; nada é sobrescrito).
func (g *Gerenciador) SalvarMapa(m jogo.Mapa) error {
	if !nomeValido.MatchString(m.Nome) {
		return fmt.Errorf("%w: mapa %q deve ter de 1 a 64 caracteres entre a-z, 0-9 e -", ErrPedidoInvalido, m.Nome)
	}
	if err := jogo.VerificarMapa(m); err != nil {
		return fmt.Errorf("%w: %v", ErrPedidoInvalido, err)
	}
	dados, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	// D2: grava num temporário e liga ao nome final; o Link falha se o
	// arquivo já existe, e ninguém lê um mapa pela metade.
	tmp, err := os.CreateTemp(g.cfg.DirMapas, "."+m.Nome+"-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err = tmp.Chmod(0o644); err == nil { // CreateTemp cria com 0600
		_, err = tmp.Write(append(dados, '\n'))
	}
	if errFechar := tmp.Close(); err == nil {
		err = errFechar
	}
	if err != nil {
		return err
	}
	if err := os.Link(tmp.Name(), filepath.Join(g.cfg.DirMapas, m.Nome+".json")); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("%w: mapa %q", ErrNomeRepetido, m.Nome)
		}
		return err
	}
	return nil
}

// rodar é o laço em tempo real da partida (D1): avança até agora e espera o
// próximo instante em que algo muda, até a partida acabar ou o gerenciador
// ser encerrado.
func (g *Gerenciador) rodar(p *Partida) {
	defer g.lacos.Done()
	for {
		proximo, encerrada := p.Avancar(g.cfg.Relogio.Agora())
		if encerrada {
			return
		}
		espera := time.NewTimer(proximo.Sub(g.cfg.Relogio.Agora()))
		select {
		case <-g.ctx.Done():
			espera.Stop()
			return
		case <-espera.C:
		}
	}
}

// Obter devolve a partida com o nome dado, ou um erro que embrulha
// ErrPartidaDesconhecida.
func (g *Gerenciador) Obter(nome string) (*Partida, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if p, ok := g.partidas[nome]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("%w: %q", ErrPartidaDesconhecida, nome)
}

// Listar devolve as partidas na ordem de criação (API-10).
func (g *Gerenciador) Listar() []*Partida {
	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]*Partida{}, g.ordem...)
}

// Versoes devolve as versões do catálogo, em ordem alfabética (API-09).
func (g *Gerenciador) Versoes() []string { return g.cfg.Catalogo.Versoes() }

// Relogio devolve o relógio do gerenciador.
func (g *Gerenciador) Relogio() Relogio { return g.cfg.Relogio }

// Encerrar para os lacos das partidas e espera que terminem.
func (g *Gerenciador) Encerrar() {
	g.cancelar()
	g.lacos.Wait()
}
