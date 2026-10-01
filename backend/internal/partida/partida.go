package partida

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/fila"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// Partida é uma partida em tempo real, guiada pelo instante que recebe (D1).
// Segura para uso concorrente.
type Partida struct {
	mu        sync.Mutex
	cfg       Config
	fila      fila.Fila
	planos    fila.Consumidor // planos-enviados desta partida (D4)
	jogadores []jogo.Bot
	estado    jogo.Estado // início do turno atual
	fase      Fase
	inicio    time.Time // início da fase atual
	resolvido *turnoResolvido
	historico []RegistroTurno
	desfecho  *jogo.Desfecho
}

// turnoResolvido é o turno resolvido no fim do planejamento, ainda sendo
// liberado etapa a etapa (D6).
type turnoResolvido struct {
	registro RegistroTurno
	proximo  jogo.Estado
}

// Nova cria a partida e começa o planejamento do turno 1 em agora.
// jogadores[i] joga com estado.Jogadores[i].
func Nova(cfg Config, estado jogo.Estado, jogadores []jogo.Bot, f fila.Fila, agora time.Time) (*Partida, error) {
	if len(jogadores) != len(estado.Jogadores) {
		return nil, fmt.Errorf("%d bots para %d jogadores", len(jogadores), len(estado.Jogadores))
	}
	planos, err := f.Consumir(fila.PlanosEnviados)
	if err != nil {
		return nil, err
	}
	p := &Partida{
		cfg:       cfg,
		fila:      f,
		planos:    planos,
		jogadores: jogadores,
		estado:    estado.Copiar(),
		fase:      Planejamento,
		inicio:    agora,
		historico: []RegistroTurno{},
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.iniciarTurno(agora)
	return p, nil
}

// Avancar leva a partida até agora e devolve o próximo instante em que algo
// muda (fim da fase) e se ela está encerrada.
func (p *Partida) Avancar(agora time.Time) (proximo time.Time, encerrada bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.avancar(agora)
}

// avancar atravessa as fases que terminaram até agora, em ordem (D7).
func (p *Partida) avancar(agora time.Time) (time.Time, bool) {
	for {
		switch p.fase {
		case Planejamento:
			fim := p.inicio.Add(p.prazo())
			if agora.Before(fim) {
				return fim, false
			}
			p.resolverTurno(fim)
		case Execucao:
			fim := p.fimDaExecucao()
			if agora.Before(fim) {
				return fim, false
			}
			p.fecharTurno(fim)
		default:
			return time.Time{}, true
		}
	}
}

// Visao avança até agora e devolve o que se vê da partida.
func (p *Partida) Visao(agora time.Time) Visao {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.avancar(agora)
	v := Visao{Nome: p.cfg.Nome, Fase: p.fase, Etapas: []jogo.RelatorioEtapa{}}
	switch p.fase {
	case Planejamento:
		v.Turno, v.Estado = p.estado.Turno, p.estado.Copiar()
		v.FimDaFase = p.inicio.Add(p.prazo())
		v.Desfecho = jogo.VerificarFim(p.estado)
	case Execucao:
		etapas := p.resolvido.registro.Etapas
		k := len(etapas)
		if d := p.duracaoEtapa(); d > 0 {
			k = min(k, 1+int(agora.Sub(p.inicio)/d))
		}
		v.Turno, v.Etapa, v.Estado = p.estado.Turno, k, p.estado.Copiar()
		v.Etapas = append(v.Etapas, etapas[:k]...)
		v.FimDaFase = p.fimDaExecucao()
		v.Desfecho = jogo.VerificarFim(p.estado)
	case Encerrada:
		ultimo := p.historico[len(p.historico)-1]
		v.Turno, v.Etapa, v.Estado = ultimo.Turno, len(ultimo.Etapas), ultimo.Estado.Copiar()
		v.Etapas = append(v.Etapas, ultimo.Etapas...)
		v.Desfecho = *p.desfecho
	}
	return v
}

// Historico devolve uma cópia do histórico: só os turnos totalmente
// executados (D6).
func (p *Partida) Historico() Historico {
	p.mu.Lock()
	defer p.mu.Unlock()
	h := Historico{
		Nome:    p.cfg.Nome,
		Mapa:    p.cfg.Mapa,
		Bots:    append([]string{}, p.cfg.Bots...),
		Semente: p.cfg.Semente,
		Turnos:  append([]RegistroTurno{}, p.historico...),
	}
	if p.desfecho != nil {
		d := *p.desfecho
		h.Desfecho = &d
	}
	return h
}

func (p *Partida) prazo() time.Duration {
	return time.Duration(p.estado.Config.PrazoPlanejamentoMs) * time.Millisecond
}

func (p *Partida) duracaoEtapa() time.Duration {
	return time.Duration(p.estado.Config.DuracaoEtapaMs) * time.Millisecond
}

func (p *Partida) fimDaExecucao() time.Time {
	return p.inicio.Add(time.Duration(len(p.resolvido.registro.Etapas)) * p.duracaoEtapa())
}

// iniciarTurno começa o planejamento do turno atual em agora: dispara um
// bot por jogador vivo, e cada um publica o resultado em planos-enviados (D2, D3).
func (p *Partida) iniciarTurno(agora time.Time) {
	p.fase, p.inicio = Planejamento, agora
	for i, j := range p.estado.Jogadores {
		if j.Status != jogo.Vivo {
			continue
		}
		go p.chamarBot(p.jogadores[i], p.estado.Copiar(), j.ID)
	}
}

// chamarBot roda o bot sem prazo (o prazo é o fim da fase) e publica o plano.
func (p *Partida) chamarBot(bot jogo.Bot, estado jogo.Estado, jogadorID string) {
	r := bots.Chamar(bot, estado, jogadorID, 0)
	valor, err := json.Marshal(PlanoEnviado{
		Partida:   p.cfg.Nome,
		Turno:     estado.Turno,
		JogadorID: jogadorID,
		Acoes:     r.Acoes,
		Falha:     r.Falha,
		Detalhe:   r.Detalhe,
	})
	if err != nil {
		return
	}
	p.fila.Publicar(fila.PlanosEnviados, fila.Mensagem{Chave: p.cfg.Nome, Valor: valor})
}

// resolverTurno fecha o planejamento em fim: lê os planos da fila, valida e
// resolve o turno inteiro (PAR-03, D6).
func (p *Partida) resolverTurno(fim time.Time) {
	recebidos := p.lerPlanos()
	r := RegistroTurno{Partida: p.cfg.Nome, Turno: p.estado.Turno, Estado: p.estado.Copiar(), Jogadores: []RegistroJogador{}}
	var planos []jogo.Plano
	for _, j := range p.estado.Jogadores {
		if j.Status != jogo.Vivo {
			continue
		}
		rj := RegistroJogador{JogadorID: j.ID, Planejadas: []jogo.Acao{}}
		if pe, ok := recebidos[j.ID]; ok {
			rj.Planejadas = append(rj.Planejadas, pe.Acoes...)
			rj.Falha, rj.Detalhe = pe.Falha, pe.Detalhe
		} else {
			rj.Falha = bots.PrazoEstourado // PAR-04
		}
		validado, infracoes := jogo.Validar(p.estado, jogo.Plano{JogadorID: j.ID, Turno: p.estado.Turno, Acoes: rj.Planejadas})
		rj.Validadas = append([]jogo.Acao{}, validado.Acoes...)
		rj.Infracoes = append([]jogo.Infracao{}, infracoes...)
		planos = append(planos, validado)
		r.Jogadores = append(r.Jogadores, rj)
	}
	proximo, etapas := jogo.ResolverTurno(p.estado, planos)
	r.Etapas = etapas
	for i := range r.Jogadores {
		r.Jogadores[i].Executadas = executadas(r.Jogadores[i].JogadorID, etapas)
	}
	p.resolvido = &turnoResolvido{registro: r, proximo: proximo}
	p.fase, p.inicio = Execucao, fim
}

// lerPlanos lê o consumidor e fica, para cada jogador, com o primeiro plano
// desta partida e deste turno (D4, RES-01). Planos atrasados de turnos
// anteriores são descartados aqui.
func (p *Partida) lerPlanos() map[string]PlanoEnviado {
	recebidos := map[string]PlanoEnviado{}
	mensagens, _ := p.planos.Ler()
	for _, m := range mensagens {
		var pe PlanoEnviado
		if m.Chave != p.cfg.Nome || json.Unmarshal(m.Valor, &pe) != nil {
			continue
		}
		if _, repetido := recebidos[pe.JogadorID]; pe.Partida == p.cfg.Nome && pe.Turno == p.estado.Turno && !repetido {
			recebidos[pe.JogadorID] = pe
		}
	}
	return recebidos
}

// executadas tira dos relatórios a ação e o resultado de cada etapa do jogador (DEC-09).
func executadas(jogadorID string, etapas []jogo.RelatorioEtapa) []AcaoExecutada {
	r := []AcaoExecutada{}
	for _, e := range etapas {
		for _, j := range e.Jogadores {
			if j.ID == jogadorID {
				r = append(r, AcaoExecutada{Etapa: e.Etapa, Acao: j.Acao, Resultado: j.Resultado})
			}
		}
	}
	return r
}

// fecharTurno termina a execução em fim: o turno entra no histórico e em
// turno-resolvido, e começa o turno seguinte ou a partida se encerra (D6).
func (p *Partida) fecharTurno(fim time.Time) {
	r := p.resolvido
	p.resolvido = nil
	p.historico = append(p.historico, r.registro)
	p.publicar(fila.TurnoResolvido, r.registro)
	p.estado = r.proximo
	if d := jogo.VerificarFim(p.estado); d.Terminada {
		p.fase, p.desfecho = Encerrada, &d
		p.publicar(fila.PartidaFinalizada, Finalizada{Partida: p.cfg.Nome, Desfecho: d})
		p.planos.Fechar()
		return
	}
	p.iniciarTurno(fim)
}

func (p *Partida) publicar(topico string, v any) {
	if valor, err := json.Marshal(v); err == nil {
		p.fila.Publicar(topico, fila.Mensagem{Chave: p.cfg.Nome, Valor: valor})
	}
}
