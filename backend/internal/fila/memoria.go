package fila

import (
	"errors"
	"sync"
)

// ErrFechado é devolvido por um consumidor depois de Fechar.
var ErrFechado = errors.New("consumidor fechado")

// Memoria é a fila em memória: Publicar entrega a mensagem na hora a todos
// os consumidores do tópico. Segura para uso concorrente.
type Memoria struct {
	mu           sync.Mutex
	consumidores map[string][]*consumidorMemoria
}

var _ Fila = (*Memoria)(nil)

// NovaMemoria cria uma fila em memória vazia.
func NovaMemoria() *Memoria {
	return &Memoria{consumidores: map[string][]*consumidorMemoria{}}
}

// Publicar entrega uma cópia da mensagem a cada consumidor do tópico.
func (f *Memoria) Publicar(topico string, m Mensagem) error {
	m.Valor = append([]byte(nil), m.Valor...)
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.consumidores[topico] {
		c.receber(m)
	}
	return nil
}

// Consumir cria um consumidor do tópico.
func (f *Memoria) Consumir(topico string) (Consumidor, error) {
	c := &consumidorMemoria{fila: f, topico: topico}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.consumidores[topico] = append(f.consumidores[topico], c)
	return c, nil
}

// remover tira o consumidor da lista do tópico.
func (f *Memoria) remover(c *consumidorMemoria) {
	f.mu.Lock()
	defer f.mu.Unlock()
	lista := f.consumidores[c.topico]
	for i, outro := range lista {
		if outro == c {
			f.consumidores[c.topico] = append(lista[:i:i], lista[i+1:]...)
			return
		}
	}
}

type consumidorMemoria struct {
	fila     *Memoria
	topico   string
	mu       sync.Mutex
	pendente []Mensagem
	fechado  bool
}

func (c *consumidorMemoria) receber(m Mensagem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.fechado {
		c.pendente = append(c.pendente, m)
	}
}

func (c *consumidorMemoria) Ler() ([]Mensagem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fechado {
		return nil, ErrFechado
	}
	lidas := c.pendente
	c.pendente = nil
	return lidas, nil
}

func (c *consumidorMemoria) Fechar() error {
	c.mu.Lock()
	c.fechado, c.pendente = true, nil
	c.mu.Unlock()
	c.fila.remover(c)
	return nil
}
