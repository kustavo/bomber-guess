# Marco 16: bot aleatorio-v4 (prevê bombas dos adversários)

**Status**: concluída (com as revisões R1 a R3, 2026-10-01)
**Roadmap**: `docs/ROADMAP.md`, marco 16
**Documentos de referência**: `docs/BOTS.md`, `docs/REGRAS.md` (bombas e ordem da etapa); `specs/03-bot-simples/spec.md`, `specs/13-aleatorio-v2/spec.md` e `specs/15-aleatorio-v3/spec.md` (as versões que o v4 estende)

## Objetivo

Entregar `aleatorio-v4`, um v3 que não morre tanto por bombas que não viu. Os bots aleatórios preveem o turno supondo que os adversários **ficam parados** (marco 3), então nunca enxergam as bombas que os outros plantam no mesmo turno. Medido em 200 partidas com 4 `aleatorio-v3` no mapa de exemplo (8 792 turnos), as mortes foram:

| Causa | Mortes |
|---|---|
| bomba de adversário plantada no mesmo turno | 349 |
| fechamento do tabuleiro | 103 |
| própria bomba junto com a de um adversário | 45 |
| bomba de adversário que já estava no tabuleiro | 22 |
| só a própria bomba | 20 |

O v4 calcula uma **zona de ameaça**: em cada etapa, as casas que as bombas que um adversário ainda pode plantar neste turno conseguem alcançar. Ele prefere planos que fiquem fora dela, sem abrir mão de nada do que o v3 já garante.

v1, v2 e v3 não mudam.

### Termos usados nesta spec

Valem os termos dos marcos 3, 13 e 15 (**simulação do turno**, **plano seguro**, **plano sobrevivente**, **janela de antecipação**...). Além deles:

- **Alcance de um adversário na etapa s**: as casas que ele pode ocupar ao fim da etapa s, andando por casas livres a partir da posição do início do turno, sem atravessar blocos fixos nem destrutíveis, com no máximo `min(s, acoes_por_turno)` passos.
- **Bomba possível**: uma bomba que um adversário vivo, com `bombas_por_turno` ≥ 1, poderia plantar na etapa p (1 ≤ p ≤ `acoes_por_turno` dele) numa casa do seu alcance na etapa p − 1. Pela BOM-05, ela explode na etapa p + `pavio_padrao` dele.
- **Ameaça da etapa t**: a união das chamas (BOM-06, BOM-07, com a `potencia` do adversário e os blocos do início do turno) das bombas possíveis que explodem na etapa t.
- **Ameaça final**: a união das chamas das bombas possíveis que explodem depois da última etapa do turno (atravessam o turno). É a parte da ameaça que vale para a casa onde o bot termina.
- **Plano protegido**: plano seguro (marco 3) em que, além disso, o bot não está em nenhuma casa da ameaça da etapa t ao fim de nenhuma etapa t, e não termina numa casa da ameaça final.

## Escopo

**Dentro**
- Versão `aleatorio-v4` no mesmo subpacote, registrada no catálogo.
- Cálculo da zona de ameaça (lógica pura, testável).
- Ordem de preferência dos planos: protegido > seguro > sobrevivente. Os planos com bomba do próprio bot também precisam ser protegidos para ganhar dos planos sem bomba protegidos (decisão 3).
- Tudo o que o v3 garante (CA-01 a CA-12 do marco 15, e por herança os dos marcos 3 e 13), exceto onde a decisão 3 muda a escolha.
- Comparação com o v3 no mapa de exemplo, com metas (decisão 6).

**Fora** (fica para outro marco)
- Andar com propósito (ir atrás dos adversários ou de áreas abertas). As bombas com mira entraram na revisão R2.
- Prever o que o adversário **vai** fazer (só o que ele **pode** fazer).
- Pilhas e reações em cadeia entre bombas possíveis (decisão 2).
- Mudanças em v1, v2, v3 e nas regras do jogo.

## Regras cobertas

- As do marco 15, que o v4 herda.
- `BOM-05`, `ORD-03`: etapa em que a bomba possível explode.
- `BOM-06`, `BOM-07`: alcance das chamas da bomba possível, parando nos blocos.
- `BOM-02`, `EST-04`: só adversário com `bombas_por_turno` ≥ 1 ameaça.
- `MOV-02`, `MOV-03`: alcance do adversário só por casas livres (bloco destrutível do início do turno bloqueia, decisão 2).
- `FIM-01`: adversário morto não ameaça.

## Critérios de aceitação

Cada critério é verificável por um teste automatizado. Salvo menção contrária, vale para as sementes 1 a 50, em tabuleiros pequenos montados à mão.

### Contrato e herança

- **CA-01** (BOT-01): `Versao()` devolve `"aleatorio-v4"`.
- **CA-02** (marco 15, CA-01 a CA-12): os testes herdados das versões anteriores (tabela de versões do pacote) rodam também com o v4 e passam. Exceção (R3): o caso "campo aberto, turno curto" do CA-21 do marco 3 não vale para o v4, porque ali todo plano com bomba fica num nível de proteção pior que o melhor plano sem bomba (decisão 3); para o v4, o mesmo caso confere que ele **não** planta.
- **CA-03** (API-09): o catálogo lista `aleatorio-v1` a `aleatorio-v4`, e todos podem ser criados por nome.
- **CA-04**: **Dado** um estado em que nenhum adversário vivo tem bomba possível que alcance casa alguma que o bot possa ocupar no turno (ex.: adversário a mais de `acoes_por_turno` + `potencia` casas, ou cercado por blocos fixos), **quando** `Planejar` é chamado no v3 e no v4 com a mesma semente, **então** os planos são idênticos.
- **CA-05** (BOT-02): no mapa de exemplo, cada chamada a `Planejar` do v4 termina em menos de 10% do `prazo_planejamento_ms`, nos mesmos estados do teste de tempo do marco 13.

### Zona de ameaça (lógica pura)

- **CA-06** (BOM-05, BOM-06, ORD-03): **Dado** um corredor livre com um adversário de `potencia` 1, `pavio_padrao` 3, `acoes_por_turno` 3, **quando** a ameaça é calculada para um turno de 7 etapas, **então**:
  - nas etapas 1 a 3, a ameaça é vazia (nenhuma bomba possível explode antes da etapa 4);
  - na etapa 4, é a cruz de potência 1 em torno da casa inicial do adversário;
  - na etapa 5, a união das cruzes em torno das casas a até 1 passo; na etapa 6, a até 2 passos;
  - na etapa 7 e na ameaça final, nenhuma casa nova (ele só tem 3 ações: a última bomba possível é na etapa 3 e explode na etapa 6).
- **CA-07** (BOM-07, MOV-02, MOV-03): **Dado** um adversário separado do bot por um bloco fixo ou destrutível, **quando** a ameaça é calculada, **então** as chamas param no bloco e o alcance do adversário não atravessa o bloco.
- **CA-08** (BOM-02, FIM-01): **Dado** um adversário morto, ou o próprio bot, **quando** a ameaça é calculada, **então** eles não contribuem para a ameaça.
- **CA-09** (BOM-11): **Dado** um adversário com `pavio_padrao` maior que as etapas que restam depois do plantio, **quando** a ameaça é calculada, **então** as chamas dessas bombas possíveis entram na ameaça final, e não em nenhuma etapa do turno.

### Preferência pelo plano protegido

- **CA-10**: **Dado** um tabuleiro em que existe plano protegido e em que o v3, na mesma semente, termina dentro da ameaça, **quando** `Planejar` é chamado no v4, **então** o plano é protegido em todas as sementes.
- **CA-11** (marco 3, CA-18): **Dado** um tabuleiro sem plano protegido mas com plano seguro, **quando** `Planejar` é chamado no v4, **então** o plano é seguro (o v4 nunca escolhe pior que o v3 por causa da ameaça).
- **CA-12**: **Dado** um tabuleiro em que só existe plano com bomba protegido e plano sem bomba protegido, **quando** o v4 decide plantar, **então** o plano com bomba devolvido é protegido; se nenhum plano com bomba for protegido, o v4 fica com o plano sem bomba protegido em vez de um plano com bomba só seguro (decisão 3).

### Ameaça graduada (revisão R1)

- **CA-16** (BOM-05, BOM-06): **Dado** o corredor do CA-06, **quando** a ameaça é calculada, **então** cada casa ameaçada numa etapa tem um **peso**: quantas bombas possíveis (adversário, etapa de plantio, casa de origem) a alcançam naquela etapa. Os pesos do corredor batem com a contagem feita à mão.
- **CA-17** (marco 3, CA-18): **Dado** um tabuleiro sem plano protegido, mas com plano seguro que só passa por casas de peso baixo, **quando** `Planejar` é chamado no v4, **então** o plano é seguro e é protegido no menor nível de peso possível (decisão 7): nenhum outro plano seguro do mesmo tabuleiro é protegido num nível menor.

### Bombas com mira (revisão R2)

- **CA-18**: **Dada** uma bomba do bot e um adversário, **quando** a **mira** é calculada, **então** ela é a fração das casas onde o adversário pode estar na etapa da explosão que as chamas alcançam (decisão 8). Casos testados: adversário fora do alcance (0); adversário num beco sem saída, todo coberto (1); adversário com 4 casas possíveis, 1 coberta (0,25); bomba que só explode no turno seguinte, contra as casas que ele pode ocupar ao fim do turno.
- **CA-19**: **Dado** um tabuleiro em que existe plano com bomba protegido que encurrala um adversário (mira 1), **quando** `Planejar` é chamado no v4, **então** em pelo menos 45 das sementes 1 a 50 o plano tem bomba com mira maior que 0.
- **CA-20** (BOM-02): **Dado** um tabuleiro em que o único plano com bomba de mira maior que 0 não é protegido, **quando** `Planejar` é chamado no v4, **então** o plano devolvido não tem essa bomba: a mira nunca vence a proteção (decisão 9).

### Comparação no mapa de exemplo (sementes 1 a 200)

- **CA-13**: **Dadas** partidas com 4 `aleatorio-v4`, **quando** são jogadas até o fim, **então** as mortes por bomba de adversário plantada no mesmo turno são no máximo 70% das medidas com 4 `aleatorio-v3` nas mesmas sementes.
- **CA-14**: **Dadas** partidas com 2 `aleatorio-v3` e 2 `aleatorio-v4`, alternando as posições entre as sementes, **quando** são jogadas até o fim, **então** o v4 vence pelo menos 1,5 vez o número de partidas do v3.
- **CA-15**: **Nas** partidas do CA-13 e do CA-14, o teste registra (`t.Log`) as mortes por causa (a tabela do objetivo), as vitórias e os empates.

## Decisões

1. **Nome e base.** Decisão: `aleatorio-v4`, com o v3 inteiro como base (antecipação do fechamento, bomba de abertura e várias bombas por turno). Dentro da janela de antecipação, a ameaça também vale: o anel alvo continua sendo o objetivo, mas, entre planos que chegam lá, prefere os protegidos.
2. **Modelo da ameaça.** Decisão: cada bomba possível isolada, com a `potencia` e o `pavio_padrao` do adversário, blocos do início do turno, sem pilhas (o adversário plantar duas na mesma casa) e sem reações em cadeia entre bombas possíveis ou com as bombas que já estão no tabuleiro. O alcance do adversário não atravessa bloco destrutível, mesmo que ele seja destruído no meio do turno. Fica mais simples, rápido e já cobre o caso que mais mata; os casos de fora entram se a medida mostrar que valem.
3. **Ordem de preferência.** Decisão: plano protegido > plano seguro > plano sobrevivente, valendo para os planos com e sem bomba. No turno em que decide plantar, o v4 só fica com um plano com bomba se ele for protegido, ou se não houver plano protegido nenhum (nesse caso, segue a ordem do v3). Assim a bomba nunca é a razão de ele terminar exposto.
4. **Ameaça muito grande.** Um adversário com 7 ações alcança muitas casas, e no mapa de exemplo pode ser que quase nunca haja plano protegido perto dele. Decisão: aceitar isso neste marco (o v4 cai para o comportamento do v3 quando não acha plano protegido) e medir. Se a medida mostrar que ele quase nunca acha plano protegido, a revisão seria uma ameaça graduada (preferir o plano que passa menos etapas na ameaça), proposta com números, como no marco 13.
5. **Os próprios planos dos adversários.** A ameaça usa só a posição do início do turno, os atributos e os blocos. Não usa o histórico nem a versão do bot adversário (o `Planejar` não recebe histórico, e usar a versão seria "colar" da implementação dele).
6. **Metas.** Decisão: mortes por bomba de adversário do mesmo turno ≤ 70% das do v3 (CA-13) e v4 vencendo pelo menos 1,5 vez o v3 no 2 contra 2 (CA-14). Se a implementação ficar longe disso, os números voltam para revisão, em vez de afrouxar o teste em silêncio.

## Revisões durante a implementação

Medido com o v4 da primeira versão (só a zona de ameaça), nas sementes 1 a 200:

| Medida | v3 | v4 | Meta |
|---|---|---|---|
| Mortes por bomba de adversário do mesmo turno (4 iguais) | 349 | 251 (72 %) | ≤ 70 % |
| Empates (4 iguais) | 76 | 165 | — |
| Vitórias no 2 contra 2 | 43 | 40 (117 empates) | v4 ≥ 1,5 × v3 |

O v4 achou plano protegido em 94 % das decisões (7 834 de 8 359). Das 251 mortes por bomba do mesmo turno que sobraram, 234 foram em casa que estava na ameaça: são os turnos sem plano protegido, em que o v4 caía no v3 e ignorava a ameaça. Pilhas, cadeias e blocos destruídos explicaram só 17. E sobreviver mais virou empate, não vitória.

- **R1** (aprovada em 2026-10-01): ameaça graduada. Cada casa ameaçada numa etapa ganha um peso (CA-16). Sem plano protegido, o v4 procura plano protegido só contra as casas de peso ≥ w, para w = 2, 3, 5, 8… até o maior peso, e fica com o primeiro nível que tiver plano (CA-17). Só se nenhum nível servir é que cai no v3.
- **R2** (aprovada em 2026-10-01): bombas com mira (CA-18 a CA-20). O CA-14 (vencer 1,5 vez o v3) continua como meta.

7. **Níveis da ameaça graduada (R1).** Decisão: níveis w = 1 (o protegido de antes), 2, 3, 5, 8, 13… até passar do maior peso. É uma sequência curta (poucas buscas a mais) e corta primeiro as casas que só uma bomba possível alcança.
8. **Mira (R2).** Decisão: para cada bomba do plano, plantada na etapa p, a explosão é na etapa t = p + `pavio_padrao` do bot (sem cadeias). As casas onde o adversário pode estar na etapa t são o alcance dele em `min(t, acoes_por_turno)` passos; se t passa do turno, o alcance em `acoes_por_turno` passos. A mira do plano é a soma, por adversário vivo, da maior fração coberta entre as bombas do plano. Um adversário encurralado vale 1.
9. **Quando e como plantar (R2).** Decisão: o v4 avalia planos com bomba **em todo turno** (não só na metade sorteada), com o dobro de tentativas do v3 (16), e fica com o de maior mira entre os protegidos. Se o melhor tiver mira maior que 0, planta. Se nenhum tiver mira, volta ao sorteio de hoje (metade dos turnos, plantando a esmo, o que ainda abre caminho entre blocos). A ordem de preferência continua: protegido > protegido em nível maior (R1) > seguro > sobrevivente, e a mira só desempata entre planos do mesmo nível.
- **R3** (aprovada em 2026-10-01): exceção no CA-02. No caso "campo aberto, turno curto" do CA-21 do marco 3 (bot com 3 ações, adversário com 7), todo plano com bomba ficou no nível 13 da ameaça graduada e o melhor sem bomba no nível 3; pela decisão 3, o v4 não planta ali. O teste herdado passa a valer só para v1 a v3, e para o v4 confere o contrário.

