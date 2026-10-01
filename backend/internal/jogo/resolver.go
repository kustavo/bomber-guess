package jogo

import (
	"cmp"
	"slices"
)

// bombaViva é uma bomba no tabuleiro durante a resolução.
type bombaViva struct {
	Bomba
	plantadaNaEtapa int // 0 para bombas de turnos anteriores
}

// jogadorVivo é um jogador durante a resolução, com o plano do turno.
type jogadorVivo struct {
	Jogador
	acoes     []Acao // no máximo acoes_por_turno (D3)
	abortado  bool   // MOV-05: o resto do plano vira ESPERAR
	plantadas int    // bombas plantadas neste turno (RES-02)
}

// mesa é o estado de trabalho de ResolverTurno. É montada a partir de cópias,
// então pode ser alterada à vontade.
type mesa struct {
	turno, etapa       int
	config             Config
	fixos              map[Posicao]bool
	destrutiveis       map[Posicao]bool
	blocosFixos        []Posicao // ordem original, para o novo Estado (D7)
	blocosDestrutiveis []Posicao // ordem original, para o novo Estado (D7)
	bombas             []bombaViva
	jogadores          []jogadorVivo
}

// ResolverTurno resolve o turno etapa a etapa (ORD-01 a ORD-06) e devolve o
// estado do turno seguinte e um relatório por etapa executada. Para no fim da
// etapa em que a partida termina (DEC-06). Para uma partida já terminada,
// devolve uma cópia do estado e nenhum relatório. Não altera as entradas.
func ResolverTurno(estado Estado, planos []Plano) (Estado, []RelatorioEtapa) {
	if VerificarFim(estado).Terminada {
		return estado.Copiar(), []RelatorioEtapa{}
	}
	m := novaMesa(estado, planos)
	etapas := CalcularEtapas(estado.Jogadores) // D4
	relatorios := []RelatorioEtapa{}
	for m.etapa = 1; m.etapa <= etapas; m.etapa++ {
		relatorios = append(relatorios, m.resolverEtapa())
		if m.vivos() <= 1 { // DEC-06
			break
		}
	}
	if m.vivos() > 1 { // FEC-07
		m.fechar(&relatorios[len(relatorios)-1]) // ORD-08
	}
	return m.novoEstado(), relatorios
}

// novaMesa copia o estado e associa a cada jogador vivo o primeiro plano com
// o seu id (RES-01, D8).
func novaMesa(estado Estado, planos []Plano) *mesa {
	m := &mesa{
		turno:              estado.Turno,
		config:             estado.Config,
		fixos:              conjunto(estado.BlocosFixos),
		destrutiveis:       conjunto(estado.BlocosDestrutiveis),
		blocosFixos:        estado.BlocosFixos,
		blocosDestrutiveis: estado.BlocosDestrutiveis,
	}
	for _, b := range estado.Bombas {
		m.bombas = append(m.bombas, bombaViva{Bomba: b})
	}
	acoesPorJogador := map[string][]Acao{}
	for _, p := range planos {
		if _, repetido := acoesPorJogador[p.JogadorID]; !repetido {
			acoesPorJogador[p.JogadorID] = p.Acoes
		}
	}
	for _, j := range estado.Copiar().Jogadores {
		jv := jogadorVivo{Jogador: j}
		if j.Status == Vivo {
			a := acoesPorJogador[j.ID]
			jv.acoes = a[:min(len(a), j.AcoesPorTurno)]
		}
		m.jogadores = append(m.jogadores, jv)
	}
	return m
}

// resolverEtapa executa a etapa atual na ordem ORD-01 a ORD-06.
func (m *mesa) resolverEtapa() RelatorioEtapa {
	r := RelatorioEtapa{
		Turno:                m.turno,
		Etapa:                m.etapa,
		Jogadores:            make([]JogadorEtapa, len(m.jogadores)),
		Mortes:               []string{},
		MovimentosBloqueados: []string{},
		BlocosFechados:       []Posicao{},
	}
	var plantios []Bomba
	for i := range m.jogadores {
		j := &m.jogadores[i]
		acao, resultado, plantar := m.executar(j) // ORD-01
		if resultado == Bloqueada {
			r.MovimentosBloqueados = append(r.MovimentosBloqueados, j.ID)
		}
		if plantar { // BOM-01: casa atual, atributos atuais
			plantios = append(plantios, Bomba{Posicao: j.Posicao, JogadorID: j.ID, Potencia: j.Potencia, PavioRestante: j.PavioPadrao})
		}
		r.Jogadores[i] = JogadorEtapa{ID: m.jogadores[i].ID, Acao: acao, Resultado: resultado}
	}
	m.plantar(plantios) // ORD-02
	m.queimarPavio()    // ORD-03

	bombas := make([]Bomba, len(m.bombas))
	for i, b := range m.bombas {
		bombas[i] = b.Bomba
	}
	ex := explodir(m.config, m.fixos, m.destrutiveis, bombas) // ORD-04
	m.bombas = slices.DeleteFunc(m.bombas, func(b bombaViva) bool { return ex.explodidas[b.Posicao] })

	for i := range m.jogadores { // ORD-05
		j := &m.jogadores[i]
		if j.Status == Vivo && ex.chamas[j.Posicao] {
			j.Status = Morto
			j.Morte = &Morte{Turno: m.turno, Etapa: m.etapa} // FIM-01, DEC-02
			r.Mortes = append(r.Mortes, j.ID)
		}
	}
	for p := range ex.destruidos { // ORD-06
		delete(m.destrutiveis, p)
	}

	for i, j := range m.jogadores {
		r.Jogadores[i].Posicao, r.Jogadores[i].Status = j.Posicao, j.Status
	}
	r.Bombas = m.bombasOrdenadas()
	r.Explosoes = ex.explosoes
	r.Chamas = posicoesOrdenadas(ex.chamas)
	r.BlocosDestruidos = posicoesOrdenadas(ex.destruidos)
	return r
}

// executar aplica o movimento da etapa (ORD-01) e decide se o jogador planta.
// Devolve a ação do plano para a etapa e o que aconteceu com ela (DEC-09).
func (m *mesa) executar(j *jogadorVivo) (acao Acao, resultado ResultadoAcao, plantar bool) {
	acao = Acao{Tipo: Esperar}
	if i := m.etapa - 1; i < len(j.acoes) {
		acao = j.acoes[i]
	}
	acao.Etapa = m.etapa // D3
	switch {
	case j.Status != Vivo:
		return Acao{Etapa: m.etapa, Tipo: Esperar}, Descartada, false // FIM-01
	case j.abortado:
		return acao, Abortada, false // MOV-05
	case acao.Tipo == Esperar:
		return acao, Executada, false
	case acao.Tipo == Plantar:
		if j.plantadas >= j.BombasPorTurno { // BOM-02, RES-02
			return acao, Ignorada, false
		}
		j.plantadas++
		return acao, Executada, true
	case acao.Tipo == Mover && direcaoValida(acao.Direcao):
		destino := j.Posicao.Vizinha(acao.Direcao)
		switch {
		case !m.config.NoTabuleiro(destino) || m.fixos[destino]: // MOV-02, RES-02
			return acao, Ignorada, false
		case m.destrutiveis[destino]: // MOV-03, MOV-05
			j.abortado = true
			return acao, Bloqueada, false
		}
		j.Posicao = destino // MOV-01, MOV-04
		return acao, Executada, false
	}
	return acao, Ignorada, false // RES-02: tipo ou direção desconhecidos
}

// plantar coloca as bombas da etapa (ORD-02), na ordem dos jogadores.
func (m *mesa) plantar(bombas []Bomba) {
	for _, b := range bombas {
		m.bombas = append(m.bombas, bombaViva{Bomba: b, plantadaNaEtapa: m.etapa})
	}
}

// queimarPavio diminui o pavio das bombas não plantadas nesta etapa (ORD-03, BOM-05).
func (m *mesa) queimarPavio() {
	for i := range m.bombas {
		if m.bombas[i].plantadaNaEtapa != m.etapa {
			m.bombas[i].PavioRestante--
		}
	}
}

// vivos conta os jogadores vivos.
func (m *mesa) vivos() int {
	n := 0
	for _, j := range m.jogadores {
		if j.Status == Vivo {
			n++
		}
	}
	return n
}

// bombasOrdenadas devolve as bombas por posição, jogador e pavio (D7).
func (m *mesa) bombasOrdenadas() []Bomba {
	bombas := make([]Bomba, len(m.bombas))
	for i, b := range m.bombas {
		bombas[i] = b.Bomba
	}
	slices.SortStableFunc(bombas, func(a, b Bomba) int {
		return cmp.Or(compararPosicoes(a.Posicao, b.Posicao), cmp.Compare(a.JogadorID, b.JogadorID), cmp.Compare(a.PavioRestante, b.PavioRestante))
	})
	return bombas
}

// novoEstado monta o Estado do turno seguinte (D5, BOM-11, EST-03, DEC-03).
func (m *mesa) novoEstado() Estado {
	e := Estado{
		Turno:              m.turno + 1,
		Config:             m.config,
		BlocosFixos:        []Posicao{},
		BlocosDestrutiveis: []Posicao{},
		Bombas:             []Bomba{},
		Jogadores:          []Jogador{},
	}
	e.BlocosFixos = append(e.BlocosFixos, m.blocosFixos...)
	for _, p := range m.blocosDestrutiveis {
		if m.destrutiveis[p] {
			e.BlocosDestrutiveis = append(e.BlocosDestrutiveis, p)
		}
	}
	for _, b := range m.bombas {
		e.Bombas = append(e.Bombas, b.Bomba)
	}
	for _, j := range m.jogadores {
		e.Jogadores = append(e.Jogadores, j.Jogador)
	}
	e.EtapasNesteTurno = CalcularEtapas(e.Jogadores)
	return e
}
