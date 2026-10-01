package bots

import (
	"fmt"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// Falha diz por que a chamada ao bot não produziu um plano (BOT-02).
type Falha string

const (
	SemFalha       Falha = ""
	PrazoEstourado Falha = "PRAZO_ESTOURADO"
	Panico         Falha = "PANICO"
)

// Resposta é o resultado de uma chamada protegida.
type Resposta struct {
	Acoes   []jogo.Acao // saída bruta do bot; vazia em caso de falha
	Falha   Falha
	Detalhe string // valor do panic, em texto; vazio nos outros casos
}

// Chamar entrega ao bot uma cópia do estado (BOT-01) e espera o plano até o
// prazo. Prazo estourado ou panic devolvem Acoes vazias e a falha (BOT-02).
// prazo <= 0 significa sem limite. Ao estourar o prazo, não espera a
// goroutine do bot terminar.
func Chamar(bot jogo.Bot, estado jogo.Estado, jogadorID string, prazo time.Duration) Resposta {
	copia := estado.Copiar()
	pronto := make(chan Resposta, 1) // com espaço: o bot atrasado não fica preso
	go func() {
		defer func() {
			if v := recover(); v != nil {
				pronto <- Resposta{Acoes: []jogo.Acao{}, Falha: Panico, Detalhe: fmt.Sprint(v)}
			}
		}()
		pronto <- Resposta{Acoes: bot.Planejar(copia, jogadorID)}
	}()
	if prazo <= 0 {
		return <-pronto
	}
	limite := time.NewTimer(prazo)
	defer limite.Stop()
	select {
	case r := <-pronto:
		return r
	case <-limite.C:
		return Resposta{Acoes: []jogo.Acao{}, Falha: PrazoEstourado}
	}
}
