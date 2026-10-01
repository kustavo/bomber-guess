package partida

import (
	"reflect"
	"testing"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

const doisCantos = `
	1...
	....
	...2
`

func TestPlanejamento(t *testing.T) {
	m := novaMesa(t, mapaDe(t, doisCantos), espera, espera)
	for _, agora := range []time.Time{t0, t0.Add(p / 2), t0.Add(p - time.Nanosecond)} {
		v := m.Visao(agora)
		if v.Fase != Planejamento || v.Turno != 1 || v.Etapa != 0 || len(v.Etapas) != 0 || !v.FimDaFase.Equal(t0.Add(p)) {
			t.Errorf("PAR-02 CA-01 em t0+%v: %+v", agora.Sub(t0), v)
		}
	}
	if v := m.Visao(t0); !reflect.DeepEqual(v.Estado, m.inicial) {
		t.Errorf("PAR-01 CA-01 estado diferente do inicial")
	}
}

func TestExecucaoLiberaEtapas(t *testing.T) {
	m := novaMesa(t, mapaDe(t, doisCantos), espera, espera)
	t1 := m.jogarTurno(t, t0, 2)
	const n = 7
	for k := 1; k <= n; k++ {
		for _, agora := range []time.Time{t1.Add(time.Duration(k-1) * d), t1.Add(time.Duration(k)*d - time.Nanosecond)} {
			v := m.Visao(agora)
			if v.Fase != Execucao || v.Etapa != k || len(v.Etapas) != k || !v.FimDaFase.Equal(t1.Add(n*d)) {
				t.Fatalf("PAR-03 CA-02 em t1+%v: fase %s, etapa %d, %d etapas, fim %v", agora.Sub(t1), v.Fase, v.Etapa, len(v.Etapas), v.FimDaFase)
			}
			for i, r := range v.Etapas {
				if r.Etapa != i+1 || r.Turno != 1 {
					t.Fatalf("PAR-03 CA-02 relatório %d é turno %d etapa %d", i, r.Turno, r.Etapa)
				}
			}
		}
	}
}

func TestTurnoSeguinte(t *testing.T) {
	m := novaMesa(t, mapaDe(t, doisCantos), sempre("D"), espera)
	t1 := m.jogarTurno(t, t0, 2)
	const n = 7
	v := m.Visao(t1.Add(n * d))
	esperado, _ := jogo.ResolverTurno(m.inicial, []jogo.Plano{{JogadorID: "jogador_1", Turno: 1, Acoes: acoes("D")}})
	if v.Fase != Planejamento || v.Turno != 2 || !v.FimDaFase.Equal(t1.Add(n*d+p)) {
		t.Errorf("PAR-03 CA-03 fase %s, turno %d, fim %v", v.Fase, v.Turno, v.FimDaFase)
	}
	if !reflect.DeepEqual(v.Estado, esperado) {
		t.Errorf("PAR-03 CA-03 estado:\n obtido   %+v\n esperado %+v", v.Estado, esperado)
	}
}

func TestPlanosPelaFila(t *testing.T) {
	m := novaMesa(t, mapaDe(t, doisCantos), sempre("D"), sempre("C"))
	planos := m.esperarPlanos(t, 1, 2)
	porJogador := map[string][]jogo.Acao{}
	for _, pe := range planos {
		porJogador[pe.JogadorID] = pe.Acoes
	}
	esperado := map[string][]jogo.Acao{"jogador_1": acoes("D"), "jogador_2": acoes("C")}
	if !reflect.DeepEqual(porJogador, esperado) {
		t.Errorf("FILA-02 CA-13 planos publicados %+v", porJogador)
	}
	m.Avancar(t0.Add(p))
	v := m.Visao(t0.Add(p + 7*d))
	posicoes := []jogo.Posicao{v.Estado.Jogadores[0].Posicao, v.Estado.Jogadores[1].Posicao}
	if !reflect.DeepEqual(posicoes, []jogo.Posicao{{X: 1, Y: 0}, {X: 3, Y: 1}}) {
		t.Errorf("PAR-02 CA-13 o turno não usou os planos da fila: %v", posicoes)
	}
}
