package aleatorio

import (
	"strconv"
	"sync"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// resultado é o que uma partida de comparação mede (CA-14 a CA-16).
type resultado struct {
	vencedor       string // versão do bot vencedor; vazio em empate
	porFechamento  int
	somaDosAneis   int // soma dos anéis das mortes por fechamento
	porBomba       int
	fechouComSaida int // mortes por fechamento com saída no início do turno
	// bombas conta, por versão, os jogadores-turno com 0, 1, 2… PLANTAR
	// executados (CA-12 do marco 15).
	bombas map[string]map[int]int
}

// somarBombas acumula as contagens de bombas de r em total.
func (total *resultado) somarBombas(r resultado) {
	if total.bombas == nil {
		total.bombas = map[string]map[int]int{}
	}
	for v, porN := range r.bombas {
		if total.bombas[v] == nil {
			total.bombas[v] = map[int]int{}
		}
		for n, c := range porN {
			total.bombas[v][n] += c
		}
	}
}

// anelMedio é o anel médio das mortes por fechamento (0 se não houve).
func (r resultado) anelMedio() float64 {
	if r.porFechamento == 0 {
		return 0
	}
	return float64(r.somaDosAneis) / float64(r.porFechamento)
}

// jogarComparacao joga uma partida no mapa com o bot de cada posição inicial
// até o fim e classifica as mortes (D11): por fechamento é a de quem não está
// nas chamas da etapa e está numa casa de blocos_fechados.
func jogarComparacao(t *testing.T, m jogo.Mapa, bots []*Bot) resultado {
	t.Helper()
	versoes := make([]string, len(bots))
	for i, b := range bots {
		versoes[i] = b.Versao()
	}
	e, err := jogo.EstadoInicial(m, versoes)
	if err != nil {
		t.Fatal(err)
	}
	r := resultado{bombas: map[string]map[int]int{}}
	for !jogo.VerificarFim(e).Terminada {
		planos := make([]jogo.Plano, 0, len(bots))
		for i, j := range e.Jogadores {
			if j.Status != jogo.Vivo {
				continue
			}
			plano, infracoes := jogo.Validar(e, jogo.Plano{JogadorID: j.ID, Turno: e.Turno, Acoes: bots[i].Planejar(e.Copiar(), j.ID)})
			if len(infracoes) > 0 {
				t.Fatalf("VAL-01 turno %d, %s: infrações %+v", e.Turno, j.ID, infracoes)
			}
			planos = append(planos, plano)
		}
		inicio := e
		var relatorios []jogo.RelatorioEtapa
		e, relatorios = jogo.ResolverTurno(e, planos)
		plantados := map[string]int{}
		for _, rel := range relatorios {
			for _, j := range rel.Jogadores {
				if j.Acao.Tipo == jogo.Plantar && j.Resultado == jogo.Executada {
					plantados[j.ID]++
				}
			}
		}
		for _, j := range inicio.Jogadores {
			if j.Status == jogo.Vivo {
				if r.bombas[j.BotVersao] == nil {
					r.bombas[j.BotVersao] = map[int]int{}
				}
				r.bombas[j.BotVersao][plantados[j.ID]]++
			}
		}
		for _, rel := range relatorios {
			chamas := conjuntoDe(rel.Chamas)
			fechadas := conjuntoDe(rel.BlocosFechados)
			for _, id := range rel.Mortes {
				eu, _ := encontrar(e, id)
				switch {
				case !chamas[eu.Posicao] && fechadas[eu.Posicao]:
					r.porFechamento++
					r.somaDosAneis += e.Config.Anel(eu.Posicao)
					if tinhaSaida(inicio, id) {
						r.fechouComSaida++
					}
				default:
					r.porBomba++
				}
			}
		}
	}
	if d := jogo.VerificarFim(e); !d.Empate {
		eu, _ := encontrar(e, d.Vencedor)
		r.vencedor = eu.BotVersao
	}
	return r
}

// tinhaSaida diz se, no início do turno, o jogador alcançava uma casa que não
// fecha por casas livres em até acoes_por_turno passos.
func tinhaSaida(e jogo.Estado, id string) bool {
	eu, _ := encontrar(e, id)
	fecha := conjuntoDe(jogo.CasasQueFecham(e))
	for casa, d := range distancias(bloqueios(e), e.Config, []jogo.Posicao{eu.Posicao}) {
		if d <= eu.AcoesPorTurno && !fecha[casa] {
			return true
		}
	}
	return false
}

// sementesDeComparacao são as sementes 1 a 200 da spec; com -short, 1 a 50.
func sementesDeComparacao() uint64 {
	if testing.Short() {
		return 50
	}
	return 200
}

// compararEmParalelo joga uma partida por semente, em paralelo, com os bots
// que montar devolve, e soma os resultados.
func compararEmParalelo(t *testing.T, m jogo.Mapa, montar func(semente uint64) []*Bot) (total resultado, vitorias map[string]int, empates int) {
	t.Helper()
	var mu sync.Mutex
	vitorias = map[string]int{}
	t.Run("sementes", func(t *testing.T) {
		for semente := uint64(1); semente <= sementesDeComparacao(); semente++ {
			t.Run(strconv.FormatUint(semente, 10), func(t *testing.T) {
				t.Parallel()
				r := jogarComparacao(t, m, montar(semente))
				mu.Lock()
				defer mu.Unlock()
				total.porFechamento += r.porFechamento
				total.somaDosAneis += r.somaDosAneis
				total.porBomba += r.porBomba
				total.fechouComSaida += r.fechouComSaida
				total.somarBombas(r)
				if r.vencedor == "" {
					empates++
				} else {
					vitorias[r.vencedor]++
				}
			})
		}
	})
	return total, vitorias, empates
}

func quatro(novo func(uint64) *Bot) func(uint64) []*Bot {
	return func(s uint64) []*Bot { return []*Bot{novo(s), novo(s), novo(s), novo(s)} }
}

func TestComparacaoNoMapaExemplo(t *testing.T) {
	m := lerMapaExemplo(t)

	v1, _, empatesV1 := compararEmParalelo(t, m, quatro(Novo))
	v2, _, empatesV2 := compararEmParalelo(t, m, quatro(NovoV2))
	t.Logf("CA-16 4×v1: %d mortes por fechamento (anel médio %.2f), %d por bomba, %d empates", v1.porFechamento, v1.anelMedio(), v1.porBomba, empatesV1)
	t.Logf("CA-16 4×v2: %d mortes por fechamento (anel médio %.2f), %d por bomba, %d empates", v2.porFechamento, v2.anelMedio(), v2.porBomba, empatesV2)
	if 4*v2.porFechamento > v1.porFechamento { // R3
		t.Errorf("FEC-04 CA-14 v2 com %d mortes por fechamento, limite 25%% de %d", v2.porFechamento, v1.porFechamento)
	}
	if v2.fechouComSaida > 0 {
		t.Errorf("FEC-04 CA-14 v2 morreu %d vezes no fechamento com saída disponível", v2.fechouComSaida)
	}

	_, vitorias, empates := compararEmParalelo(t, m, func(s uint64) []*Bot {
		if s%2 == 1 { // D12: alterna as posições
			return []*Bot{Novo(s), NovoV2(s), Novo(s), NovoV2(s)}
		}
		return []*Bot{NovoV2(s), Novo(s), NovoV2(s), Novo(s)}
	})
	t.Logf("CA-16 2×v1 + 2×v2: vitórias %v, %d empates", vitorias, empates)
	if vitorias[VersaoV2] < 2*vitorias[Versao] {
		t.Errorf("CA-15 v2 venceu %d, v1 venceu %d: esperado pelo menos o dobro", vitorias[VersaoV2], vitorias[Versao])
	}
}

func TestComparacaoV3NoMapaExemplo(t *testing.T) {
	m := lerMapaExemplo(t)
	v3, _, empatesV3 := compararEmParalelo(t, m, quatro(NovoV3))
	t.Logf("CA-12 4×v3: bombas por jogador-turno %v; %d mortes por fechamento, %d por bomba, %d empates",
		v3.bombas[VersaoV3], v3.porFechamento, v3.porBomba, empatesV3)

	total, vitorias, empates := compararEmParalelo(t, m, func(s uint64) []*Bot {
		if s%2 == 1 { // alterna as posições, como no CA-15 do marco 13
			return []*Bot{NovoV2(s), NovoV3(s), NovoV2(s), NovoV3(s)}
		}
		return []*Bot{NovoV3(s), NovoV2(s), NovoV3(s), NovoV2(s)}
	})
	t.Logf("CA-12 2×v2 + 2×v3: vitórias %v, %d empates; bombas v2 %v, v3 %v; %d mortes por fechamento, %d por bomba",
		vitorias, empates, total.bombas[VersaoV2], total.bombas[VersaoV3], total.porFechamento, total.porBomba)
}
