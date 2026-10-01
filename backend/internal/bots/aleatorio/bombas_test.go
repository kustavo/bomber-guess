package aleatorio

import (
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// plantios conta as ações PLANTAR do plano.
func plantios(plano []jogo.Acao) int {
	n := 0
	for _, a := range plano {
		if a.Tipo == jogo.Plantar {
			n++
		}
	}
	return n
}

func TestSortearEtapas(t *testing.T) {
	casos := []struct {
		nome     string
		acoes, n int
	}{
		{"BOM-02 CA-07 duas em sete", 7, 2},
		{"BOM-02 CA-07 três em cinco", 5, 3},
		{"BOM-02 CA-07 todas as etapas", 4, 4},
		{"BOM-02 CA-07 uma só", 3, 1},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			for semente := uint64(1); semente <= 50; semente++ {
				ks := sortearEtapas(rand.New(rand.NewPCG(semente, 0)), c.acoes, c.n)
				if len(ks) != c.n || !slices.IsSorted(ks) || len(slices.Compact(slices.Clone(ks))) != c.n {
					t.Fatalf("semente %d: %v", semente, ks)
				}
				if ks[0] < 1 || ks[len(ks)-1] > c.acoes {
					t.Fatalf("semente %d: fora de 1..%d: %v", semente, c.acoes, ks)
				}
			}
		})
	}
}

func TestTentarPlantarVarias(t *testing.T) {
	t.Run("BOM-03 BOM-04 BOM-11 CA-09 planos com 2 bombas são seguros", func(t *testing.T) {
		e := montar(t, campoAberto)
		eu, _ := encontrar(e, "jogador_1")
		etapas := jogo.CalcularEtapas(e.Jogadores)
		achou, atravessa := 0, 0
		for semente := uint64(1); semente <= 50; semente++ {
			passos, ok := tentarPlantarVarias(rand.New(rand.NewPCG(semente, 0)), e, eu, etapas, 2, nil)
			if !ok {
				continue
			}
			achou++
			plano := numerar(passos[:eu.AcoesPorTurno])
			if n := plantios(plano); n != 2 {
				t.Fatalf("semente %d: %d PLANTAR: %+v", semente, n, plano)
			}
			if c := classificar(e, "jogador_1", plano); c != seguro {
				t.Fatalf("semente %d: plano %s: %+v", semente, c, plano)
			}
			if final, _ := simular(e, "jogador_1", plano); len(final.Bombas) > 0 {
				atravessa++
			}
		}
		if achou == 0 {
			t.Fatal("nenhuma semente achou plano com 2 bombas no campo aberto")
		}
		t.Logf("planos com 2 bombas: %d de 50; com bomba que atravessa o turno: %d", achou, atravessa)
	})

	t.Run("BOM-02 CA-09 cercado: sem plano com 2 bombas", func(t *testing.T) {
		e := montar(t, `
			###.
			#1#.
			###2
		`)
		eu, _ := encontrar(e, "jogador_1")
		for semente := uint64(1); semente <= 20; semente++ {
			if passos, ok := tentarPlantarVarias(rand.New(rand.NewPCG(semente, 0)), e, eu, jogo.CalcularEtapas(e.Jogadores), 2, nil); ok {
				t.Fatalf("semente %d: plano no cercado: %+v", semente, passos)
			}
		}
	})
}

func TestV3NoLimiteDeBombas(t *testing.T) {
	for _, b := range []int{1, 2, 3} {
		t.Run("BOM-02 VAL-03 CA-07 bombas_por_turno "+string(rune('0'+b)), func(t *testing.T) {
			com := func(a *jogo.Atributos) { a.BombasPorTurno = b }
			estados := []jogo.Estado{
				montar(t, campoAberto, comAtributos("jogador_1", com)),
				montar(t, campoAberto, comAtributos("jogador_1", com), comBomba(3, 2, "jogador_2", 2, 3)),
				montar(t, campoAberto, comAtributos("jogador_1", com), comAtributos("jogador_1", func(a *jogo.Atributos) { a.AcoesPorTurno = 3 })),
			}
			for i, e := range estados {
				for semente := uint64(1); semente <= 50; semente++ {
					plano := planejarValido(t, NovoV3, e, semente) // sem infração (VAL-03)
					if n := plantios(plano); n > b {
						t.Fatalf("estado %d semente %d: %d PLANTAR com bombas_por_turno %d", i, semente, n, b)
					}
				}
			}
		})
	}
}

func TestV3UsaASegundaBomba(t *testing.T) {
	e := montar(t, campoAberto)
	duas := 0
	for semente := uint64(1); semente <= 50; semente++ {
		plano := planejarValido(t, NovoV3, e, semente)
		if plantios(plano) == 2 {
			duas++
			if c := classificar(e, "jogador_1", plano); c != seguro {
				t.Fatalf("BOM-11 CA-09 semente %d: plano com 2 bombas %s: %+v", semente, c, plano)
			}
		}
	}
	if duas == 0 {
		t.Fatal("BOM-02 CA-08 nenhuma semente plantou 2 bombas no campo aberto")
	}
	t.Logf("CA-08 planos com 2 bombas: %d de 50", duas)
}

func TestV3CaiParaUmaBomba(t *testing.T) {
	// Com 3 ações, qualquer plano com 2 bombas termina sobre uma bomba ou no
	// alcance da pilha; com 1, plantar e virar a esquina em (1,1) é seguro.
	tres := func(a *jogo.Atributos) { a.AcoesPorTurno = 3 }
	e := montar(t, `
		1.#
		#.2
	`, comAtributos("jogador_1", tres), comAtributos("jogador_2", tres))
	uma := 0
	for semente := uint64(1); semente <= 50; semente++ {
		plano := planejarValido(t, NovoV3, e, semente)
		switch plantios(plano) {
		case 0:
		case 1:
			uma++
			if c := classificar(e, "jogador_1", plano); c != seguro {
				t.Fatalf("semente %d: plano com 1 bomba %s: %+v", semente, c, plano)
			}
		default:
			t.Fatalf("BOM-02 CA-11 semente %d: plantou %d sem plano seguro com 2: %+v", semente, plantios(plano), plano)
		}
	}
	if uma == 0 {
		t.Fatal("BOM-02 CA-11 nenhuma semente caiu para 1 bomba")
	}
}

func TestV3IgualAoV2(t *testing.T) {
	uma := func(a *jogo.Atributos) { a.BombasPorTurno = 1 }
	comUmaBomba := func(e jogo.Estado) jogo.Estado {
		for i := range e.Jogadores {
			uma(&e.Jogadores[i].Atributos)
		}
		return e
	}
	exemplo, err := jogo.EstadoInicial(lerMapaExemplo(t), []string{VersaoV3, VersaoV3, VersaoV3, VersaoV3})
	if err != nil {
		t.Fatal(err)
	}
	exemploUma := comUmaBomba(exemplo.Copiar())
	exemploUma.Turno = 10
	naJanelaExemplo := exemplo.Copiar() // bombas_por_turno 2, dentro da janela
	naJanelaExemplo.Turno = exemplo.Config.TurnoFechamento
	casos := []struct {
		nome   string
		estado jogo.Estado
	}{
		{"BOM-02 CA-04 uma bomba por turno, campo aberto", comUmaBomba(montar(t, campoAberto))},
		{"BOM-02 CA-04 uma bomba por turno, com bombas", comUmaBomba(montar(t, campoAberto, comBomba(1, 2, "jogador_2", 2, 2), comBomba(5, 1, "jogador_2", 1, 9)))},
		{"BOM-02 CA-04 uma bomba por turno, mapa de exemplo", exemploUma},
		{"EST-09 FEC-02 CA-05 dentro da janela, campo aberto", montar(t, campoAberto, comFechamento(1))},
		{"EST-09 FEC-02 CA-05 dentro da janela, mapa de exemplo", naJanelaExemplo},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			for semente := uint64(1); semente <= 50; semente++ {
				for _, j := range c.estado.Jogadores {
					v2 := NovoV2(semente).Planejar(c.estado.Copiar(), j.ID)
					v3 := NovoV3(semente).Planejar(c.estado.Copiar(), j.ID)
					if !reflect.DeepEqual(v2, v3) {
						t.Fatalf("semente %d, %s: v2 %+v, v3 %+v", semente, j.ID, v2, v3)
					}
				}
			}
		})
	}
}

func TestV3NuncaMorrePelasPropriasBombas(t *testing.T) {
	// D7 do marco 15: só as bombas do v3 existem; morrer seria errar a previsão.
	atributos := func(a *jogo.Atributos) {
		*a = jogo.Atributos{BombasPorTurno: 2, Potencia: 2, PavioPadrao: 3, AcoesPorTurno: 5}
	}
	atravessou := 0
	for semente := uint64(1); semente <= 50; semente++ {
		e := montar(t, `
			1......
			.......
			.......
			.......
			.......
			.......
			......2
		`, comAtributos("jogador_1", atributos), comAtributos("jogador_2", atributos))
		bot := NovoV3(semente)
		for !jogo.VerificarFim(e).Terminada {
			acoes := bot.Planejar(e.Copiar(), "jogador_1")
			plano, infracoes := jogo.Validar(e, jogo.Plano{JogadorID: "jogador_1", Turno: e.Turno, Acoes: acoes})
			if len(infracoes) > 0 {
				t.Fatalf("VAL-03 semente %d turno %d: infrações %+v", semente, e.Turno, infracoes)
			}
			e, _ = jogo.ResolverTurno(e, []jogo.Plano{plano}) // jogador_2 só espera
			if eu, _ := encontrar(e, "jogador_1"); eu.Status != jogo.Vivo {
				t.Fatalf("BOM-11 ORD-07 CA-10 semente %d: v3 morreu no turno %d (%+v)", semente, eu.Morte.Turno, eu.Morte)
			}
			for _, b := range e.Bombas {
				if b.JogadorID == "jogador_1" {
					atravessou++
					break
				}
			}
		}
	}
	if atravessou == 0 {
		t.Fatal("BOM-11 CA-10 nenhuma bomba do v3 atravessou o turno em 50 sementes")
	}
	t.Logf("CA-10 turnos que terminaram com bomba do v3 no tabuleiro: %d", atravessou)
}
