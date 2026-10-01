# Marco 16: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

O v4 é mais um modo do mesmo `Bot` do pacote `internal/bots/aleatorio`, com `antecipa`, `variasBombas` e o novo `preveAmeaca`. Uma função pura, `calcularAmeaca`, monta a **zona de ameaça** a partir dos adversários vivos: chamas por etapa e ameaça final.

A ameaça entra pela previsão. `preverLinha` ganha o parâmetro `a *ameaca` e, quando ele não é `nil`, soma a ameaça da etapa t às chamas da etapa t e a ameaça final à zona de perigo final. Com isso, toda a lógica do v3 (fuga, bombas, janela) passa a procurar **planos protegidos** sem mudar nada além da linha do tempo que ela consulta. Os v1, v2 e v3 passam `nil` e continuam exatamente iguais.

O `Planejar` do v4 tem duas passadas:
1. Planeja como o v3, mas com a ameaça na previsão. Se o plano obtido for protegido (conferido por uma previsão independente, D5), devolve esse plano, desde que, dentro da janela, ele não termine num anel pior que o plano do v3 (D6).
2. Se não for protegido, devolve o plano do v3, calculado com um gerador novo da mesma semente. Ele é idêntico ao do v3 (CA-11: nunca pior que o v3).

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/internal/bots/aleatorio/ameaca.go` | tipo `ameaca`, `calcularAmeaca`, `vazia`, `alcanceDaBomba` (CA-06 a CA-09) |
| `backend/internal/bots/aleatorio/ameaca_test.go` | CA-04, CA-06 a CA-12 |
| `backend/internal/bots/aleatorio/previsao.go` | `preverLinha(..., a *ameaca)` soma a ameaça (D2) |
| `backend/internal/bots/aleatorio/aleatorio.go` | `VersaoV4`, `NovoV4`, campo `preveAmeaca`; corpo de `Planejar` vira `planejarCom(rng, estado, eu, etapas, a)`; duas passadas do v4 (D4 a D6) |
| `antecipacao.go`, `bombas.go`, `aleatorio.go` | `tentarPlantar`, `tentarPlantarVarias`, `planejarComBombas`, `planejarNaJanela`, `tentarAbrir` recebem `a *ameaca` e o repassam à `preverLinha` (D3) |
| `backend/internal/bots/aleatorio/tabuleiro_test.go` | v4 na tabela `versoesDeTeste` (CA-02, CA-05) |
| `backend/internal/bots/aleatorio/aleatorio_test.go` | `Versao()` do v4 (CA-01) |
| `backend/internal/bots/aleatorio/comparacao_test.go` | `causaDaMorte` (a classificação do diagnóstico) e v4 contra v3 com as metas (CA-13 a CA-15) |
| `backend/internal/bots/catalogo.go` (+ teste) | registra `aleatorio-v4` (CA-03) |

## Tipos e assinaturas

```go
// aleatorio.go
const VersaoV4 = "aleatorio-v4"

type Bot struct {
	semente      uint64
	antecipa     bool // v2, v3, v4
	variasBombas bool // v3, v4
	preveAmeaca  bool // v4: evita a zona de ameaça (spec 16)
}

func NovoV4(semente uint64) *Bot

// planejarCom é o corpo do Planejar do v1 ao v3, com a ameaça opcional na
// previsão (nil = sem ameaça). Devolve os passos de todas as etapas.
func (b *Bot) planejarCom(rng *rand.Rand, estado jogo.Estado, eu jogo.Jogador, etapas int, a *ameaca) []jogo.Acao

// ameaca.go

// ameaca é a zona de ameaça das bombas possíveis dos adversários (spec 16).
type ameaca struct {
	chamas []map[jogo.Posicao]bool // índice = etapa (1..etapas); [0] vazio
	final  map[jogo.Posicao]bool   // bombas possíveis que explodem depois do turno
}

// calcularAmeaca monta a ameaça dos adversários vivos de `eu`, com
// bombas_por_turno ≥ 1, para um turno de `etapas` etapas.
func calcularAmeaca(estado jogo.Estado, eu jogo.Jogador, etapas int) ameaca

// vazia informa se nenhuma casa está ameaçada.
func (a ameaca) vazia() bool

// alcanceDaBomba devolve as casas das chamas de uma bomba em `origem` com a potência
// dada, parando nos blocos (BOM-06, BOM-07), incluindo a casa do bloco.
func alcanceDaBomba(bloq map[jogo.Posicao]bool, c jogo.Config, origem jogo.Posicao, potencia int) []jogo.Posicao

// previsao.go
func preverLinha(estado jogo.Estado, etapas int, eu *jogo.Jogador, acoes []jogo.Acao, a *ameaca) linha

// protegido informa se o plano completo de `eu` é protegido: previsão com a
// ameaça, sem chamas na casa de cada etapa e fora da zona de perigo final.
func protegido(estado jogo.Estado, eu jogo.Jogador, etapas int, plano []jogo.Acao, a ameaca) bool
```

## Decisões

- **D1** (decisão 2): `calcularAmeaca` faz, para cada adversário: BFS de distâncias a partir da posição dele por casas livres (`bloqueios(estado)`: blocos fixos e destrutíveis do início do turno). Para cada etapa de plantio p de 1 a `acoes_por_turno` dele, cada casa a distância ≤ p − 1 é origem de uma bomba possível que explode em t = p + `pavio_padrao`. As chamas são a `alcanceDaBomba` com a `potencia` dele. Se t ≤ etapas, vão para `chamas[t]`; senão, para `final`. Sem pilhas e sem reações em cadeia. **Motivo**: é a definição da spec, calculada em O(adversários × casas × potência), bem abaixo do custo de uma previsão.
- **D2**: `preverLinha` com `a != nil` soma `a.chamas[t]` a `l.chamas[t]` e `a.final` a `l.perigo`. Não muda a simulação, só a linha devolvida. **Motivo**: toda a busca (`ir`, `buscaSegura`, `viveAte`, `sobrevive`, a janela) já consulta a linha, então os planos achados ficam protegidos sem mudar a busca.
- **D3**: As funções de planejamento ganham o parâmetro `a *ameaca`, repassado a cada chamada de `preverLinha`. O v1 ao v3 passam `nil`. **Motivo**: troca mecânica, sem estado global (o bot roda em paralelo com outros) e sem mexer na ordem do gerador. Os testes de igualdade v1 = v2 e v2 = v3 seguem valendo e servem de guarda.
- **D4**: No `Planejar` do v4, se `calcularAmeaca` é vazia, a chamada vai direto para `planejarCom(…, nil)`, idêntica ao v3 (CA-04). Se não é vazia, primeira passada com `&a`. **Motivo**: CA-04 sem custo extra e sem depender de a busca não pisar em casas ameaçadas.
- **D5**: O resultado da primeira passada é aceito só se `protegido(...)` confirmar, com uma previsão independente do plano completo. **Motivo**: os itens de último recurso do v3 (sobrevivente, "o que sobrevive mais etapas") também rodam com a ameaça na linha e podem devolver um plano não protegido; a conferência separa os dois casos sem rotular cada caminho do código.
- **D6** (decisão 1): dentro da janela de antecipação, o plano protegido só é aceito se terminar num anel ≥ `min(anel do plano do v3, anel alvo)`. Por isso, na janela, o plano do v3 é calculado antes (gerador próprio) para a comparação. **Motivo**: "o anel alvo continua sendo o objetivo": o v4 não troca chegar ao anel alvo por ficar protegido mais para fora.
- **D7**: Na segunda passada (fallback), o gerador é recriado com `b.gerador(turno, jogador)`, como o v3 faria. **Motivo**: o plano devolvido é exatamente o do v3 (CA-11, e serve de teste).
- **D8**: `causaDaMorte` vira helper de teste em `comparacao_test.go`, com as categorias do diagnóstico (própria, adversário com bomba já no tabuleiro, adversário com bomba do mesmo turno, própria + adversário, fechamento). Para cada morte, olha as explosões da etapa que alcançam a casa e o dono das bombas nas origens (as bombas do fim da etapa anterior). **Motivo**: CA-13 e CA-15 medem exatamente a tabela do objetivo.

## Revisões R1 e R2 (aprovadas em 2026-10-01)

- **D9** (R1, CA-16): `ameaca` guarda o peso: `chamas []map[jogo.Posicao]int` e `final map[jogo.Posicao]int`, somando 1 por bomba possível. `nivel(w) ameaca` devolve a ameaça só com as casas de peso ≥ w. `preverLinha` e `protegido` passam a receber a ameaça já filtrada pelo nível.
- **D10** (R1, CA-17): `planejarV4` percorre os níveis 1, 2, 3, 5, 8, 13… (Fibonacci) até passar do maior peso. Em cada nível, faz a passada com a ameaça daquele nível e aceita o plano se ele for protegido naquele nível (e respeitar o D6 na janela). Se nenhum nível servir, devolve o plano do v3 (D7). Cada nível recria o gerador, para o resultado não depender dos níveis anteriores.
- **D11** (R2, CA-18): `mira(estado, eu, plano) float64` em `mira.go`. Para cada `PLANTAR` do plano na etapa p, na casa onde o bot está naquela etapa (do passeio do plano), explosão em t = p + `pavio_padrao`. As chamas são `alcanceDaBomba(bloqueios, casa, potencia)`. Para cada adversário vivo, as casas possíveis são as de distância ≤ `min(t, acoes)` (≤ `acoes` se t passa do turno), e a fração coberta é |chamas ∩ casas| / |casas|. A mira é a soma, por adversário, da maior fração entre as bombas.
- **D12** (R2, CA-19, CA-20): no v4, `planejarCom` recebe também `mirar bool`. Com mira, a parte das bombas (antes do sorteio de 50 %) chama `melhorPlanoComMira`: gera até 16 planos com bomba (os mesmos geradores de `tentarPlantar`/`tentarPlantarVarias`, sem parar no primeiro), mantém os seguros na linha com a ameaça do nível e escolhe o de maior mira (desempate: o primeiro gerado). Se a mira for > 0, devolve esse plano; senão, segue o caminho de hoje (sorteio de 50 % e plantio a esmo). Como a ameaça já está na linha, todo plano com bomba aceito é protegido no nível (CA-20).
- **D13** (tempo, CA-05): o custo extra é de até 16 previsões por nível tentado. Se o CA-05 apertar, a mira roda só no nível em que o plano foi aceito, e os níveis param em 8.
- **D14** (CA-13 a CA-15): o teste de comparação continua igual. Se as metas ainda não forem atingidas, paro de novo e volto com os números.

## Riscos

- **Tempo**: com ameaça não vazia, o v4 faz até duas planificações (protegida e v3) e, na janela, três. No pior caso é cerca de 3 vezes o custo do v3. O CA-05 mede; a margem do v3 hoje é desconhecida. Mitigação, se necessário: só calcular o plano do v3 na janela quando o plano protegido terminar abaixo do anel alvo.
- **Ameaça grande demais** (decisão 4): perto de adversários com 7 ações, pode quase nunca haver plano protegido, e o v4 vira o v3. As metas do CA-13 e do CA-14 mostram isso; se ficarem longe, volto com os números e a proposta de ameaça graduada.
- **Pavio curto e adversário colado**: a ameaça pode cobrir todas as casas vizinhas, e o v4 cai para o v3. É o comportamento esperado.
