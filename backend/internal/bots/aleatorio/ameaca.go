package aleatorio

import "github.com/kustavo/bomber-guess/backend/internal/jogo"

// ameaca é a zona de ameaça das bombas possíveis dos adversários: as que eles
// ainda podem plantar neste turno (spec 16, "Termos"). O valor de cada casa é
// o peso: quantas bombas possíveis a alcançam (R1, D9).
type ameaca struct {
	chamas []map[jogo.Posicao]int // índice = etapa (1..etapas); [0] vazio
	final  map[jogo.Posicao]int   // bombas possíveis que explodem depois do turno
}

// calcularAmeaca monta a ameaça dos adversários vivos de `eu`, com
// bombas_por_turno ≥ 1, para um turno de `etapas` etapas (D1 do marco 16).
// Cada bomba possível é isolada: sem pilhas e sem reações em cadeia
// (decisão 2); o alcance do adversário não atravessa blocos.
func calcularAmeaca(estado jogo.Estado, eu jogo.Jogador, etapas int) ameaca {
	a := ameaca{chamas: make([]map[jogo.Posicao]int, etapas+1), final: map[jogo.Posicao]int{}}
	for i := range a.chamas {
		a.chamas[i] = map[jogo.Posicao]int{}
	}
	bloq := bloqueios(estado)
	for _, adv := range estado.Jogadores {
		if adv.ID == eu.ID || adv.Status != jogo.Vivo || adv.BombasPorTurno < 1 { // FIM-01, BOM-02
			continue
		}
		dist := distancias(bloq, estado.Config, []jogo.Posicao{adv.Posicao})
		for casa, d := range dist {
			chamas := alcanceDaBomba(bloq, estado.Config, casa, adv.Potencia)
			// Plantada na etapa p (d+1 ≤ p ≤ acoes), explode em p + pavio (BOM-05).
			for p := d + 1; p <= adv.AcoesPorTurno; p++ {
				destino := a.final
				if t := p + adv.PavioPadrao; t <= etapas {
					destino = a.chamas[t]
				}
				for _, c := range chamas {
					destino[c]++
				}
			}
		}
	}
	return a
}

// nivel devolve a ameaça só com as casas de peso ≥ w (R1, D9).
func (a ameaca) nivel(w int) ameaca {
	filtrar := func(m map[jogo.Posicao]int) map[jogo.Posicao]int {
		r := map[jogo.Posicao]int{}
		for c, p := range m {
			if p >= w {
				r[c] = p
			}
		}
		return r
	}
	n := ameaca{chamas: make([]map[jogo.Posicao]int, len(a.chamas)), final: filtrar(a.final)}
	for i, m := range a.chamas {
		n.chamas[i] = filtrar(m)
	}
	return n
}

// maiorPeso devolve o maior peso de uma casa da ameaça (0 se vazia).
func (a ameaca) maiorPeso() int {
	maior := 0
	for _, m := range append([]map[jogo.Posicao]int{a.final}, a.chamas...) {
		for _, p := range m {
			maior = max(maior, p)
		}
	}
	return maior
}

// niveis devolve os níveis da ameaça graduada: 1, 2, 3, 5, 8… até o
// primeiro que passa de maior (decisão 7).
func niveis(maior int) []int {
	r := []int{}
	for a, b := 1, 2; a <= maior; a, b = b, a+b {
		r = append(r, a)
	}
	return r
}

// vazia informa se nenhuma casa está ameaçada.
func (a ameaca) vazia() bool {
	if len(a.final) > 0 {
		return false
	}
	for _, casas := range a.chamas {
		if len(casas) > 0 {
			return false
		}
	}
	return true
}

// alcanceDaBomba devolve as casas das chamas de uma bomba em `origem` com a potência
// dada, parando nos blocos (BOM-06, BOM-07), incluindo a casa do bloco.
func alcanceDaBomba(bloq map[jogo.Posicao]bool, c jogo.Config, origem jogo.Posicao, potencia int) []jogo.Posicao {
	casas := []jogo.Posicao{origem}
	for _, d := range direcoes {
		p := origem
		for range potencia {
			p = p.Vizinha(d)
			if !c.NoTabuleiro(p) {
				break
			}
			casas = append(casas, p)
			if bloq[p] {
				break
			}
		}
	}
	return casas
}

// protegido informa se o plano completo de `eu` é protegido (D5 do marco 16):
// na simulação do turno, com os adversários parados, o jogador fica vivo, não
// está em casa da ameaça (nem em chamas) ao fim de nenhuma etapa e termina
// fora da zona de perigo final somada à ameaça final. Devolve também a casa
// em que o jogador termina.
func protegido(estado jogo.Estado, eu jogo.Jogador, etapas int, plano []jogo.Acao, a ameaca) (bool, jogo.Posicao) {
	l := preverLinha(estado, etapas, &eu, plano, &a)
	aux := auxiliar(estado, etapas)
	aux.Jogadores = append(aux.Jogadores, eu)
	_, relatorios := jogo.ResolverTurno(aux, []jogo.Plano{{JogadorID: eu.ID, Turno: aux.Turno, Acoes: plano}})
	casa, ok := eu.Posicao, true
	for _, r := range relatorios {
		for _, j := range r.Jogadores {
			if j.ID == eu.ID {
				casa = j.Posicao
				if j.Status != jogo.Vivo || l.chamas[r.Etapa][casa] {
					ok = false
				}
			}
		}
	}
	return ok && !l.perigo[casa], casa
}
