package jogo

import (
	"reflect"
	"testing"
)

// comFechamento liga o fechamento a partir do turno dado (FEC-01), com área
// mínima 1×1 para que os tabuleiros pequenos dos testes fechem (FEC-08).
func comFechamento(turno int) opcao {
	return func(e *Estado) {
		e.Config.TurnoFechamento = turno
		e.Config.AreaMinima = Area{Largura: 1, Altura: 1}
	}
}

func TestAnel(t *testing.T) {
	c := Config{Largura: 5, Altura: 4}
	casos := []struct {
		nome string
		casa Posicao
		anel int
	}{
		{"FEC-02 canto é anel 0", p(0, 0), 0},
		{"FEC-02 borda direita é anel 0", p(4, 2), 0},
		{"FEC-02 borda de baixo é anel 0", p(2, 3), 0},
		{"FEC-02 casa interna é anel 1", p(1, 1), 1},
		{"FEC-02 centro em tabuleiro de altura par é anel 1", p(2, 2), 1},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if a := c.Anel(caso.casa); a != caso.anel {
				t.Errorf("anel = %d, esperado %d", a, caso.anel)
			}
		})
	}
}

func TestCasasQueFecham(t *testing.T) {
	desenho := `
		1....
		.#...
		....2
	`
	casos := []struct {
		nome   string
		estado Estado
		casas  []Posicao
	}{
		{"FEC-01 desligado sem turno_fechamento", montar(t, desenho, comTurno(9)), []Posicao{}},
		{"FEC-03 turno anterior ao fechamento não fecha", montar(t, desenho, comFechamento(3), comTurno(2)), []Posicao{}},
		{"FEC-03 no turno_fechamento fecha a borda", montar(t, desenho, comFechamento(3), comTurno(3)), []Posicao{
			p(0, 0), p(1, 0), p(2, 0), p(3, 0), p(4, 0), p(0, 1), p(4, 1), p(0, 2), p(1, 2), p(2, 2), p(3, 2), p(4, 2),
		}},
		{"FEC-08 o anel 1 não fecha: sobraria menos que 1×1", montar(t, desenho, comFechamento(3), comTurno(4)), []Posicao{}},
		{"FEC-03 FEC-08 depois de parar, não fecha mais nada", montar(t, desenho, comFechamento(3), comTurno(5)), []Posicao{}},
		{"FEC-03 um turno depois fecha o anel 1, menos os blocos fixos", montar(t, `
			1......
			.......
			..#....
			......2
			.......
		`, comFechamento(3), comTurno(4)), []Posicao{p(1, 1), p(2, 1), p(3, 1), p(4, 1), p(5, 1), p(1, 2), p(5, 2), p(1, 3), p(2, 3), p(3, 3), p(4, 3), p(5, 3)}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if casas := CasasQueFecham(c.estado); !reflect.DeepEqual(casas, c.casas) {
				t.Errorf("casas = %v, esperado %v", casas, c.casas)
			}
		})
	}
}

func TestAreaMinima(t *testing.T) {
	c := Config{Largura: 15, Altura: 13}
	casos := []struct {
		nome  string
		area  Area
		anel  int
		fecha bool
	}{
		{"EST-10 FEC-08 padrão 5×5: anel 3 fecha, sobram 7×5", Area{}, 3, true},
		{"EST-10 FEC-08 padrão 5×5: anel 4 não fecha, sobrariam 5×3", Area{}, 4, false},
		{"FEC-08 anel 5 também não fecha", Area{}, 5, false},
		{"FEC-08 área 1×1: anel 5 fecha, sobram 3×1", Area{Largura: 1, Altura: 1}, 5, true},
		{"FEC-08 área 1×1: anel 6 não fecha", Area{Largura: 1, Altura: 1}, 6, false},
		{"FEC-08 largura e altura comparadas sem girar: 9×7 cabe em 9×7", Area{Largura: 9, Altura: 7}, 2, true},
		{"FEC-08 largura e altura comparadas sem girar: 7×9 não cabe em 9×7", Area{Largura: 7, Altura: 9}, 2, false},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			c.AreaMinima = caso.area
			if f := c.AnelFecha(caso.anel); f != caso.fecha {
				t.Errorf("AnelFecha(%d) = %v, esperado %v", caso.anel, f, caso.fecha)
			}
		})
	}

	t.Run("FEC-08 quem está na área mínima não morre e o fechamento para", func(t *testing.T) {
		e := montar(t, `
			.......
			.1.....
			.......
			.......
			.......
			.....2.
			.......
		`, comAcoes(1))
		e.Config.TurnoFechamento = 1
		for range 3 {
			e, _ = ResolverTurno(e, nil)
		}
		if len(e.BlocosFixos) != 24 {
			t.Errorf("%d blocos fixos, esperado só a borda (24)", len(e.BlocosFixos))
		}
		if d := VerificarFim(e); d.Terminada {
			t.Errorf("desfecho = %+v, esperado partida em andamento", d)
		}
	})
}

func TestFechamento(t *testing.T) {
	t.Run("FEC-03 FEC-04 ORD-08 quem está na borda morre na última etapa", func(t *testing.T) {
		e := montar(t, `
			1....
			..3..
			.2...
			.....
		`, comFechamento(2), comTurno(2), comAcoes(3))
		final, relatorios := ResolverTurno(e, []Plano{plano(t, e, "jogador_2", "B")})
		if len(relatorios) != 3 {
			t.Fatalf("%d relatórios, esperado 3", len(relatorios))
		}
		for _, r := range relatorios[:2] {
			if len(r.BlocosFechados) != 0 || len(r.Mortes) != 0 {
				t.Errorf("etapa %d: fechamento antes da última etapa: %+v", r.Etapa, r)
			}
		}
		ultimo := relatorios[2]
		if !reflect.DeepEqual(ultimo.Mortes, []string{"jogador_1", "jogador_2"}) {
			t.Errorf("mortes = %v", ultimo.Mortes)
		}
		if len(ultimo.BlocosFechados) != 14 {
			t.Errorf("%d casas fechadas, esperado 14", len(ultimo.BlocosFechados))
		}
		if s := situacao(t, ultimo, "jogador_1"); s.Status != Morto {
			t.Errorf("relatório: jogador_1 %s", s.Status)
		}
		for _, j := range final.Jogadores {
			switch j.ID {
			case "jogador_3":
				if j.Status != Vivo {
					t.Errorf("jogador_3 morreu fora da borda")
				}
			default:
				if j.Status != Morto || !reflect.DeepEqual(j.Morte, &Morte{Turno: 2, Etapa: 3}) {
					t.Errorf("%s: status %s, morte %+v", j.ID, j.Status, j.Morte)
				}
			}
		}
		if len(final.BlocosFixos) != 14 {
			t.Errorf("%d blocos fixos no novo estado, esperado 14", len(final.BlocosFixos))
		}
		if d := VerificarFim(final); d.Vencedor != "jogador_3" {
			t.Errorf("desfecho = %+v, esperado vitória de jogador_3", d)
		}
	})

	t.Run("FEC-04 FIM-03 todos na borda é empate", func(t *testing.T) {
		e := montar(t, `
			1.2
			...
			...
		`, comFechamento(1), comAcoes(1))
		final, _ := ResolverTurno(e, nil)
		if d := VerificarFim(final); !d.Empate {
			t.Errorf("desfecho = %+v, esperado empate", d)
		}
	})

	t.Run("FEC-05 FEC-06 bloco destrutível vira fixo e bomba some sem explodir", func(t *testing.T) {
		e := montar(t, `
			+....
			.1...
			...2.
			.....
		`, comFechamento(1), comAcoes(1), comBomba(4, 0, "jogador_1", 2, 2))
		final, relatorios := ResolverTurno(e, nil)
		if len(final.Bombas) != 0 || len(relatorios[0].Bombas) != 0 {
			t.Errorf("bombas = %v, relatório %v; esperado nenhuma", final.Bombas, relatorios[0].Bombas)
		}
		if len(relatorios[0].Explosoes) != 0 || len(relatorios[0].BlocosDestruidos) != 0 {
			t.Errorf("explosões %v, destruídos %v; esperado nada", relatorios[0].Explosoes, relatorios[0].BlocosDestruidos)
		}
		if len(final.BlocosDestrutiveis) != 0 {
			t.Errorf("destrutíveis = %v, esperado nenhum", final.BlocosDestrutiveis)
		}
		if !conjunto(final.BlocosFixos)[p(0, 0)] {
			t.Errorf("(0,0) não virou bloco fixo")
		}
		if VerificarFim(final).Terminada {
			t.Errorf("partida terminou; os dois estão no anel 1")
		}
	})

	t.Run("FEC-07 partida terminada no meio do turno não fecha", func(t *testing.T) {
		e := montar(t, `
			1....
			.....
			....2
		`, comFechamento(1), comAcoes(3), comBomba(0, 1, "jogador_2", 1, 1))
		final, relatorios := ResolverTurno(e, nil)
		if len(relatorios) != 1 || len(relatorios[0].BlocosFechados) != 0 || len(final.BlocosFixos) != 0 {
			t.Errorf("relatórios %d, fechados %v, fixos %v", len(relatorios), relatorios[0].BlocosFechados, final.BlocosFixos)
		}
		if j := final.Jogadores[1]; j.Status != Vivo {
			t.Errorf("jogador_2 morreu pelo fechamento depois do fim")
		}
	})

	t.Run("FEC-03 não altera o estado recebido", func(t *testing.T) {
		e := montar(t, `
			1#...
			.....
			....2
		`, comFechamento(1), comAcoes(1))
		e.BlocosFixos = append(make([]Posicao, 0, 10), e.BlocosFixos...) // capacidade sobrando
		antes := e.Copiar()
		ResolverTurno(e, nil)
		if !reflect.DeepEqual(e, antes) {
			t.Errorf("estado alterado")
		}
		if e.BlocosFixos[:2][1] != (Posicao{}) {
			t.Errorf("fechamento escreveu na capacidade do slice recebido")
		}
	})
}
