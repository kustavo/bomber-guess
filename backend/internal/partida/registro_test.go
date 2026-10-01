package partida

import (
	"reflect"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

const emL = `
	1.2+
	....
`

// registroDe devolve o registro do jogador no turno dado do histórico.
func registroDe(t *testing.T, h Historico, turno int, jogadorID string) RegistroJogador {
	t.Helper()
	for _, r := range h.Turnos {
		if r.Turno != turno {
			continue
		}
		for _, j := range r.Jogadores {
			if j.JogadorID == jogadorID {
				return j
			}
		}
	}
	t.Fatalf("sem registro de %s no turno %d", jogadorID, turno)
	return RegistroJogador{}
}

// soEspera confere que todas as ações executadas foram ESPERAR.
func soEspera(t *testing.T, r RegistroJogador) {
	t.Helper()
	if len(r.Executadas) == 0 {
		t.Fatal("nenhuma ação executada")
	}
	for _, a := range r.Executadas {
		if a.Acao.Tipo != jogo.Esperar || a.Resultado != jogo.Executada {
			t.Errorf("etapa %d: %s %s, esperado ESPERAR", a.Etapa, a.Acao.Tipo, a.Resultado)
		}
	}
}

func TestPrazoEstourado(t *testing.T) {
	lento, liberar := bloqueado(t, "D")
	m := novaMesa(t, mapaDe(t, doisCantos), lento, espera)
	t1 := m.jogarTurno(t, t0, 1) // só o jogador_2 respondeu
	liberar()                    // o plano do turno 1 chega atrasado
	m.esperarPlanos(t, 1, 1)
	m.Avancar(t1.Add(7 * d))
	r := registroDe(t, m.Historico(), 1, "jogador_1")
	if r.Falha != bots.PrazoEstourado || len(r.Planejadas) != 0 {
		t.Errorf("PAR-04 CA-04 falha %q, planejadas %+v", r.Falha, r.Planejadas)
	}
	soEspera(t, r)
	if v := m.Visao(t1.Add(7 * d)); v.Estado.Jogadores[0].Posicao != (jogo.Posicao{}) {
		t.Errorf("BOT-02 CA-04 o jogador se moveu: %v", v.Estado.Jogadores[0].Posicao)
	}
}

func TestPanico(t *testing.T) {
	m := novaMesa(t, mapaDe(t, doisCantos), panico, espera)
	t1 := m.jogarTurno(t, t0, 2)
	v := m.Visao(t1.Add(7 * d))
	r := registroDe(t, m.Historico(), 1, "jogador_1")
	if r.Falha != bots.Panico || r.Detalhe != "bum" {
		t.Errorf("BOT-02 CA-05 falha %q, detalhe %q", r.Falha, r.Detalhe)
	}
	soEspera(t, r)
	if v.Fase != Planejamento || v.Turno != 2 {
		t.Errorf("BOT-02 CA-05 a partida não continuou: %s, turno %d", v.Fase, v.Turno)
	}
}

func TestBotQueAlteraOEstado(t *testing.T) {
	jogar := func(outro jogo.Bot) Historico {
		m := novaMesa(t, mapaDe(t, emL), noTurno1("P B D"), outro)
		m.jogarAteOFim(t)
		return m.Historico()
	}
	comVandalo, comEspera := jogar(vandalo), jogar(espera)
	if !reflect.DeepEqual(comVandalo, comEspera) {
		t.Errorf("BOT-01 CA-06 históricos diferentes:\n%+v\n%+v", comVandalo, comEspera)
	}
	if comEspera.Desfecho == nil || comEspera.Desfecho.Vencedor != "jogador_1" {
		t.Errorf("BOT-01 CA-06 desfecho inesperado: %+v", comEspera.Desfecho)
	}
}

func TestTresVersoesDasAcoes(t *testing.T) {
	m := novaMesa(t, mapaDe(t, doisCantos), noTurno1("D C D"), espera)
	t1 := m.jogarTurno(t, t0, 2)
	m.Avancar(t1.Add(7 * d))
	r := registroDe(t, m.Historico(), 1, "jogador_1")
	if !reflect.DeepEqual(r.Planejadas, acoes("D C D")) {
		t.Errorf("BOT-03 CA-07 planejadas %+v", r.Planejadas)
	}
	if !reflect.DeepEqual(r.Validadas, acoes("D . .")) {
		t.Errorf("VAL-05 CA-07 validadas %+v", r.Validadas)
	}
	if len(r.Infracoes) != 1 || r.Infracoes[0].Regra != "VAL-02" || r.Infracoes[0].Etapa != 2 {
		t.Errorf("VAL-02 CA-07 infrações %+v", r.Infracoes)
	}
	if len(r.Executadas) != 7 {
		t.Fatalf("DEC-09 CA-07 %d ações executadas, esperado 7", len(r.Executadas))
	}
	primeira := AcaoExecutada{Etapa: 1, Acao: acoes("D")[0], Resultado: jogo.Executada}
	if r.Executadas[0] != primeira || r.Executadas[1].Acao.Tipo != jogo.Esperar || r.Executadas[6].Etapa != 7 {
		t.Errorf("DEC-09 CA-07 executadas %+v", r.Executadas)
	}
	if r.Falha != "" {
		t.Errorf("CA-07 falha inesperada %q", r.Falha)
	}
}
