package jogo

import (
	"cmp"
	"slices"
)

// explosoesDaEtapa é o resultado da fase de explosões de uma etapa (ORD-04).
type explosoesDaEtapa struct {
	explosoes  []Explosao       // ordenadas pela origem
	explodidas map[Posicao]bool // casas cujas pilhas explodiram
	chamas     map[Posicao]bool
	destruidos map[Posicao]bool // blocos destrutíveis atingidos (DEC-01)
}

// explodir calcula as explosões da etapa a partir das pilhas com alguma bomba
// de pavio ≤ 0, com reação em cadeia (BOM-04, BOM-06 a BOM-09). As bombas e os
// blocos recebidos são os do início da fase e servem de obstáculo do começo
// ao fim, mesmo que explodam ou sejam destruídos nela (DEC-05, ORD-06), por
// isso a ordem das bombas não muda o resultado (D6).
func explodir(config Config, fixos, destrutiveis map[Posicao]bool, bombas []Bomba) explosoesDaEtapa {
	r := explosoesDaEtapa{
		explosoes:  []Explosao{},
		explodidas: map[Posicao]bool{},
		chamas:     map[Posicao]bool{},
		destruidos: map[Posicao]bool{},
	}
	pilhas := map[Posicao][]Bomba{}
	var fila []Posicao
	for _, b := range bombas {
		pilhas[b.Posicao] = append(pilhas[b.Posicao], b)
		if b.PavioRestante <= 0 {
			fila = append(fila, b.Posicao)
		}
	}
	for len(fila) > 0 {
		origem := fila[0]
		fila = fila[1:]
		if r.explodidas[origem] {
			continue
		}
		r.explodidas[origem] = true
		e := Explosao{Origem: origem, Potencia: potenciaDaPilha(pilhas[origem]), Chamas: []Posicao{origem}}
		for _, d := range []Direcao{Cima, Baixo, Esquerda, Direita} {
			p := origem
			for range e.Potencia {
				p = p.Vizinha(d)
				if !config.NoTabuleiro(p) || fixos[p] {
					break
				}
				e.Chamas = append(e.Chamas, p)
				if destrutiveis[p] {
					r.destruidos[p] = true
					break
				}
				if len(pilhas[p]) > 0 {
					fila = append(fila, p) // BOM-09
					break
				}
			}
		}
		ordenarPosicoes(e.Chamas)
		for _, p := range e.Chamas {
			r.chamas[p] = true
		}
		r.explosoes = append(r.explosoes, e)
	}
	slices.SortFunc(r.explosoes, func(a, b Explosao) int { return compararPosicoes(a.Origem, b.Origem) })
	return r
}

// potenciaDaPilha é a maior potência da pilha mais 1 por bomba adicional (BOM-04).
func potenciaDaPilha(pilha []Bomba) int {
	maior := 0
	for _, b := range pilha {
		maior = max(maior, b.Potencia)
	}
	return maior + len(pilha) - 1
}

// compararPosicoes ordena por y e depois por x (D7).
func compararPosicoes(a, b Posicao) int {
	return cmp.Or(cmp.Compare(a.Y, b.Y), cmp.Compare(a.X, b.X))
}

func ordenarPosicoes(p []Posicao) { slices.SortFunc(p, compararPosicoes) }

// posicoesOrdenadas devolve as posições do conjunto ordenadas (D7).
func posicoesOrdenadas(c map[Posicao]bool) []Posicao {
	p := make([]Posicao, 0, len(c))
	for pos := range c {
		p = append(p, pos)
	}
	ordenarPosicoes(p)
	return p
}
