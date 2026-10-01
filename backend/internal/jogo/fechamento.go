package jogo

import "slices"

// Anel devolve o índice do anel da casa p: a distância até a borda mais
// próxima (FEC-02). A borda é o anel 0.
func (c Config) Anel(p Posicao) int {
	return min(p.X, p.Y, c.Largura-1-p.X, c.Altura-1-p.Y)
}

// CasasQueFecham devolve as casas que viram bloco fixo ao fim do turno do
// estado (FEC-01, FEC-03), ordenadas por y e depois por x. Casas que já têm
// bloco fixo ficam de fora. Vazio se o fechamento estiver desligado, se o
// turno for anterior a turno_fechamento ou se o tabuleiro já tiver fechado.
func CasasQueFecham(estado Estado) []Posicao {
	c := estado.Config
	casas := []Posicao{}
	if c.TurnoFechamento < 1 || estado.Turno < c.TurnoFechamento {
		return casas
	}
	anel := estado.Turno - c.TurnoFechamento
	fixos := conjunto(estado.BlocosFixos)
	for y := range c.Altura {
		for x := range c.Largura {
			p := Posicao{X: x, Y: y}
			if c.Anel(p) == anel && !fixos[p] {
				casas = append(casas, p)
			}
		}
	}
	return casas
}

// fechar aplica o fechamento do fim do turno (ORD-08, FEC-03 a FEC-06) e o
// registra no relatório da última etapa executada.
func (m *mesa) fechar(r *RelatorioEtapa) {
	casas := CasasQueFecham(Estado{Turno: m.turno, Config: m.config, BlocosFixos: m.blocosFixos})
	if len(casas) == 0 {
		return
	}
	fecha := conjunto(casas)
	m.blocosFixos = append(slices.Clip(m.blocosFixos), casas...) // não altera o estado recebido
	for _, p := range casas {
		m.fixos[p] = true
		delete(m.destrutiveis, p) // FEC-05
	}
	m.bombas = slices.DeleteFunc(m.bombas, func(b bombaViva) bool { return fecha[b.Posicao] }) // FEC-06
	for i := range m.jogadores {                                                               // FEC-04
		j := &m.jogadores[i]
		if j.Status == Vivo && fecha[j.Posicao] {
			j.Status = Morto
			j.Morte = &Morte{Turno: m.turno, Etapa: r.Etapa}
			r.Mortes = append(r.Mortes, j.ID)
			r.Jogadores[i].Status = Morto
		}
	}
	r.Bombas = m.bombasOrdenadas()
	r.BlocosFechados = casas
}
