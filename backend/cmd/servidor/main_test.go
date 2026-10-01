package main

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestServidor(t *testing.T) {
	cliente := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()
	enderecos := make(chan string, 1)
	codigo := make(chan int, 1)
	var erros strings.Builder
	go func() {
		codigo <- rodar(ctx, []string{"-porta", "0", "-mapas", "../../../mapas"}, &erros, func(e string) { enderecos <- e })
	}()
	var base string
	select {
	case e := <-enderecos:
		base = "http://" + e
	case c := <-codigo:
		t.Fatalf("CA-25 o servidor saiu com %d: %s", c, erros.String())
	}

	resp, err := cliente.Get(base + "/bots")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("CA-25 GET /bots: %v %v", resp, err)
	}
	resp.Body.Close()
	resp, err = cliente.Post(base+"/partidas", "application/json", strings.NewReader(`{"nome": "teste", "mapa": "exemplo", "bots": ["aleatorio-v1"]}`))
	if err != nil || resp.StatusCode != 201 {
		t.Fatalf("CA-25 POST /partidas: %v %v", resp, err)
	}
	resp.Body.Close()

	cancelar()
	select {
	case c := <-codigo:
		if c != 0 {
			t.Errorf("CA-25 código %d: %s", c, erros.String())
		}
	case <-time.After(2*tempoParaEncerrar + time.Second):
		t.Fatal("CA-25 o servidor não encerrou")
	}
	if _, err := cliente.Get(base + "/bots"); err == nil {
		t.Error("CA-25 o servidor ainda aceita requisições")
	}
}

func TestDiretorioDeMapasInexistente(t *testing.T) {
	var erros strings.Builder
	c := rodar(context.Background(), []string{"-porta", "0", "-mapas", "nao/existe"}, &erros, func(string) {})
	if c != 2 || !strings.Contains(erros.String(), "nao/existe") {
		t.Errorf("código %d, erros %q", c, erros.String())
	}
}
