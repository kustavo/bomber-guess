package aleatorio

import (
	"math"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// planoDe monta um plano numerado de jogador_1 com `acoes` ações: PLANTAR nas
// etapas dadas e ESPERAR no resto.
func planoDe(acoes int, plantarEm ...int) []jogo.Acao {
	p := make([]jogo.Acao, acoes)
	for i := range p {
		p[i] = jogo.Acao{Tipo: jogo.Esperar}
	}
	for _, k := range plantarEm {
		p[k-1] = jogo.Acao{Tipo: jogo.Plantar}
	}
	return numerar(p)
}

func TestMira(t *testing.T) {
	adversario := func(acoes int) opcao {
		return comAtributos("jogador_2", func(a *jogo.Atributos) { a.AcoesPorTurno = acoes })
	}
	comPavio := func(pavio int) opcao {
		return comAtributos("jogador_1", func(a *jogo.Atributos) { a.PavioPadrao = pavio })
	}
	casos := []struct {
		nome     string
		desenho  string
		opcoes   []opcao
		plano    []jogo.Acao
		esperado float64
	}{
		{"BOM-06 CA-18 adversário fora do alcance", "1...........2", []opcao{adversario(1)}, planoDe(7, 1), 0},
		{"BOM-07 CA-18 beco sem saída: tudo coberto", "1.2#", []opcao{adversario(7)}, planoDe(7, 1), 1},
		{"BOM-06 CA-18 uma de quatro casas coberta", `
			..2..
			....1
			.....
		`, []opcao{adversario(1)}, planoDe(7, 1), 0.25},
		{"CA-18 plano sem bomba", "1.2#", []opcao{adversario(7)}, planoDe(7), 0},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e := montar(t, c.desenho, c.opcoes...)
			eu, _ := encontrar(e, "jogador_1")
			if m := mira(e, eu, e.EtapasNesteTurno, c.plano); math.Abs(m-c.esperado) > 1e-9 {
				t.Errorf("mira = %v, esperado %v", m, c.esperado)
			}
		})
	}

	t.Run("BOM-11 CA-18 bomba do turno seguinte mira onde ele pode terminar", func(t *testing.T) {
		// Adversário com 3 ações. Com pavio 1, a bomba plantada na etapa 1
		// explode na etapa 2 (ele andou até 2 casas); plantada na etapa 7 com
		// pavio 3, explode depois do turno (até 3 casas, igual a uma bomba que
		// explode na etapa 3).
		desenho := `
			.......
			...1...
			.......
			...2...
			.......
		`
		cedo := montar(t, desenho, adversario(3), comPavio(1))
		tarde := montar(t, desenho, adversario(3))
		meio := montar(t, desenho, adversario(3), comPavio(2)) // etapa 1 + pavio 2: explode na etapa 3
		mira := func(e jogo.Estado, plano []jogo.Acao) float64 {
			eu, _ := encontrar(e, "jogador_1")
			return mira(e, eu, e.EtapasNesteTurno, plano)
		}
		mCedo, mTarde, mMeio := mira(cedo, planoDe(7, 1)), mira(tarde, planoDe(7, 7)), mira(meio, planoDe(7, 1))
		// Contado à mão: cedo, 3 de 12 casas a até 2 passos; tarde, 6 de 21 a até 3.
		if math.Abs(mCedo-3.0/12) > 1e-9 || math.Abs(mTarde-6.0/21) > 1e-9 || mTarde != mMeio {
			t.Errorf("cedo %v (esperado 3/12), tarde %v (esperado 6/21), explodindo na etapa 3 %v", mCedo, mTarde, mMeio)
		}
	})

	t.Run("CA-18 soma por adversário, a melhor bomba de cada um", func(t *testing.T) {
		e := montar(t, "2#1.3#", adversario(7), comAtributos("jogador_3", func(a *jogo.Atributos) { a.AcoesPorTurno = 7 }))
		eu, _ := encontrar(e, "jogador_1")
		// jogador_2 está atrás de um bloco fixo (fração 0); jogador_3 só tem as
		// casas 2 a 4, todas no alcance da bomba em 2 (fração 1).
		if m := mira(e, eu, e.EtapasNesteTurno, planoDe(7, 1)); math.Abs(m-1) > 1e-9 {
			t.Errorf("mira = %v, esperado 1", m)
		}
	})
}

func TestV4Encurrala(t *testing.T) {
	// Adversário num beco (bloco em x=12) com 1 ação: uma bomba do bot em
	// x=9 ou x=10 cobre todas as casas dele, e dá para fugir para a esquerda.
	e := montar(t, ".........1.2#", comAtributos("jogador_2", func(a *jogo.Atributos) {
		*a = jogo.Atributos{BombasPorTurno: 1, Potencia: 1, PavioPadrao: 3, AcoesPorTurno: 1}
	}))
	eu, _ := encontrar(e, "jogador_1")
	comMira := 0
	for semente := uint64(1); semente <= 50; semente++ {
		plano := planejarValido(t, NovoV4, e, semente)
		if !fiscalizar(t, e, plano) {
			t.Fatalf("CA-19 semente %d: plano não protegido: %+v", semente, plano)
		}
		if planta(plano) && mira(e, eu, e.EtapasNesteTurno, plano) > 0 {
			comMira++
		}
	}
	if comMira < 45 {
		t.Errorf("CA-19 só %d de 50 sementes plantaram com mira", comMira)
	}
	t.Logf("CA-19 %d de 50 sementes plantaram com mira", comMira)
}

func TestV4MiraNaoVenceAProtecao(t *testing.T) {
	// Encostado na parede, com o adversário a 2 casas: uma bomba que o alcança
	// obriga a fugir pelas casas que ele ameaça. Sem bomba, dá para escapar.
	e := montar(t, "1.2......", comAtributos("jogador_2", func(a *jogo.Atributos) {
		*a = jogo.Atributos{BombasPorTurno: 1, Potencia: 1, PavioPadrao: 3, AcoesPorTurno: 2}
	}))
	eu, _ := encontrar(e, "jogador_1")
	exposta := 0 // premissa: o v3 acha bomba com mira segura, mas não protegida
	for semente := uint64(1); semente <= 50; semente++ {
		if p := NovoV3(semente).Planejar(e.Copiar(), "jogador_1"); planta(p) && mira(e, eu, e.EtapasNesteTurno, p) > 0 &&
			classificar(e, "jogador_1", p) == seguro && !fiscalizar(t, e, p) {
			exposta++
		}
		plano := planejarValido(t, NovoV4, e, semente)
		if !fiscalizar(t, e, plano) {
			t.Fatalf("BOM-02 CA-20 semente %d: plano do v4 não protegido: %+v", semente, plano)
		}
	}
	if exposta == 0 {
		t.Fatal("CA-20 premissa: nenhum plano do v3 tinha bomba com mira segura e exposta")
	}
}
