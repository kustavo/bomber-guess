package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// resumo é o que jogar devolve para os testes conferirem a saída.
type resumo struct {
	desfecho   jogo.Desfecho
	relatorios int // total de relatórios de etapa (CA-16)
}

// jogar roda a partida do estado inicial até o fim, imprimindo em saida. O
// bot i joga com o jogador i do estado. Antes de desenhar cada etapa, espera
// o atraso (D5).
func jogar(estado jogo.Estado, jogadores []jogo.Bot, atraso time.Duration, saida io.Writer) resumo {
	q := novoQuadro(estado)
	q.inicio(saida, estado)
	prazo := time.Duration(estado.Config.PrazoPlanejamentoMs) * time.Millisecond
	var ultimo jogo.RelatorioEtapa
	total := 0
	for !jogo.VerificarFim(estado).Terminada {
		var planos []jogo.Plano
		var avisos []string
		for i, j := range estado.Jogadores { // D4: um de cada vez, na ordem do estado
			if j.Status != jogo.Vivo {
				continue
			}
			r := bots.Chamar(jogadores[i], estado, j.ID, prazo)
			switch r.Falha {
			case bots.PrazoEstourado:
				avisos = append(avisos, fmt.Sprintf("aviso: %s estourou o prazo de %v no turno %d (BOT-02)", j.ID, prazo, estado.Turno))
			case bots.Panico:
				avisos = append(avisos, fmt.Sprintf("aviso: %s entrou em pânico no turno %d (BOT-02): %s", j.ID, estado.Turno, r.Detalhe))
			}
			plano, infracoes := jogo.Validar(estado, jogo.Plano{JogadorID: j.ID, Turno: estado.Turno, Acoes: r.Acoes})
			for _, inf := range infracoes {
				avisos = append(avisos, fmt.Sprintf("infração: %s, turno %d, etapa %d, %s: %s", inf.JogadorID, estado.Turno, inf.Etapa, inf.Regra, inf.Motivo))
			}
			planos = append(planos, plano)
		}
		if len(avisos) > 0 {
			fmt.Fprintf(saida, "\n%s\n", strings.Join(avisos, "\n"))
		}
		var relatorios []jogo.RelatorioEtapa
		estado, relatorios = jogo.ResolverTurno(estado, planos)
		for _, r := range relatorios {
			if atraso > 0 {
				time.Sleep(atraso)
			}
			fmt.Fprintln(saida)
			q.etapa(saida, r)
			ultimo = r
		}
		total += len(relatorios)
	}
	d := jogo.VerificarFim(estado)
	fmt.Fprintf(saida, "\n%s\n", linhaDoFim(d, ultimo))
	return resumo{desfecho: d, relatorios: total}
}

// linhaDoFim descreve o desfecho, com o turno e a etapa do último relatório (D7).
func linhaDoFim(d jogo.Desfecho, ultimo jogo.RelatorioEtapa) string {
	switch {
	case d.Vencedor != "":
		return fmt.Sprintf("Fim: vitória de %s no turno %d, etapa %d", d.Vencedor, ultimo.Turno, ultimo.Etapa)
	case len(d.Sobreviventes) == 0:
		return fmt.Sprintf("Fim: empate, todos morreram no turno %d, etapa %d", ultimo.Turno, ultimo.Etapa)
	default:
		return "Fim: empate por limite de turnos entre " + strings.Join(d.Sobreviventes, ", ")
	}
}
