package jogo

import "fmt"

// Vizinha devolve a casa vizinha de p na direção d (TAB-03, TAB-04).
// Para uma direção desconhecida, devolve p.
func (p Posicao) Vizinha(d Direcao) Posicao {
	switch d {
	case Cima:
		p.Y--
	case Baixo:
		p.Y++
	case Esquerda:
		p.X--
	case Direita:
		p.X++
	}
	return p
}

// NoTabuleiro informa se p está dentro do tabuleiro (TAB-01).
func (c Config) NoTabuleiro(p Posicao) bool {
	return p.X >= 0 && p.X < c.Largura && p.Y >= 0 && p.Y < c.Altura
}

// String devolve a posição no formato (x,y), usado em mensagens de erro.
func (p Posicao) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}
