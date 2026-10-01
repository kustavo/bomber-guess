package jogo

// Bot é um jogador controlado por código (docs/BOTS.md).
type Bot interface {
	// Versao é o identificador único da versão, ex.: "claude-v1".
	Versao() string

	// Planejar recebe uma cópia do estado (BOT-01) e o id do próprio jogador,
	// e devolve o plano do turno: no máximo acoes_por_turno ações.
	Planejar(estado Estado, jogadorID string) []Acao
}
