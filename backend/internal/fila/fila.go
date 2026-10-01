// Package fila é a fila de mensagens entre o loop das partidas e quem mais
// precisar dos eventos (docs/ARQUITETURA.md, seção 3). A implementação em
// memória vem primeiro; Kafka depois, sem mudar quem usa a interface (FILA-01).
package fila

// Tópicos (FILA-02).
const (
	PlanosEnviados    = "planos-enviados"
	TurnoResolvido    = "turno-resolvido"
	PartidaFinalizada = "partida-finalizada"
)

// Mensagem é um item publicado num tópico. Chave é o nome da partida.
type Mensagem struct {
	Chave string
	Valor []byte // JSON
}

// Fila publica mensagens em tópicos (FILA-01).
type Fila interface {
	Publicar(topico string, m Mensagem) error
	// Consumir cria um consumidor que recebe as mensagens publicadas no
	// tópico a partir de agora, na ordem de publicação.
	Consumir(topico string) (Consumidor, error)
}

// Consumidor lê as mensagens de um tópico.
type Consumidor interface {
	// Ler devolve, sem bloquear, as mensagens que chegaram desde a última leitura.
	Ler() ([]Mensagem, error)
	Fechar() error
}
