package jogo

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func lerMapaExemplo(t *testing.T) Mapa {
	t.Helper()
	dados, err := os.ReadFile(filepath.Join(raizDoProjeto, "mapas", "exemplo.json"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := LerMapa(dados)
	if err != nil {
		t.Fatalf("MAP-05 mapas/exemplo.json inválido: %v", err)
	}
	return m
}

func TestMapaExemploLayout(t *testing.T) {
	m := lerMapaExemplo(t)
	impares := map[Posicao]bool{}
	for x := 1; x < 15; x += 2 {
		for y := 1; y < 13; y += 2 {
			impares[Posicao{X: x, Y: y}] = true
		}
	}
	cantos := []Posicao{{X: 0, Y: 0}, {X: 14, Y: 0}, {X: 0, Y: 12}, {X: 14, Y: 12}}
	casos := []struct {
		nome     string
		obtido   any
		esperado any
	}{
		{"TAB-01 15×13", [2]int{m.Config.Largura, m.Config.Altura}, [2]int{15, 13}},
		{"MAP-05 4 posições iniciais nos cantos", conjunto(m.PosicoesIniciais), conjunto(cantos)},
		{"MAP-05 posições iniciais sem repetição", len(m.PosicoesIniciais), 4},
		{"TAB-01 blocos fixos em todas as casas com x e y ímpares", conjunto(m.BlocosFixos), impares},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if !reflect.DeepEqual(c.obtido, c.esperado) {
				t.Errorf("obtido %v, esperado %v", c.obtido, c.esperado)
			}
		})
	}

	blocos := conjunto(append(append([]Posicao{}, m.BlocosFixos...), m.BlocosDestrutiveis...))
	for _, canto := range cantos {
		t.Run("MAP-05 canto livre em L "+canto.String(), func(t *testing.T) {
			livres := []Posicao{canto}
			for _, d := range []Direcao{Cima, Baixo, Esquerda, Direita} {
				if v := canto.Vizinha(d); m.Config.NoTabuleiro(v) {
					livres = append(livres, v)
				}
			}
			if len(livres) != 3 {
				t.Fatalf("canto %s com %d casas no L", canto, len(livres))
			}
			for _, p := range livres {
				if blocos[p] {
					t.Errorf("casa %s do L tem bloco", p)
				}
			}
		})
	}
}

func TestEstadoInicialMapaExemplo(t *testing.T) {
	m := lerMapaExemplo(t)
	versoes := []string{"v1", "v2", "v3", "v4"}
	e, err := EstadoInicial(m, versoes)
	if err != nil {
		t.Fatal(err)
	}
	esperados := make([]Jogador, len(versoes))
	for i, v := range versoes {
		esperados[i] = Jogador{
			ID:        []string{"jogador_1", "jogador_2", "jogador_3", "jogador_4"}[i],
			Posicao:   m.PosicoesIniciais[i],
			Status:    Vivo,
			Atributos: m.JogadorPadrao,
			BotVersao: v,
		}
	}
	casos := []struct {
		nome     string
		obtido   any
		esperado any
	}{
		{"MAP-04 turno 1", e.Turno, 1},
		{"MAP-01 config e dimensões do mapa", e.Config, m.Config},
		{"MAP-02 MAP-03 MAP-04 jogadores na ordem das posições iniciais", e.Jogadores, esperados},
		{"MAP-04 sem bombas", e.Bombas, []Bomba{}},
		{"MAP-01 blocos fixos do mapa", e.BlocosFixos, m.BlocosFixos},
		{"MAP-01 blocos destrutíveis do mapa", e.BlocosDestrutiveis, m.BlocosDestrutiveis},
		{"EST-03 etapas = acoes_por_turno padrão", e.EtapasNesteTurno, m.JogadorPadrao.AcoesPorTurno},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if !reflect.DeepEqual(c.obtido, c.esperado) {
				t.Errorf("obtido %+v, esperado %+v", c.obtido, c.esperado)
			}
		})
	}
}
