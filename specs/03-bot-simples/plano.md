# Marco 03: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

O bot fica em `internal/bots/aleatorio` e usa só a API pública de `internal/jogo` e a biblioteca padrão. O pacote `jogo` não muda.

A ideia central: com os adversários parados, as explosões de um turno não dependem de onde o bot anda. Elas só dependem das bombas já no tabuleiro e da bomba que o bot planta, se plantar. Por isso o bot faz assim:

1. Calcula uma **linha do tempo**: as chamas de cada etapa e a **zona de perigo final** (as casas que as bombas que sobram ao fim do turno alcançam). Ela sai de `jogo.ResolverTurno` rodado em um estado auxiliar, sem os jogadores reais e com dois **fantasmas** fora do tabuleiro, que mantêm a partida em andamento.
2. Faz uma **busca em profundidade** sobre os pares (casa, etapa), sorteando a ordem dos passos (`ESPERAR` e as 4 direções). Ela encontra um caminho que evita as chamas de cada etapa e termina fora da zona de perigo final. São no máximo `largura × altura × etapas` estados (1365 no mapa de exemplo), então a busca é exaustiva: se existe um plano seguro só com movimentos, ela o encontra (CA-18). Como a ordem é sorteada, o caminho escolhido varia com a semente (decisão 6).
3. Se não há plano seguro, repete a busca exigindo só sobreviver (CA-19). Se nem isso existe, fica com o caminho que sobrevive mais etapas (CA-20).
4. Em cerca de metade dos turnos (decisão 5), antes do passo 2, tenta até 8 planos com bomba. Para cada um, sorteia a etapa `k` em que vai plantar e um prefixo aleatório de `k-1` movimentos válidos. Recalcula então a linha do tempo, desta vez com o bot no estado auxiliar executando prefixo + `PLANTAR`, e procura a continuação a partir da etapa `k+1` com a mesma busca, exigindo plano seguro. O primeiro plano encontrado vence. Se nenhum servir, o bot não planta (decisão 4).

O gerador de números aleatórios é criado a cada chamada a partir de (semente, turno, `jogadorID`) (decisão 1). Assim o bot não guarda estado entre chamadas (CA-09).

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/internal/bots/aleatorio/aleatorio.go` | `Bot`, `Novo`, `Versao`, `Planejar`: orquestra as etapas 1 a 4 e monta as ações numeradas; `tentarPlantar` (etapa 4). |
| `backend/internal/bots/aleatorio/previsao.go` | `preverLinha` (chamas por etapa + zona de perigo final) calculada com o estado auxiliar de fantasmas. |
| `backend/internal/bots/aleatorio/busca.go` | Busca em profundidade com ordem sorteada sobre (casa, etapa), nos três modos: seguro, sobrevivente, mais longo. |
| `backend/internal/bots/aleatorio/tabuleiro_test.go` | Helpers de teste: `montar` (desenho ASCII, como o de `internal/jogo`), `comBomba`, `simular` e `classificar` (seguro / sobrevivente / morto, pela definição da spec, com `jogo.ResolverTurno` e o estado real). |
| `backend/internal/bots/aleatorio/previsao_test.go` | Testes de `preverLinha` (chamas por etapa, pilha, cadeia, bloco, zona de perigo final). |
| `backend/internal/bots/aleatorio/aleatorio_test.go` | CA-01 a CA-05 (contrato). |
| `backend/internal/bots/aleatorio/bomba_test.go` | CA-21 (planos com bomba são seguros) e CA-22. |
| `backend/internal/bots/aleatorio/partida_test.go` | `jogarPartida(semente)` no mapa de exemplo; CA-06, CA-07, CA-12, CA-21. |
| `backend/internal/bots/aleatorio/aleatoriedade_test.go` | CA-09 a CA-11. |
| `backend/internal/bots/aleatorio/fuga_test.go` | CA-08, CA-13a a CA-20. |

## Tipos e assinaturas

```go
package aleatorio

// Versao é o identificador desta versão do bot.
const Versao = "aleatorio-v1"

// Bot anda aleatoriamente, planta bombas de vez em quando e foge das
// explosões que consegue prever (specs/03-bot-simples). Não guarda estado
// entre chamadas: o plano depende só da semente, do estado e do jogador.
type Bot struct {
    semente uint64
}

// Novo cria o bot com a semente dada (decisão 1).
func Novo(semente uint64) *Bot

// Versao devolve "aleatorio-v1".
func (b *Bot) Versao() string

// Planejar devolve exatamente acoes_por_turno ações numeradas 1, 2, 3…,
// ou nenhuma se o jogador não existe ou está morto. Não altera o estado.
func (b *Bot) Planejar(estado jogo.Estado, jogadorID string) []jogo.Acao

var _ jogo.Bot = (*Bot)(nil)
```

Internos (sem exportar), para orientar a implementação:

```go
// linha é a previsão de um turno com os adversários parados.
type linha struct {
    chamas []map[jogo.Posicao]bool // índice = etapa (1..etapas); [0] vazio
    perigo map[jogo.Posicao]bool   // zona de perigo final
}

// preverLinha roda ResolverTurno no estado auxiliar: blocos e bombas do
// estado, dois fantasmas em (-1,-1) com acoes_por_turno = etapas e, se
// eu != nil, o próprio jogador com as ações dadas.
func preverLinha(estado jogo.Estado, etapas int, eu *jogo.Jogador, acoes []jogo.Acao) linha

// busca é a busca em profundidade com ordem sorteada; ir procura passos da
// etapa dada até a última, partindo da casa, no modo buscaSegura ou
// buscaSobrevivente; sem sucesso, devolve o caminho que sobrevive mais etapas.
func (b *busca) ir(casa jogo.Posicao, etapa int) (passos []jogo.Acao, vivas int, ok bool)

// fugir encadeia os modos: seguro, senão sobrevivente, senão o mais longo.
func fugir(rng *rand.Rand, estado jogo.Estado, l linha, casa jogo.Posicao, de, acoes, etapas int) []jogo.Acao

// tentarPlantar sorteia até 8 planos com bomba e devolve o primeiro seguro.
func tentarPlantar(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas int) ([]jogo.Acao, bool)
```

## Decisões

- **D1**: A previsão usa `jogo.ResolverTurno` em um estado auxiliar, sem reimplementar as regras de explosão. **Motivo**: uma só fonte da verdade para pilhas, reação em cadeia, DEC-05 e blocos. O bot herda os testes do marco 2 e continua importando só `internal/jogo`, como pede `docs/BOTS.md`.
- **D2**: O estado auxiliar troca todos os jogadores por dois fantasmas vivos em `(-1,-1)`, fora do tabuleiro, com `turno` 1 e `limite_turnos` alto. **Motivo**: nenhuma chama alcança uma casa fora do tabuleiro, então os fantasmas nunca morrem e a simulação nunca termina antes da hora (DEC-06, DEC-07). Assim a linha do tempo fica completa mesmo que o bot ou um adversário morresse nela. As explosões não dependem dos jogadores, então tirá-los não muda as chamas.
- **D3**: A zona de perigo final vem de uma segunda chamada a `ResolverTurno`: o estado final da primeira, com todas as bombas com `pavio_restante` 1 e fantasmas com `acoes_por_turno` 1. **Motivo**: todas as bombas explodem na mesma etapa, então a união das chamas é exatamente o conjunto de casas que alguma bomba restante alcança, já com pilhas e blocos do fim do turno (decisão 3).
- **D4**: Com bomba, o bot entra no estado auxiliar com o próprio plano (prefixo + `PLANTAR`). **Motivo**: a linha do tempo fica exata, inclusive para a própria bomba servindo de obstáculo e reagindo em cadeia. Se o bot morrer no prefixo, a bomba não é plantada e a busca da continuação falha, o que descarta o candidato.
- **D5**: A busca não entra em casa com bloco destrutível do início do turno, mesmo que ele seja destruído antes. **Motivo**: a spec não aposta em blocos (CA-07, CA-08) e o validador nunca reclama de movimento para casa livre, nem dentro do tabuleiro nem fora de bloco fixo (CA-06).
- **D6**: A busca roda sobre `etapas = jogo.CalcularEtapas(estado.Jogadores)`. Nas etapas depois de `acoes_por_turno` do bot, o único passo possível é ficar parado. **Motivo**: o turno pode ser mais longo que o plano do bot (EST-03), e as chamas dessas etapas também o atingem.
- **D7**: O gerador é `rand.New(rand.NewPCG(semente, h))`, com `h` = FNV-1a de `jogadorID` e `turno`, de `math/rand/v2`. **Motivo**: biblioteca padrão, determinístico e independente de chamadas anteriores (CA-09). Jogadores diferentes com a mesma semente sorteiam coisas diferentes (CA-11).
- **D8**: O bot planta no máximo 1 bomba por turno e só se `bombas_por_turno` ≥ 1. **Motivo**: decisão 5; garante VAL-03 sem precisar contar.
- **D9**: Os helpers de teste (`montar`, `comBomba`) são copiados de `internal/jogo/tabuleiro_test.go`, em vez de virarem um pacote compartilhado. **Motivo**: os testes internos de `jogo` não podem importar um pacote que importa `jogo` (ciclo de importação), e são poucas linhas.
- **D10**: Os testes de partida (CA-06, CA-07, CA-12, CA-21) jogam as 50 sementes em subtestes paralelos, com os 4 jogadores do mapa de exemplo usando o mesmo bot `Novo(semente)`. **Motivo**: cobre o caso de bots iguais na mesma partida (decisão 1) e mantém o tempo do `go test` baixo.

## Riscos

- **Tempo dos testes de partida**: 50 sementes × até 50 turnos × 4 jogadores × até 18 chamadas a `ResolverTurno`. Medido na implementação: ~0,2 s com subtestes paralelos; o atalho de `testing.Short()` não foi necessário.
- **CA-12 depende da sorte**: com cerca de 50% de tentativas por turno e 50 sementes, é praticamente certo que haja `PLANTAR`. Se a fixação das sementes der azar, o teste falha de forma estável (e não aleatória), e basta revisar os candidatos com bomba.
- **CA-05 em máquina lenta ou CI**: o limite de 100 ms é folgado diante da estimativa (bem menos de 10 ms por chamada). Medir com a média de várias chamadas, não com uma só.
- **Ficar sempre parado**: se a busca sortear `ESPERAR` com a mesma chance dos movimentos, o bot anda pouco. Aceitável pela decisão 6, mas, se as partidas no terminal (marco 4) ficarem paradas demais, dá para pesar a ordem do sorteio sem mudar a spec.
