import { describe, expect, test } from 'vitest';
import { calcularTabuleiro } from './tabuleiro';
import { estado, jogador, jogadorEtapa, p, relatorio } from '../testes/fabricas';
import type { Posicao } from './tipos';

const ordenar = (ps: Posicao[]) => [...ps].sort((a, b) => a.y - b.y || a.x - b.x);

describe('calcularTabuleiro', () => {
  test('API-05 CA-01 planejamento: tudo nas posições do estado', () => {
    const e = estado({
      bombas: [{ posicao: p(1, 0), jogador_id: 'jogador_1', potencia: 2, pavio_restante: 2 }],
    });
    const t = calcularTabuleiro(e, [], 0);
    expect(t.largura).toBe(5);
    expect(t.altura).toBe(5);
    expect(t.turno).toBe(1);
    expect(t.etapa).toBe(0);
    expect(t.blocosFixos).toEqual([p(2, 2)]);
    expect(t.blocosDestrutiveis).toEqual([p(2, 0)]);
    expect(t.bombas).toEqual(e.bombas);
    expect(t.chamas).toEqual([]);
    expect(t.jogadores.map((j) => [j.id, j.indice, j.posicao, j.status, j.bot_versao])).toEqual([
      ['jogador_1', 0, p(0, 0), 'VIVO', 'aleatorio-v1'],
      ['jogador_2', 1, p(4, 4), 'VIVO', 'aleatorio-v1'],
    ]);
    expect(t.jogadores.every((j) => !j.bloqueado && j.acao === undefined)).toBe(true);
  });

  const bomba = { posicao: p(1, 1), jogador_id: 'jogador_2', potencia: 1, pavio_restante: 1 };
  const etapas = [
    relatorio(1, {
      jogadores: [
        jogadorEtapa('jogador_1', p(1, 0), { acao: { etapa: 1, tipo: 'MOVER', direcao: 'DIREITA' } }),
        jogadorEtapa('jogador_2', p(4, 4), { acao: { etapa: 1, tipo: 'PLANTAR' } }),
      ],
      bombas: [bomba],
      chamas: [p(3, 0), p(2, 0)],
      blocos_destruidos: [p(2, 0)],
    }),
    relatorio(2, {
      jogadores: [
        jogadorEtapa('jogador_1', p(1, 1), { acao: { etapa: 2, tipo: 'MOVER', direcao: 'BAIXO' } }),
        jogadorEtapa('jogador_2', p(4, 3), { acao: { etapa: 2, tipo: 'MOVER', direcao: 'CIMA' } }),
      ],
      bombas: [],
      chamas: [p(1, 1)],
      blocos_destruidos: [p(0, 4)],
    }),
  ];
  const comDoisBlocos = estado({ blocos_destrutiveis: [p(2, 0), p(0, 4), p(4, 0)] });

  test.each([
    {
      nome: 'API-05 CA-02 etapa 1: posições, bombas e chamas do relatório 1',
      ate: 1,
      posicoes: [p(1, 0), p(4, 4)],
      bombas: [bomba],
      chamas: [p(2, 0), p(3, 0)],
      destrutiveis: [p(4, 0), p(0, 4)],
    },
    {
      nome: 'API-05 CA-02 etapa 2: chamas só da etapa 2 e blocos destruídos acumulados',
      ate: 2,
      posicoes: [p(1, 1), p(4, 3)],
      bombas: [],
      chamas: [p(1, 1)],
      destrutiveis: [p(4, 0)],
    },
    {
      nome: 'API-05 CA-02 ate maior que os relatórios: para no último liberado',
      ate: 9,
      posicoes: [p(1, 1), p(4, 3)],
      bombas: [],
      chamas: [p(1, 1)],
      destrutiveis: [p(4, 0)],
    },
  ])('$nome', ({ ate, posicoes, bombas, chamas, destrutiveis }) => {
    const t = calcularTabuleiro(comDoisBlocos, etapas, ate);
    expect(t.etapa).toBe(Math.min(ate, etapas.length));
    expect(t.jogadores.map((j) => j.posicao)).toEqual(posicoes);
    expect(t.bombas).toEqual(bombas);
    expect(ordenar(t.chamas)).toEqual(ordenar(chamas));
    expect(ordenar(t.blocosDestrutiveis)).toEqual(ordenar(destrutiveis));
  });

  test('DEC-09 CA-02 ação e resultado da etapa vêm do relatório', () => {
    const t = calcularTabuleiro(estado(), etapas, 1);
    expect(t.jogadores[0].acao).toEqual({ etapa: 1, tipo: 'MOVER', direcao: 'DIREITA' });
    expect(t.jogadores[0].resultado).toBe('EXECUTADA');
  });

  test('API-05 CA-02 não altera o estado nem os relatórios recebidos', () => {
    const e = estado();
    const antes = JSON.stringify([e, etapas]);
    calcularTabuleiro(e, etapas, 2);
    expect(JSON.stringify([e, etapas])).toBe(antes);
  });

  test.each([
    {
      nome: 'EST-07 CA-03 morto no estado do início do turno: com morte, fora das ocupantes',
      e: estado({
        jogadores: [
          jogador('jogador_1', p(0, 0)),
          jogador('jogador_2', p(4, 4), { status: 'MORTO', morte: { turno: 3, etapa: 2 } }),
        ],
      }),
      etapas: [],
      ate: 0,
      morte: { turno: 3, etapa: 2 },
    },
    {
      nome: 'FIM-01 CA-03 morto na etapa k: status do relatório e morte no turno e etapa k',
      e: estado({ turno: 4 }),
      etapas: [
        relatorio(1, {
          turno: 4,
          jogadores: [
            jogadorEtapa('jogador_1', p(0, 0)),
            jogadorEtapa('jogador_2', p(4, 4), { status: 'MORTO' }),
          ],
          mortes: ['jogador_2'],
        }),
      ],
      ate: 1,
      morte: { turno: 4, etapa: 1 },
    },
  ])('$nome', ({ e, etapas, ate, morte }) => {
    const t = calcularTabuleiro(e, etapas, ate);
    const morto = t.jogadores.find((j) => j.id === 'jogador_2')!;
    expect(morto.status).toBe('MORTO');
    expect(morto.morte).toEqual(morte);
    expect(morto.posicao).toEqual(p(4, 4)); // EST-08: a casa onde morreu
    expect(t.jogadores.filter((j) => j.status === 'VIVO').map((j) => j.id)).toEqual(['jogador_1']);
  });

  describe('fechamento do tabuleiro', () => {
    // Borda do 5×5 fecha ao fim da etapa 2; jogador_1 (0,0) morre nela, jogador_2 já está em (3,3).
    const borda = [
      ...[0, 1, 2, 3, 4].map((x) => p(x, 0)),
      ...[1, 2, 3].flatMap((y) => [p(0, y), p(4, y)]),
      ...[0, 1, 2, 3, 4].map((x) => p(x, 4)),
    ];
    const e = estado({
      turno: 7,
      config: { ...estado().config, turno_fechamento: 7 },
      jogadores: [jogador('jogador_1', p(0, 0)), jogador('jogador_2', p(3, 3))],
    });
    const etapasFechamento = [
      relatorio(1, { turno: 7, jogadores: [jogadorEtapa('jogador_1', p(0, 0)), jogadorEtapa('jogador_2', p(3, 3))] }),
      relatorio(2, {
        turno: 7,
        jogadores: [jogadorEtapa('jogador_1', p(0, 0), { status: 'MORTO' }), jogadorEtapa('jogador_2', p(3, 3))],
        mortes: ['jogador_1'],
        blocos_fechados: borda,
      }),
    ];

    test.each([
      { nome: 'FEC-03 CA-17 antes da etapa do fechamento: tabuleiro ainda aberto', ate: 1, fixos: [p(2, 2)], destrutiveis: [p(2, 0)] },
      { nome: 'FEC-03 FEC-05 CA-17 na etapa do fechamento: borda fixa, destrutível some', ate: 2, fixos: [p(2, 2), ...borda], destrutiveis: [] },
    ])('$nome', ({ ate, fixos, destrutiveis }) => {
      const t = calcularTabuleiro(e, etapasFechamento, ate);
      expect(ordenar(t.blocosFixos)).toEqual(ordenar(fixos));
      expect(t.blocosDestrutiveis).toEqual(destrutiveis);
    });

    test('FEC-04 CA-17 morto pelo fechamento sai do tabuleiro, com a morte na etapa do fechamento', () => {
      const t = calcularTabuleiro(e, etapasFechamento, 2);
      const morto = t.jogadores.find((j) => j.id === 'jogador_1')!;
      expect(morto.status).toBe('MORTO');
      expect(morto.morte).toEqual({ turno: 7, etapa: 2 });
      expect(t.jogadores.filter((j) => j.status === 'VIVO').map((j) => j.id)).toEqual(['jogador_2']);
    });
  });

  test('DEC-09 CA-04 bloqueado a partir de movimentos_bloqueados', () => {
    const r = relatorio(1, {
      jogadores: [
        jogadorEtapa('jogador_1', p(0, 0), { acao: { etapa: 1, tipo: 'MOVER', direcao: 'CIMA' }, resultado: 'BLOQUEADA' }),
        jogadorEtapa('jogador_2', p(4, 4)),
      ],
      movimentos_bloqueados: ['jogador_1'],
    });
    const t = calcularTabuleiro(estado(), [r], 1);
    expect(t.jogadores.map((j) => j.bloqueado)).toEqual([true, false]);
    expect(t.jogadores[0].resultado).toBe('BLOQUEADA');
  });
});
