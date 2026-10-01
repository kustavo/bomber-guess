package jogo

import "fmt"

// Infracao é uma ação rejeitada pelo validador (VAL-05). Cada plano gera no
// máximo uma.
type Infracao struct {
	JogadorID string `json:"jogador_id"`
	Etapa     int    `json:"etapa"`  // posição (1, 2, 3…) da ação rejeitada; 1 para VAL-07 e VAL-08
	Regra     string `json:"regra"`  // "VAL-01" … "VAL-08"
	Motivo    string `json:"motivo"` // texto para humanos
}

// Validar devolve o plano executável e as infrações (VAL-01 a VAL-08, DEC-08).
// As ações além de acoes_por_turno são removidas; a primeira ação inválida e
// todas as seguintes viram ESPERAR (D2). Não altera o estado nem o plano
// recebidos.
func Validar(estado Estado, plano Plano) (Plano, []Infracao) {
	validado := Plano{JogadorID: plano.JogadorID, Turno: plano.Turno, Acoes: []Acao{}}
	infracao := func(etapa int, regra, motivo string, args ...any) []Infracao {
		return []Infracao{{JogadorID: plano.JogadorID, Etapa: etapa, Regra: regra, Motivo: fmt.Sprintf(motivo, args...)}}
	}

	jogador, ok := buscarJogador(estado, plano.JogadorID)
	if !ok {
		return validado, infracao(1, "VAL-08", "jogador %q não existe", plano.JogadorID)
	}
	if jogador.Status != Vivo {
		return validado, nil
	}

	n := min(len(plano.Acoes), jogador.AcoesPorTurno)
	esperarDesde := func(i int) {
		for ; i < n; i++ {
			validado.Acoes = append(validado.Acoes, Acao{Etapa: i + 1, Tipo: Esperar})
		}
	}
	if plano.Turno != estado.Turno {
		esperarDesde(0)
		return validado, infracao(1, "VAL-07", "plano do turno %d, turno atual %d", plano.Turno, estado.Turno)
	}

	fixos := conjunto(estado.BlocosFixos)
	posicao, bombas := jogador.Posicao, 0
	for i, a := range plano.Acoes {
		etapa := i + 1
		var regra, motivo string
		switch {
		case i >= jogador.AcoesPorTurno:
			return validado, infracao(etapa, "VAL-01", "%d ações, limite %d", len(plano.Acoes), jogador.AcoesPorTurno)
		case a.Etapa != etapa:
			regra, motivo = "VAL-06", fmt.Sprintf("ação na posição %d com etapa %d", etapa, a.Etapa)
		case a.Tipo == Mover && !direcaoValida(a.Direcao):
			regra, motivo = "VAL-01", fmt.Sprintf("MOVER com direção %q inválida", a.Direcao)
		case a.Tipo == Mover:
			destino := posicao.Vizinha(a.Direcao)
			if !estado.Config.NoTabuleiro(destino) {
				regra, motivo = "VAL-02", fmt.Sprintf("movimento para %s fora do tabuleiro", destino)
			} else if fixos[destino] {
				regra, motivo = "VAL-02", fmt.Sprintf("movimento para bloco fixo em %s", destino)
			} else {
				posicao = destino
			}
		case a.Tipo == Plantar:
			bombas++
			if bombas > jogador.BombasPorTurno {
				regra, motivo = "VAL-03", fmt.Sprintf("%dª bomba no turno, limite %d", bombas, jogador.BombasPorTurno)
			}
		case a.Tipo != Esperar:
			regra, motivo = "VAL-01", fmt.Sprintf("tipo de ação %q inválido", a.Tipo)
		}
		if regra != "" {
			esperarDesde(i)
			return validado, infracao(etapa, regra, "%s", motivo)
		}
		if a.Tipo != Mover {
			a.Direcao = "" // DEC-08
		}
		validado.Acoes = append(validado.Acoes, a)
	}
	return validado, nil
}

// buscarJogador devolve o jogador com o id dado.
func buscarJogador(estado Estado, id string) (Jogador, bool) {
	for _, j := range estado.Jogadores {
		if j.ID == id {
			return j, true
		}
	}
	return Jogador{}, false
}

// direcaoValida informa se d é uma das quatro direções (ACA-01).
func direcaoValida(d Direcao) bool {
	return d == Cima || d == Baixo || d == Esquerda || d == Direita
}

// conjunto devolve as posições como conjunto.
func conjunto(posicoes []Posicao) map[Posicao]bool {
	c := make(map[Posicao]bool, len(posicoes))
	for _, p := range posicoes {
		c[p] = true
	}
	return c
}
