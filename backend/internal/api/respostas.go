package api

import (
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
	"github.com/kustavo/bomber-guess/backend/internal/partida"
)

// horario é um instante no formato da API: RFC 3339, UTC, com milissegundos (D10).
type horario time.Time

const formatoHorario = "2006-01-02T15:04:05.000Z"

func (h horario) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(h).UTC().Format(formatoHorario) + `"`), nil
}

// fimDaFase devolve o fim da fase, ou nil se a partida acabou (decisão 5).
func fimDaFase(v partida.Visao) *horario {
	if v.FimDaFase.IsZero() {
		return nil
	}
	h := horario(v.FimDaFase)
	return &h
}

type respostaErro struct {
	Erro            string  `json:"erro"`
	HorarioServidor horario `json:"horario_servidor"`
}

// respostaMapa é a resposta de POST /mapas (API-03).
type respostaMapa struct {
	Nome            string  `json:"nome"`
	HorarioServidor horario `json:"horario_servidor"`
}

type respostaBots struct {
	Bots            []string `json:"bots"`
	HorarioServidor horario  `json:"horario_servidor"`
}

// respostaEstado é a resposta de estado (API-05, decisão 5).
type respostaEstado struct {
	Nome            string                `json:"nome"`
	Fase            partida.Fase          `json:"fase"`
	Turno           int                   `json:"turno"`
	Etapa           int                   `json:"etapa"`
	HorarioServidor horario               `json:"horario_servidor"`
	FimDaFase       *horario              `json:"fim_da_fase,omitempty"`
	Estado          jogo.Estado           `json:"estado"`
	Etapas          []jogo.RelatorioEtapa `json:"etapas"`
	Desfecho        jogo.Desfecho         `json:"desfecho"`
}

func novaRespostaEstado(v partida.Visao, agora time.Time) respostaEstado {
	return respostaEstado{
		Nome:            v.Nome,
		Fase:            v.Fase,
		Turno:           v.Turno,
		Etapa:           v.Etapa,
		HorarioServidor: horario(agora),
		FimDaFase:       fimDaFase(v),
		Estado:          v.Estado,
		Etapas:          v.Etapas,
		Desfecho:        v.Desfecho,
	}
}

// resumoPartida é um item de GET /partidas (API-10).
type resumoPartida struct {
	Nome      string        `json:"nome"`
	Fase      partida.Fase  `json:"fase"`
	Turno     int           `json:"turno"`
	FimDaFase *horario      `json:"fim_da_fase,omitempty"`
	Desfecho  jogo.Desfecho `json:"desfecho"`
}

type respostaPartidas struct {
	Partidas        []resumoPartida `json:"partidas"`
	HorarioServidor horario         `json:"horario_servidor"`
}

// respostaHistorico é o histórico com a fase atual (API-07, API-02).
type respostaHistorico struct {
	partida.Historico
	Fase            partida.Fase `json:"fase"`
	FimDaFase       *horario     `json:"fim_da_fase,omitempty"`
	HorarioServidor horario      `json:"horario_servidor"`
}
