package aleatorio

import (
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// lerMapaExemplo carrega mapas/exemplo.json.
func lerMapaExemplo(t *testing.T) jogo.Mapa {
	t.Helper()
	dados, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "mapas", "exemplo.json"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := jogo.LerMapa(dados)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// jogarPartida joga uma partida no mapa com um bot aleatorio-v1 da semente
// dada em cada posição inicial (D10), até o fim, verificando a cada turno os
// CA-06, CA-07 e CA-21. Devolve quantos planos tinham PLANTAR.
func jogarPartida(t *testing.T, novo func(uint64) *Bot, m jogo.Mapa, semente uint64) int {
	t.Helper()
	versoes := make([]string, len(m.PosicoesIniciais))
	for i := range versoes {
		versoes[i] = Versao
	}
	e, err := jogo.EstadoInicial(m, versoes)
	if err != nil {
		t.Fatal(err)
	}
	bot := novo(semente)
	plantios := 0
	for !jogo.VerificarFim(e).Terminada {
		var planos []jogo.Plano
		for _, j := range e.Jogadores {
			if j.Status != jogo.Vivo {
				continue
			}
			acoes := bot.Planejar(e.Copiar(), j.ID)
			plano, infracoes := jogo.Validar(e, jogo.Plano{JogadorID: j.ID, Turno: e.Turno, Acoes: acoes})
			if len(infracoes) > 0 {
				t.Fatalf("BOT-04 CA-06 turno %d, %s: infrações %+v no plano %+v", e.Turno, j.ID, infracoes, acoes)
			}
			if planta(acoes) {
				plantios++
				if c := classificar(e, j.ID, acoes); c != seguro {
					t.Fatalf("BOM-11 CA-21 turno %d, %s: plano com bomba %s: %+v", e.Turno, j.ID, c, acoes)
				}
			}
			planos = append(planos, plano)
		}
		var relatorios []jogo.RelatorioEtapa
		e, relatorios = jogo.ResolverTurno(e, planos)
		for _, r := range relatorios {
			for _, j := range r.Jogadores {
				if j.Resultado == jogo.Bloqueada || j.Resultado == jogo.Abortada {
					t.Fatalf("MOV-05 CA-07 turno %d, etapa %d, %s: ação %s", r.Turno, r.Etapa, j.ID, j.Resultado)
				}
			}
		}
	}
	return plantios
}

func TestPartidaNoMapaExemplo(t *testing.T) {
	paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot) {
		m := lerMapaExemplo(t)
		const sementes = 50
		var plantios atomic.Int64
		t.Run("sementes", func(t *testing.T) {
			for semente := uint64(1); semente <= sementes; semente++ {
				t.Run("VAL-01 CA-06 semente "+strconv.FormatUint(semente, 10), func(t *testing.T) {
					t.Parallel()
					plantios.Add(int64(jogarPartida(t, novo, m, semente)))
				})
			}
		})
		if plantios.Load() == 0 {
			t.Error("BOM-01 CA-12 nenhum plano com PLANTAR em nenhuma partida")
		}
	})
}
