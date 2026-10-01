package jogo

// Posicao é uma casa do tabuleiro (TAB-02, TAB-03).
type Posicao struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Config são os parâmetros da partida, copiados do mapa (MAP-01).
type Config struct {
	Largura             int `json:"largura"`
	Altura              int `json:"altura"`
	LimiteTurnos        int `json:"limite_turnos"`
	PrazoPlanejamentoMs int `json:"prazo_planejamento_ms"` // EST-01
	DuracaoEtapaMs      int `json:"duracao_etapa_ms"`      // EST-02
}

// Atributos que valem para as bombas e o plano de um jogador.
// É o "jogador_padrao" do mapa e vai embutido em Jogador.
type Atributos struct {
	BombasPorTurno int `json:"bombas_por_turno"` // EST-04
	Potencia       int `json:"potencia"`         // EST-05
	PavioPadrao    int `json:"pavio_padrao"`     // EST-05
	AcoesPorTurno  int `json:"acoes_por_turno"`
}

// Morte registra o turno e a etapa em que o jogador morreu (EST-07).
type Morte struct {
	Turno int `json:"turno"`
	Etapa int `json:"etapa"`
}

// Jogador é um participante da partida, controlado por um bot.
type Jogador struct {
	ID        string  `json:"id"`
	Posicao   Posicao `json:"posicao"` // EST-08: se morto, a casa onde morreu
	Status    Status  `json:"status"`
	Morte     *Morte  `json:"morte,omitempty"` // EST-07: nil enquanto vivo
	Atributos         // campos achatados no JSON
	BotVersao string  `json:"bot_versao"`
}

// Bomba é uma bomba plantada no tabuleiro.
type Bomba struct {
	Posicao       Posicao `json:"posicao"`
	JogadorID     string  `json:"jogador_id"`
	Potencia      int     `json:"potencia"`
	PavioRestante int     `json:"pavio_restante"`
}

// Estado é a foto completa da partida no início de um turno (REGRAS.md, seção 2).
type Estado struct {
	Turno              int       `json:"turno"`
	Config             Config    `json:"config"`
	EtapasNesteTurno   int       `json:"etapas_neste_turno"` // EST-03
	BlocosFixos        []Posicao `json:"blocos_fixos"`
	BlocosDestrutiveis []Posicao `json:"blocos_destrutiveis"`
	Bombas             []Bomba   `json:"bombas"`
	Jogadores          []Jogador `json:"jogadores"`
}

// Copiar devolve uma cópia profunda do estado (BOT-01): alterar a cópia não
// altera o original. Os slices da cópia nunca são nulos, para que o JSON
// tenha [] em vez de null.
func (e Estado) Copiar() Estado {
	copia := e
	copia.BlocosFixos = append([]Posicao{}, e.BlocosFixos...)
	copia.BlocosDestrutiveis = append([]Posicao{}, e.BlocosDestrutiveis...)
	copia.Bombas = append([]Bomba{}, e.Bombas...)
	copia.Jogadores = append([]Jogador{}, e.Jogadores...)
	for i, j := range copia.Jogadores {
		if j.Morte != nil {
			morte := *j.Morte
			copia.Jogadores[i].Morte = &morte
		}
	}
	return copia
}
