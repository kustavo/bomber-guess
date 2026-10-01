# Marco 12: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01 (registrado depois da implementação)

## Visão geral

O fechamento é um passo a mais no fim de `ResolverTurno`, depois do laço de etapas, que só roda se a partida não terminou no meio do turno. Ele usa `CasasQueFecham`, uma função pública e pura que também serve aos bots. As casas que fecham vão para `blocos_fixos` e saem dos destrutíveis. As bombas nelas somem, e os jogadores nelas morrem. O resultado é registrado no relatório da última etapa. Terminal e bot só consomem o novo campo e a nova função.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/internal/jogo/estado.go` | `Config.TurnoFechamento` (`turno_fechamento,omitempty`) |
| `backend/internal/jogo/fechamento.go` | `Config.Anel`, `CasasQueFecham` e `mesa.fechar` |
| `backend/internal/jogo/resolver.go` | chamada de `fechar` depois do laço (ORD-08, FEC-07); `BlocosFechados: []` em cada relatório |
| `backend/internal/jogo/relatorio.go` | `RelatorioEtapa.BlocosFechados` |
| `backend/internal/jogo/mapa.go` | `VerificarMapa` valida `turno_fechamento` (MAP-05) |
| `backend/internal/jogo/fechamento_test.go` | CA-01 a CA-08 |
| `backend/internal/jogo/mapa_test.go`, `resolver_test.go` | CA-09, CA-10 |
| `backend/cmd/terminal/desenho.go`, `desenho_test.go` | evento e blocos fixos novos no quadro (CA-11) |
| `backend/internal/bots/aleatorio/previsao.go` | casas que fecham entram na zona de perigo; fechamento desligado no estado auxiliar |
| `backend/internal/bots/aleatorio/fuga_test.go`, `tabuleiro_test.go` | CA-12; oráculo `classificar` ignora o fechamento do turno seguinte |
| `mapas/exemplo.json` | `turno_fechamento` 30 (CA-13) |
| `docs/REGRAS.md`, `docs/EDITOR.md`, `docs/ARQUITETURA.md` | EST-09, ORD-08, FEC-01 a FEC-07, MAP-05, `blocos_fechados` |

## Tipos e assinaturas

```go
type Config struct {
	// ...
	TurnoFechamento int `json:"turno_fechamento,omitempty"` // EST-09, FEC-01: 0 desliga
}

type RelatorioEtapa struct {
	// ...
	BlocosFechados []Posicao `json:"blocos_fechados"` // FEC-03: só na última etapa do turno
}

func (c Config) Anel(p Posicao) int        // FEC-02
func CasasQueFecham(estado Estado) []Posicao // FEC-01, FEC-03: ordenadas por y e x, sem as que já são fixas
```

## Decisões

- **D1**: O fechamento é registrado no relatório da última etapa executada, não num relatório próprio. **Motivo**: `RelatorioEtapa` é por etapa, e a morte precisa de uma etapa (EST-07, FEC-04). Uma "etapa extra" quebraria `etapas_neste_turno` e o ritmo do loop do servidor.
- **D2**: `CasasQueFecham` é pública e recebe um `Estado`. **Motivo**: bots precisam prever o fechamento sem duplicar a regra. O `jogo` continua sem depender de ninguém.
- **D3**: `turno_fechamento` usa `omitempty`. **Motivo**: estados e mapas sem fechamento continuam com o JSON de antes.
- **D4**: As casas novas são acrescentadas ao fim de `blocos_fixos`, depois de `slices.Clip`. **Motivo**: manter a ordem original (D7 do marco 2) e não escrever na capacidade do slice recebido (pureza, CA-08).
- **D5**: O estado auxiliar do `aleatorio-v1` desliga o fechamento, e as casas que fecham entram na zona de perigo final. **Motivo**: o auxiliar usa `turno` 1 e fantasmas fora do tabuleiro, então o fechamento simulado ali seria o do turno errado.
- **D6**: O oráculo `classificar` dos testes do bot desliga o fechamento ao explodir as bombas que sobram. **Motivo**: ele mede só o perigo das bombas; o fechamento do turno seguinte é decisão de outro turno.

## Riscos

- O frontend do marco 6 ignora `blocos_fechados` até o CA-17 ser implementado; os blocos novos só aparecem com o estado do turno seguinte.
- Bots de outras IAs (marco 10) precisam ler a regra FEC em `docs/REGRAS.md`; ela já está lá.
