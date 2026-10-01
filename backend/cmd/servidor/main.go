// Comando servidor sobe a API HTTP do jogo e roda as partidas em tempo real
// (specs/05-servidor).
//
// Uso: go run ./cmd/servidor [-porta 8080] [-mapas diretorio]
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kustavo/bomber-guess/backend/internal/api"
	"github.com/kustavo/bomber-guess/backend/internal/bots"
	"github.com/kustavo/bomber-guess/backend/internal/fila"
	"github.com/kustavo/bomber-guess/backend/internal/partida"
)

// Códigos de saída.
const (
	codigoOK        = 0
	codigoFalha     = 1
	codigoErroDeUso = 2
)

// dirsMapasPadrao são os diretórios tentados quando -mapas não é dado (decisão 11).
var dirsMapasPadrao = []string{"mapas", "../mapas"}

// tempoParaEncerrar é quanto o servidor espera as requisições em andamento.
const tempoParaEncerrar = 2 * time.Second

func main() {
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()
	os.Exit(rodar(ctx, os.Args[1:], os.Stderr, func(endereco string) {
		fmt.Fprintf(os.Stderr, "servidor: escutando em %s\n", endereco)
	}))
}

// rodar sobe o servidor até ctx ser cancelado; pronto recebe o endereço em
// que ele escuta (útil com -porta 0).
func rodar(ctx context.Context, args []string, erros io.Writer, pronto func(endereco string)) int {
	fs := flag.NewFlagSet("servidor", flag.ContinueOnError)
	fs.SetOutput(erros)
	porta := fs.Int("porta", 8080, "porta HTTP (0 escolhe uma livre)")
	dirMapas := fs.String("mapas", "", "diretório dos mapas (padrão: "+strings.Join(dirsMapasPadrao, " ou ")+")")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return codigoOK
		}
		return codigoErroDeUso
	}
	if *dirMapas == "" {
		*dirMapas = dirPadrao()
	}
	if info, err := os.Stat(*dirMapas); err != nil || !info.IsDir() {
		fmt.Fprintf(erros, "servidor: diretório de mapas %q não encontrado\n", *dirMapas)
		return codigoErroDeUso
	}

	g := partida.NovoGerenciador(partida.ConfigGerenciador{
		Catalogo:   bots.Padrao(),
		DirMapas:   *dirMapas,
		Fila:       fila.NovaMemoria(),
		Automatico: true,
	})
	defer g.Encerrar()

	ouvinte, err := net.Listen("tcp", fmt.Sprintf(":%d", *porta))
	if err != nil {
		fmt.Fprintf(erros, "servidor: %v\n", err)
		return codigoFalha
	}
	srv := &http.Server{Handler: api.Novo(g), ReadHeaderTimeout: 5 * time.Second}
	falhou := make(chan error, 1)
	go func() { falhou <- srv.Serve(ouvinte) }()
	pronto(ouvinte.Addr().String())

	select {
	case <-ctx.Done():
		encerrar, cancelar := context.WithTimeout(context.Background(), tempoParaEncerrar)
		defer cancelar()
		if err := srv.Shutdown(encerrar); err != nil {
			// Conexões que nunca mandaram nada só contam como ociosas depois
			// de 5 s no net/http; o que sobrar é fechado à força.
			fmt.Fprintf(erros, "servidor: encerrando à força: %v\n", err)
			srv.Close()
		}
		return codigoOK
	case err := <-falhou:
		fmt.Fprintf(erros, "servidor: %v\n", err)
		return codigoFalha
	}
}

// dirPadrao devolve o primeiro diretório de mapas que existe, ou o primeiro
// da lista, para a mensagem de erro.
func dirPadrao() string {
	for _, d := range dirsMapasPadrao {
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			return d
		}
	}
	return dirsMapasPadrao[0]
}
