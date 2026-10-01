// Package api é a API HTTP do servidor (docs/ARQUITETURA.md, seção 4): sem
// WebSockets (API-01), com o horário do servidor em toda resposta (API-02).
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/partida"
)

// tamanhoMaximoDoCorpo limita o corpo de POST /partidas.
const tamanhoMaximoDoCorpo = 1 << 20

type api struct {
	g *partida.Gerenciador
}

// Novo devolve o handler da API sobre o gerenciador (API-01 a API-10).
func Novo(g *partida.Gerenciador) http.Handler {
	a := &api{g: g}
	mux := http.NewServeMux()
	// D9: rotas sem método; cada uma despacha o método e responde 405 em JSON.
	mux.HandleFunc("/bots", a.metodos(map[string]http.HandlerFunc{"GET": a.bots}))
	mux.HandleFunc("/partidas", a.metodos(map[string]http.HandlerFunc{"GET": a.listar, "POST": a.criar}))
	mux.HandleFunc("/partidas/{nome}/estado", a.metodos(map[string]http.HandlerFunc{"GET": a.estado}))
	mux.HandleFunc("/partidas/{nome}/historico", a.metodos(map[string]http.HandlerFunc{"GET": a.historico}))
	mux.HandleFunc("/mapas", a.metodos(map[string]http.HandlerFunc{"POST": a.depois("POST /mapas", "marco 7")}))
	mux.HandleFunc("/partidas/{nome}/turnos/{n}/plano", a.metodos(map[string]http.HandlerFunc{"POST": a.depois("POST /partidas/{nome}/turnos/{n}/plano", "versão 2")}))
	mux.HandleFunc("/ranking", a.metodos(map[string]http.HandlerFunc{"GET": a.depois("GET /ranking", "marco 9")}))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		a.erro(w, http.StatusNotFound, fmt.Sprintf("rota %s não existe", r.URL.Path))
	})
	return mux
}

// metodos despacha a requisição pelo método, ou responde 405.
func (a *api) metodos(handlers map[string]http.HandlerFunc) http.HandlerFunc {
	permitidos := make([]string, 0, len(handlers))
	for m := range handlers {
		permitidos = append(permitidos, m)
	}
	slices.Sort(permitidos)
	return func(w http.ResponseWriter, r *http.Request) {
		if h, ok := handlers[r.Method]; ok {
			h(w, r)
			return
		}
		w.Header().Set("Allow", strings.Join(permitidos, ", "))
		a.erro(w, http.StatusMethodNotAllowed, fmt.Sprintf("método %s não é aceito em %s", r.Method, r.URL.Path))
	}
}

// depois responde 501 para endpoints de outros marcos.
func (a *api) depois(rota, quando string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		a.erro(w, http.StatusNotImplemented, fmt.Sprintf("%s fica para o %s", rota, quando))
	}
}

func (a *api) agora() time.Time { return a.g.Relogio().Agora() }

func (a *api) bots(w http.ResponseWriter, _ *http.Request) {
	responder(w, http.StatusOK, respostaBots{Bots: a.g.Versoes(), HorarioServidor: horario(a.agora())})
}

func (a *api) listar(w http.ResponseWriter, _ *http.Request) {
	agora := a.agora()
	r := respostaPartidas{Partidas: []resumoPartida{}, HorarioServidor: horario(agora)}
	for _, p := range a.g.Listar() {
		v := p.Visao(agora)
		r.Partidas = append(r.Partidas, resumoPartida{Nome: v.Nome, Fase: v.Fase, Turno: v.Turno, FimDaFase: fimDaFase(v), Desfecho: v.Desfecho})
	}
	responder(w, http.StatusOK, r)
}

func (a *api) criar(w http.ResponseWriter, r *http.Request) {
	var pedido partida.Pedido
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, tamanhoMaximoDoCorpo)).Decode(&pedido); err != nil {
		a.erro(w, http.StatusBadRequest, "corpo inválido: "+err.Error())
		return
	}
	p, err := a.g.Criar(pedido)
	switch {
	case errors.Is(err, partida.ErrPedidoInvalido):
		a.erro(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, partida.ErrNomeRepetido):
		a.erro(w, http.StatusConflict, err.Error())
	case err != nil:
		a.erro(w, http.StatusInternalServerError, err.Error())
	default:
		agora := a.agora()
		responder(w, http.StatusCreated, novaRespostaEstado(p.Visao(agora), agora))
	}
}

func (a *api) estado(w http.ResponseWriter, r *http.Request) {
	p, ok := a.partida(w, r)
	if !ok {
		return
	}
	agora := a.agora()
	responder(w, http.StatusOK, novaRespostaEstado(p.Visao(agora), agora))
}

func (a *api) historico(w http.ResponseWriter, r *http.Request) {
	p, ok := a.partida(w, r)
	if !ok {
		return
	}
	agora := a.agora()
	v := p.Visao(agora) // avança antes de copiar o histórico
	responder(w, http.StatusOK, respostaHistorico{Historico: p.Historico(), Fase: v.Fase, FimDaFase: fimDaFase(v), HorarioServidor: horario(agora)})
}

// partida busca a partida do caminho, ou responde 404.
func (a *api) partida(w http.ResponseWriter, r *http.Request) (*partida.Partida, bool) {
	p, err := a.g.Obter(r.PathValue("nome"))
	if err != nil {
		a.erro(w, http.StatusNotFound, err.Error())
		return nil, false
	}
	return p, true
}

func (a *api) erro(w http.ResponseWriter, status int, mensagem string) {
	responder(w, status, respostaErro{Erro: mensagem, HorarioServidor: horario(a.agora())})
}

func responder(w http.ResponseWriter, status int, corpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(corpo)
}
