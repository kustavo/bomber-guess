package partida

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/bots/aleatorio"
	"github.com/kustavo/bomber-guess/backend/internal/fila"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

func TestFimNoMeioDoTurno(t *testing.T) {
	m := novaMesa(t, mapaDe(t, emL), noTurno1("P B D"), espera)
	t1 := m.jogarTurno(t, t0, 2)
	for _, agora := range []time.Time{t1.Add(3 * d), t1.Add(4*d - time.Nanosecond)} {
		if v := m.Visao(agora); v.Fase != Execucao || v.Etapa != 4 || len(v.Etapas) != 4 {
			t.Errorf("DEC-06 CA-08 em t1+%v: fase %s, etapa %d", agora.Sub(t1), v.Fase, v.Etapa)
		}
	}
	for _, agora := range []time.Time{t1.Add(4 * d), t1.Add(time.Hour)} {
		v := m.Visao(agora)
		if v.Fase != Encerrada || v.Turno != 1 || v.Etapa != 4 || len(v.Etapas) != 4 || !v.FimDaFase.IsZero() {
			t.Errorf("DEC-06 CA-08 em t1+%v: %s, turno %d, etapa %d, fim %v", agora.Sub(t1), v.Fase, v.Turno, v.Etapa, v.FimDaFase)
		}
		if v.Desfecho.Vencedor != "jogador_1" {
			t.Errorf("FIM-02 CA-08 desfecho %+v", v.Desfecho)
		}
	}
	if proximo, encerrada := m.Avancar(t1.Add(time.Hour)); !encerrada || !proximo.IsZero() {
		t.Errorf("CA-08 Avancar depois do fim: %v, %v", proximo, encerrada)
	}
	if n := len(m.Historico().Turnos); n != 1 {
		t.Errorf("CA-08 %d turnos no histórico", n)
	}
}

func TestLimiteDeTurnos(t *testing.T) {
	mapa := mapaDe(t, doisCantos)
	mapa.Config.LimiteTurnos = 2
	m := novaMesa(t, mapa, espera, espera)
	fim := m.jogarAteOFim(t)
	v := m.Visao(fim)
	h := m.Historico()
	if v.Fase != Encerrada || !v.Desfecho.Empate || len(v.Desfecho.Sobreviventes) != 2 {
		t.Errorf("FIM-04 CA-09 %s, desfecho %+v", v.Fase, v.Desfecho)
	}
	if len(h.Turnos) != 2 || h.Desfecho == nil || !h.Desfecho.Empate {
		t.Errorf("DEC-07 CA-09 %d turnos, desfecho %+v", len(h.Turnos), h.Desfecho)
	}
}

func TestPartidasIndependentes(t *testing.T) {
	f := fila.NovaMemoria()
	const atraso = 500 * time.Millisecond
	a := novaMesaNa(t, f, "a", mapaDe(t, doisCantos), t0, sempre("D"), espera)
	b := novaMesaNa(t, f, "b", mapaDe(t, doisCantos), t0.Add(atraso), sempre("B"), espera)
	a.esperarPlanos(t, 1, 2)
	b.esperarPlanos(t, 1, 2)
	if va, vb := a.Visao(t0.Add(p)), b.Visao(t0.Add(p)); va.Fase != Execucao || vb.Fase != Planejamento {
		t.Errorf("PAR-01 CA-10 em t0+p: a %s, b %s", va.Fase, vb.Fase)
	}
	fimA, fimB := t0.Add(p+7*d), t0.Add(atraso+p+7*d)
	va, vb := a.Visao(fimA), b.Visao(fimB)
	if va.Turno != 2 || vb.Turno != 2 {
		t.Fatalf("PAR-01 CA-10 turnos %d e %d", va.Turno, vb.Turno)
	}
	if pa, pb := va.Estado.Jogadores[0].Posicao, vb.Estado.Jogadores[0].Posicao; pa != (jogo.Posicao{X: 1}) || pb != (jogo.Posicao{Y: 1}) {
		t.Errorf("PAR-01 CA-10 cada partida usou os planos da outra: a %v, b %v", pa, pb)
	}
}

func TestIgualASimulacaoSincrona(t *testing.T) {
	dados, err := os.ReadFile(filepath.Join("..", "..", "..", "mapas", "exemplo.json"))
	if err != nil {
		t.Fatal(err)
	}
	mapa, err := jogo.LerMapa(dados)
	if err != nil {
		t.Fatal(err)
	}
	bot := aleatorio.Novo(1)
	m := novaMesa(t, mapa, bot, bot, bot, bot)
	m.jogarAteOFim(t)
	h := m.Historico()

	estado := m.inicial
	for i, r := range h.Turnos {
		var planos []jogo.Plano
		for _, j := range estado.Jogadores {
			if j.Status == jogo.Vivo {
				p, _ := jogo.Validar(estado, jogo.Plano{JogadorID: j.ID, Turno: estado.Turno, Acoes: bot.Planejar(estado.Copiar(), j.ID)})
				planos = append(planos, p)
			}
		}
		if !reflect.DeepEqual(r.Estado, estado) {
			t.Fatalf("RES-03 CA-11 turno %d: estado inicial diferente", i+1)
		}
		for k, rj := range r.Jogadores {
			if !reflect.DeepEqual(rj.Validadas, append([]jogo.Acao{}, planos[k].Acoes...)) {
				t.Fatalf("RES-03 CA-11 turno %d, %s: planos diferentes", i+1, rj.JogadorID)
			}
		}
		var etapas []jogo.RelatorioEtapa
		estado, etapas = jogo.ResolverTurno(estado, planos)
		if !reflect.DeepEqual(r.Etapas, etapas) {
			t.Fatalf("RES-03 CA-11 turno %d: relatórios diferentes", i+1)
		}
	}
	if d := jogo.VerificarFim(estado); !d.Terminada || !reflect.DeepEqual(*h.Desfecho, d) {
		t.Errorf("RES-03 CA-11 desfecho %+v, esperado %+v", h.Desfecho, d)
	}
}

func TestTopicosDeSaida(t *testing.T) {
	f := fila.NovaMemoria()
	resolvidos, _ := f.Consumir(fila.TurnoResolvido)
	finalizadas, _ := f.Consumir(fila.PartidaFinalizada)
	mapa := mapaDe(t, doisCantos)
	mapa.Config.LimiteTurnos = 3
	m := novaMesaNa(t, f, "teste", mapa, t0, sempre("D"), espera)
	m.jogarAteOFim(t)
	h := m.Historico()

	mensagens, _ := resolvidos.Ler()
	if len(mensagens) != len(h.Turnos) {
		t.Fatalf("FILA-02 CA-16 %d mensagens em turno-resolvido, %d turnos", len(mensagens), len(h.Turnos))
	}
	for i, msg := range mensagens {
		var r RegistroTurno
		if err := json.Unmarshal(msg.Valor, &r); err != nil {
			t.Fatal(err)
		}
		esperado, _ := json.Marshal(h.Turnos[i])
		if obtido, _ := json.Marshal(r); msg.Chave != "teste" || string(obtido) != string(esperado) {
			t.Errorf("FILA-02 CA-16 turno-resolvido %d diferente do histórico", i+1)
		}
	}
	mensagens, _ = finalizadas.Ler()
	if len(mensagens) != 1 {
		t.Fatalf("FILA-02 CA-16 %d mensagens em partida-finalizada", len(mensagens))
	}
	var fim Finalizada
	if err := json.Unmarshal(mensagens[0].Valor, &fim); err != nil {
		t.Fatal(err)
	}
	if fim.Partida != "teste" || !reflect.DeepEqual(fim.Desfecho, *h.Desfecho) {
		t.Errorf("FILA-02 CA-16 partida-finalizada %+v", fim)
	}
}
