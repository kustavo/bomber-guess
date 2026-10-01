package aleatorio

import (
	"math/rand/v2"
	"slices"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// planejarComBombas sorteia n em 1..min(bombas_por_turno, acoes_por_turno)
// e tenta planos seguros com n, n-1, ..., 2 bombas; com 1 usa tentarPlantar,
// como o v2 (decisão 2 e D2, D3 do marco 15).
func planejarComBombas(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas int) ([]jogo.Acao, bool) {
	n := 1 + rng.IntN(min(eu.BombasPorTurno, eu.AcoesPorTurno))
	for m := n; m >= 2; m-- {
		if passos, ok := tentarPlantarVarias(rng, estado, eu, etapas, m); ok {
			return passos, true
		}
	}
	return tentarPlantar(rng, estado, eu, etapas)
}

// tentarPlantarVarias sorteia até tentativasComBomba planos com exatamente n
// bombas e devolve o primeiro seguro (D4 do marco 15). Cada plano: etapas de
// plantio sorteadas, passeios aleatórios entre elas, previsão com todas as
// bombas (pilhas e reações incluídas) e continuação segura depois da última.
func tentarPlantarVarias(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas, n int) ([]jogo.Acao, bool) {
	sorteio := novaBusca(rng, estado, linha{}, eu.AcoesPorTurno, etapas, buscaSegura)
	for range tentativasComBomba {
		passos := []jogo.Acao{}
		casas := []jogo.Posicao{eu.Posicao} // casas[i]: casa ao fim da etapa i
		casa := eu.Posicao
		for _, k := range sortearEtapas(rng, eu.AcoesPorTurno, n) {
			trecho, cs := sorteio.passeio(casa, k-1-len(passos))
			passos, casas = append(passos, trecho...), append(casas, cs...)
			if len(cs) > 0 {
				casa = cs[len(cs)-1]
			}
			passos, casas = append(passos, jogo.Acao{Tipo: jogo.Plantar}), append(casas, casa) // parado na etapa da planta
		}
		ultima := len(passos)
		l := preverLinha(estado, etapas, &eu, numerar(append([]jogo.Acao{}, passos...)))
		if !viveAte(l, casas, ultima) {
			continue
		}
		resto, _, ok := novaBusca(rng, estado, l, eu.AcoesPorTurno, etapas, buscaSegura).ir(casa, ultima+1)
		if ok {
			return append(passos, resto...), true
		}
	}
	return nil, false
}

// viveAte informa se o jogador, em casas[i] ao fim da etapa i, escapa das
// chamas das etapas 1 a ultima.
func viveAte(l linha, casas []jogo.Posicao, ultima int) bool {
	for etapa := 1; etapa <= ultima; etapa++ {
		if l.chamas[etapa][casas[etapa]] {
			return false
		}
	}
	return true
}

// sortearEtapas devolve n etapas distintas em 1..acoes, em ordem crescente.
func sortearEtapas(rng *rand.Rand, acoes, n int) []int {
	ks := rng.Perm(acoes)[:n]
	for i := range ks {
		ks[i]++
	}
	slices.Sort(ks)
	return ks
}
