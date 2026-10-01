package main

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

const emL = `
	1.2+
	....
`

func TestChamadaProtegidaNaPartida(t *testing.T) {
	t.Run("BOT-01 CA-04 bot que altera o estado não muda a partida", func(t *testing.T) {
		comVandalo, r1 := partida(t, mapaDe(t, emL), "bombista", "vandalo")
		comEspera, r2 := partida(t, mapaDe(t, emL), "bombista", "espera")
		if comVandalo != comEspera || !reflect.DeepEqual(r1, r2) {
			t.Errorf("saídas diferentes:\n%s\n---\n%s", comVandalo, comEspera)
		}
	})

	casos := []struct {
		nome  string
		bot   string
		aviso string
	}{
		{"BOT-02 CA-05 prazo estourado", "dorminhoco", "aviso: jogador_1 estourou o prazo de 50ms no turno 1 (BOT-02)"},
		{"BOT-02 CA-06 panic", "panico", "aviso: jogador_1 entrou em pânico no turno 1 (BOT-02): bum"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			inicio := time.Now()
			saida, r := partida(t, mapaDe(t, "1.2", comLimite(2), comPrazo(50)), c.bot, "espera")
			if time.Since(inicio) > time.Second {
				t.Errorf("a partida esperou o bot: %v", time.Since(inicio))
			}
			if !strings.Contains(saida, c.aviso+"\n") {
				t.Errorf("aviso ausente: %q\n%s", c.aviso, saida)
			}
			if linha := tabuleiroDa(saida, "Turno 1, etapa 7", 1); !reflect.DeepEqual(linha, []string{"1.2"}) {
				t.Errorf("o jogador não esperou no turno 1: %v", linha)
			}
			if !r.desfecho.Terminada || r.desfecho.Vencedor != "" {
				t.Errorf("a partida não continuou até o limite: %+v", r.desfecho)
			}
		})
	}
}

func TestLacoDaPartida(t *testing.T) {
	casos := []struct {
		nome      string
		desenho   string
		limite    int
		versoes   []string
		contem    []string
		tabuleiro map[string][]string // cabeçalho → linhas
		ausente   []string
		fim       string
	}{
		{
			nome:    "VAL-05 CA-07 infração impressa e plano validado executado",
			desenho: "1..\n...\n..2",
			limite:  1,
			versoes: []string{"fora", "espera"},
			contem:  []string{"infração: jogador_1, turno 1, etapa 2, VAL-02: "},
			tabuleiro: map[string][]string{
				"Turno 1, etapa 1": {".1.", "...", "..2"},
				"Turno 1, etapa 3": {".1.", "...", "..2"}, // a 3ª ação (DIREITA) virou ESPERAR
			},
			fim: "Fim: empate por limite de turnos entre jogador_1, jogador_2",
		},
		{
			nome:      "FIM-02 CA-10 vitória com turno e etapa do fim",
			desenho:   emL,
			versoes:   []string{"bombista", "espera"},
			contem:    []string{"mortes: jogador_2"},
			tabuleiro: map[string][]string{"Turno 1, etapa 4": {"***+", "*1.."}},
			fim:       "Fim: vitória de jogador_1 no turno 1, etapa 4",
		},
		{
			nome:    "DEC-06 CA-09 a partida termina no meio do turno",
			desenho: emL,
			versoes: []string{"bombista", "espera"},
			ausente: []string{"Turno 1, etapa 5", "Turno 2"},
			fim:     "Fim: vitória de jogador_1 no turno 1, etapa 4",
		},
		{
			nome:    "FIM-03 CA-11 empate por morte simultânea",
			desenho: "1.2",
			versoes: []string{"suicida", "suicida"},
			contem:  []string{"mortes: jogador_1, jogador_2"},
			fim:     "Fim: empate, todos morreram no turno 1, etapa 4",
		},
		{
			nome:    "FIM-04 CA-12 empate por limite de turnos",
			desenho: "1.2",
			limite:  2,
			versoes: []string{"espera", "espera"},
			contem:  []string{"Turno 2, etapa 7"},
			ausente: []string{"Turno 3"},
			fim:     "Fim: empate por limite de turnos entre jogador_1, jogador_2",
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			m := mapaDe(t, c.desenho)
			if c.limite > 0 {
				m.Config.LimiteTurnos = c.limite
			}
			saida, _ := partida(t, m, c.versoes...)
			for _, s := range c.contem {
				if !strings.Contains(saida, s) {
					t.Errorf("falta %q", s)
				}
			}
			for _, s := range c.ausente {
				if strings.Contains(saida, s) {
					t.Errorf("não devia ter %q", s)
				}
			}
			for cab, esperado := range c.tabuleiro {
				if obtido := tabuleiroDa(saida, cab, len(esperado)); !reflect.DeepEqual(obtido, esperado) {
					t.Errorf("%s: %v, esperado %v", cab, obtido, esperado)
				}
			}
			if ultimaLinha(saida) != c.fim {
				t.Errorf("última linha %q, esperado %q", ultimaLinha(saida), c.fim)
			}
			if t.Failed() {
				t.Log("\n" + saida)
			}
		})
	}
}
