# Marco 13: bot aleatorio-v2 (sobrevive ao fechamento)

**Status**: concluída
**Roadmap**: `docs/ROADMAP.md`, marco 13
**Documentos de referência**: `docs/BOTS.md`, `docs/REGRAS.md` (em especial a seção 7, "Fechamento do tabuleiro"); `specs/03-bot-simples/spec.md` (comportamento do `aleatorio-v1`, que o v2 estende)

## Objetivo

Entregar `aleatorio-v2`, um bot que mantém o jeito do `aleatorio-v1` mas se prepara para o fechamento do tabuleiro (marco 12). O v1 só olha o turno atual: escapa do anel que fecha naquele turno, mas fica preso em bolsões cercados por blocos destrutíveis e morre quando o anel deles fecha. Em 200 partidas com 4 `aleatorio-v1` no mapa de exemplo, 631 de 685 mortes (92%) foram pelo fechamento, e em todas elas já não havia saída no início do turno. O v2 deve ir para o centro antes da hora e abrir caminho com bombas quando estiver preso.

O `aleatorio-v1` não muda. Ele continua no catálogo como referência mais fraca do ranking.

### Termos usados nesta spec

Valem os termos do marco 3: **simulação do turno**, **plano seguro** e **plano sobrevivente**. Além deles:

- **Casas que fecham**: as casas que viram bloco fixo ao fim do turno atual (`CasasQueFecham`, FEC-03).
- **Anel alvo** do turno `T`: o menor anel em que o bot quer terminar o turno, com uma margem sobre o que fecha no turno seguinte. Ver decisão 1.
- **Janela de antecipação**: os turnos em que o anel alvo vale. Antes e depois dela, o bot anda como o v1. Ver decisão 1.
- **Região**: as casas que o bot alcança a partir da casa atual andando só por casas livres, sem atravessar blocos, ignorando bombas e quantidade de etapas.
- **Preso**: a região do bot não tem nenhuma casa no anel alvo.
- **Bloco de abertura**: o primeiro bloco destrutível do caminho mais curto até o anel alvo, contando cada bloco destrutível atravessado como custo. Destruí-lo é o que tira o bot da prisão.

## Escopo

**Dentro**
- Subpacote de bot com a versão `aleatorio-v2`, implementando a interface `Bot`, registrado no catálogo (`GET /bots`, terminal).
- Tudo o que o `aleatorio-v1` garante: contrato, validade, determinismo e fuga (CA-01 a CA-22 do marco 3), agora também para o v2.
- Antecipação do fechamento: dentro da janela, terminar o turno no anel alvo quando possível e, senão, o mais perto dele.
- Abertura de caminho: quando preso, plantar bomba que destrua o bloco de abertura, se houver plano seguro com ela.
- No último recurso (sem plano seguro), preferir um plano sobrevivente que não termine em casa que fecha.
- Testes de comparação com o v1 no mapa de exemplo.

**Fora** (fica para outro marco)
- Mudanças no `aleatorio-v1` e nas regras do jogo.
- Prever adversários, caçá-los ou disputar o centro com eles: bots do marco 10.
- Atravessar blocos destrutíveis apostando que alguém os destrói (MOV-03): o v2, como o v1, nunca planeja movimento bloqueado.
- Ranking: marco 9. Aqui só se mede nos testes.

## Regras cobertas

- `BOT-01` a `BOT-04`, `VAL-01` a `VAL-06`, `ACA-01`, `ACA-02`, `MOV-02`, `MOV-03`, `MOV-05`, `BOM-01` a `BOM-12`, `ORD-01` a `ORD-07`, `EST-03`, `EST-04`, `EST-08`, `FIM-01`: como no marco 3.
- `EST-09`, `EST-10`, `FEC-01` a `FEC-04`, `FEC-08`: previsão do fechamento no turno atual e nos seguintes, até a área mínima.
- `FEC-05`: bloco destrutível que fecha deixa de ser bloco de abertura útil.
- `FEC-07`, `DEC-06`: fim antecipado na simulação conta como seguro, como no v1.

## Critérios de aceitação

Cada critério é verificável por um teste automatizado. Salvo menção contrária, vale para as sementes 1 a 50, em tabuleiros pequenos montados à mão.

### Contrato e herança do v1

- **CA-01** (BOT-01): **Dado** o bot criado com qualquer semente, **quando** `Versao()` é chamada, **então** devolve `"aleatorio-v2"`.
- **CA-02** (marco 3, CA-02 a CA-22): **Dados** os critérios de aceitação do `aleatorio-v1`, **quando** os mesmos testes rodam com o `aleatorio-v2`, **então** todos passam. O CA-06 do marco 3 (partida no mapa de exemplo sem infrações) roda só com bots v2.
- **CA-03** (API-09): **Dado** o catálogo padrão, **quando** as versões são listadas, **então** aparecem `aleatorio-v1` e `aleatorio-v2`, e as duas podem ser criadas por nome.
- **CA-04** (EST-09, FEC-01): **Dado** um estado com o fechamento desligado (`turno_fechamento` 0), **quando** `Planejar` é chamado no v1 e no v2 com a mesma semente, **então** os planos são idênticos (questão 2).
- **CA-05** (BOT-02): **Dado** o mapa de exemplo em um turno dentro da janela, com o bot preso, **quando** `Planejar` é chamado para cada jogador, **então** cada chamada termina em menos de 10% do `prazo_planejamento_ms` do mapa.

### Antecipação

- **CA-06** (FEC-03): **Dado** um turno anterior à janela de antecipação, em campo aberto, **quando** `Planejar` é chamado, **então** o plano é igual ao do v1 com a mesma semente.
- **CA-07** (FEC-02, FEC-03): **Dado** um turno dentro da janela, com o bot na borda e uma casa do anel alvo alcançável por casas livres neste turno, sem bombas, **quando** o plano é simulado, **então** o bot termina o turno numa casa do anel alvo, e o plano é seguro.
- **CA-08** (FEC-02): **Dado** um turno dentro da janela em que o anel alvo está longe demais para este turno (mais casas que `acoes_por_turno`), **quando** o plano é simulado, **então** o bot termina mais perto do anel alvo do que começou (distância por casas livres).
- **CA-09** (BOM-11): **Dado** um turno dentro da janela em que a única casa do anel alvo alcançável está no alcance de uma bomba que sobra para o turno seguinte, e há casa segura fora do anel alvo, **quando** o plano é simulado, **então** o plano é seguro: a segurança vem antes do anel alvo.
- **CA-10** (FEC-04): **Dado** um bot sem plano seguro e com dois planos sobreviventes, um terminando em casa que fecha e outro não, **quando** `Planejar` é chamado, **então** o plano devolvido termina fora das casas que fecham.

### Abertura de caminho

- **CA-11** (BOM-07, FEC-02): **Dado** um bot preso, com um único bloco destrutível entre a região dele e o anel alvo, sem bombas no tabuleiro, **quando** `Planejar` é chamado, **então** o plano tem `PLANTAR` numa casa cuja explosão alcança esse bloco e é seguro.
- **CA-12** (BOM-02): **Dado** um bot preso cujo plano com bomba de abertura não seria seguro (por exemplo, sem casa para fugir do próprio fogo), **quando** `Planejar` é chamado, **então** o plano não tem `PLANTAR` e é o melhor plano sem bomba (seguro, senão sobrevivente).
- **CA-13** (FEC-03, FEC-04, ORD-06): **Dado** um tabuleiro pequeno em que o bot começa preso atrás de um bloco destrutível, com o fechamento ligado e um adversário parado no centro, **quando** os turnos são jogados com `Validar` e `ResolverTurno` até o anel do bot fechar, **então** o v2 destrói o bloco, passa para o lado de dentro e está vivo depois que o anel em que começou fecha. Com o v1 no lugar, no mesmo tabuleiro, ele morre em pelo menos uma semente (o teste mostra a diferença).

### Comparação no mapa de exemplo

Partidas jogadas com `Validar` e `ResolverTurno`, como no CA-06 do marco 3, sementes 1 a 200. Os limites estão na questão 4.

- **CA-14** (FEC-04): **Dadas** partidas com 4 `aleatorio-v2`, **quando** são jogadas até o fim, **então** as mortes por fechamento são no máximo 25% das medidas com 4 `aleatorio-v1` nas mesmas sementes, e nenhuma morte por fechamento acontece com saída disponível no início do turno (casa que não fecha alcançável por casas livres em até `acoes_por_turno` passos).
- **CA-15**: **Dadas** partidas com 2 `aleatorio-v1` e 2 `aleatorio-v2`, alternando as posições iniciais entre as sementes, **quando** são jogadas até o fim, **então** o v2 vence pelo menos o dobro de partidas que o v1.
- **CA-16**: **Nas** partidas do CA-14, **então** o teste registra (`t.Log`) a quantidade de empates e de mortes por fechamento e por bomba, para comparar com o v1. Não há limite para empates (questão 5).

## Decisões

Questões levantadas no rascunho e aprovadas pelo usuário em 2026-10-01, com as propostas como estavam.

1. **Janela e anel alvo** (revista na implementação, ver R1 e R2). Com o fechamento ligado, `f = turno_fechamento` e `k` = primeiro anel que nunca fecha (FEC-08), a janela vai do turno `f − 3` até o turno `f + k − 1`, em que fecha o último anel que pode fechar. No turno `T` dentro dela, o anel alvo é `max(0, T − f + 1) + 2`, limitado a `k` e ao maior anel do tabuleiro. Ou seja: o anel que fecha no turno seguinte mais duas casas de margem. Exemplo no mapa de exemplo (`f` 30, `k` 4): nos turnos 27 a 29 o alvo é o anel 2; no turno 30, o anel 3; do turno 31 ao 33, o anel 4; do turno 34 em diante a janela acabou e o v2 joga como o v1.
2. **v2 igual ao v1 sem fechamento.** Proposta: sim (CA-04 e CA-06). Assim, a diferença entre as versões no ranking mede só a estratégia contra o fechamento. O custo é o v2 herdar todas as limitações do v1 fora da janela.
3. **Ordem de preferência dos planos dentro da janela.** Proposta, do melhor para o pior:
   1. seguro e no anel alvo;
   2. seguro, com bomba de abertura (só quando preso);
   3. seguro e mais perto do anel alvo;
   4. sobrevivente e fora das casas que fecham;
   5. sobrevivente;
   6. o que sobrevive mais etapas.

   Fora da janela, vale a ordem do v1, mais o item 4 (CA-10), que só muda algo quando há casas que fecham.
4. **Limites dos testes de comparação** (ver R1 e R3). Mortes por fechamento com v2 ≤ 25% das do v1 nas mesmas sementes, e v2 com pelo menos o dobro de vitórias do v1 no confronto 2 contra 2. Se a implementação ficar longe disso, os números voltam para revisão em vez de afrouxar o teste em silêncio.
5. **Efeito nos empates.** Sem área mínima, bots que sobrevivem ao fechamento tendem a chegar juntos ao centro e morrer juntos quando o último anel fecha. Com 4 v2 os empates podem até aumentar em relação ao v1 (85 de 200). Proposta: não pôr meta de empates neste marco, só registrar (CA-16). Se os empates continuarem altos, a solução é de regra (por exemplo, desempate por quem morreu por último ou por blocos destruídos), discutida à parte.
6. **Bombas aleatórias.** O v1 tenta plantar em cerca de metade dos turnos. Proposta: dentro da janela, a bomba de abertura (quando preso) substitui o sorteio; fora da janela, nada muda.

## Revisões durante a implementação

- **R1** (aprovada em 2026-10-01): margem do anel alvo de 1 para 2 (decisão 1). Com margem 1, o v2 ficava preso atrás de blocos no anel 3 enquanto o fechamento avançava um anel por turno. Na época, sem área mínima, o CA-14 também mudou para "anel médio das mortes por fechamento pelo menos 1 anel mais fundo que o do v1", porque todo sobrevivente acabava morrendo no fechamento do centro e a contagem de mortes não podia cair.
- **R2** (regra nova pedida pelo usuário em 2026-10-01): área mínima que nunca fecha (EST-10, FEC-08; adendo do marco 12). Consequências no v2:
  - o anel alvo é limitado ao primeiro anel que nunca fecha;
  - a janela termina quando fecha o último anel que pode fechar. Sem isso, o v2 ficava na janela para sempre, sem as bombas aleatórias (decisão 6), e 196 de 200 partidas com 4 v2 empatavam por `limite_turnos`. Com a janela terminando, ele volta a jogar como o v1 dentro da área mínima.
- **R3** (aprovada em 2026-10-01): com a área mínima, a contagem de mortes por fechamento voltou a medir o que importa, e o anel médio deixou de medir: os anéis que podem fechar vão só de 0 a 3. Decisão: voltar o CA-14 ao texto original (≤ 25% das mortes por fechamento do v1, e nenhuma com saída disponível). Medido nas sementes 1 a 200:

  | | Mortes por fechamento | Anel médio | Mortes por bomba | Empates (4 iguais) | 2 contra 2 |
  |---|---|---|---|---|---|
  | v1 | 587 | 2,34 | 76 | 69 | 18 vitórias |
  | v2 | 91 (15,5%) | 2,63 | 441 | 76 | 82 vitórias |

