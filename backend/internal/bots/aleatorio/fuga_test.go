package aleatorio

import (
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// planejarValido chama Planejar para jogador_1 e falha se o plano não cumprir
// o contrato (CA-03) ou se Validar apontar infração.
func planejarValido(t *testing.T, e jogo.Estado, semente uint64) []jogo.Acao {
	t.Helper()
	plano := Novo(semente).Planejar(e.Copiar(), "jogador_1")
	conferirContrato(t, e, "jogador_1", plano)
	if _, infracoes := jogo.Validar(e, jogo.Plano{JogadorID: "jogador_1", Turno: e.Turno, Acoes: plano}); len(infracoes) > 0 {
		t.Fatalf("semente %d: infrações %+v no plano %+v", semente, infracoes, plano)
	}
	return plano
}

func TestFuga(t *testing.T) {
	aberto := `
		.....
		.....
		.....
		.....
		....2
	`
	casos := []struct {
		nome     string
		desenho  string
		opcoes   []opcao
		esperado categoria // o mínimo aceito para o plano
	}{
		{
			nome: "BOM-05 CA-13a sobre bomba de pavio 2 sai pela única casa livre",
			desenho: `
				###
				#1#
				#.#
				#.#
				#2#
			`,
			opcoes:   []opcao{comBomba(1, 1, "jogador_2", 1, 2)},
			esperado: seguro,
		},
		{
			nome:     "ORD-07 CA-13b a 1 casa de bomba de pavio 1 foge na etapa da explosão",
			desenho:  "#.1..2",
			opcoes:   []opcao{comBomba(1, 0, "jogador_2", 1, 1)},
			esperado: seguro,
		},
		{
			nome: "BOM-04 CA-14 pilha de alcance 2 com o jogador a 2 casas",
			desenho: `
				..1..
				.....
				.....
				.....
				....2
			`,
			opcoes:   []opcao{comBomba(2, 2, "jogador_2", 1, 2), comBomba(2, 2, "jogador_2", 1, 2)},
			esperado: seguro,
		},
		{
			nome: "BOM-09 CA-15 reação em cadeia prevista na etapa 1",
			desenho: `
				.....
				.....
				...1.
				.....
				....2
			`,
			opcoes:   []opcao{comBomba(0, 2, "jogador_2", 1, 1), comBomba(1, 2, "jogador_2", 2, 6)},
			esperado: seguro,
		},
		{
			nome: "BOM-07 CA-16 bloco destrutível protege o jogador",
			desenho: `
				#######
				#1+.###
				#######
				#####2#
			`,
			opcoes:   []opcao{comBomba(3, 1, "jogador_2", 3, 3)},
			esperado: seguro,
		},
		{
			nome:     "BOM-11 CA-17 termina fora do alcance da bomba que sobra",
			desenho:  aberto,
			opcoes:   []opcao{comBomba(2, 2, "jogador_2", 1, 9), func(e *jogo.Estado) { e.Jogadores = append(e.Jogadores, jogador1Em(2, 1)) }},
			esperado: seguro,
		},
		{
			nome: "CA-18 único caminho seguro, longo, é sempre encontrado",
			desenho: `
				1.....
				#####.
				2.....
			`,
			opcoes:   []opcao{comBomba(0, 0, "jogador_2", 6, 7)},
			esperado: seguro,
		},
		{
			nome:     "CA-19 sem plano seguro, devolve plano sobrevivente",
			desenho:  "1.+..2",
			opcoes:   []opcao{comBomba(1, 0, "jogador_2", 2, 9)},
			esperado: sobrevivente,
		},
		{
			nome: "CA-20 sem saída, devolve plano válido",
			desenho: `
				###.
				#1#2
				###.
			`,
			opcoes:   []opcao{comBomba(1, 1, "jogador_2", 1, 1)},
			esperado: morre,
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e := montar(t, c.desenho, c.opcoes...)
			for semente := uint64(1); semente <= 50; semente++ {
				plano := planejarValido(t, e, semente)
				if obtido := classificar(e, "jogador_1", plano); obtido < c.esperado {
					t.Fatalf("semente %d: plano %s, esperado %s: %+v", semente, obtido, c.esperado, plano)
				}
			}
		})
	}
}

// jogador1Em acrescenta jogador_1 na casa dada (para desenhos sem o '1').
func jogador1Em(x, y int) jogo.Jogador {
	return jogo.Jogador{ID: "jogador_1", Posicao: jogo.Posicao{X: x, Y: y}, Status: jogo.Vivo, Atributos: atributosDeTeste, BotVersao: Versao}
}

func TestNaoEntraEmBlocoDestrutivel(t *testing.T) {
	e := montar(t, `
		#+#.
		+1+.
		#+#2
	`)
	for semente := uint64(1); semente <= 50; semente++ {
		casa := e.Jogadores[0].Posicao
		for _, a := range planejarValido(t, e, semente) {
			if a.Tipo != jogo.Mover {
				continue
			}
			casa = casa.Vizinha(a.Direcao)
			for _, b := range e.BlocosDestrutiveis {
				if b == casa {
					t.Fatalf("MOV-03 CA-08 semente %d: movimento para bloco destrutível em %s", semente, casa)
				}
			}
		}
	}
}
