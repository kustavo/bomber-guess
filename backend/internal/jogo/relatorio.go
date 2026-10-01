package jogo

// ResultadoAcao diz o que aconteceu com a ação planejada de um jogador em uma
// etapa (DEC-09).
type ResultadoAcao string

const (
	Executada  ResultadoAcao = "EXECUTADA"  // a ação aconteceu como planejada
	Bloqueada  ResultadoAcao = "BLOQUEADA"  // MOV-05: ficou parado; o resto do plano foi abortado
	Abortada   ResultadoAcao = "ABORTADA"   // MOV-05: etapa depois de um bloqueio; executada como ESPERAR
	Descartada ResultadoAcao = "DESCARTADA" // FIM-01: jogador morto; nada executado
	Ignorada   ResultadoAcao = "IGNORADA"   // RES-02: ação impossível; executada como ESPERAR
)

// JogadorEtapa é a situação de um jogador ao fim de uma etapa.
type JogadorEtapa struct {
	ID        string        `json:"id"`
	Posicao   Posicao       `json:"posicao"`
	Status    Status        `json:"status"`
	Acao      Acao          `json:"acao"` // a ação do plano para esta etapa (ESPERAR se não havia)
	Resultado ResultadoAcao `json:"resultado"`
}

// Explosao é a explosão de uma pilha (BOM-04, BOM-06).
type Explosao struct {
	Origem   Posicao   `json:"origem"`
	Potencia int       `json:"potencia"` // já considerando a pilha
	Chamas   []Posicao `json:"chamas"`   // inclui a origem
}

// RelatorioEtapa é o que aconteceu em uma etapa, usado no replay e no
// registro (docs/ARQUITETURA.md, seção 1.2). As listas de posições vêm
// ordenadas por y e depois por x (D7).
type RelatorioEtapa struct {
	Turno                int            `json:"turno"`
	Etapa                int            `json:"etapa"`
	Jogadores            []JogadorEtapa `json:"jogadores"` // todos, na ordem do Estado
	Bombas               []Bomba        `json:"bombas"`    // no tabuleiro ao fim da etapa
	Explosoes            []Explosao     `json:"explosoes"`
	Chamas               []Posicao      `json:"chamas"`                // união das chamas, sem repetição
	Mortes               []string       `json:"mortes"`                // ids dos que morreram nesta etapa
	BlocosDestruidos     []Posicao      `json:"blocos_destruidos"`     // sem repetição (DEC-01)
	MovimentosBloqueados []string       `json:"movimentos_bloqueados"` // ids
}
