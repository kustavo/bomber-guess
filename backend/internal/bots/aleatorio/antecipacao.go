package aleatorio

import (
	"math/rand/v2"
	"slices"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// planejarNaJanela escolhe o plano do v2 dentro da janela de antecipação,
// pela ordem da decisão 3 do marco 13. Devolve os passos de todas as etapas
// do turno.
func planejarNaJanela(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas int, a *ameaca) []jogo.Acao {
	c := estado.Config
	l := preverLinha(estado, etapas, nil, nil, a)
	bloq := bloqueios(estado)
	alvo := anelAlvo(c, estado.Turno)
	reg := regiao(bloq, c, eu.Posicao)

	var destinos []jogo.Posicao
	for casa := range reg {
		if c.Anel(casa) >= alvo {
			destinos = append(destinos, casa)
		}
	}
	preso := len(destinos) == 0
	if preso {
		if bloco, ok := blocoDeAbertura(estado, eu.Posicao, alvo); ok {
			destinos = casasDePlantio(estado, reg, bloco, eu.Potencia)
		}
	}
	slices.SortFunc(destinos, compararPosicoes) // ordem fixa antes de sortear
	dist := distancias(bloq, c, destinos)

	// Item 2: preso, bomba de abertura (D7).
	if preso && len(destinos) > 0 && eu.BombasPorTurno >= 1 {
		if passos, ok := tentarAbrir(rng, estado, eu, etapas, destinos, bloq, a); ok {
			return completar(passos, etapas)
		}
	}

	buscar := func(obj objetivo) ([]jogo.Acao, bool) {
		passos, _, ok := novaBusca(rng, estado, l, eu.AcoesPorTurno, etapas, obj).ir(eu.Posicao, 1)
		return completar(passos, etapas), ok
	}
	// Itens 1 e 3: seguro, o mais perto possível dos destinos (D4).
	if atual, ok := dist[eu.Posicao]; ok {
		for d := 0; d <= atual; d++ {
			perto := func(l linha, casa jogo.Posicao) bool {
				dc, alcanca := dist[casa]
				return alcanca && dc <= d && !l.perigo[casa]
			}
			if passos, ok := buscar(perto); ok {
				return passos
			}
		}
	}
	if passos, ok := buscar(buscaSegura); ok {
		return passos
	}
	// Item 4: sobrevivente fora das casas que fecham (D9).
	if fecham := jogo.CasasQueFecham(estado); len(fecham) > 0 {
		fecha := map[jogo.Posicao]bool{}
		for _, p := range fecham {
			fecha[p] = true
		}
		if passos, ok := buscar(func(_ linha, casa jogo.Posicao) bool { return !fecha[casa] }); ok {
			return passos
		}
	}
	// Itens 5 e 6: sobrevivente, senão o que sobrevive mais etapas.
	passos, _ := buscar(buscaSobrevivente)
	return passos
}

// tentarAbrir tenta plantar a bomba de abertura numa das casas de plantio,
// da mais perto para a mais longe (empates sorteados), indo pelo caminho mais
// curto e plantando na etapa seguinte à chegada. Aceita o primeiro plano em
// que o jogador sobrevive até plantar e tem continuação segura (D7).
func tentarAbrir(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas int, plantio []jogo.Posicao, bloq map[jogo.Posicao]bool, a *ameaca) ([]jogo.Acao, bool) {
	c := estado.Config
	deMim := distancias(bloq, c, []jogo.Posicao{eu.Posicao})
	var candidatas []jogo.Posicao
	for _, casa := range plantio {
		if d, ok := deMim[casa]; ok && d < eu.AcoesPorTurno {
			candidatas = append(candidatas, casa)
		}
	}
	rng.Shuffle(len(candidatas), func(i, j int) { candidatas[i], candidatas[j] = candidatas[j], candidatas[i] })
	slices.SortStableFunc(candidatas, func(a, b jogo.Posicao) int { return deMim[a] - deMim[b] })
	for _, casa := range candidatas[:min(len(candidatas), tentativasComBomba)] {
		passos, casas := caminhoAte(bloq, c, eu.Posicao, casa)
		k := len(passos) + 1
		passos = append(passos, jogo.Acao{Tipo: jogo.Plantar})
		l := preverLinha(estado, etapas, &eu, numerar(append([]jogo.Acao{}, passos...)), a)
		if !sobrevive(l, casas, k) {
			continue
		}
		resto, _, ok := novaBusca(rng, estado, l, eu.AcoesPorTurno, etapas, buscaSegura).ir(casas[k-1], k+1)
		if ok {
			return append(passos, resto...), true
		}
	}
	return nil, false
}

// caminhoAte devolve os movimentos do caminho mais curto por casas livres de
// `de` até `ate`, com desempate pela ordem fixa das direções, e as casas
// visitadas a partir de `de` (casas[i]: casa ao fim da etapa i).
func caminhoAte(bloq map[jogo.Posicao]bool, c jogo.Config, de, ate jogo.Posicao) ([]jogo.Acao, []jogo.Posicao) {
	dist := distancias(bloq, c, []jogo.Posicao{ate})
	passos := []jogo.Acao{}
	casas := []jogo.Posicao{de}
	for casa := de; casa != ate; {
		for _, d := range direcoes {
			v := casa.Vizinha(d)
			if dv, ok := dist[v]; ok && dv == dist[casa]-1 {
				passos = append(passos, jogo.Acao{Tipo: jogo.Mover, Direcao: d})
				casas = append(casas, v)
				casa = v
				break
			}
		}
	}
	return passos, casas
}
