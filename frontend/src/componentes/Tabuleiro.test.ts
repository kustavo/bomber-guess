import { render } from '@testing-library/svelte';
import { describe, expect, test } from 'vitest';
import Tabuleiro from './Tabuleiro.svelte';
import PainelJogadores from './PainelJogadores.svelte';
import { calcularTabuleiro } from '../lib/tabuleiro';
import { estado, jogador, jogadorEtapa, p, relatorio } from '../testes/fabricas';

const casas = (c: HTMLElement, tipo: string) =>
  [...c.querySelectorAll(`[data-tipo="${tipo}"]`)].map((e) => [Number(e.getAttribute('data-x')), Number(e.getAttribute('data-y'))]);

describe('Tabuleiro', () => {
  test('API-05 CA-01 desenha largura × altura e cada elemento na sua casa', () => {
    const e = estado({
      bombas: [{ posicao: p(1, 0), jogador_id: 'jogador_1', potencia: 2, pavio_restante: 2 }],
    });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [], 0) });
    expect(container.querySelector('svg')?.getAttribute('viewBox')).toBe('0 0 5 5');
    expect(casas(container, 'bloco-fixo')).toEqual([[2, 2]]);
    expect(casas(container, 'bloco-destrutivel')).toEqual([[2, 0]]);
    expect(casas(container, 'bomba')).toEqual([[1, 0]]);
    expect(container.querySelector('[data-tipo="bomba"]')?.textContent).toBe('2'); // pavio_restante
    expect(casas(container, 'jogador')).toEqual([[0, 0], [4, 4]]);
    expect(casas(container, 'chama')).toEqual([]);
  });

  test('API-05 CA-02 chamas da etapa exibida', () => {
    const t = calcularTabuleiro(estado(), [relatorio(1, { chamas: [p(0, 1), p(0, 2)] })], 1);
    const { container } = render(Tabuleiro, { tabuleiro: t });
    expect(casas(container, 'chama')).toEqual([[0, 1], [0, 2]]);
  });

  test('FIM-01 CA-03 jogador morto não ocupa casa', () => {
    const e = estado({
      jogadores: [
        jogador('jogador_1', p(0, 0)),
        jogador('jogador_2', p(4, 4), { status: 'MORTO', morte: { turno: 3, etapa: 2 } }),
      ],
    });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [], 0) });
    expect([...container.querySelectorAll('[data-tipo="jogador"]')].map((j) => j.getAttribute('data-jogador'))).toEqual([
      'jogador_1',
    ]);
  });

  test('DEC-09 CA-04 jogador bloqueado tem a marca BLOQUEADA', () => {
    const r = relatorio(1, {
      jogadores: [
        jogadorEtapa('jogador_1', p(0, 0), { acao: { etapa: 1, tipo: 'MOVER', direcao: 'CIMA' }, resultado: 'BLOQUEADA' }),
        jogadorEtapa('jogador_2', p(4, 4)),
      ],
      movimentos_bloqueados: ['jogador_1'],
    });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(estado(), [r], 1) });
    const j1 = container.querySelector('[data-jogador="jogador_1"]')!;
    const j2 = container.querySelector('[data-jogador="jogador_2"]')!;
    expect(j1.getAttribute('data-bloqueado')).toBe('true');
    expect(j1.classList.contains('bloqueado')).toBe(true);
    expect(j1.querySelector('title')?.textContent).toContain('BLOQUEADA');
    expect(j2.getAttribute('data-bloqueado')).toBe('false');
  });
});

describe('PainelJogadores', () => {
  test('EST-07 CA-03 morto aparece no painel com turno e etapa da morte', () => {
    const e = estado({
      jogadores: [
        jogador('jogador_1', p(0, 0), { bot_versao: 'claude-v1' }),
        jogador('jogador_2', p(4, 4), { status: 'MORTO', morte: { turno: 3, etapa: 2 } }),
      ],
    });
    const { container } = render(PainelJogadores, { jogadores: calcularTabuleiro(e, [], 0).jogadores });
    const j1 = container.querySelector('[data-jogador="jogador_1"]')!;
    const j2 = container.querySelector('[data-jogador="jogador_2"]')!;
    expect(j1.textContent).toContain('claude-v1');
    expect(j1.textContent).toContain('Vivo');
    expect(j2.textContent).toContain('Morto no turno 3, etapa 2');
  });

  test('DEC-09 CA-04 painel mostra a ação e o resultado bloqueada', () => {
    const r = relatorio(1, {
      jogadores: [
        jogadorEtapa('jogador_1', p(0, 0), { acao: { etapa: 1, tipo: 'MOVER', direcao: 'CIMA' }, resultado: 'BLOQUEADA' }),
        jogadorEtapa('jogador_2', p(4, 4)),
      ],
      movimentos_bloqueados: ['jogador_1'],
    });
    const { container } = render(PainelJogadores, { jogadores: calcularTabuleiro(estado(), [r], 1).jogadores });
    expect(container.querySelector('[data-jogador="jogador_1"] .acao')?.textContent).toBe('Mover ↑ · bloqueada');
  });
});
