# Marco 13: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

O v2 fica no mesmo pacote do v1, `internal/bots/aleatorio`, como uma segunda versão do mesmo `Bot`. Só muda o caminho de `Planejar` dentro da janela de antecipação. Fora dela, o código executado é exatamente o do v1, inclusive no consumo do gerador aleatório. Por isso os planos são idênticos (CA-04, CA-06) sem duplicar nada.

Dentro da janela, a cada chamada o v2 faz assim:

1. Calcula o **anel alvo** e a **região** do jogador. A região é feita por busca em largura nas casas livres do início do turno, sem blocos fixos nem destrutíveis.
2. Escolhe os **destinos**:
   - se a região tem casa no anel alvo, são essas casas;
   - se não tem, o bot está **preso**. Uma busca 0-1 acha o **bloco de abertura**, em que cada bloco destrutível atravessado custa 1 e cada casa livre custa 0. Os destinos passam a ser as **casas de plantio**: casas da região de onde uma bomba do jogador alcança esse bloco.
3. Calcula, também por busca em largura nas casas livres, a distância de cada casa até os destinos.
4. Escolhe o plano pela ordem da decisão 3 da spec. Para isso, reaproveita a mesma busca em profundidade do v1, trocando só o teste da casa final (o **objetivo**):
   1. se não está preso: seguro e no anel alvo;
   2. se está preso: bomba de abertura. O bot vai pelo caminho mais curto até uma casa de plantio, planta e procura continuação segura, como o `tentarPlantar` do v1, só que com prefixo calculado em vez de sorteado;
   3. seguro, com a menor distância possível aos destinos. Tenta distância 0, 1, 2… até a distância atual. Sem destinos alcançáveis, aceita qualquer plano seguro;
   4. sobrevivente, terminando fora das casas que fecham;
   5. sobrevivente;
   6. o que sobrevive mais etapas.

A zona de perigo final já contém as casas que fecham (marco 12). Por isso "seguro" já exclui terminar nelas.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/internal/bots/aleatorio/aleatorio.go` | `VersaoV2`, `NovoV2`, campo `antecipa` em `Bot`; `Versao()` devolve a versão certa; `Planejar` desvia para `planejarNaJanela` quando `antecipa` e `naJanela` |
| `backend/internal/bots/aleatorio/busca.go` | `modoBusca` vira um `objetivo func(jogo.Posicao) bool` na casa final; `buscaSegura` e `buscaSobrevivente` passam a ser objetivos prontos. Sem mudança de comportamento no v1 |
| `backend/internal/bots/aleatorio/fechamento.go` | geometria do fechamento para o bot: `naJanela`, `anelAlvo`, `regiao`, `distancias`, `blocoDeAbertura`, `casasDePlantio` |
| `backend/internal/bots/aleatorio/antecipacao.go` | `planejarNaJanela` (ordem da decisão 3) e `tentarAbrir` (bomba de abertura) |
| `backend/internal/bots/catalogo.go` | `aleatorio.VersaoV2` no catálogo padrão |
| `backend/internal/bots/catalogo_test.go` | CA-03 |
| `backend/internal/bots/aleatorio/tabuleiro_test.go` | `versoesDeTeste` e `paraCadaVersao`; `comFechamento`; `morteNoFechamento` e `tinhaSaida` (classificação de mortes do CA-14) |
| `backend/internal/bots/aleatorio/*_test.go` existentes | passam a rodar para as duas versões por `paraCadaVersao` (CA-02); `aleatorio_test.go` ganha CA-01 do v2 e CA-05 dentro da janela |
| `backend/internal/bots/aleatorio/fechamento_test.go` | testes de tabela de `anelAlvo`, `naJanela`, `regiao`, `blocoDeAbertura`, `casasDePlantio` |
| `backend/internal/bots/aleatorio/antecipacao_test.go` | CA-04, CA-06 a CA-13 |
| `backend/internal/bots/aleatorio/partida_test.go` | o CA-06 do marco 3 passa a rodar também com só v2 (CA-02) |
| `backend/internal/bots/aleatorio/comparacao_test.go` | `jogarComparacao` com um bot por posição, classificação das mortes e as comparações do CA-14 ao CA-16 (registrado na implementação: arquivo próprio em vez de mudar `jogarPartida`) |

## Tipos e assinaturas

```go
package aleatorio

const (
	Versao   = "aleatorio-v1" // sem mudança
	VersaoV2 = "aleatorio-v2"
)

type Bot struct {
	semente  uint64
	antecipa bool // v2: prepara-se para o fechamento (specs/13-aleatorio-v2)
}

func Novo(semente uint64) *Bot   // v1, sem mudança
func NovoV2(semente uint64) *Bot // v2
func (b *Bot) Versao() string    // Versao ou VersaoV2
func (b *Bot) Planejar(estado jogo.Estado, jogadorID string) []jogo.Acao
```

Internos, para orientar a implementação:

```go
// fechamento.go
const janelaAntecipacao = 3 // turnos antes de turno_fechamento (decisão 1)
const margemAlvo = 2        // anéis além do que fecha no turno seguinte (R1)

func naJanela(c jogo.Config, turno int) bool         // f ≥ 1 e f − 3 ≤ turno < f + primeiro anel que nunca fecha (R2)
func primeiroAnelQueNuncaFecha(c jogo.Config) int    // FEC-08
func anelAlvo(c jogo.Config, turno int) int          // min(max(0, turno − f + 1) + 2, primeiro anel que nunca fecha, maior anel)
func regiao(bloqueadas map[jogo.Posicao]bool, c jogo.Config, de jogo.Posicao) map[jogo.Posicao]bool
func distancias(bloqueadas map[jogo.Posicao]bool, c jogo.Config, destinos []jogo.Posicao) map[jogo.Posicao]int
// blocoDeAbertura: busca 0-1 do jogador até o anel alvo; destrutíveis que fecham neste turno contam como fixos (FEC-05).
func blocoDeAbertura(estado jogo.Estado, de jogo.Posicao, alvo int) (jogo.Posicao, bool)
// casasDePlantio: casas da região em linha reta com o bloco, a até `potencia` casas, sem bloco nem bomba no meio (BOM-06, BOM-07).
func casasDePlantio(estado jogo.Estado, reg map[jogo.Posicao]bool, bloco jogo.Posicao, potencia int) []jogo.Posicao

// busca.go
type objetivo func(l linha, casa jogo.Posicao) bool
var buscaSegura, buscaSobrevivente objetivo
func novaBusca(rng *rand.Rand, estado jogo.Estado, l linha, acoes, etapas int, obj objetivo) *busca

// antecipacao.go
func planejarNaJanela(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas int) []jogo.Acao
func tentarAbrir(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas int, plantio []jogo.Posicao, dist map[jogo.Posicao]int) ([]jogo.Acao, bool)
```

## Decisões

- **D1**: O v2 fica no pacote `aleatorio`, como `NovoV2`, e não num pacote `aleatorio2`. **Motivo**: `docs/BOTS.md` pede que o bot importe só `internal/jogo`, então um pacote novo não poderia reusar o v1. Copiar o v1 abriria espaço para os dois divergirem fora da janela, e o mesmo pacote garante o CA-04 e o CA-06 por construção.
- **D2**: Fora da janela, `Planejar` do v2 executa o mesmo código do v1, na mesma ordem de chamadas ao gerador. Dentro da janela, o sorteio de bomba do v1 (`rng.IntN(2)`) não acontece (decisão 6). **Motivo**: o plano depende do consumo do gerador; qualquer chamada extra mudaria os planos e quebraria o CA-06.
- **D3**: A busca do v1 troca `modoBusca` por um objetivo na casa final, e `buscaSegura` e `buscaSobrevivente` passam a ser dois objetivos prontos. **Motivo**: cada item da decisão 3 é a mesma busca com outro objetivo. A ordem dos passos e o uso do gerador não mudam, então o v1 continua igual (os testes do marco 3 provam isso).
- **D4**: "Mais perto" (item 3) é a distância por casas livres até os destinos, calculada uma vez por chamada. O plano sai da primeira busca bem-sucedida com objetivo `seguro && dist ≤ d`, para d = 0, 1, … até a distância da casa atual. Depois disso, a busca aceita qualquer plano seguro. **Motivo**: reaproveita a busca sem um modo novo de otimização. São no máximo ~25 buscas de até `largura × altura × etapas` nós, bem abaixo do limite do CA-05.
- **D5**: A região e as distâncias consideram como bloqueados os blocos fixos e destrutíveis do início do turno e ignoram bombas. **Motivo**: é o mesmo critério de movimento da busca do v1 (D5 do marco 3). Bombas não impedem movimento (MOV-04), e o perigo delas já é tratado pela busca.
- **D6**: O bloco de abertura sai de uma busca 0-1 com custo (blocos atravessados, passos), sem sorteio e com desempate pela ordem fixa das direções. Blocos destrutíveis em casas que fecham neste turno contam como fixos. **Motivo**: determinismo sem gastar o gerador. Um bloco que vira fixo ao fim do turno não abre caminho (FEC-05).
- **D7**: A bomba de abertura usa o caminho mais curto até a casa de plantio, planta na etapa seguinte à chegada e só é aceita se o prefixo sobrevive e há continuação segura. A verificação é a mesma do `tentarPlantar` do v1, com `preverLinha` incluindo a própria bomba. As casas de plantio são tentadas da mais perto para a mais longe, com empates embaralhados pelo gerador, até `tentativasComBomba` (8). **Motivo**: a segurança fica igual à das bombas do v1 (CA-21 do marco 3), e o custo por chamada fica limitado.
- **D8**: A bomba de abertura não exige que o bloco seja destruído no mesmo turno. Basta a explosão alcançá-lo, mesmo que o pavio passe para o turno seguinte. **Motivo**: o CA-11 pede só que a explosão alcance o bloco; com pavio 3 e 7 etapas, quase sempre explode no turno.
- **D9**: O item 4 (sobrevivente fora das casas que fecham) só roda quando há casas que fecham neste turno. **Motivo**: sem elas o item 4 seria igual ao 5, e a busca extra gastaria o gerador à toa.
- **D10**: Os testes do v1 passam a rodar para as duas versões com `paraCadaVersao(t, func(t *testing.T, novo func(uint64) *Bot))`, em subtestes com o nome da versão. **Motivo**: cumpre o CA-02 com uma mudança mecânica e deixa explícito, no nome do teste, qual versão falhou.
- **D11**: Os testes de comparação (CA-14 a CA-16) jogam as configurações 4×v1, 4×v2 e 2×v1 + 2×v2 nas sementes 1 a 200, em subtestes paralelos. A base do v1 é medida no próprio teste, não fixada no código. Morte por fechamento é a de um jogador fora das `chamas` da etapa e numa casa de `blocos_fechados`. **Motivo**: o limite de 25% fica relativo e não quebra se o v1 mudar de forma legítima, e a classificação de morte não depende de detalhes internos.
- **D12**: No confronto 2 contra 2, as sementes ímpares usam a ordem v1, v2, v1, v2 nas posições iniciais e as pares, v2, v1, v2, v1. **Motivo**: tira a vantagem de posição do resultado (CA-15).

- **D13** (R2): a janela termina no turno em que fecha o último anel que pode fechar (FEC-08), e o anel alvo nunca passa do primeiro anel que nunca fecha. **Motivo**: dentro da área mínima não há mais o que antecipar; voltar ao código do v1 devolve as bombas aleatórias e evita empates por passividade.
- **D14**: o teste do CA-05 tira o estado "dentro da janela, com jogador preso" de uma partida real com 4 v2, em vez de montá-lo a partir das posições iniciais. **Motivo**: no início da partida ninguém está preso no mapa de exemplo.

## Riscos

- **Tempo dos testes**: as comparações somam 600 partidas, e o v2 faz mais buscas dentro da janela. Estimativa: alguns segundos com subtestes paralelos. Se passar de ~10 s, os CA-14 a CA-16 rodam com 200 sementes só fora de `testing.Short()` e com 50 dentro.
- **Limites da decisão 4 não atingidos**: se o v2 ficar acima de 25% das mortes do v1 ou abaixo do dobro de vitórias, a causa provável é o caminho até o centro passar por mais de um bloco de abertura, ou a abertura atrasar demais. Nesse caso os números voltam para revisão com você, como decidido, em vez de o teste ser afrouxado.
- **Empates**: podem subir com 4×v2 (decisão 5). O teste só registra.
- **CA-13 depende do tabuleiro**: o tabuleiro tem que fazer o v1 morrer em pelo menos uma semente e o v2 sobreviver em todas. Ele é desenhado à mão na implementação e, se não existir um tabuleiro pequeno com essa propriedade, a spec volta para revisão.
