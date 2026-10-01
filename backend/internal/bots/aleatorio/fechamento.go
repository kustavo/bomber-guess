package aleatorio

import (
	"cmp"
	"slices"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// janelaAntecipacao é quantos turnos antes de turno_fechamento o v2 começa a
// ir para o centro (decisão 1 do marco 13).
const janelaAntecipacao = 3

// direcoes é a ordem fixa das direções nas buscas sem sorteio (D6).
var direcoes = []jogo.Direcao{jogo.Cima, jogo.Baixo, jogo.Esquerda, jogo.Direita}

// naJanela informa se o turno está na janela de antecipação: de
// janelaAntecipacao turnos antes de turno_fechamento até o turno em que fecha
// o último anel que pode fechar (decisão 1, FEC-08).
func naJanela(c jogo.Config, turno int) bool {
	return c.TurnoFechamento >= 1 && turno >= c.TurnoFechamento-janelaAntecipacao &&
		turno < c.TurnoFechamento+primeiroAnelQueNuncaFecha(c)
}

// primeiroAnelQueNuncaFecha é o menor anel que FEC-08 não deixa fechar.
func primeiroAnelQueNuncaFecha(c jogo.Config) int {
	n := 0
	for c.AnelFecha(n) {
		n++
	}
	return n
}

// margemAlvo é quantos anéis além do que fecha no turno seguinte o v2 quer
// estar ao fim do turno (decisão 1 do marco 13).
const margemAlvo = 2

// anelAlvo é o anel que fecha no turno seguinte mais a margem, limitado ao
// primeiro anel que nunca fecha (FEC-08) e ao maior anel do tabuleiro
// (decisão 1, FEC-02).
func anelAlvo(c jogo.Config, turno int) int {
	maior := (min(c.Largura, c.Altura) - 1) / 2
	return min(max(0, turno-c.TurnoFechamento+1)+margemAlvo, primeiroAnelQueNuncaFecha(c), maior)
}

// bloqueios devolve as casas com bloco fixo ou destrutível (D5).
func bloqueios(estado jogo.Estado) map[jogo.Posicao]bool {
	b := map[jogo.Posicao]bool{}
	for _, p := range estado.BlocosFixos {
		b[p] = true
	}
	for _, p := range estado.BlocosDestrutiveis {
		b[p] = true
	}
	return b
}

// regiao devolve as casas alcançáveis a partir de `de` andando só por casas
// livres, sem limite de passos (D5).
func regiao(bloq map[jogo.Posicao]bool, c jogo.Config, de jogo.Posicao) map[jogo.Posicao]bool {
	dist := distancias(bloq, c, []jogo.Posicao{de})
	r := make(map[jogo.Posicao]bool, len(dist))
	for p := range dist {
		r[p] = true
	}
	return r
}

// distancias devolve, para cada casa livre que alcança algum destino, a
// quantidade de passos até o destino mais próximo (D4, D5).
func distancias(bloq map[jogo.Posicao]bool, c jogo.Config, destinos []jogo.Posicao) map[jogo.Posicao]int {
	dist := map[jogo.Posicao]int{}
	var fila []jogo.Posicao
	for _, p := range destinos {
		if _, visto := dist[p]; !visto && c.NoTabuleiro(p) && !bloq[p] {
			dist[p] = 0
			fila = append(fila, p)
		}
	}
	for len(fila) > 0 {
		p := fila[0]
		fila = fila[1:]
		for _, d := range direcoes {
			v := p.Vizinha(d)
			if _, visto := dist[v]; visto || !c.NoTabuleiro(v) || bloq[v] {
				continue
			}
			dist[v] = dist[p] + 1
			fila = append(fila, v)
		}
	}
	return dist
}

// custoAbertura é o custo de um caminho na busca do bloco de abertura:
// blocos destrutíveis atravessados e depois passos (D6).
type custoAbertura struct {
	blocos, passos int
	primeiro       jogo.Posicao // primeiro bloco destrutível do caminho
	temBloco       bool
}

func (a custoAbertura) menor(b custoAbertura) bool {
	return a.blocos < b.blocos || a.blocos == b.blocos && a.passos < b.passos
}

// blocoDeAbertura devolve o primeiro bloco destrutível do caminho mais barato
// de `de` até uma casa de anel ≥ alvo, contando blocos atravessados e depois
// passos. Blocos destrutíveis que fecham neste turno contam como fixos
// (FEC-05). Devolve false se o anel alvo é alcançável sem atravessar blocos
// ou se não é alcançável de jeito nenhum.
func blocoDeAbertura(estado jogo.Estado, de jogo.Posicao, alvo int) (jogo.Posicao, bool) {
	c := estado.Config
	fixos := map[jogo.Posicao]bool{}
	for _, p := range estado.BlocosFixos {
		fixos[p] = true
	}
	destrutiveis := map[jogo.Posicao]bool{}
	for _, p := range estado.BlocosDestrutiveis {
		destrutiveis[p] = true
	}
	for _, p := range jogo.CasasQueFecham(estado) {
		if destrutiveis[p] {
			fixos[p] = true
		}
	}
	melhor := map[jogo.Posicao]custoAbertura{de: {}}
	fechados := map[jogo.Posicao]bool{}
	for {
		// Dijkstra simples: o tabuleiro é pequeno. Empate pela posição (y, x).
		var atual jogo.Posicao
		achou := false
		for p, cu := range melhor {
			if fechados[p] {
				continue
			}
			if !achou || cu.menor(melhor[atual]) || !melhor[atual].menor(cu) && compararPosicoes(p, atual) < 0 {
				atual, achou = p, true
			}
		}
		if !achou {
			return jogo.Posicao{}, false
		}
		cu := melhor[atual]
		if c.Anel(atual) >= alvo {
			return cu.primeiro, cu.temBloco
		}
		fechados[atual] = true
		for _, d := range direcoes {
			v := atual.Vizinha(d)
			if !c.NoTabuleiro(v) || fixos[v] || fechados[v] {
				continue
			}
			novo := custoAbertura{blocos: cu.blocos, passos: cu.passos + 1, primeiro: cu.primeiro, temBloco: cu.temBloco}
			if destrutiveis[v] {
				novo.blocos++
				if !novo.temBloco {
					novo.primeiro, novo.temBloco = v, true
				}
			}
			if anterior, visto := melhor[v]; !visto || novo.menor(anterior) {
				melhor[v] = novo
			}
		}
	}
}

// casasDePlantio devolve as casas da região de onde uma bomba de potência
// dada alcança o bloco: em linha reta, a até `potencia` casas, sem bloco nem
// bomba no meio (BOM-06, BOM-07). Ordenadas por y e depois por x.
func casasDePlantio(estado jogo.Estado, reg map[jogo.Posicao]bool, bloco jogo.Posicao, potencia int) []jogo.Posicao {
	bloq := bloqueios(estado)
	comBomba := map[jogo.Posicao]bool{}
	for _, b := range estado.Bombas {
		comBomba[b.Posicao] = true
	}
	var casas []jogo.Posicao
	for _, d := range direcoes {
		p := bloco
		for range potencia {
			p = p.Vizinha(d)
			if !estado.Config.NoTabuleiro(p) || bloq[p] || comBomba[p] {
				break
			}
			if reg[p] {
				casas = append(casas, p)
			}
		}
	}
	slices.SortFunc(casas, compararPosicoes)
	return casas
}

// compararPosicoes ordena por y e depois por x.
func compararPosicoes(a, b jogo.Posicao) int {
	return cmp.Or(cmp.Compare(a.Y, b.Y), cmp.Compare(a.X, b.X))
}
