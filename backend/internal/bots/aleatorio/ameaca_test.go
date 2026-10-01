package aleatorio

import (
	"reflect"
	"slices"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// xs devolve as colunas (na linha 0) das casas ameaçadas, em ordem.
func xs(casas map[jogo.Posicao]int) []int {
	r := []int{}
	for c := range casas {
		r = append(r, c.X)
	}
	slices.Sort(r)
	return r
}

func TestCalcularAmeaca(t *testing.T) {
	adversario := func(potencia, pavio, acoes int) opcao {
		return comAtributos("jogador_2", func(a *jogo.Atributos) {
			*a = jogo.Atributos{BombasPorTurno: 1, Potencia: potencia, PavioPadrao: pavio, AcoesPorTurno: acoes}
		})
	}
	sete := comAtributos("jogador_1", func(a *jogo.Atributos) { a.AcoesPorTurno = 7 })
	casos := []struct {
		nome     string
		desenho  string
		opcoes   []opcao
		porEtapa map[int][]int // etapa → colunas ameaçadas; etapas ausentes: vazias
		final    []int
	}{
		{
			nome:     "BOM-05 BOM-06 ORD-03 CA-06 corredor: etapas 4 a 6 e nada depois",
			desenho:  "1.....2....",
			opcoes:   []opcao{sete, adversario(1, 3, 3)},
			porEtapa: map[int][]int{4: {5, 6, 7}, 5: {4, 5, 6, 7, 8}, 6: {3, 4, 5, 6, 7, 8, 9}},
			final:    []int{},
		},
		{
			nome:     "BOM-07 MOV-02 CA-07 bloco fixo para as chamas e o alcance",
			desenho:  "1..#2....",
			opcoes:   []opcao{sete, adversario(1, 3, 3)},
			porEtapa: map[int][]int{4: {3, 4, 5}, 5: {3, 4, 5, 6}, 6: {3, 4, 5, 6, 7}},
			final:    []int{},
		},
		{
			nome:     "BOM-07 MOV-03 CA-07 bloco destrutível para as chamas e o alcance",
			desenho:  "1..+2....",
			opcoes:   []opcao{sete, adversario(1, 3, 3)},
			porEtapa: map[int][]int{4: {3, 4, 5}, 5: {3, 4, 5, 6}, 6: {3, 4, 5, 6, 7}},
			final:    []int{},
		},
		{
			nome:     "FIM-01 CA-08 adversário morto não ameaça",
			desenho:  "1.....2....",
			opcoes:   []opcao{sete, adversario(1, 3, 3), comMorto("jogador_2")},
			porEtapa: map[int][]int{},
			final:    []int{},
		},
		{
			nome:     "BOM-11 CA-09 pavio longo: o que explode depois do turno vai para a ameaça final",
			desenho:  "1.....2....",
			opcoes:   []opcao{sete, adversario(1, 6, 3)},
			porEtapa: map[int][]int{7: {5, 6, 7}},
			final:    []int{3, 4, 5, 6, 7, 8, 9},
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e := montar(t, c.desenho, c.opcoes...)
			eu, _ := encontrar(e, "jogador_1")
			etapas := jogo.CalcularEtapas(e.Jogadores)
			a := calcularAmeaca(e, eu, etapas)
			if len(a.chamas) != etapas+1 {
				t.Fatalf("chamas com %d posições, esperado %d", len(a.chamas), etapas+1)
			}
			for etapa := 1; etapa <= etapas; etapa++ {
				esperado := c.porEtapa[etapa]
				if esperado == nil {
					esperado = []int{}
				}
				if obtido := xs(a.chamas[etapa]); !slices.Equal(obtido, esperado) {
					t.Errorf("etapa %d: %v, esperado %v", etapa, obtido, esperado)
				}
			}
			if obtido := xs(a.final); !slices.Equal(obtido, c.final) {
				t.Errorf("final: %v, esperado %v", obtido, c.final)
			}
			if a.vazia() != (len(c.porEtapa) == 0 && len(c.final) == 0) {
				t.Errorf("vazia() = %v", a.vazia())
			}
		})
	}
}

func TestAmeacaIgnoraOProprioBot(t *testing.T) {
	// CA-08: a ameaça de jogador_2 não inclui as bombas possíveis dele mesmo,
	// só as de jogador_1, que está longe demais para alcançar alguma casa.
	e := montar(t, "1......................2", comAtributos("jogador_1", func(a *jogo.Atributos) { a.AcoesPorTurno = 2 }))
	eu, _ := encontrar(e, "jogador_2")
	a := calcularAmeaca(e, eu, jogo.CalcularEtapas(e.Jogadores))
	for etapa, casas := range a.chamas {
		for c := range casas {
			if c.X > 10 {
				t.Fatalf("BOM-02 CA-08 etapa %d: casa %v perto do próprio bot está na ameaça", etapa, c)
			}
		}
	}
}

// fiscalizar diz, independente da implementação, se o plano de jogador_1 é
// protegido: seguro pela simulação (classificar) e, em cada etapa t, fora da
// ameaça da etapa t, terminando fora da ameaça final.
func fiscalizar(t *testing.T, e jogo.Estado, plano []jogo.Acao) bool {
	t.Helper()
	return fiscalizarNoNivel(t, e, plano, 1)
}

// fiscalizarNoNivel é o fiscalizar contra a ameaça só com as casas de peso ≥ w.
func fiscalizarNoNivel(t *testing.T, e jogo.Estado, plano []jogo.Acao, w int) bool {
	t.Helper()
	if classificar(e, "jogador_1", plano) != seguro {
		return false
	}
	eu, _ := encontrar(e, "jogador_1")
	a := calcularAmeaca(e, eu, e.EtapasNesteTurno).nivel(w)
	_, relatorios := simular(e, "jogador_1", plano)
	var casa jogo.Posicao
	for _, r := range relatorios {
		for _, j := range r.Jogadores {
			if j.ID == "jogador_1" {
				casa = j.Posicao
			}
		}
		if a.chamas[r.Etapa][casa] > 0 {
			return false
		}
	}
	return a.final[casa] == 0
}

// corredorComAdversario: jogador_1 em x=4, adversário em x=7 com potência 1,
// pavio 3 e 2 ações: ameaça nas etapas 4 e 5, perto de x=7, e nada no fim.
func corredorComAdversario(t *testing.T) jogo.Estado {
	return montar(t, "....1..2.......",
		comAtributos("jogador_2", func(a *jogo.Atributos) {
			*a = jogo.Atributos{BombasPorTurno: 1, Potencia: 1, PavioPadrao: 3, AcoesPorTurno: 2}
		}))
}

func TestV4IgualAoV3SemAmeacaAlcancavel(t *testing.T) {
	casos := []struct {
		nome   string
		estado jogo.Estado
	}{
		{"CA-04 adversário longe demais", montar(t, "1.......................2",
			comAtributos("jogador_2", func(a *jogo.Atributos) { a.AcoesPorTurno = 2 }))},
		{"CA-04 adversário cercado por blocos fixos", montar(t, `
			1......###
			.......#2#
			.......###
		`)},
		{"FIM-01 CA-04 adversário morto", montar(t, campoAberto, comMorto("jogador_2"))},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			for semente := uint64(1); semente <= 50; semente++ {
				v3 := NovoV3(semente).Planejar(c.estado.Copiar(), "jogador_1")
				v4 := NovoV4(semente).Planejar(c.estado.Copiar(), "jogador_1")
				if !reflect.DeepEqual(v3, v4) {
					t.Fatalf("semente %d: v3 %+v, v4 %+v", semente, v3, v4)
				}
			}
		})
	}
}

func TestV4PrefereProtegido(t *testing.T) {
	e := corredorComAdversario(t)
	v3Exposto, comBomba := 0, 0
	for semente := uint64(1); semente <= 50; semente++ {
		if !fiscalizar(t, e, NovoV3(semente).Planejar(e.Copiar(), "jogador_1")) {
			v3Exposto++
		}
		plano := planejarValido(t, NovoV4, e, semente)
		if !fiscalizar(t, e, plano) {
			t.Fatalf("BOM-05 CA-10 semente %d: plano do v4 não protegido: %+v", semente, plano)
		}
		if planta(plano) {
			comBomba++ // CA-12: plano com bomba só se protegido (conferido acima)
		}
	}
	if v3Exposto == 0 {
		t.Fatal("CA-10 premissa: o v3 nunca ficou exposto neste tabuleiro")
	}
	if comBomba == 0 {
		t.Fatal("CA-12 nenhum plano protegido com bomba neste tabuleiro")
	}
	t.Logf("v3 exposto em %d de 50 sementes; v4 protegido em todas, %d com bomba", v3Exposto, comBomba)
}

func TestV4SemProtegidoFicaComOSeguro(t *testing.T) {
	// Tabuleiro pequeno: o adversário ameaça todas as casas em alguma etapa.
	e := montar(t, "1.2", comAtributos("jogador_2", func(a *jogo.Atributos) {
		*a = jogo.Atributos{BombasPorTurno: 1, Potencia: 2, PavioPadrao: 3, AcoesPorTurno: 7}
	}))
	for semente := uint64(1); semente <= 50; semente++ {
		v4 := planejarValido(t, NovoV4, e, semente)
		if c := classificar(e, "jogador_1", v4); c != seguro {
			t.Fatalf("CA-11 semente %d: plano %s: %+v", semente, c, v4)
		}
		if n3, n4 := nivelDoPlano(t, e, NovoV3(semente).Planejar(e.Copiar(), "jogador_1")), nivelDoPlano(t, e, v4); n4 > n3 {
			t.Fatalf("CA-11 semente %d: v4 no nível %d, pior que o v3 (%d)", semente, n4, n3)
		}
	}
}

func TestV4NaJanelaNaoTrocaOAnelAlvo(t *testing.T) {
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
		v3 := NovoV3(semente).Planejar(e.Copiar(), "jogador_1")
		v4 := planejarValido(t, NovoV4, e, semente)
		anel := func(p []jogo.Acao) int { return e.Config.Anel(casaFinal(t, e, p)) }
		if anel(v4) < min(anel(v3), alvo) {
			t.Fatalf("FEC-02 decisão 1 semente %d: v4 termina no anel %d, v3 no %d, alvo %d", semente, anel(v4), anel(v3), alvo)
		}
	}
}

func TestPesoDaAmeaca(t *testing.T) {
	// Corredor do CA-06: adversário em x=6, potência 1, pavio 3, 3 ações.
	e := montar(t, "1.....2....",
		comAtributos("jogador_1", func(a *jogo.Atributos) { a.AcoesPorTurno = 7 }),
		comAtributos("jogador_2", func(a *jogo.Atributos) {
			*a = jogo.Atributos{BombasPorTurno: 1, Potencia: 1, PavioPadrao: 3, AcoesPorTurno: 3}
		}))
	eu, _ := encontrar(e, "jogador_1")
	a := calcularAmeaca(e, eu, jogo.CalcularEtapas(e.Jogadores))
	esperado := map[int]map[int]int{ // etapa → coluna → peso
		4: {5: 1, 6: 1, 7: 1},
		5: {4: 1, 5: 2, 6: 3, 7: 2, 8: 1},
		6: {3: 1, 4: 2, 5: 3, 6: 3, 7: 3, 8: 2, 9: 1},
	}
	for etapa := 1; etapa < len(a.chamas); etapa++ {
		obtido := map[int]int{}
		for c, p := range a.chamas[etapa] {
			obtido[c.X] = p
		}
		quer := esperado[etapa]
		if quer == nil {
			quer = map[int]int{}
		}
		if !reflect.DeepEqual(obtido, quer) {
			t.Errorf("BOM-05 BOM-06 CA-16 etapa %d: pesos %v, esperado %v", etapa, obtido, quer)
		}
	}
	if a.maiorPeso() != 3 {
		t.Errorf("CA-16 maiorPeso = %d, esperado 3", a.maiorPeso())
	}
	if n := a.nivel(3); !reflect.DeepEqual(xs(n.chamas[6]), []int{5, 6, 7}) || len(n.chamas[4]) != 0 {
		t.Errorf("CA-16 nivel(3): etapa 6 %v, etapa 4 %v", xs(n.chamas[6]), xs(n.chamas[4]))
	}
	if got := niveis(13); !reflect.DeepEqual(got, []int{1, 2, 3, 5, 8, 13}) {
		t.Errorf("decisão 7 niveis(13) = %v", got)
	}
	if got := niveis(0); len(got) != 0 {
		t.Errorf("decisão 7 niveis(0) = %v", got)
	}
}

// nivelDoPlano devolve o menor nível em que o plano é protegido (0 se não é
// seguro, maior+1 se não é protegido em nível nenhum).
func nivelDoPlano(t *testing.T, e jogo.Estado, plano []jogo.Acao) int {
	t.Helper()
	if classificar(e, "jogador_1", plano) != seguro {
		return 0
	}
	eu, _ := encontrar(e, "jogador_1")
	maior := calcularAmeaca(e, eu, e.EtapasNesteTurno).maiorPeso()
	for _, w := range niveis(maior) {
		if fiscalizarNoNivel(t, e, plano, w) {
			return w
		}
	}
	return maior + 1
}

func TestV4NoMenorNivel(t *testing.T) {
	// Nenhum plano é protegido no nível 1; o v4 acha o menor nível que dá. O
	// pavio 9 do bot faz as bombas dele explodirem só depois do turno, para a
	// simulação medir a esquiva, e não o adversário morto.
	e := montar(t, "..1..2..",
		comAtributos("jogador_1", func(a *jogo.Atributos) { a.PavioPadrao = 9 }),
		comAtributos("jogador_2", func(a *jogo.Atributos) {
			*a = jogo.Atributos{BombasPorTurno: 1, Potencia: 1, PavioPadrao: 3, AcoesPorTurno: 5}
		}))
	// Um plano que mata o adversário na simulação (em que ele fica parado)
	// apaga a ameaça dele da conta; esses planos ficam fora da comparação.
	mataAdversario := func(p []jogo.Acao) bool {
		final, _ := simular(e, "jogador_1", p)
		adv, _ := encontrar(final, "jogador_2")
		return adv.Status != jogo.Vivo
	}
	menor := 1 << 30
	var n4 []int
	for semente := uint64(1); semente <= 50; semente++ {
		p3 := NovoV3(semente).Planejar(e.Copiar(), "jogador_1")
		p4 := planejarValido(t, NovoV4, e, semente)
		n := nivelDoPlano(t, e, p4)
		if n == 0 {
			t.Fatalf("CA-17 semente %d: plano do v4 não é seguro", semente)
		}
		if mataAdversario(p4) {
			continue
		}
		menor = min(menor, n)
		if !mataAdversario(p3) {
			n3 := nivelDoPlano(t, e, p3)
			if n > n3 {
				t.Fatalf("CA-17 semente %d: v4 no nível %d, v3 no %d", semente, n, n3)
			}
			menor = min(menor, n3)
		}
		n4 = append(n4, n)
	}
	if menor < 2 {
		t.Fatalf("CA-17 premissa: algum plano foi protegido no nível %d", menor)
	}
	for i, n := range n4 {
		if n != menor {
			t.Fatalf("CA-17 plano %d: v4 no nível %d, o menor possível é %d", i+1, n, menor)
		}
	}
	if len(n4) < 25 {
		t.Fatalf("CA-17 só %d planos do v4 entraram na comparação", len(n4))
	}
}
