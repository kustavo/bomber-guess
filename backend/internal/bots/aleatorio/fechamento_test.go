package aleatorio

import (
	"reflect"
	"slices"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// comFechamento liga o fechamento a partir do turno dado (FEC-01), com área
// mínima 1×1 para que os tabuleiros pequenos dos testes fechem (FEC-08).
func comFechamento(turno int) opcao {
	return func(e *jogo.Estado) {
		e.Config.TurnoFechamento = turno
		e.Config.AreaMinima = jogo.Area{Largura: 1, Altura: 1}
	}
}

func p(x, y int) jogo.Posicao { return jogo.Posicao{X: x, Y: y} }

func TestJanelaEAnelAlvo(t *testing.T) {
	c := jogo.Config{Largura: 15, Altura: 13, TurnoFechamento: 30}
	casos := []struct {
		nome   string
		config jogo.Config
		turno  int
		janela bool
		alvo   int
	}{
		{"FEC-01 sem fechamento não há janela", jogo.Config{Largura: 15, Altura: 13}, 40, false, 0},
		{"FEC-03 antes da janela", c, 26, false, 0},
		{"FEC-03 início da janela: margem de 2 anéis", c, 27, true, 2},
		{"FEC-03 véspera do fechamento", c, 29, true, 2},
		{"FEC-03 turno do fechamento da borda", c, 30, true, 3},
		{"FEC-03 turno seguinte", c, 31, true, 4},
		{"FEC-08 limitado ao primeiro anel que nunca fecha (área 5×5)", c, 33, true, 4},
		{"FEC-08 depois do último anel que fecha, a janela acaba", c, 34, false, 0},
		{"FEC-02 limitado ao maior anel (área 1×1)", jogo.Config{Largura: 15, Altura: 13, TurnoFechamento: 30, AreaMinima: jogo.Area{Largura: 1, Altura: 1}}, 35, true, 6},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if j := naJanela(caso.config, caso.turno); j != caso.janela {
				t.Errorf("naJanela = %v, esperado %v", j, caso.janela)
			}
			if caso.janela {
				if a := anelAlvo(caso.config, caso.turno); a != caso.alvo {
					t.Errorf("anelAlvo = %d, esperado %d", a, caso.alvo)
				}
			}
		})
	}
}

func TestRegiaoEDistancias(t *testing.T) {
	e := montar(t, `
		1.+..
		.#+..
		..+.2
	`)
	reg := regiao(bloqueios(e), e.Config, p(0, 0))
	if len(reg) != 5 || reg[p(3, 0)] || !reg[p(1, 2)] {
		t.Errorf("MOV-03 região = %v, esperado as 5 casas livres à esquerda dos blocos", reg)
	}
	dist := distancias(bloqueios(e), e.Config, []jogo.Posicao{p(1, 2)})
	if dist[p(0, 0)] != 3 || dist[p(1, 0)] != 4 {
		t.Errorf("distâncias = %v", dist)
	}
	if _, ok := dist[p(4, 2)]; ok {
		t.Errorf("MOV-03 casa atrás dos blocos não deveria ter distância")
	}
}

func TestBlocoDeAbertura(t *testing.T) {
	casos := []struct {
		nome    string
		desenho string
		opcoes  []opcao
		de      jogo.Posicao
		alvo    int
		bloco   jogo.Posicao
		preso   bool
	}{
		{
			nome: "FEC-02 alvo alcançável por casas livres: não está preso",
			desenho: `
				1....
				.....
				.....
				.....
				....2
			`,
			de: p(0, 0), alvo: 1,
		},
		{
			nome: "MOV-03 preso na borda atrás de um bloco",
			desenho: `
				1.#..
				+#...
				.....
				.....
				....2
			`,
			de: p(0, 0), alvo: 1, bloco: p(0, 1), preso: true,
		},
		{
			nome: "MOV-03 dois caminhos: escolhe o de menos blocos",
			desenho: `
				1+...
				+#...
				+#...
				.#...
				.#..2
			`,
			de: p(0, 0), alvo: 1, bloco: p(1, 0), preso: true,
		},
		{
			nome: "FEC-05 bloco que fecha neste turno não abre caminho",
			desenho: `
				1+...
				.####
				.#...
				.#...
				.#..2
			`,
			opcoes: []opcao{comFechamento(1)},
			de:     p(0, 0), alvo: 1,
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e := montar(t, c.desenho, c.opcoes...)
			bloco, preso := blocoDeAbertura(e, c.de, c.alvo)
			if preso != c.preso || preso && bloco != c.bloco {
				t.Errorf("bloco %v, preso %v; esperado %v, %v", bloco, preso, c.bloco, c.preso)
			}
		})
	}
}

func TestCasasDePlantio(t *testing.T) {
	e := montar(t, `
		1...#
		....+
		.....
		..#.2
	`, comBomba(4, 3, "jogador_2", 1, 9))
	reg := regiao(bloqueios(e), e.Config, p(0, 0))
	casos := []struct {
		nome     string
		bloco    jogo.Posicao
		potencia int
		casas    []jogo.Posicao
	}{
		{"BOM-06 BOM-07 potência 2, cortada por bloco fixo e bomba", p(4, 1), 2, []jogo.Posicao{p(2, 1), p(3, 1), p(4, 2)}},
		{"BOM-06 BOM-07 potência 3, até a borda e até o bloco fixo", p(2, 0), 3, []jogo.Posicao{p(0, 0), p(1, 0), p(3, 0), p(2, 1), p(2, 2)}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if casas := casasDePlantio(e, reg, c.bloco, c.potencia); !reflect.DeepEqual(casas, ordenadas(c.casas)) {
				t.Errorf("casas = %v, esperado %v", casas, ordenadas(c.casas))
			}
		})
	}
}

func ordenadas(ps []jogo.Posicao) []jogo.Posicao {
	r := slices.Clone(ps)
	slices.SortFunc(r, compararPosicoes)
	return r
}
