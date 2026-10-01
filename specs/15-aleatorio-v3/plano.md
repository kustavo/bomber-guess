# Marco 15: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

O v3 é mais um modo do mesmo `Bot` do pacote `internal/bots/aleatorio`, como o v2 (marco 13, D1): o v3 liga a antecipação do fechamento (`antecipa`) e uma nova opção, `variasBombas`. Dentro da janela, nada muda (decisão 3). Fora dela, no turno em que o bot decide plantar (o mesmo sorteio de hoje, `rng.IntN(2) == 0`), o v3 sorteia quantas bombas tentar e procura um plano seguro com elas. Se não achar, tenta com uma a menos e, ao chegar a uma bomba, usa exatamente o `tentarPlantar` de hoje.

O plano com várias bombas generaliza o `tentarPlantar`. Ele sorteia as etapas de plantio em ordem crescente e, entre elas, passeios aleatórios. Depois prevê a linha do tempo com todo o prefixo (as bombas entram de verdade no `ResolverTurno` do estado auxiliar) e confere que o bot não pisa em chamas até a última bomba. Por fim, procura uma continuação segura a partir daí, com a mesma busca de hoje. Como a zona de perigo final já é a explosão de **todas** as bombas que sobram (marco 3, D3), bombas que atravessam o turno e pilhas na mesma casa já são levadas em conta (decisão 4).

As regras do jogo e o pacote `jogo` não mudam.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/internal/bots/aleatorio/aleatorio.go` | `VersaoV3`, `NovoV3`, campo `variasBombas`; `Versao()` e `Planejar` despacham para o modo do v3 |
| `backend/internal/bots/aleatorio/bombas.go` | `planejarComBombas`, `tentarPlantarVarias`, `sortearEtapas` (D2 a D5) |
| `backend/internal/bots/aleatorio/bombas_test.go` | CA-07 a CA-11 |
| `backend/internal/bots/aleatorio/aleatorio_test.go` | `Versao()` do v3; tempo de planejamento com o v3 (CA-01, CA-06) |
| `backend/internal/bots/aleatorio/antecipacao_test.go` | v3 igual ao v2 com uma bomba por turno e dentro da janela (CA-04, CA-05) |
| testes herdados (`fuga_test.go`, `bomba_test.go`, `aleatoriedade_test.go`, `partida_test.go`, `antecipacao_test.go`...) | rodam também com `NovoV3`, por uma tabela de construtores (D6) (CA-02) |
| `backend/internal/bots/aleatorio/comparacao_test.go` | v3 contra v2 no mapa de exemplo, só `t.Log` (CA-12) |
| `backend/internal/bots/catalogo.go` (+ teste) | registra `aleatorio-v3` (CA-03) |

## Tipos e assinaturas

```go
// aleatorio.go
const VersaoV3 = "aleatorio-v3"

type Bot struct {
	semente      uint64
	antecipa     bool // v2 e v3
	variasBombas bool // v3: até bombas_por_turno bombas por turno (spec 15)
}

// NovoV3 cria o aleatorio-v3 com a semente dada.
func NovoV3(semente uint64) *Bot

// bombas.go

// planejarComBombas sorteia n em 1..min(bombas_por_turno, acoes_por_turno)
// e tenta planos seguros com n, n-1, ..., 2 bombas; com 1 usa tentarPlantar
// (decisão 2, D3).
func planejarComBombas(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas int) ([]jogo.Acao, bool)

// tentarPlantarVarias sorteia até tentativasComBomba planos com exatamente n
// bombas e devolve o primeiro seguro (D4).
func tentarPlantarVarias(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas, n int) ([]jogo.Acao, bool)

// sortearEtapas devolve n etapas distintas em 1..acoes, em ordem crescente.
func sortearEtapas(rng *rand.Rand, acoes, n int) []int
```

## Decisões

- **D1**: O v3 é o mesmo `Bot` com `antecipa: true` e `variasBombas: true`. O despacho fica em `Planejar`: só fora da janela, no turno em que o sorteio de hoje decide plantar, e só com `bombas_por_turno` ≥ 2 é que o v3 entra em `planejarComBombas`; em qualquer outro caso ele segue o caminho do v2. **Motivo**: decisões 1 e 3, e CA-04/CA-05. Com `bombas_por_turno` 1, o v3 não faz nenhuma chamada a mais ao gerador, e os planos saem idênticos aos do v2.
- **D2**: n é sorteado com `1 + rng.IntN(min(bombas_por_turno, acoes_por_turno))`. Se n for 1, a chamada vai direto para o `tentarPlantar` de hoje. **Motivo**: decisão 2 (mesma chance para cada valor). Não dá para plantar mais bombas do que ações.
- **D3**: A queda de n para n − 1 só gasta tentativas novas; o gerador continua o mesmo (não é recriado). Ao chegar a 1, `tentarPlantar` roda como no v2, e, se ele falhar, o bot foge sem plantar, como hoje. **Motivo**: decisão 2 e CA-11.
- **D4**: Cada tentativa de `tentarPlantarVarias`:
  1. sorteia as etapas `k1 < … < kn` (`sortearEtapas`);
  2. monta o prefixo até `kn`: passeio aleatório (`busca.passeio`, que ignora chamas) entre as plantas e `PLANTAR` nas etapas sorteadas, ficando parado na etapa da planta;
  3. prevê a linha com o prefixo inteiro (`preverLinha` com o próprio jogador executando o prefixo), o que põe todas as bombas no tabuleiro e calcula pilhas e reações (BOM-03, BOM-04, BOM-09);
  4. confere que, em cada etapa de 1 a `kn`, a casa do bot não está em chamas;
  5. procura com `buscaSegura` a continuação de `kn + 1` até o fim.

  Fica com a primeira tentativa que passar. Como as bombas não bloqueiam movimento (MOV-04), o passeio pode passar por cima da primeira bomba, e a previsão decide se isso é seguro. **Motivo**: generaliza o `tentarPlantar` (marco 3, plano, passo 4) sem mudar a busca.
- **D5**: O número de tentativas por n é o mesmo `tentativasComBomba` (8). No pior caso o v3 faz 8 × (n − 1) previsões a mais que o v2 por turno. Com n ≤ 2 no mapa de exemplo, são 8 previsões a mais. **Motivo**: CA-06 (tempo). Se o tempo apertar, o número cai para o v3 sem mudar o v2.
- **D6**: Os testes herdados recebem uma tabela `versoesQueHerdam = []func(uint64) *Bot{NovoV2, NovoV3}` (ou `Novo, NovoV2, NovoV3`, onde o teste já rodava com o v1) e rodam por versão com `t.Run`. Nos que contam bombas, "no máximo 1" vira "no máximo `bombas_por_turno`". **Motivo**: CA-02 sem duplicar testes.
- **D7**: O CA-10 usa um tabuleiro aberto de 7 × 7, com o v3 num canto e um bot de teste que só espera no canto oposto, `bombas_por_turno` 2, `pavio_padrao` 3 e `acoes_por_turno` 5. Joga até o `limite_turnos` com `Validar` + `ResolverTurno` e conta as bombas que sobram ao fim de cada turno com `jogador_id` do v3. **Motivo**: com pavio 3 e 5 ações, plantar na etapa 3 ou depois atravessa o turno, então o caso "bomba que atravessa" aparece naturalmente.
- **D8**: O CA-12 reaproveita `compararEmParalelo` do marco 13 e soma, por versão, os jogadores-turno com 0, 1 e 2 `PLANTAR` executados (a partir dos relatórios). Só registra; não falha. **Motivo**: decisão 5.

## Riscos

- **Tempo de planejamento**: tentativas com 2 bombas que falham caem para 1, o que custa até 16 previsões num turno ruim. Mitigação: o CA-06 mede; se passar do limite, reduzir as tentativas com n ≥ 2 (D5).
- **Taxa baixa de planos com 2 bombas**: o passeio sorteado ignora chamas, e muitas tentativas podem morrer na própria bomba. O CA-08 exige só um plano com duas bombas em 50 sementes, num tabuleiro aberto; a taxa real no mapa de exemplo aparece no CA-12.
- **Mais suicídio por cadeia**: duas bombas aumentam a área de perigo e as reações em cadeia. A previsão já cobre isso com os adversários parados; o risco que sobra é o mesmo do v1 (adversários que se movem), e o CA-12 mostra o efeito.
