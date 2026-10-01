// Comando terminal roda uma partida inteira entre bots e imprime o tabuleiro
// em ASCII ao fim de cada etapa (specs/04-terminal).
//
// Uso: go run ./cmd/terminal [-mapa arquivo] [-bots v1,v2,...] [-semente N] [-atraso 200ms] [-listar]
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// Códigos de saída (decisão 6).
const (
	codigoOK        = 0
	codigoErroDeUso = 2
)

// maxJogadores é o limite do desenho, que usa um dígito por jogador (decisão 3).
const maxJogadores = 9

// mapasPadrao são os caminhos tentados quando -mapa não é dado (D8).
var mapasPadrao = []string{"mapas/exemplo.json", "../mapas/exemplo.json"}

func main() {
	os.Exit(rodar(os.Args[1:], os.Stdout, os.Stderr, bots.Padrao()))
}

// rodar executa o terminal com os argumentos dados e devolve o código de
// saída: 0 para partida terminada ou listagem, 2 para erro de uso. Erros de
// uso são detectados antes de escrever qualquer coisa em saida.
func rodar(args []string, saida, erros io.Writer, catalogo bots.Catalogo) int {
	fs := flag.NewFlagSet("terminal", flag.ContinueOnError)
	fs.SetOutput(erros)
	caminho := fs.String("mapa", "", "arquivo do mapa (padrão: "+strings.Join(mapasPadrao, " ou ")+")")
	versoes := fs.String("bots", "aleatorio-v1", "versões dos bots, separadas por vírgula; uma só joga em todas as posições")
	semente := fs.Uint64("semente", 1, "semente de todos os bots")
	atraso := fs.Duration("atraso", -1, "espera antes de cada etapa, ex.: 200ms ou 0 (padrão: duracao_etapa_ms do mapa)")
	listar := fs.Bool("listar", false, "lista as versões do catálogo e sai")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return codigoOK
		}
		return codigoErroDeUso
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(erros, "terminal: argumento inesperado %q\n", fs.Arg(0))
		return codigoErroDeUso
	}
	if *listar {
		for _, v := range catalogo.Versoes() {
			fmt.Fprintln(saida, v)
		}
		return codigoOK
	}
	estado, jogadores, err := preparar(*caminho, *versoes, *semente, catalogo)
	if err != nil {
		fmt.Fprintf(erros, "terminal: %v\n", err)
		return codigoErroDeUso
	}
	if *atraso < 0 { // D5
		*atraso = time.Duration(estado.Config.DuracaoEtapaMs) * time.Millisecond
	}
	jogar(estado, jogadores, *atraso, saida)
	return codigoOK
}

// preparar lê o mapa, cria o estado inicial (MAP-03) e os bots.
func preparar(caminho, versoes string, semente uint64, catalogo bots.Catalogo) (jogo.Estado, []jogo.Bot, error) {
	if caminho == "" {
		caminho = mapaPadrao()
	}
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return jogo.Estado{}, nil, err
	}
	m, err := jogo.LerMapa(dados)
	if err != nil {
		return jogo.Estado{}, nil, fmt.Errorf("%s: %w", caminho, err)
	}
	if n := len(m.PosicoesIniciais); n > maxJogadores {
		return jogo.Estado{}, nil, fmt.Errorf("%s: %d posições iniciais, o terminal desenha no máximo %d jogadores", caminho, n, maxJogadores)
	}
	lista := strings.Split(versoes, ",")
	for i := range lista {
		lista[i] = strings.TrimSpace(lista[i])
	}
	if len(lista) == 1 {
		lista = slices.Repeat(lista[:1], len(m.PosicoesIniciais))
	}
	estado, err := jogo.EstadoInicial(m, lista) // MAP-06
	if err != nil {
		return jogo.Estado{}, nil, fmt.Errorf("%s: %w", caminho, err)
	}
	jogadores := make([]jogo.Bot, len(lista))
	for i, v := range lista {
		if jogadores[i], err = catalogo.Criar(v, semente); err != nil {
			return jogo.Estado{}, nil, err
		}
	}
	return estado, jogadores, nil
}

// mapaPadrao devolve o primeiro mapa padrão que existe, ou o primeiro da
// lista, para a mensagem de erro (D8).
func mapaPadrao() string {
	for _, c := range mapasPadrao {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return mapasPadrao[0]
}
