package jogo

// Acao é o que um jogador faz em uma etapa (ACA-01). Direcao só é usada
// (e só aparece no JSON) quando Tipo é MOVER.
type Acao struct {
	Etapa   int      `json:"etapa"`
	Tipo    TipoAcao `json:"tipo"`
	Direcao Direcao  `json:"direcao,omitempty"`
}

// Plano é a lista de ações que um jogador envia para um turno (REGRAS.md, seção 3).
type Plano struct {
	JogadorID string `json:"jogador_id"`
	Turno     int    `json:"turno"`
	Acoes     []Acao `json:"acoes"`
}
