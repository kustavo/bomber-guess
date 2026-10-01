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

// Versões do bot: o v2 é o v1 preparado para o fechamento do tabuleiro
// (specs/13-aleatorio-v2); o v3 é o v2 sem o limite de uma bomba por turno
// (specs/15-aleatorio-v3).
const (
	Versao   = "aleatorio-v1"
	VersaoV2 = "aleatorio-v2"
	VersaoV3 = "aleatorio-v3"
)

// Bot anda aleatoriamente, planta bombas de vez em quando e foge das
// explosões que consegue prever. Não guarda estado entre chamadas: o plano
// depende só da semente, do estado e do jogador.
type Bot struct {
	semente      uint64
	antecipa     bool // v2 e v3: prepara-se para o fechamento (D1 do marco 13)
	variasBombas bool // v3: até bombas_por_turno bombas por turno (D1 do marco 15)
}

var _ jogo.Bot = (*Bot)(nil)

// Novo cria o bot com a semente dada (decisão 1).
func Novo(semente uint64) *Bot {
	return &Bot{semente: semente}
}

// NovoV2 cria o aleatorio-v2 com a semente dada.
func NovoV2(semente uint64) *Bot {
	return &Bot{semente: semente, antecipa: true}
}

// NovoV3 cria o aleatorio-v3 com a semente dada.
func NovoV3(semente uint64) *Bot {
	return &Bot{semente: semente, antecipa: true, variasBombas: true}
}

// Versao devolve "aleatorio-v1", "aleatorio-v2" ou "aleatorio-v3".
func (b *Bot) Versao() string {
	if b.variasBombas {
		return VersaoV3
	}
	if b.antecipa {
		return VersaoV2
	}
	return Versao
}

// Planejar devolve exatamente acoes_por_turno ações numeradas 1, 2, 3…,
// ou nenhuma se o jogador não existe ou está morto. Não altera o estado.
func (b *Bot) Planejar(estado jogo.Estado, jogadorID string) []jogo.Acao {
	eu, ok := encontrar(estado, jogadorID)
	if !ok || eu.Status != jogo.Vivo || eu.AcoesPorTurno < 1 {
		return []jogo.Acao{}
	}
	rng := b.gerador(estado.Turno, jogadorID)
	etapas := jogo.CalcularEtapas(estado.Jogadores)
	if b.antecipa && naJanela(estado.Config, estado.Turno) { // D2 do marco 13
		return numerar(planejarNaJanela(rng, estado, eu, etapas)[:eu.AcoesPorTurno])
	}
	if eu.BombasPorTurno >= 1 && rng.IntN(2) == 0 { // decisão 5, D8
		plantar := tentarPlantar
		if b.variasBombas && eu.BombasPorTurno >= 2 { // v3: D1 do marco 15
			plantar = planejarComBombas
		}
		if passos, ok := plantar(rng, estado, eu, etapas); ok {
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
