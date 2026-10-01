package main

import (
	"strings"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

func pos(x, y int) jogo.Posicao { return jogo.Posicao{X: x, Y: y} }

// estadoDesenho: 5×3 com um bloco fixo em (1,1) e destrutíveis em (2,0) e (3,1).
func estadoDesenho() jogo.Estado {
	return jogo.Estado{
		Turno:              3,
		Config:             jogo.Config{Largura: 5, Altura: 3},
		BlocosFixos:        []jogo.Posicao{pos(1, 1)},
		BlocosDestrutiveis: []jogo.Posicao{pos(2, 0), pos(3, 1)},
		Bombas:             []jogo.Bomba{{Posicao: pos(4, 2), Potencia: 1, PavioRestante: 2}},
		Jogadores: []jogo.Jogador{
			{ID: "jogador_1", Posicao: pos(0, 0), Status: jogo.Vivo},
			{ID: "jogador_2", Posicao: pos(4, 0), Status: jogo.Vivo},
			{ID: "jogador_3", Posicao: pos(0, 2), Status: jogo.Vivo},
		},
	}
}

func jogadorEtapa(id string, p jogo.Posicao, s jogo.Status) jogo.JogadorEtapa {
	return jogo.JogadorEtapa{ID: id, Posicao: p, Status: s}
}

func TestDesenharInicio(t *testing.T) {
	var b strings.Builder
	novoQuadro(estadoDesenho()).inicio(&b, estadoDesenho())
	esperado := "Turno 3, início\n" +
		"1.+.2\n" +
		".#.+.\n" +
		"3...o\n"
	if b.String() != esperado {
		t.Errorf("obtido:\n%s\nesperado:\n%s", b.String(), esperado)
	}
}

func TestDesenharEtapa(t *testing.T) {
	casos := []struct {
		nome      string
		relatorio jogo.RelatorioEtapa
		esperado  string
	}{
		{
			nome: "EST-08 CA-14 legenda, prioridade, jogador morto fora e eventos",
			relatorio: jogo.RelatorioEtapa{
				Turno: 3, Etapa: 2,
				Jogadores: []jogo.JogadorEtapa{
					jogadorEtapa("jogador_1", pos(0, 1), jogo.Vivo),  // casa com chama: o jogador tem prioridade
					jogadorEtapa("jogador_2", pos(4, 2), jogo.Vivo),  // sobre a bomba
					jogadorEtapa("jogador_3", pos(1, 2), jogo.Morto), // morto não aparece
				},
				Bombas:               []jogo.Bomba{{Posicao: pos(4, 2)}, {Posicao: pos(3, 0)}},
				Chamas:               []jogo.Posicao{pos(0, 0), pos(1, 0), pos(2, 0), pos(0, 1), pos(1, 2)},
				Mortes:               []string{"jogador_3"},
				BlocosDestruidos:     []jogo.Posicao{pos(2, 0)},
				MovimentosBloqueados: []string{"jogador_2"},
			},
			esperado: "Turno 3, etapa 2\n" +
				"***o.\n" +
				"1#.+.\n" +
				".*..2\n" +
				"mortes: jogador_3\n" +
				"blocos destruídos: (2,0)\n" +
				"bloqueados: jogador_2\n",
		},
		{
			nome: "MOV-04 CA-14 dois jogadores na mesma casa e etapa sem eventos",
			relatorio: jogo.RelatorioEtapa{
				Turno: 3, Etapa: 1,
				Jogadores: []jogo.JogadorEtapa{
					jogadorEtapa("jogador_1", pos(4, 1), jogo.Vivo),
					jogadorEtapa("jogador_2", pos(4, 1), jogo.Vivo),
					jogadorEtapa("jogador_3", pos(0, 2), jogo.Vivo),
				},
			},
			esperado: "Turno 3, etapa 1\n" +
				"..+..\n" +
				".#.+&\n" +
				"3....\n",
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			var b strings.Builder
			novoQuadro(estadoDesenho()).etapa(&b, c.relatorio)
			if b.String() != c.esperado {
				t.Errorf("obtido:\n%s\nesperado:\n%s", b.String(), c.esperado)
			}
		})
	}
}

func TestBlocoDestruidoViraCasaLivre(t *testing.T) {
	q := novoQuadro(estadoDesenho())
	vivos := []jogo.JogadorEtapa{jogadorEtapa("jogador_1", pos(0, 0), jogo.Vivo)}
	var etapa2, etapa3 strings.Builder
	q.etapa(&etapa2, jogo.RelatorioEtapa{Turno: 3, Etapa: 2, Jogadores: vivos, Chamas: []jogo.Posicao{pos(3, 1)}, BlocosDestruidos: []jogo.Posicao{pos(3, 1)}})
	q.etapa(&etapa3, jogo.RelatorioEtapa{Turno: 3, Etapa: 3, Jogadores: vivos})
	if linha := strings.Split(etapa2.String(), "\n")[2]; linha != ".#.*." {
		t.Errorf("ORD-06 CA-15 etapa 2: %q, esperado chama", linha)
	}
	if linha := strings.Split(etapa3.String(), "\n")[2]; linha != ".#..." {
		t.Errorf("ORD-06 CA-15 etapa 3: %q, esperado casa livre", linha)
	}
}
