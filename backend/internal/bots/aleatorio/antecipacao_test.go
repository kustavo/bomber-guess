package aleatorio

import (
	"reflect"
	"slices"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// planejarV2 chama o v2 para jogador_1, confere o contrato e a validade
// (como planejarValido) e devolve o plano.
func planejarV2(t *testing.T, e jogo.Estado, semente uint64) []jogo.Acao {
	t.Helper()
	return planejarValido(t, NovoV2, e, semente)
}

// casaFinal simula o plano e devolve onde jogador_1 termina o turno.
func casaFinal(t *testing.T, e jogo.Estado, plano []jogo.Acao) jogo.Posicao {
	t.Helper()
	final, _ := simular(e, "jogador_1", plano)
	eu, _ := encontrar(final, "jogador_1")
	return eu.Posicao
}

func TestV2IgualAoV1ForaDaJanela(t *testing.T) {
	exemplo, err := jogo.EstadoInicial(lerMapaExemplo(t), []string{Versao, Versao, Versao, Versao})
	if err != nil {
		t.Fatal(err)
	}
	semFechamento := exemplo.Copiar()
	semFechamento.Config.TurnoFechamento = 0
	semFechamento.Turno = 40
	antesDaJanela := exemplo.Copiar()
	antesDaJanela.Turno = exemplo.Config.TurnoFechamento - janelaAntecipacao - 1
	casos := []struct {
		nome   string
		estado jogo.Estado
	}{
		{"FEC-01 EST-09 CA-04 sem fechamento, campo aberto", montar(t, campoAberto, comTurno(40))},
		{"FEC-01 EST-09 CA-04 sem fechamento, com bombas", montar(t, campoAberto, comTurno(40), comBomba(1, 2, "jogador_2", 2, 2), comBomba(5, 1, "jogador_2", 1, 9))},
		{"FEC-01 EST-09 CA-04 sem fechamento, mapa de exemplo", semFechamento},
		{"FEC-03 CA-06 antes da janela, campo aberto", montar(t, campoAberto, comFechamento(30), comTurno(26))},
		{"FEC-03 CA-06 antes da janela, mapa de exemplo", antesDaJanela},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			for semente := uint64(1); semente <= 50; semente++ {
				for _, j := range c.estado.Jogadores {
					v1 := Novo(semente).Planejar(c.estado.Copiar(), j.ID)
					v2 := NovoV2(semente).Planejar(c.estado.Copiar(), j.ID)
					if !reflect.DeepEqual(v1, v2) {
						t.Fatalf("semente %d, %s: v1 %+v, v2 %+v", semente, j.ID, v1, v2)
					}
				}
			}
		})
	}
}

func TestAntecipacao(t *testing.T) {
	t.Run("FEC-02 FEC-03 CA-07 da borda para o anel alvo", func(t *testing.T) {
		e := montar(t, `
			1......
			.......
			.......
			.......
			.......
			.......
			......2
		`, comFechamento(1))
		alvo := anelAlvo(e.Config, e.Turno)
		for semente := uint64(1); semente <= 50; semente++ {
			plano := planejarV2(t, e, semente)
			if c := classificar(e, "jogador_1", plano); c != seguro {
				t.Fatalf("semente %d: plano %s: %+v", semente, c, plano)
			}
			if casa := casaFinal(t, e, plano); e.Config.Anel(casa) < alvo {
				t.Fatalf("semente %d: termina em %v, anel %d < alvo %d", semente, casa, e.Config.Anel(casa), alvo)
			}
		}
	})

	t.Run("FEC-02 CA-08 anel alvo longe: termina mais perto", func(t *testing.T) {
		e := montar(t, `
			1..........
			##########.
			...........
			...........
			...........
			...........
			..........2
		`, comFechamento(3))
		bloq := bloqueios(e)
		var destinos []jogo.Posicao
		for casa := range regiao(bloq, e.Config, p(0, 0)) {
			if e.Config.Anel(casa) >= anelAlvo(e.Config, e.Turno) {
				destinos = append(destinos, casa)
			}
		}
		dist := distancias(bloq, e.Config, destinos)
		if dist[p(0, 0)] <= atributosDeTeste.AcoesPorTurno {
			t.Fatalf("tabuleiro errado: alvo a %d casas", dist[p(0, 0)])
		}
		for semente := uint64(1); semente <= 50; semente++ {
			plano := planejarV2(t, e, semente)
			if casa := casaFinal(t, e, plano); dist[casa] >= dist[p(0, 0)] {
				t.Fatalf("semente %d: termina em %v, a %d casas do alvo (começou a %d)", semente, casa, dist[casa], dist[p(0, 0)])
			}
		}
	})

	t.Run("BOM-11 CA-09 segurança antes do anel alvo", func(t *testing.T) {
		// O anel 1 livre é (1,1)–(3,1), todo no alcance da bomba que sobra.
		e := montar(t, `
			1....
			.....
			.###.
			.###.
			....2
		`, comFechamento(3), comBomba(2, 1, "jogador_2", 1, 9))
		for semente := uint64(1); semente <= 50; semente++ {
			if plano := planejarV2(t, e, semente); classificar(e, "jogador_1", plano) != seguro {
				t.Fatalf("semente %d: plano não seguro: %+v", semente, plano)
			}
		}
	})

	t.Run("FEC-04 CA-10 sem plano seguro, não termina em casa que fecha", func(t *testing.T) {
		// A borda fecha neste turno; o resto livre está no alcance da bomba que sobra.
		e := montar(t, `
			1....
			.....
			.###.
			.###.
			....2
		`, comFechamento(1), comBomba(2, 1, "jogador_2", 2, 9))
		fecham := map[jogo.Posicao]bool{}
		for _, c := range jogo.CasasQueFecham(e) {
			fecham[c] = true
		}
		for semente := uint64(1); semente <= 50; semente++ {
			plano := planejarV2(t, e, semente)
			if casa := casaFinal(t, e, plano); fecham[casa] {
				t.Fatalf("semente %d: termina em %v, que fecha: %+v", semente, casa, plano)
			}
			if c := classificar(e, "jogador_1", plano); c == morre {
				t.Fatalf("semente %d: plano %s", semente, c)
			}
		}
	})
}

func TestAbertura(t *testing.T) {
	t.Run("BOM-07 FEC-02 CA-11 preso planta a bomba que alcança o bloco", func(t *testing.T) {
		e := montar(t, `
			1....#.
			+#####.
			.......
			.......
			.......
			.......
			......2
		`, comFechamento(3))
		for semente := uint64(1); semente <= 50; semente++ {
			plano := planejarV2(t, e, semente)
			if !planta(plano) {
				t.Fatalf("semente %d: não plantou: %+v", semente, plano)
			}
			if c := classificar(e, "jogador_1", plano); c != seguro {
				t.Fatalf("semente %d: plano %s: %+v", semente, c, plano)
			}
			final, _ := simular(e, "jogador_1", plano)
			if slices.Contains(final.BlocosDestrutiveis, p(0, 1)) {
				t.Fatalf("semente %d: o bloco de abertura não foi destruído: %+v", semente, plano)
			}
		}
	})

	t.Run("BOM-02 CA-12 bomba de abertura sem fuga: não planta", func(t *testing.T) {
		e := montar(t, `
			1.#....
			+#.....
			.......
			.......
			.......
			.......
			......2
		`, comFechamento(3))
		for semente := uint64(1); semente <= 50; semente++ {
			plano := planejarV2(t, e, semente)
			if planta(plano) {
				t.Fatalf("semente %d: plantou sem fuga: %+v", semente, plano)
			}
			if c := classificar(e, "jogador_1", plano); c != seguro {
				t.Fatalf("semente %d: plano %s: %+v", semente, c, plano)
			}
		}
	})
}

// sobreviveAoFechamento joga turnos com o bot como jogador_1 e jogador_2
// parado, até o anel 0 fechar (fim do turno turno_fechamento), e diz se
// jogador_1 está vivo depois disso. Falha o teste se houver infração.
func sobreviveAoFechamento(t *testing.T, e jogo.Estado, bot jogo.Bot) bool {
	t.Helper()
	for e.Turno <= e.Config.TurnoFechamento && !jogo.VerificarFim(e).Terminada {
		acoes := bot.Planejar(e.Copiar(), "jogador_1")
		plano, infracoes := jogo.Validar(e, jogo.Plano{JogadorID: "jogador_1", Turno: e.Turno, Acoes: acoes})
		if len(infracoes) > 0 {
			t.Fatalf("VAL-01 turno %d: infrações %+v", e.Turno, infracoes)
		}
		e, _ = jogo.ResolverTurno(e, []jogo.Plano{plano})
	}
	eu, _ := encontrar(e, "jogador_1")
	return eu.Status == jogo.Vivo
}

func TestSaiDoBolsao(t *testing.T) {
	// jogador_1 começa na borda, num bolsão cuja única saída é o bloco (0,1).
	e := montar(t, `
		1....#.
		+#####.
		.......
		...2...
		.......
		.......
		.......
	`, comFechamento(3))
	v1Morreu := false
	for semente := uint64(1); semente <= 50; semente++ {
		if !sobreviveAoFechamento(t, e, NovoV2(semente)) {
			t.Fatalf("FEC-03 FEC-04 ORD-06 CA-13 semente %d: o v2 morreu", semente)
		}
		if !sobreviveAoFechamento(t, e, Novo(semente)) {
			v1Morreu = true
		}
	}
	if !v1Morreu {
		t.Error("FEC-04 CA-13 o v1 sobreviveu em todas as sementes: o tabuleiro não mostra a diferença")
	}
}
