package jogo

import (
	"reflect"
	"testing"
)

func TestVerificarFim(t *testing.T) {
	desenho := `
		1.2
		..3
	`
	casos := []struct {
		nome     string
		estado   Estado
		esperado Desfecho
	}{
		{"FIM-02 partida em andamento", montar(t, desenho),
			Desfecho{Sobreviventes: []string{"jogador_1", "jogador_2", "jogador_3"}}},
		{"FIM-02 último vivo vence", montar(t, desenho, comMorto("jogador_1", 1, 1), comMorto("jogador_3", 1, 2)),
			Desfecho{Terminada: true, Vencedor: "jogador_2", Sobreviventes: []string{"jogador_2"}}},
		{"FIM-03 nenhum vivo é empate", montar(t, desenho, comMorto("jogador_1", 1, 1), comMorto("jogador_2", 1, 1), comMorto("jogador_3", 1, 1)),
			Desfecho{Terminada: true, Empate: true, Sobreviventes: []string{}}},
		{"FIM-04 DEC-07 turno no limite ainda é jogado", montar(t, desenho, comTurno(50)),
			Desfecho{Sobreviventes: []string{"jogador_1", "jogador_2", "jogador_3"}}},
		{"FIM-04 DEC-07 turno além do limite é empate entre os vivos", montar(t, desenho, comTurno(51), comMorto("jogador_2", 50, 1)),
			Desfecho{Terminada: true, Empate: true, Sobreviventes: []string{"jogador_1", "jogador_3"}}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if d := VerificarFim(c.estado); !reflect.DeepEqual(d, c.esperado) {
				t.Errorf("desfecho = %+v, esperado %+v", d, c.esperado)
			}
		})
	}
}

func TestFimDaPartida(t *testing.T) {
	casos := []struct {
		nome       string
		estado     Estado
		relatorios int
		esperado   Desfecho
		bombas     int // bombas que sobram no novo estado (DEC-04)
	}{
		{
			// Bomba em (0,1) com pavio 3 mata jogador_1 na etapa 3; sobra outra bomba.
			nome: "FIM-02 DEC-06 DEC-04 vitória no meio do turno para a resolução",
			estado: montar(t, `
				.....2
				.1....
			`, comBomba(0, 1, "jogador_2", 2, 3), comBomba(5, 1, "jogador_2", 1, 9)),
			relatorios: 3,
			esperado:   Desfecho{Terminada: true, Vencedor: "jogador_2", Sobreviventes: []string{"jogador_2"}},
			bombas:     1,
		},
		{
			nome: "FIM-03 DEC-06 os dois últimos morrem na mesma etapa",
			estado: montar(t, `
				......
				1.2...
			`, comBomba(1, 1, "jogador_1", 2, 2)),
			relatorios: 2,
			esperado:   Desfecho{Terminada: true, Empate: true, Sobreviventes: []string{}},
		},
		{
			nome: "FIM-04 DEC-07 turno igual ao limite com 2 vivos termina em empate",
			estado: montar(t, `
				1....
				....2
			`, comTurno(50)),
			relatorios: 7,
			esperado:   Desfecho{Terminada: true, Empate: true, Sobreviventes: []string{"jogador_1", "jogador_2"}},
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			novo, relatorios := ResolverTurno(c.estado, nil)
			if len(relatorios) != c.relatorios {
				t.Errorf("%d relatórios, esperado %d", len(relatorios), c.relatorios)
			}
			if d := VerificarFim(novo); !reflect.DeepEqual(d, c.esperado) {
				t.Errorf("desfecho = %+v, esperado %+v", d, c.esperado)
			}
			if len(novo.Bombas) != c.bombas {
				t.Errorf("bombas = %+v, esperado %d", novo.Bombas, c.bombas)
			}
		})
	}
}

func TestPartidaTerminada(t *testing.T) {
	e := montar(t, `
		1.2
	`, comMorto("jogador_2", 3, 1), comBomba(1, 0, "jogador_1", 1, 1))
	novo, relatorios := ResolverTurno(e, []Plano{plano(t, e, "jogador_1", "D")})
	t.Run("DEC-06 partida terminada não muda nem gera relatório", func(t *testing.T) {
		if !reflect.DeepEqual(novo, e) || len(relatorios) != 0 {
			t.Errorf("novo = %+v, relatórios = %d", novo, len(relatorios))
		}
	})
}

func TestNovoEstado(t *testing.T) {
	// jogador_1 (7 ações) morre na etapa 1; o bloco (3,0) é destruído; a bomba
	// em (5,1) segue para o turno seguinte.
	e := montar(t, `
		1.#+.
		...2.
		....3
	`, comTurno(4),
		comAtributos("jogador_2", func(a *Atributos) { a.AcoesPorTurno = 5 }),
		comAtributos("jogador_3", func(a *Atributos) { a.AcoesPorTurno = 3 }),
		comBomba(0, 1, "jogador_2", 1, 1), comBomba(3, 1, "jogador_2", 1, 9), comBomba(4, 0, "jogador_3", 1, 1))
	novo, relatorios := ResolverTurno(e, nil)
	esperado := Estado{
		Turno:              5,
		Config:             e.Config,
		EtapasNesteTurno:   5,
		BlocosFixos:        []Posicao{p(2, 0)},
		BlocosDestrutiveis: []Posicao{},
		Bombas:             []Bomba{{Posicao: p(3, 1), JogadorID: "jogador_2", Potencia: 1, PavioRestante: 2}},
		Jogadores: []Jogador{
			{ID: "jogador_1", Posicao: p(0, 0), Status: Morto, Morte: &Morte{Turno: 4, Etapa: 1}, Atributos: atributosDeTeste, BotVersao: "teste"},
			e.Jogadores[1],
			e.Jogadores[2],
		},
	}
	t.Run("EST-03 DEC-03 BOM-11 ORD-06 novo estado do turno seguinte", func(t *testing.T) {
		if len(relatorios) != 7 {
			t.Errorf("%d relatórios, esperado 7", len(relatorios))
		}
		if !reflect.DeepEqual(novo, esperado) {
			t.Errorf("novo estado:\n %+v\n esperado:\n %+v", novo, esperado)
		}
	})
}
