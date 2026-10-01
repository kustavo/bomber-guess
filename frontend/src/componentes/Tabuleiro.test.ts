import { fireEvent, render } from '@testing-library/svelte';
import { describe, expect, test } from 'vitest';
import Tabuleiro from './Tabuleiro.svelte';
import PainelJogadores from './PainelJogadores.svelte';
import { calcularTabuleiro } from '../lib/tabuleiro';
import { estado, explosao, jogador, jogadorEtapa, p, relatorio } from '../testes/fabricas';
import { corJogador } from '../lib/textos';

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

// sprite: o sprite usado dentro do elemento de data-tipo na casa (x, y).
const sprite = (c: HTMLElement, tipo: string, x: number, y: number) =>
  c.querySelector(`[data-tipo="${tipo}"][data-x="${x}"][data-y="${y}"] use`)?.getAttribute('data-sprite');

describe('Tabuleiro com sprites (marco 14)', () => {
  test('CA-04 cada elemento usa o sprite do seu tipo; o piso cobre todas as casas', () => {
    const e = estado({ bombas: [{ posicao: p(1, 0), jogador_id: 'jogador_1', potencia: 2, pavio_restante: 2 }] });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [], 0) });
    expect(sprite(container, 'bloco-fixo', 2, 2)).toBe('bloco-fixo');
    expect(sprite(container, 'bloco-destrutivel', 2, 0)).toBe('bloco-destrutivel');
    expect(sprite(container, 'bomba', 1, 0)).toBe('bomba');
    expect(sprite(container, 'jogador', 0, 0)).toBe('jogador');
    const pisos = [...container.querySelectorAll('use[data-sprite="piso"]')];
    expect(pisos).toHaveLength(25);
    expect(pisos.every((u) => !u.closest('[data-tipo]'))).toBe(true); // piso não é elemento (D4)
    for (const u of container.querySelectorAll('use')) expect(u.getAttribute('href')).toBe(`#sprite-${u.getAttribute('data-sprite')}`);
  });

  test('CA-02 o tabuleiro não pede nada à rede', () => {
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(estado(), [], 0) });
    expect(container.querySelector('image')).toBeNull();
    for (const u of container.querySelectorAll('[href]')) expect(u.getAttribute('href')?.startsWith('#')).toBe(true);
  });

  test('CA-05 cada jogador com a sua cor e o seu número', () => {
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(estado(), [], 0) });
    const js = [...container.querySelectorAll('[data-tipo="jogador"]')];
    expect(js.map((j) => (j.querySelector('use') as SVGElement).style.color)).toEqual(
      [corJogador(0), corJogador(1)].map((c) => {
        const d = document.createElement('div');
        d.style.color = c;
        return d.style.color;
      }),
    );
    expect(js.map((j) => j.querySelector('.numero')?.textContent)).toEqual(['1', '2']);
  });

  test('CA-06 bomba usa o sprite e mostra o pavio', () => {
    const e = estado({ bombas: [{ posicao: p(3, 3), jogador_id: 'jogador_1', potencia: 2, pavio_restante: 3 }] });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [], 0) });
    expect(sprite(container, 'bomba', 3, 3)).toBe('bomba');
    expect(container.querySelector('[data-tipo="bomba"]')?.textContent).toBe('3');
  });

  test('DEC-09 CA-07 marca de bloqueio sobre o sprite do jogador', () => {
    const r = relatorio(1, { movimentos_bloqueados: ['jogador_2'] });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(estado(), [r], 1) });
    expect(container.querySelector('[data-jogador="jogador_2"] .marca-bloqueio')).not.toBeNull();
    expect(container.querySelector('[data-jogador="jogador_1"] .marca-bloqueio')).toBeNull();
  });

  test('FEC-03 FEC-05 CA-08 casa fechada usa o sprite de bloco fixo', () => {
    const r = relatorio(1, { blocos_fechados: [p(0, 2), p(2, 0)] });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(estado(), [r], 1) });
    expect(sprite(container, 'bloco-fixo', 0, 2)).toBe('bloco-fixo');
    expect(sprite(container, 'bloco-fixo', 2, 0)).toBe('bloco-fixo');
    expect(container.querySelector('[data-tipo="bloco-destrutivel"]')).toBeNull();
  });

  test('CA-09 cada chama usa o sprite da sua forma', () => {
    const e = explosao(p(1, 1), [p(0, 1), p(2, 1), p(3, 1), p(1, 0)]);
    const r = relatorio(1, { explosoes: [e], chamas: e.chamas });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(estado(), [r], 1) });
    expect(sprite(container, 'chama', 1, 1)).toBe('chama-centro');
    expect(sprite(container, 'chama', 0, 1)).toBe('chama-ponta-esquerda');
    expect(sprite(container, 'chama', 2, 1)).toBe('chama-horizontal');
    expect(sprite(container, 'chama', 3, 1)).toBe('chama-ponta-direita');
    expect(sprite(container, 'chama', 1, 0)).toBe('chama-ponta-cima');
  });

  test('CA-12 bombas e chamas com a classe de animação; jogador com deslize', () => {
    const e = estado({ bombas: [{ posicao: p(1, 0), jogador_id: 'jogador_1', potencia: 2, pavio_restante: 2 }] });
    const r = relatorio(1, { chamas: [p(3, 3)], bombas: e.bombas });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [r], 1), duracaoEtapaMs: 400 });
    expect(container.querySelector('[data-tipo="bomba"] use')?.classList.contains('animada-bomba')).toBe(true);
    expect(container.querySelector('[data-tipo="chama"] use')?.classList.contains('animada-chama')).toBe(true);
    const j = container.querySelector('[data-tipo="jogador"]') as SVGElement;
    expect(j.classList.contains('deslize')).toBe(true);
    expect(j.style.getPropertyValue('--transicao')).toBe('200ms');
  });

  test('CA-15 a bomba pulsa em torno do centro da sua casa', () => {
    const e = estado({ bombas: [{ posicao: p(3, 1), jogador_id: 'jogador_1', potencia: 2, pavio_restante: 2 }] });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [], 0) });
    const use = container.querySelector('[data-tipo="bomba"] use') as SVGElement;
    expect(use.style.transformOrigin).toBe('3.5px 1.5px');
  });

  test('CA-16 grade clara entre as casas, entre os blocos e o resto, sem receber cliques', () => {
    const e = estado({ bombas: [{ posicao: p(1, 0), jogador_id: 'jogador_1', potencia: 2, pavio_restante: 2 }] });
    const r = relatorio(1, { chamas: [p(3, 3)], bombas: e.bombas });
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [r], 1) });
    const grade = container.querySelector('[data-grade]')!;
    expect(grade.getAttribute('d')).toBe('M1 0V5M2 0V5M3 0V5M4 0V5M0 1H5M0 2H5M0 3H5M0 4H5');
    expect(grade.getAttribute('pointer-events')).toBe('none');
    const filhos = [...container.querySelector('svg')!.children];
    const pos = (sel: string) => filhos.findIndex((f) => f.matches(sel));
    const ultimo = (sel: string) => Math.max(...filhos.map((f, i) => (f.matches(sel) ? i : -1)));
    expect(pos('[data-grade]')).toBeGreaterThan(ultimo('[data-tipo^="bloco"]'));
    for (const t of ['chama', 'bomba', 'jogador']) expect(pos(`[data-tipo="${t}"]`)).toBeGreaterThan(pos('[data-grade]'));
  });
});

describe('cartão do jogador (marco 14, reabertura)', () => {
  const plantou = (etapa: number) =>
    relatorio(etapa, {
      jogadores: [
        jogadorEtapa('jogador_1', p(0, 0), { acao: { etapa, tipo: 'PLANTAR' }, resultado: 'EXECUTADA' }),
        jogadorEtapa('jogador_2', p(4, 4)),
      ],
    });
  const e = estado({ jogadores: [jogador('jogador_1', p(0, 0), { bombas_por_turno: 2 }), jogador('jogador_2', p(4, 4), { bot_versao: 'aleatorio-v2' })] });
  const campos = (c: HTMLElement) =>
    Object.fromEntries([...c.querySelectorAll('[data-cartao] [data-campo]')].map((el) => [el.getAttribute('data-campo'), el.textContent]));

  test('EST-04 BOM-02 CA-17 mouse sobre o jogador mostra nome, IA e bombas; sair esconde', async () => {
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [plantou(1)], 1) });
    expect(container.querySelector('[data-cartao]')).toBeNull();
    const j1 = container.querySelector('[data-jogador="jogador_1"]')!;
    await fireEvent.mouseEnter(j1);
    expect(campos(container)).toEqual({
      nome: 'jogador_1',
      bot: 'aleatorio-v1',
      restantes: 'Bombas restantes: 1',
      usadas: 'usou 1 de 2 (1/2)',
    });
    await fireEvent.mouseLeave(j1);
    expect(container.querySelector('[data-cartao]')).toBeNull();
  });

  test('CA-17 o cartão também abre com o foco do teclado', async () => {
    const { container } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [], 0) });
    const j2 = container.querySelector('[data-jogador="jogador_2"]') as SVGElement;
    expect(j2.getAttribute('tabindex')).toBe('0');
    await fireEvent.focus(j2);
    expect(campos(container)).toMatchObject({ nome: 'jogador_2', bot: 'aleatorio-v2', usadas: 'usou 0 de 1 (0/1)' });
    await fireEvent.blur(j2);
    expect(container.querySelector('[data-cartao]')).toBeNull();
  });

  test('CA-17 com o cartão aberto, os números acompanham a etapa exibida', async () => {
    const rs = [plantou(1), plantou(2)];
    const { container, rerender } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, rs, 1) });
    await fireEvent.mouseEnter(container.querySelector('[data-jogador="jogador_1"]')!);
    expect(campos(container).usadas).toBe('usou 1 de 2 (1/2)');
    await rerender({ tabuleiro: calcularTabuleiro(e, rs, 2) });
    expect(campos(container)).toMatchObject({ restantes: 'Bombas restantes: 0', usadas: 'usou 2 de 2 (2/2)' });
  });

  test('FIM-01 CA-17 jogador que morre com o cartão aberto: o cartão some', async () => {
    const morte = relatorio(1, { mortes: ['jogador_1'], jogadores: [jogadorEtapa('jogador_1', p(0, 0), { status: 'MORTO' }), jogadorEtapa('jogador_2', p(4, 4))] });
    const { container, rerender } = render(Tabuleiro, { tabuleiro: calcularTabuleiro(e, [morte], 0) });
    await fireEvent.mouseEnter(container.querySelector('[data-jogador="jogador_1"]')!);
    expect(container.querySelector('[data-cartao]')).not.toBeNull();
    await rerender({ tabuleiro: calcularTabuleiro(e, [morte], 1) });
    expect(container.querySelector('[data-cartao]')).toBeNull();
  });
});
