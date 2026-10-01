# Marco 15: bot aleatorio-v3 (mais de uma bomba por turno)

**Status**: concluída
**Roadmap**: `docs/ROADMAP.md`, marco 15
**Documentos de referência**: `docs/BOTS.md`, `docs/REGRAS.md` (seção das bombas); `specs/03-bot-simples/spec.md` (comportamento do `aleatorio-v1`) e `specs/13-aleatorio-v2/spec.md` (o v2, que o v3 estende)

## Objetivo

Entregar `aleatorio-v3`, igual ao `aleatorio-v2` mas sem o limite de **uma bomba por turno**, que a decisão 5 do marco 3 impôs ao v1 só por simplicidade (o v2 herdou). A regra do jogo permite até `bombas_por_turno` (BOM-02), e no mapa de exemplo esse limite é 2. Hoje nenhum bot do catálogo chega nele: na partida `teste-visual` (20 turnos, 4 bots), houve 42 turnos com uma bomba e nenhum com duas.

O v3 **pode** plantar mais de uma bomba quando houver plano seguro, inclusive bombas cujo pavio passa para o turno seguinte. Não é obrigado: quantas bombas plantar continua sendo sorteado, e na dúvida ele planta menos.

O `aleatorio-v1` e o `aleatorio-v2` não mudam.

### Termos usados nesta spec

Valem os termos do marco 3 (**simulação do turno**, **plano seguro**, **plano sobrevivente**) e do marco 13 (**janela de antecipação**, **preso**, **bloco de abertura**). Além deles:

- **Plano com n bombas**: plano com exatamente n ações `PLANTAR`, em etapas sorteadas.
- **Bomba que atravessa o turno**: bomba plantada tão tarde que o pavio não acaba neste turno (BOM-11). Ela explode no turno seguinte.

## Escopo

**Dentro**
- Versão `aleatorio-v3` no mesmo subpacote dos bots aleatórios, registrada no catálogo (`GET /bots`, terminal, editor).
- Tudo o que o v2 garante (CA-01 a CA-16 do marco 13, e por herança os do marco 3), agora também para o v3, exceto o limite de uma bomba por turno.
- Fora da janela de antecipação: nos turnos em que decide plantar, sortear quantas bombas tentar, de 1 a `bombas_por_turno`, e procurar um plano seguro com essa quantidade; se não achar, tentar com uma a menos, até zero.
- Bombas que atravessam o turno são permitidas, desde que o plano seja seguro (o jogador termina o turno fora do alcance de todas as bombas que sobram, como no marco 3).
- Testes de comparação com o v2 no mapa de exemplo, só registrados.

**Fora** (fica para outro marco)
- Mudanças no `aleatorio-v1`, no `aleatorio-v2` e nas regras do jogo.
- Usar bombas para atacar adversários: os bots aleatórios continuam sem mirar ninguém.
- Mais de uma bomba de abertura dentro da janela de antecipação (decisão 3).
- Ranking: marco 9.

## Regras cobertas

- `BOT-01` a `BOT-04`, `VAL-01` a `VAL-06`, `MOV-02`, `MOV-03`, `MOV-05`, `BOM-01` a `BOM-12`, `ORD-01` a `ORD-07`, `EST-03`, `EST-04`, `EST-08`, `FIM-01`, `EST-09`, `EST-10`, `FEC-01` a `FEC-08`, `DEC-06`: como no marco 13.
- `BOM-02`: no máximo `bombas_por_turno` bombas por turno, agora podendo chegar ao limite.
- `BOM-03`, `BOM-04`: duas bombas do bot na mesma casa formam pilha, com potência maior, e a previsão considera isso.
- `BOM-11`: bomba que atravessa o turno continua no tabuleiro, e o plano seguro já exige terminar fora do alcance dela.

## Critérios de aceitação

Cada critério é verificável por um teste automatizado. Salvo menção contrária, vale para as sementes 1 a 50, em tabuleiros pequenos montados à mão.

### Contrato e herança

- **CA-01** (BOT-01): **Dado** o bot criado com qualquer semente, **quando** `Versao()` é chamada, **então** devolve `"aleatorio-v3"`.
- **CA-02** (marco 13, CA-01 a CA-16): **Dados** os critérios do `aleatorio-v2`, **quando** os mesmos testes rodam com o `aleatorio-v3`, **então** todos passam, exceto os que fixam "no máximo uma bomba por turno", que passam a valer com "no máximo `bombas_por_turno`". O CA-06 do marco 3 (partida no mapa de exemplo sem infrações) roda só com bots v3.
- **CA-03** (API-09): **Dado** o catálogo padrão, **quando** as versões são listadas, **então** aparecem `aleatorio-v1`, `aleatorio-v2` e `aleatorio-v3`, e as três podem ser criadas por nome.
- **CA-04**: **Dado** um estado com `bombas_por_turno` 1, **quando** `Planejar` é chamado no v2 e no v3 com a mesma semente, **então** os planos são idênticos (com uma bomba só, o v3 é o v2).
- **CA-05** (EST-09, FEC-02): **Dado** um turno dentro da janela de antecipação, **quando** `Planejar` é chamado no v2 e no v3 com a mesma semente, **então** os planos são idênticos (decisão 3).
- **CA-06** (BOT-02): **Dado** o mapa de exemplo com `bombas_por_turno` 2, **quando** `Planejar` é chamado para cada jogador, **então** cada chamada termina em menos de 10% do `prazo_planejamento_ms` do mapa.

### Mais de uma bomba

- **CA-07** (BOM-02, VAL-03): **Dados** estados variados com `bombas_por_turno` de 1 a 3, **quando** `Planejar` é chamado, **então** o plano nunca tem mais `PLANTAR` do que `bombas_por_turno`, e `Validar` não aponta infração.
- **CA-08** (BOM-02): **Dado** um tabuleiro aberto, sem adversários perto, com `bombas_por_turno` 2, **quando** `Planejar` é chamado nas sementes 1 a 50, **então** pelo menos um plano tem 2 `PLANTAR` (o v3 usa a segunda bomba quando pode).
- **CA-09** (BOM-03, BOM-04, BOM-11): **Dado** um plano do v3 com 2 bombas, **quando** o turno é simulado, **então** o plano é seguro: o bot termina vivo e fora do alcance de todas as bombas que sobram, contando a pilha se as duas estiverem na mesma casa e as bombas que atravessam o turno.
- **CA-10** (BOM-11, ORD-07): **Dado** um tabuleiro pequeno com o v3 e um adversário que só espera, longe, com `bombas_por_turno` 2, **quando** os turnos são jogados com `Validar` e `ResolverTurno` até o limite de turnos, nas sementes 1 a 50, **então** o v3 nunca morre (as únicas bombas são as dele: morrer seria errar a previsão), e em pelo menos uma semente ele planta uma bomba que atravessa o turno.
- **CA-11** (BOM-02): **Dado** um bot sem plano seguro com 2 bombas mas com plano seguro com 1, **quando** `Planejar` é chamado, **então** o plano tem 1 `PLANTAR` (tenta com uma a menos, decisão 2).

### Comparação no mapa de exemplo

- **CA-12**: **Dadas** partidas com 4 `aleatorio-v3` e com 2 `aleatorio-v2` contra 2 `aleatorio-v3`, nas sementes 1 a 200, **quando** são jogadas até o fim, **então** o teste registra (`t.Log`) os turnos com 0, 1 e 2 bombas por jogador, as vitórias de cada versão, os empates e as mortes por bomba e por fechamento. Não há meta (decisão 5).

## Decisões

1. **Base do v3.** Decisão: o v2 inteiro (antecipação do fechamento e bomba de abertura), mudando só a quantidade de bombas fora da janela. Assim, a diferença entre v2 e v3 no ranking mede só o uso da segunda bomba.
2. **Quantas bombas tentar.** Decisão: nos turnos em que decide plantar (cerca de metade, como hoje), sorteia n de 1 a `bombas_por_turno`, com a mesma chance para cada valor. Tenta planos com n bombas; se nenhum for seguro, tenta com n − 1, e assim até 1; se nem com 1 houver plano seguro, não planta. As etapas das bombas são sorteadas em ordem crescente, e a mesma casa pode receber duas bombas (pilha, BOM-03).
3. **Dentro da janela de antecipação.** Decisão: igual ao v2 (no máximo a bomba de abertura). Abrir caminho com duas bombas é possível, mas complica a lógica que mais importa para sobreviver ao fechamento; fica para depois, se o ranking mostrar que vale.
4. **Bombas que atravessam o turno.** Decisão: nenhuma regra nova. O plano seguro do marco 3 já exige terminar fora do alcance de toda bomba que sobra, e no turno seguinte a fuga já considera as bombas existentes (CA-17 do marco 3). O CA-10 confere isso em jogo.
5. **Metas na comparação.** Decisão: só registrar (CA-12), sem meta. Duas bombas podem tanto matar mais adversários quanto fechar o próprio caminho; a medida decide se vale um ajuste depois.
6. **Nome.** Decisão: `aleatorio-v3`, seguindo o padrão das versões anteriores.
