package aleatorio

import "github.com/kustavo/bomber-guess/backend/internal/jogo"

// linha é a previsão de um turno com os adversários parados.
type linha struct {
	chamas []map[jogo.Posicao]bool // índice = etapa (1..etapas); [0] vazio
	perigo map[jogo.Posicao]bool   // zona de perigo final
}

// fora é a casa dos fantasmas: nenhuma chama a alcança (D2).
var fora = jogo.Posicao{X: -1, Y: -1}

// preverLinha roda ResolverTurno no estado auxiliar: blocos e bombas do
// estado, dois fantasmas fora do tabuleiro com acoes_por_turno = etapas e, se
// eu != nil, o próprio jogador com as ações dadas (D1, D2, D4). A zona de
// perigo final é a união das chamas de todas as bombas que sobram, explodindo
// juntas (D3).
func preverLinha(estado jogo.Estado, etapas int, eu *jogo.Jogador, acoes []jogo.Acao) linha {
	aux := auxiliar(estado, etapas)
	var planos []jogo.Plano
	if eu != nil {
		aux.Jogadores = append(aux.Jogadores, *eu)
		planos = append(planos, jogo.Plano{JogadorID: eu.ID, Turno: aux.Turno, Acoes: acoes})
	}
	final, relatorios := jogo.ResolverTurno(aux, planos)

	l := linha{chamas: make([]map[jogo.Posicao]bool, etapas+1), perigo: map[jogo.Posicao]bool{}}
	for i := range l.chamas {
		l.chamas[i] = map[jogo.Posicao]bool{}
	}
	for _, r := range relatorios {
		for _, p := range r.Chamas {
			l.chamas[r.Etapa][p] = true
		}
	}

	resto := auxiliar(final, 1)
	for i := range resto.Bombas {
		resto.Bombas[i].PavioRestante = 1
	}
	_, relatorios = jogo.ResolverTurno(resto, nil)
	for _, r := range relatorios {
		for _, p := range r.Chamas {
			l.perigo[p] = true
		}
	}
	return l
}

// auxiliar copia blocos e bombas do estado e troca os jogadores por dois
// fantasmas vivos fora do tabuleiro, que nunca morrem, de modo que a partida
// não termina antes da hora (D2, DEC-06, DEC-07).
func auxiliar(estado jogo.Estado, etapas int) jogo.Estado {
	aux := estado.Copiar()
	aux.Turno = 1
	aux.Config.LimiteTurnos = 1
	aux.EtapasNesteTurno = etapas
	aux.Jogadores = []jogo.Jogador{}
	for _, id := range []string{"~fantasma_1", "~fantasma_2"} {
		aux.Jogadores = append(aux.Jogadores, jogo.Jogador{
			ID:        id,
			Posicao:   fora,
			Status:    jogo.Vivo,
			Atributos: jogo.Atributos{AcoesPorTurno: etapas},
		})
	}
	return aux
}
