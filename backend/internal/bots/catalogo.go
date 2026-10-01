// Package bots reúne o catálogo de bots do projeto e a chamada protegida a
// um bot (docs/BOTS.md). Cada bot fica num subpacote próprio.
package bots

import (
	"fmt"
	"slices"

	"github.com/kustavo/bomber-guess/backend/internal/bots/aleatorio"
	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// Fabrica cria um bot com a semente dada.
type Fabrica func(semente uint64) jogo.Bot

// Catalogo associa cada versão de bot à sua fábrica.
type Catalogo map[string]Fabrica

// Padrao devolve o catálogo com todos os bots do projeto. Um bot novo entra
// aqui com uma linha.
func Padrao() Catalogo {
	return Catalogo{
		aleatorio.Versao: func(s uint64) jogo.Bot { return aleatorio.Novo(s) },
	}
}

// Versoes devolve as versões em ordem alfabética.
func (c Catalogo) Versoes() []string {
	versoes := make([]string, 0, len(c))
	for v := range c {
		versoes = append(versoes, v)
	}
	slices.Sort(versoes)
	return versoes
}

// Criar cria o bot da versão dada, ou devolve um erro que cita a versão se
// ela não está no catálogo.
func (c Catalogo) Criar(versao string, semente uint64) (jogo.Bot, error) {
	fabrica, ok := c[versao]
	if !ok {
		return nil, fmt.Errorf("bot %q não está no catálogo", versao)
	}
	return fabrica(semente), nil
}
