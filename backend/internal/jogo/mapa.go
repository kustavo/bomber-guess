package jogo

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Mapa é o layout do tabuleiro (docs/EDITOR.md, "Formato do mapa").
type Mapa struct {
	Nome               string    `json:"nome"`
	Config             Config    `json:"config"`         // MAP-01
	JogadorPadrao      Atributos `json:"jogador_padrao"` // MAP-02
	BlocosFixos        []Posicao `json:"blocos_fixos"`
	BlocosDestrutiveis []Posicao `json:"blocos_destrutiveis"`
	PosicoesIniciais   []Posicao `json:"posicoes_iniciais"`
}

// ErrMapaInvalido é embrulhado por todos os erros de LerMapa, VerificarMapa
// e EstadoInicial.
var ErrMapaInvalido = errors.New("mapa inválido")

// LerMapa decodifica um mapa em JSON e o verifica com VerificarMapa.
func LerMapa(dados []byte) (Mapa, error) {
	var m Mapa
	if err := json.Unmarshal(dados, &m); err != nil {
		return Mapa{}, fmt.Errorf("%w: JSON malformado: %v", ErrMapaInvalido, err)
	}
	if err := VerificarMapa(m); err != nil {
		return Mapa{}, err
	}
	return m, nil
}

// VerificarMapa devolve um erro que embrulha ErrMapaInvalido se o mapa for
// inválido (MAP-05), ou nil.
func VerificarMapa(m Mapa) error {
	if m.Config.Largura <= 0 {
		return fmt.Errorf("%w: largura %d deve ser maior que 0", ErrMapaInvalido, m.Config.Largura)
	}
	if m.Config.Altura <= 0 {
		return fmt.Errorf("%w: altura %d deve ser maior que 0", ErrMapaInvalido, m.Config.Altura)
	}
	fixos := make(map[Posicao]bool, len(m.BlocosFixos))
	for _, p := range m.BlocosFixos {
		if !m.Config.NoTabuleiro(p) {
			return fmt.Errorf("%w: bloco fixo em %s fora do tabuleiro", ErrMapaInvalido, p)
		}
		fixos[p] = true
	}
	destrutiveis := make(map[Posicao]bool, len(m.BlocosDestrutiveis))
	for _, p := range m.BlocosDestrutiveis {
		if !m.Config.NoTabuleiro(p) {
			return fmt.Errorf("%w: bloco destrutível em %s fora do tabuleiro", ErrMapaInvalido, p)
		}
		if fixos[p] {
			return fmt.Errorf("%w: casa %s tem bloco fixo e destrutível", ErrMapaInvalido, p)
		}
		destrutiveis[p] = true
	}
	if len(m.PosicoesIniciais) < 2 {
		return fmt.Errorf("%w: %d posições iniciais, mínimo 2", ErrMapaInvalido, len(m.PosicoesIniciais))
	}
	for _, p := range m.PosicoesIniciais {
		if !m.Config.NoTabuleiro(p) {
			return fmt.Errorf("%w: posição inicial %s fora do tabuleiro", ErrMapaInvalido, p)
		}
		if fixos[p] || destrutiveis[p] {
			return fmt.Errorf("%w: posição inicial %s sobre bloco", ErrMapaInvalido, p)
		}
	}
	return nil
}

// EstadoInicial cria o estado do turno 1 a partir do mapa: a versão i joga na
// posição inicial i, com id jogador_<i+1> (MAP-03, MAP-04). Falha se o mapa
// for inválido ou se a quantidade de versões for diferente da de posições
// iniciais (MAP-06). Não altera o mapa recebido.
func EstadoInicial(m Mapa, versoes []string) (Estado, error) {
	if err := VerificarMapa(m); err != nil {
		return Estado{}, err
	}
	if len(versoes) != len(m.PosicoesIniciais) {
		return Estado{}, fmt.Errorf("%w: %d versões de bot para %d posições iniciais",
			ErrMapaInvalido, len(versoes), len(m.PosicoesIniciais))
	}
	jogadores := make([]Jogador, len(versoes))
	for i, versao := range versoes {
		jogadores[i] = Jogador{
			ID:        fmt.Sprintf("jogador_%d", i+1),
			Posicao:   m.PosicoesIniciais[i],
			Status:    Vivo,
			Atributos: m.JogadorPadrao,
			BotVersao: versao,
		}
	}
	return Estado{
		Turno:              1,
		Config:             m.Config,
		EtapasNesteTurno:   CalcularEtapas(jogadores),
		BlocosFixos:        append([]Posicao{}, m.BlocosFixos...),
		BlocosDestrutiveis: append([]Posicao{}, m.BlocosDestrutiveis...),
		Bombas:             []Bomba{},
		Jogadores:          jogadores,
	}, nil
}

// CalcularEtapas devolve o maior acoes_por_turno entre os jogadores vivos,
// ou 0 se não houver nenhum (EST-03).
func CalcularEtapas(jogadores []Jogador) int {
	etapas := 0
	for _, j := range jogadores {
		if j.Status == Vivo {
			etapas = max(etapas, j.AcoesPorTurno)
		}
	}
	return etapas
}
