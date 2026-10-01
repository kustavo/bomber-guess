package fila_test

import (
	"errors"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/fila"
	"github.com/kustavo/bomber-guess/backend/internal/fila/filateste"
)

func TestMemoria(t *testing.T) {
	filateste.TestarContrato(t, func(*testing.T) fila.Fila { return fila.NovaMemoria() })
}

func TestMemoriaConsumidorFechado(t *testing.T) {
	f := fila.NovaMemoria()
	c, _ := f.Consumir("t")
	c.Fechar()
	if err := f.Publicar("t", fila.Mensagem{Valor: []byte("x")}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Ler(); !errors.Is(err, fila.ErrFechado) {
		t.Errorf("Ler depois de Fechar: %v", err)
	}
}
