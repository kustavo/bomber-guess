package jogo

// TipoAcao é o tipo de uma ação (ACA-01).
type TipoAcao string

const (
	Mover   TipoAcao = "MOVER"
	Plantar TipoAcao = "PLANTAR"
	Esperar TipoAcao = "ESPERAR"
)

// Direcao é a direção de um movimento para uma casa vizinha (TAB-04).
type Direcao string

const (
	Cima     Direcao = "CIMA"
	Baixo    Direcao = "BAIXO"
	Esquerda Direcao = "ESQUERDA"
	Direita  Direcao = "DIREITA"
)

// Status é a situação de um jogador na partida (EST-06).
type Status string

const (
	Vivo  Status = "VIVO"
	Morto Status = "MORTO"
)
