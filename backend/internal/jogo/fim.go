package jogo

// Desfecho é a situação da partida deduzida de um Estado (DEC-06).
type Desfecho struct {
	Terminada     bool     `json:"terminada"`
	Empate        bool     `json:"empate"`
	Vencedor      string   `json:"vencedor,omitempty"` // id; vazio em empate ou partida em andamento
	Sobreviventes []string `json:"sobreviventes"`      // ids dos vivos
}

// VerificarFim deduz o desfecho: 1 vivo → vitória (FIM-02); 0 vivos → empate
// (FIM-03); 2 ou mais vivos com turno > limite_turnos → empate entre eles
// (FIM-04, DEC-07).
func VerificarFim(estado Estado) Desfecho {
	d := Desfecho{Sobreviventes: []string{}}
	for _, j := range estado.Jogadores {
		if j.Status == Vivo {
			d.Sobreviventes = append(d.Sobreviventes, j.ID)
		}
	}
	switch {
	case len(d.Sobreviventes) == 0:
		d.Terminada, d.Empate = true, true
	case len(d.Sobreviventes) == 1:
		d.Terminada, d.Vencedor = true, d.Sobreviventes[0]
	case estado.Turno > estado.Config.LimiteTurnos:
		d.Terminada, d.Empate = true, true
	}
	return d
}
