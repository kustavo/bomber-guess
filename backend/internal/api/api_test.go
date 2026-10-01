package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/fila"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
	"github.com/kustavo/bomber-guess/backend/internal/partida"
)

var t0 = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

const (
	prazo = time.Second
	etapa = 100 * time.Millisecond
)

type relogioDeTeste struct{ agora time.Time }

func (r *relogioDeTeste) Agora() time.Time { return r.agora }

type botDeTeste struct {
	planejar func(e jogo.Estado) []jogo.Acao
}

func (b botDeTeste) Versao() string                               { return "teste" }
func (b botDeTeste) Planejar(e jogo.Estado, _ string) []jogo.Acao { return b.planejar(e) }

// bombista planta, desce e vai para a direita no turno 1; no mapa "eml" mata
// o jogador_2 na etapa 4.
var bombista = botDeTeste{func(e jogo.Estado) []jogo.Acao {
	if e.Turno != 1 {
		return nil
	}
	return []jogo.Acao{{Etapa: 1, Tipo: jogo.Plantar}, {Etapa: 2, Tipo: jogo.Mover, Direcao: jogo.Baixo}, {Etapa: 3, Tipo: jogo.Mover, Direcao: jogo.Direita}}
}}

var espera = botDeTeste{func(jogo.Estado) []jogo.Acao { return nil }}

// mapaEmL é o mapa "eml": 4×2, jogadores em (0,0) e (2,0), prazo 1 s, etapa 100 ms.
const mapaEmL = `{
  "nome": "eml",
  "config": {"largura": 4, "altura": 2, "limite_turnos": 5, "prazo_planejamento_ms": 1000, "duracao_etapa_ms": 100},
  "jogador_padrao": {"bombas_por_turno": 1, "potencia": 2, "pavio_padrao": 3, "acoes_por_turno": 7},
  "blocos_fixos": [],
  "blocos_destrutiveis": [{"x": 3, "y": 0}],
  "posicoes_iniciais": [{"x": 0, "y": 0}, {"x": 2, "y": 0}]
}`

// servidor é uma API de teste com relógio parado e um observador de planos.
type servidor struct {
	*httptest.Server
	relogio *relogioDeTeste
	planos  fila.Consumidor
}

func novoServidor(t *testing.T) *servidor {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "eml.json"), []byte(mapaEmL), 0o644); err != nil {
		t.Fatal(err)
	}
	f := fila.NovaMemoria()
	planos, _ := f.Consumir(fila.PlanosEnviados)
	relogio := &relogioDeTeste{t0}
	fixo := func(b jogo.Bot) bots.Fabrica { return func(uint64) jogo.Bot { return b } }
	g := partida.NovoGerenciador(partida.ConfigGerenciador{
		Catalogo: bots.Catalogo{"espera": fixo(espera), "bombista": fixo(bombista)},
		DirMapas: dir,
		Fila:     f,
		Relogio:  relogio,
	})
	s := &servidor{Server: httptest.NewServer(Novo(g)), relogio: relogio, planos: planos}
	t.Cleanup(func() { s.Close(); g.Encerrar() })
	return s
}

// pedir faz a requisição e devolve o status e o corpo decodificado.
func (s *servidor) pedir(t *testing.T, metodo, caminho, corpo string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(metodo, s.URL+caminho, strings.NewReader(corpo))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	dados, _ := io.ReadAll(resp.Body)
	var m map[string]any
	if err := json.Unmarshal(dados, &m); err != nil {
		t.Fatalf("%s %s: corpo não é JSON: %q", metodo, caminho, dados)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("%s %s: Content-Type %q", metodo, caminho, ct)
	}
	return resp.StatusCode, m
}

// criar cria a partida "final-1" no mapa eml com bombista e espera.
func (s *servidor) criar(t *testing.T) map[string]any {
	t.Helper()
	status, corpo := s.pedir(t, "POST", "/partidas", `{"nome": "final-1", "mapa": "eml", "bots": ["bombista", "espera"]}`)
	if status != http.StatusCreated {
		t.Fatalf("POST /partidas: %d %v", status, corpo)
	}
	return corpo
}

// esperarPlanos espera n planos em planos-enviados.
func (s *servidor) esperarPlanos(t *testing.T, n int) {
	t.Helper()
	vistos := 0
	limite := time.Now().Add(2 * time.Second)
	for vistos < n {
		m, _ := s.planos.Ler()
		vistos += len(m)
		if time.Now().After(limite) {
			t.Fatalf("%d planos, esperado %d", vistos, n)
		}
		time.Sleep(time.Millisecond)
	}
}

func horarioDe(t time.Time) string { return t.UTC().Format(formatoHorario) }

func TestBots(t *testing.T) {
	s := novoServidor(t)
	status, corpo := s.pedir(t, "GET", "/bots", "")
	bots, _ := corpo["bots"].([]any)
	if status != 200 || len(bots) != 2 || bots[0] != "bombista" || bots[1] != "espera" || corpo["horario_servidor"] != horarioDe(t0) {
		t.Errorf("API-09 CA-17 %d %v", status, corpo)
	}
}

func TestCriarPartida(t *testing.T) {
	s := novoServidor(t)
	corpo := s.criar(t)
	if corpo["nome"] != "final-1" || corpo["fase"] != "PLANEJAMENTO" || corpo["turno"] != 1.0 || corpo["fim_da_fase"] != horarioDe(t0.Add(prazo)) {
		t.Errorf("API-04 CA-18 %v", corpo)
	}
	status, lista := s.pedir(t, "GET", "/partidas", "")
	partidas, _ := lista["partidas"].([]any)
	if status != 200 || len(partidas) != 1 || partidas[0].(map[string]any)["nome"] != "final-1" {
		t.Errorf("API-10 CA-18 %d %v", status, lista)
	}
}

func TestCriarPartidaInvalida(t *testing.T) {
	s := novoServidor(t)
	s.criar(t)
	casos := []struct {
		nome   string
		corpo  string
		status int
	}{
		{"API-04 CA-19 JSON malformado", `{"nome": `, 400},
		{"API-04 CA-19 nome vazio", `{"mapa": "eml", "bots": ["espera"]}`, 400},
		{"API-04 CA-19 nome com caractere inválido", `{"nome": "Final 1", "mapa": "eml", "bots": ["espera"]}`, 400},
		{"API-04 CA-19 mapa inexistente", `{"nome": "x", "mapa": "nenhum", "bots": ["espera"]}`, 400},
		{"API-04 CA-19 bot desconhecido", `{"nome": "x", "mapa": "eml", "bots": ["espera", "nenhum-v9"]}`, 400},
		{"MAP-06 CA-19 quantidade de bots errada", `{"nome": "x", "mapa": "eml", "bots": ["espera", "espera", "espera"]}`, 400},
		{"API-04 CA-19 nome repetido", `{"nome": "final-1", "mapa": "eml", "bots": ["espera"]}`, 409},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			status, corpo := s.pedir(t, "POST", "/partidas", c.corpo)
			if erro, _ := corpo["erro"].(string); status != c.status || erro == "" {
				t.Errorf("%d %v, esperado %d com erro", status, corpo, c.status)
			}
		})
	}
	_, lista := s.pedir(t, "GET", "/partidas", "")
	if n := len(lista["partidas"].([]any)); n != 1 {
		t.Errorf("CA-19 %d partidas criadas", n)
	}
}

func TestEstado(t *testing.T) {
	s := novoServidor(t)
	s.criar(t)
	status, corpo := s.pedir(t, "GET", "/partidas/final-1/estado", "")
	for _, campo := range []string{"nome", "fase", "turno", "etapa", "horario_servidor", "fim_da_fase", "estado", "etapas", "desfecho"} {
		if _, ok := corpo[campo]; !ok || status != 200 {
			t.Errorf("API-05 CA-20 %d, falta %q: %v", status, campo, corpo)
		}
	}
	status, corpo = s.pedir(t, "GET", "/partidas/nenhuma/estado", "")
	if status != 404 || corpo["erro"] == nil {
		t.Errorf("API-05 CA-20 partida desconhecida: %d %v", status, corpo)
	}
}

func TestEstadoNaoMostraEtapasFuturas(t *testing.T) {
	s := novoServidor(t)
	s.criar(t)
	s.esperarPlanos(t, 2)
	t1 := t0.Add(prazo)
	for k := 1; k <= 4; k++ { // a partida termina na etapa 4
		s.relogio.agora = t1.Add(time.Duration(k-1)*etapa + etapa/2)
		_, corpo := s.pedir(t, "GET", "/partidas/final-1/estado", "")
		etapas, _ := corpo["etapas"].([]any)
		if corpo["fase"] != "EXECUCAO" || corpo["etapa"] != float64(k) || len(etapas) != k {
			t.Fatalf("API-05 CA-21 em t1+%v: fase %v, etapa %v, %d etapas", s.relogio.agora.Sub(t1), corpo["fase"], corpo["etapa"], len(etapas))
		}
		for _, e := range etapas {
			if e.(map[string]any)["etapa"].(float64) > float64(k) {
				t.Fatalf("API-05 CA-21 etapa futura liberada: %v", e)
			}
		}
		_, h := s.pedir(t, "GET", "/partidas/final-1/historico", "")
		if n := len(h["turnos"].([]any)); n != 0 {
			t.Fatalf("API-07 CA-21 histórico com %d turnos durante a execução", n)
		}
	}
}

func TestHistorico(t *testing.T) {
	s := novoServidor(t)
	s.criar(t)
	s.esperarPlanos(t, 2)
	s.relogio.agora = t0.Add(prazo + 4*etapa) // fim da partida
	status, h := s.pedir(t, "GET", "/partidas/final-1/historico", "")
	if status != 200 || h["nome"] != "final-1" || h["mapa"] != "eml" || h["semente"] != 1.0 || h["fase"] != "ENCERRADA" {
		t.Fatalf("API-07 CA-22 %d %v", status, h)
	}
	if _, ok := h["fim_da_fase"]; ok {
		t.Errorf("API-02 CA-22 fim_da_fase numa partida encerrada")
	}
	if d, _ := h["desfecho"].(map[string]any); d["vencedor"] != "jogador_1" {
		t.Errorf("API-07 CA-22 desfecho %v", h["desfecho"])
	}
	turnos, _ := h["turnos"].([]any)
	if len(turnos) != 1 {
		t.Fatalf("PAR-05 CA-22 %d turnos", len(turnos))
	}
	jogador := turnos[0].(map[string]any)["jogadores"].([]any)[0].(map[string]any)
	for _, campo := range []string{"jogador_id", "planejadas", "validadas", "executadas", "infracoes"} {
		if _, ok := jogador[campo]; !ok {
			t.Errorf("PAR-05 CA-22 falta %q no registro do jogador", campo)
		}
	}
	if status, corpo := s.pedir(t, "GET", "/partidas/nenhuma/historico", ""); status != 404 || corpo["erro"] == nil {
		t.Errorf("API-07 CA-22 partida desconhecida: %d", status)
	}
}

func TestHorarioEmTodaResposta(t *testing.T) {
	s := novoServidor(t)
	s.criar(t)
	s.relogio.agora = t0.Add(123 * time.Millisecond)
	casos := []struct {
		metodo, caminho, corpo string
		dePartida              bool
	}{
		{"GET", "/bots", "", false},
		{"GET", "/partidas", "", false},
		{"POST", "/partidas", `{"nome": "outra", "mapa": "eml", "bots": ["espera"]}`, true},
		{"GET", "/partidas/final-1/estado", "", true},
		{"GET", "/partidas/final-1/historico", "", true},
		{"GET", "/partidas/nenhuma/estado", "", false},
		{"POST", "/partidas", `{`, false},
		{"DELETE", "/bots", "", false},
		{"GET", "/rota/qualquer", "", false},
		{"GET", "/ranking", "", false},
	}
	for _, c := range casos {
		_, corpo := s.pedir(t, c.metodo, c.caminho, c.corpo)
		if corpo["horario_servidor"] != horarioDe(s.relogio.agora) {
			t.Errorf("API-02 CA-23 %s %s: horario_servidor %v", c.metodo, c.caminho, corpo["horario_servidor"])
		}
		if c.dePartida && (corpo["fase"] == nil || corpo["fim_da_fase"] == nil) {
			t.Errorf("API-02 CA-23 %s %s sem fase ou fim_da_fase", c.metodo, c.caminho)
		}
	}
	_, lista := s.pedir(t, "GET", "/partidas", "")
	for _, p := range lista["partidas"].([]any) {
		if item := p.(map[string]any); item["fase"] == nil || item["fim_da_fase"] == nil {
			t.Errorf("API-02 CA-23 item da lista sem fase ou fim_da_fase: %v", item)
		}
	}
}

func TestEndpointsDeOutrosMarcosEMetodos(t *testing.T) {
	s := novoServidor(t)
	s.criar(t)
	casos := []struct {
		nome, metodo, caminho string
		status                int
	}{
		{"API-03 CA-24 POST /mapas", "POST", "/mapas", 501},
		{"API-06 CA-24 POST plano", "POST", "/partidas/final-1/turnos/1/plano", 501},
		{"API-08 CA-24 GET /ranking", "GET", "/ranking", 501},
		{"API-09 CA-24 método errado em /bots", "POST", "/bots", 405},
		{"API-05 CA-24 método errado em estado", "PUT", "/partidas/final-1/estado", 405},
		{"API-10 CA-24 método errado em /partidas", "DELETE", "/partidas", 405},
		{"API-01 CA-24 rota inexistente", "GET", "/websocket", 404},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			status, corpo := s.pedir(t, c.metodo, c.caminho, "")
			if erro, _ := corpo["erro"].(string); status != c.status || erro == "" {
				t.Errorf("%d %v, esperado %d", status, corpo, c.status)
			}
		})
	}
}
