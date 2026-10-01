package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// Legenda do desenho (decisão 3).
const (
	casaLivre        = '.'
	blocoFixo        = '#'
	blocoDestrutivel = '+'
	bomba            = 'o'
	chama            = '*'
	variosJogadores  = '&'
)

// quadro desenha o tabuleiro e acompanha os blocos destrutíveis ao longo da
// partida.
type quadro struct {
	config       jogo.Config
	fixos        map[jogo.Posicao]bool
	destrutiveis map[jogo.Posicao]bool
}

func novoQuadro(estado jogo.Estado) *quadro {
	q := &quadro{config: estado.Config, fixos: map[jogo.Posicao]bool{}, destrutiveis: map[jogo.Posicao]bool{}}
	for _, p := range estado.BlocosFixos {
		q.fixos[p] = true
	}
	for _, p := range estado.BlocosDestrutiveis {
		q.destrutiveis[p] = true
	}
	return q
}

// inicio desenha o estado inicial da partida.
func (q *quadro) inicio(w io.Writer, estado jogo.Estado) {
	fmt.Fprintf(w, "Turno %d, início\n", estado.Turno)
	vivos := map[jogo.Posicao][]int{}
	for i, j := range estado.Jogadores {
		if j.Status == jogo.Vivo {
			vivos[j.Posicao] = append(vivos[j.Posicao], i+1)
		}
	}
	q.tabuleiro(w, vivos, estado.Bombas, nil)
}

// etapa desenha o tabuleiro ao fim da etapa e os eventos dela; depois remove
// os blocos destruídos, que ficam livres a partir da próxima etapa (ORD-06),
// e acrescenta os blocos fixos do fechamento (FEC-03).
func (q *quadro) etapa(w io.Writer, r jogo.RelatorioEtapa) {
	fmt.Fprintf(w, "Turno %d, etapa %d\n", r.Turno, r.Etapa)
	vivos := map[jogo.Posicao][]int{}
	for i, j := range r.Jogadores { // D6: na ordem do Estado
		if j.Status == jogo.Vivo {
			vivos[j.Posicao] = append(vivos[j.Posicao], i+1)
		}
	}
	q.tabuleiro(w, vivos, r.Bombas, r.Chamas)
	if len(r.Mortes) > 0 {
		fmt.Fprintf(w, "mortes: %s\n", strings.Join(r.Mortes, ", "))
	}
	if len(r.BlocosDestruidos) > 0 {
		casas := make([]string, len(r.BlocosDestruidos))
		for i, p := range r.BlocosDestruidos {
			casas[i] = p.String()
		}
		fmt.Fprintf(w, "blocos destruídos: %s\n", strings.Join(casas, " "))
	}
	if len(r.MovimentosBloqueados) > 0 {
		fmt.Fprintf(w, "bloqueados: %s\n", strings.Join(r.MovimentosBloqueados, ", "))
	}
	if len(r.BlocosFechados) > 0 {
		fmt.Fprintf(w, "fechamento: %d casas\n", len(r.BlocosFechados))
	}
	for _, p := range r.BlocosDestruidos {
		delete(q.destrutiveis, p)
	}
	for _, p := range r.BlocosFechados { // FEC-03, FEC-05
		delete(q.destrutiveis, p)
		q.fixos[p] = true
	}
}

// tabuleiro escreve uma linha por y, com a prioridade jogador > chama >
// bomba > bloco (decisão 3).
func (q *quadro) tabuleiro(w io.Writer, vivos map[jogo.Posicao][]int, bombas []jogo.Bomba, chamas []jogo.Posicao) {
	comBomba := map[jogo.Posicao]bool{}
	for _, b := range bombas {
		comBomba[b.Posicao] = true
	}
	comChama := map[jogo.Posicao]bool{}
	for _, p := range chamas {
		comChama[p] = true
	}
	var linha strings.Builder
	for y := range q.config.Altura {
		linha.Reset()
		for x := range q.config.Largura {
			p := jogo.Posicao{X: x, Y: y}
			switch n := vivos[p]; {
			case len(n) > 1:
				linha.WriteByte(variosJogadores)
			case len(n) == 1:
				linha.WriteString(strconv.Itoa(n[0]))
			case comChama[p]:
				linha.WriteByte(chama)
			case comBomba[p]:
				linha.WriteByte(bomba)
			case q.fixos[p]:
				linha.WriteByte(blocoFixo)
			case q.destrutiveis[p]:
				linha.WriteByte(blocoDestrutivel)
			default:
				linha.WriteByte(casaLivre)
			}
		}
		fmt.Fprintln(w, linha.String())
	}
}
