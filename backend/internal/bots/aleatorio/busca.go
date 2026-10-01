package aleatorio

import (
	"math/rand/v2"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// modoBusca é o objetivo da busca ao fim do turno.
type modoBusca int

const (
	buscaSegura       modoBusca = iota // vivo e fora da zona de perigo final
	buscaSobrevivente                  // só vivo
)

// passosPossiveis são os passos de uma etapa: ficar parado ou mover.
var passosPossiveis = []jogo.Acao{
	{Tipo: jogo.Esperar},
	{Tipo: jogo.Mover, Direcao: jogo.Cima},
	{Tipo: jogo.Mover, Direcao: jogo.Baixo},
	{Tipo: jogo.Mover, Direcao: jogo.Esquerda},
	{Tipo: jogo.Mover, Direcao: jogo.Direita},
}

// no é um par (casa, etapa) da busca.
type no struct {
	casa  jogo.Posicao
	etapa int
}

// caminho é o melhor caminho conhecido a partir de um nó que não atinge o
// objetivo: os passos e quantas etapas ele sobrevive.
type caminho struct {
	passos []jogo.Acao
	vivas  int
}

// busca é uma busca em profundidade com ordem sorteada sobre (casa, etapa).
type busca struct {
	rng        *rand.Rand
	config     jogo.Config
	bloqueadas map[jogo.Posicao]bool // blocos fixos e destrutíveis do início do turno (D5)
	l          linha
	acoes      int // depois desta etapa o jogador só fica parado (D6)
	etapas     int
	modo       modoBusca
	falhas     map[no]caminho
}

// novaBusca prepara a busca para o estado e a linha do tempo dados.
func novaBusca(rng *rand.Rand, estado jogo.Estado, l linha, acoes, etapas int, modo modoBusca) *busca {
	b := &busca{
		rng:        rng,
		config:     estado.Config,
		bloqueadas: map[jogo.Posicao]bool{},
		l:          l,
		acoes:      acoes,
		etapas:     etapas,
		modo:       modo,
		falhas:     map[no]caminho{},
	}
	for _, p := range estado.BlocosFixos {
		b.bloqueadas[p] = true
	}
	for _, p := range estado.BlocosDestrutiveis {
		b.bloqueadas[p] = true
	}
	return b
}

// destino devolve a casa depois do passo, ou false se o passo sai do
// tabuleiro ou entra em bloco (VAL-02, D5).
func (b *busca) destino(casa jogo.Posicao, p jogo.Acao) (jogo.Posicao, bool) {
	if p.Tipo != jogo.Mover {
		return casa, true
	}
	d := casa.Vizinha(p.Direcao)
	return d, b.config.NoTabuleiro(d) && !b.bloqueadas[d]
}

// ir procura passos da etapa dada até a última, partindo da casa. Com ok, os
// passos atingem o objetivo do modo; sem ok, são os que sobrevivem mais
// etapas (vivas). Os passos podem ser menos que as etapas restantes quando
// o jogador morre; o resto é completado com ESPERAR por quem chama.
func (b *busca) ir(casa jogo.Posicao, etapa int) (passos []jogo.Acao, vivas int, ok bool) {
	if etapa > b.etapas {
		return nil, 0, b.modo == buscaSobrevivente || !b.l.perigo[casa]
	}
	n := no{casa, etapa}
	if c, visto := b.falhas[n]; visto {
		return c.passos, c.vivas, false
	}
	opcoes := passosPossiveis[:1]
	if etapa <= b.acoes {
		opcoes = append([]jogo.Acao{}, passosPossiveis...)
		b.rng.Shuffle(len(opcoes), func(i, j int) { opcoes[i], opcoes[j] = opcoes[j], opcoes[i] })
	}
	melhor := caminho{vivas: -1}
	for _, p := range opcoes {
		destino, valido := b.destino(casa, p)
		if !valido {
			continue
		}
		if b.l.chamas[etapa][destino] {
			if melhor.vivas < 0 {
				melhor = caminho{passos: []jogo.Acao{p}, vivas: 0}
			}
			continue
		}
		resto, v, sucesso := b.ir(destino, etapa+1)
		if sucesso {
			return append([]jogo.Acao{p}, resto...), v + 1, true
		}
		if v+1 > melhor.vivas {
			melhor = caminho{passos: append([]jogo.Acao{p}, resto...), vivas: v + 1}
		}
	}
	b.falhas[n] = melhor
	return melhor.passos, melhor.vivas, false
}

// fugir devolve os passos (um por etapa, da etapa `de` até a última) do
// melhor caminho a partir da casa: seguro, senão sobrevivente, senão o que
// sobrevive mais etapas (CA-18 a CA-20).
func fugir(rng *rand.Rand, estado jogo.Estado, l linha, casa jogo.Posicao, de, acoes, etapas int) []jogo.Acao {
	passos, _, ok := novaBusca(rng, estado, l, acoes, etapas, buscaSegura).ir(casa, de)
	if !ok {
		passos, _, _ = novaBusca(rng, estado, l, acoes, etapas, buscaSobrevivente).ir(casa, de)
	}
	return completar(passos, etapas-de+1)
}

// completar devolve uma cópia dos passos com ESPERAR até ter n passos.
func completar(passos []jogo.Acao, n int) []jogo.Acao {
	r := append(make([]jogo.Acao, 0, n), passos...)
	for len(r) < n {
		r = append(r, jogo.Acao{Tipo: jogo.Esperar})
	}
	return r
}

// passeio sorteia n passos válidos a partir da casa e devolve também a casa
// depois de cada passo.
func (b *busca) passeio(casa jogo.Posicao, n int) ([]jogo.Acao, []jogo.Posicao) {
	passos := make([]jogo.Acao, 0, n)
	casas := make([]jogo.Posicao, 0, n)
	for range n {
		for {
			p := passosPossiveis[b.rng.IntN(len(passosPossiveis))]
			if d, ok := b.destino(casa, p); ok {
				passos, casas, casa = append(passos, p), append(casas, d), d
				break
			}
		}
	}
	return passos, casas
}
