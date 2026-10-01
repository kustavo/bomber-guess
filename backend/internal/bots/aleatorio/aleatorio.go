// Package aleatorio é o bot aleatorio-v1: anda aleatoriamente, planta bombas
// de vez em quando e foge das explosões que consegue prever
// (specs/03-bot-simples).
package aleatorio

import (
	"hash/fnv"
	"math/rand/v2"
	"strconv"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// Versao é o identificador desta versão do bot.
const Versao = "aleatorio-v1"

// Bot anda aleatoriamente, planta bombas de vez em quando e foge das
// explosões que consegue prever. Não guarda estado entre chamadas: o plano
// depende só da semente, do estado e do jogador.
type Bot struct {
	semente uint64
}

var _ jogo.Bot = (*Bot)(nil)

// Novo cria o bot com a semente dada (decisão 1).
func Novo(semente uint64) *Bot {
	return &Bot{semente: semente}
}

// Versao devolve "aleatorio-v1".
func (b *Bot) Versao() string { return Versao }

// Planejar devolve exatamente acoes_por_turno ações numeradas 1, 2, 3…,
// ou nenhuma se o jogador não existe ou está morto. Não altera o estado.
func (b *Bot) Planejar(estado jogo.Estado, jogadorID string) []jogo.Acao {
	eu, ok := encontrar(estado, jogadorID)
	if !ok || eu.Status != jogo.Vivo || eu.AcoesPorTurno < 1 {
		return []jogo.Acao{}
	}
	rng := b.gerador(estado.Turno, jogadorID)
	etapas := jogo.CalcularEtapas(estado.Jogadores)
	if eu.BombasPorTurno >= 1 && rng.IntN(2) == 0 { // decisão 5, D8
		if passos, ok := tentarPlantar(rng, estado, eu, etapas); ok {
			return numerar(passos[:eu.AcoesPorTurno])
		}
	}
	l := preverLinha(estado, etapas, nil, nil)
	passos := fugir(rng, estado, l, eu.Posicao, 1, eu.AcoesPorTurno, etapas)
	return numerar(passos[:eu.AcoesPorTurno])
}

// tentativasComBomba é quantos planos com bomba o bot sorteia por turno.
const tentativasComBomba = 8

// tentarPlantar sorteia a etapa k e um passeio de k-1 passos, planta na etapa
// k e procura uma continuação segura com a linha do tempo que já inclui a
// própria bomba (D4). Devolve os passos de todas as etapas do turno.
func tentarPlantar(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas int) ([]jogo.Acao, bool) {
	sorteio := novaBusca(rng, estado, linha{}, eu.AcoesPorTurno, etapas, buscaSegura)
	for range tentativasComBomba {
		k := 1 + rng.IntN(eu.AcoesPorTurno)
		prefixo, casas := sorteio.passeio(eu.Posicao, k-1)
		casas = append([]jogo.Posicao{eu.Posicao}, casas...) // casas[i]: casa ao fim da etapa i; na etapa k fica parado
		passos := append(prefixo, jogo.Acao{Tipo: jogo.Plantar})
		l := preverLinha(estado, etapas, &eu, numerar(append([]jogo.Acao{}, passos...)))
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

// sobrevive informa se o jogador escapa das chamas das etapas 1 a k, estando
// em casas[i] ao fim da etapa i (e em casas[k-1] na etapa k).
func sobrevive(l linha, casas []jogo.Posicao, k int) bool {
	for etapa := 1; etapa <= k; etapa++ {
		if l.chamas[etapa][casas[min(etapa, k-1)]] {
			return false
		}
	}
	return true
}

// gerador cria o gerador da chamada a partir da semente, do turno e do
// jogador (D7), para que o plano não dependa de chamadas anteriores.
func (b *Bot) gerador(turno int, jogadorID string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(jogadorID))
	h.Write([]byte{0})
	h.Write([]byte(strconv.Itoa(turno)))
	return rand.New(rand.NewPCG(b.semente, h.Sum64()))
}

// encontrar devolve o jogador com o id dado.
func encontrar(estado jogo.Estado, jogadorID string) (jogo.Jogador, bool) {
	for _, j := range estado.Jogadores {
		if j.ID == jogadorID {
			return j, true
		}
	}
	return jogo.Jogador{}, false
}

// numerar preenche a etapa de cada ação com a sua posição (1, 2, 3…; VAL-06).
func numerar(passos []jogo.Acao) []jogo.Acao {
	for i := range passos {
		passos[i].Etapa = i + 1
	}
	return passos
}
