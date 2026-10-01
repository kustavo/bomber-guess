// Package filateste tem o teste de contrato que toda implementação de
// fila.Fila precisa passar (CA-12 do marco 5).
package filateste

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/fila"
)

// TestarContrato verifica uma implementação de fila.Fila criada por nova.
func TestarContrato(t *testing.T, nova func(t *testing.T) fila.Fila) {
	t.Run("FILA-01 CA-12 cada consumidor recebe só o seu tópico, em ordem", func(t *testing.T) {
		f := nova(t)
		a := consumir(t, f, "topico-a")
		b := consumir(t, f, "topico-b")
		publicar(t, f, "topico-a", "p", "1")
		publicar(t, f, "topico-b", "q", "x")
		publicar(t, f, "topico-a", "p", "2")
		publicar(t, f, "topico-a", "r", "3")
		conferir(t, ler(t, a), []fila.Mensagem{msg("p", "1"), msg("p", "2"), msg("r", "3")})
		conferir(t, ler(t, b), []fila.Mensagem{msg("q", "x")})
		conferir(t, ler(t, a), nil) // já lidas
	})
	t.Run("FILA-01 CA-12 consumidor só recebe o que foi publicado depois de criado", func(t *testing.T) {
		f := nova(t)
		publicar(t, f, "topico", "", "antes")
		c := consumir(t, f, "topico")
		publicar(t, f, "topico", "", "depois")
		conferir(t, ler(t, c), []fila.Mensagem{msg("", "depois")})
	})
	t.Run("FILA-01 CA-12 dois consumidores do mesmo tópico recebem tudo", func(t *testing.T) {
		f := nova(t)
		c1, c2 := consumir(t, f, "topico"), consumir(t, f, "topico")
		publicar(t, f, "topico", "k", "v")
		conferir(t, ler(t, c1), []fila.Mensagem{msg("k", "v")})
		conferir(t, ler(t, c2), []fila.Mensagem{msg("k", "v")})
	})
	t.Run("FILA-01 CA-12 publicações concorrentes não se perdem", func(t *testing.T) {
		f := nova(t)
		c := consumir(t, f, "topico")
		var wg sync.WaitGroup
		for i := range 50 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				publicar(t, f, "topico", "k", fmt.Sprint(i))
			}()
		}
		wg.Wait()
		if n := len(ler(t, c)); n != 50 {
			t.Errorf("%d mensagens, esperado 50", n)
		}
	})
}

func consumir(t *testing.T, f fila.Fila, topico string) fila.Consumidor {
	t.Helper()
	c, err := f.Consumir(topico)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Fechar() })
	return c
}

func publicar(t *testing.T, f fila.Fila, topico, chave, valor string) {
	t.Helper()
	if err := f.Publicar(topico, fila.Mensagem{Chave: chave, Valor: []byte(valor)}); err != nil {
		t.Error(err)
	}
}

func ler(t *testing.T, c fila.Consumidor) []fila.Mensagem {
	t.Helper()
	m, err := c.Ler()
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func conferir(t *testing.T, obtido, esperado []fila.Mensagem) {
	t.Helper()
	if len(obtido) == 0 && len(esperado) == 0 {
		return
	}
	if !reflect.DeepEqual(obtido, esperado) {
		t.Errorf("obtido %q, esperado %q", obtido, esperado)
	}
}

func msg(chave, valor string) fila.Mensagem {
	return fila.Mensagem{Chave: chave, Valor: []byte(valor)}
}
