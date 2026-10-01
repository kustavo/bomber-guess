package aleatorio

import "github.com/kustavo/bomber-guess/backend/internal/jogo"

// mira mede quanto as bombas do plano ameaçam os adversários (decisão 8 e D11
// do marco 16). Para cada PLANTAR na etapa p, na casa em que o bot está, a
// explosão é na etapa t = p + pavio_padrao (sem cadeias), e as casas possíveis
// de cada adversário são as que ele alcança em min(t, acoes_por_turno) passos
// (acoes_por_turno se t passa do turno). A mira é a soma, por adversário vivo,
// da maior fração dessas casas que as chamas de alguma bomba alcançam.
func mira(estado jogo.Estado, eu jogo.Jogador, etapas int, plano []jogo.Acao) float64 {
	bloq := bloqueios(estado)
	type bomba struct {
		chamas map[jogo.Posicao]bool
		t      int
	}
	var bombas []bomba
	casa := eu.Posicao
	for i, acao := range plano {
		switch acao.Tipo {
		case jogo.Mover:
			casa = casa.Vizinha(acao.Direcao)
		case jogo.Plantar:
			bombas = append(bombas, bomba{
				chamas: conjunto(alcanceDaBomba(bloq, estado.Config, casa, eu.Potencia)),
				t:      i + 1 + eu.PavioPadrao,
			})
		}
	}
	if len(bombas) == 0 {
		return 0
	}
	total := 0.0
	for _, adv := range estado.Jogadores {
		if adv.ID == eu.ID || adv.Status != jogo.Vivo {
			continue
		}
		dist := distancias(bloq, estado.Config, []jogo.Posicao{adv.Posicao})
		melhor := 0.0
		for _, b := range bombas {
			passos := adv.AcoesPorTurno
			if b.t <= etapas {
				passos = min(b.t, adv.AcoesPorTurno)
			}
			possiveis, cobertas := 0, 0
			for c, d := range dist {
				if d > passos {
					continue
				}
				possiveis++
				if b.chamas[c] {
					cobertas++
				}
			}
			if possiveis > 0 {
				melhor = max(melhor, float64(cobertas)/float64(possiveis))
			}
		}
		total += melhor
	}
	return total
}

// conjunto transforma a lista de casas num conjunto.
func conjunto(casas []jogo.Posicao) map[jogo.Posicao]bool {
	r := make(map[jogo.Posicao]bool, len(casas))
	for _, c := range casas {
		r[c] = true
	}
	return r
}
