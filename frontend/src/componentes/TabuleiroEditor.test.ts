import { render } from '@testing-library/svelte';
import { describe, expect, test } from 'vitest';
import TabuleiroEditor from './TabuleiroEditor.svelte';
import { aplicarFerramenta, novoMapaEmEdicao, redimensionar, type Ferramenta } from '../lib/editor';
import { corJogador } from '../lib/textos';
import { p } from '../testes/fabricas';
import type { Posicao } from '../lib/tipos';

function mapa(...passos: [Ferramenta, Posicao][]) {
  return passos.reduce((m, [f, pos]) => aplicarFerramenta(m, f, pos, 'v1'), redimensionar(novoMapaEmEdicao(), 4, 3));
}

const cor = (c: string) => {
  const d = document.createElement('div');
  d.style.color = c;
  return d.style.color;
};

describe('TabuleiroEditor com sprites (marco 14)', () => {
  test('MAP-03 CA-10 blocos e posições iniciais com sprites; posições com cor e número', () => {
    const m = mapa(['bloco-fixo', p(1, 1)], ['bloco-destrutivel', p(2, 1)], ['posicao-inicial', p(0, 0)], ['posicao-inicial', p(3, 2)]);
    const { container, unmount } = render(TabuleiroEditor, { mapa: m, onCasa: () => {} });
    const spriteEm = (tipo: string, x: number, y: number) =>
      container.querySelector(`[data-tipo="${tipo}"][data-x="${x}"][data-y="${y}"] use`) as SVGElement | null;
    expect(spriteEm('bloco-fixo', 1, 1)?.getAttribute('data-sprite')).toBe('bloco-fixo');
    expect(spriteEm('bloco-destrutivel', 2, 1)?.getAttribute('data-sprite')).toBe('bloco-destrutivel');
    expect(spriteEm('posicao-inicial', 0, 0)?.getAttribute('data-sprite')).toBe('jogador');
    expect(spriteEm('posicao-inicial', 0, 0)?.style.color).toBe(cor(corJogador(0)));
    expect(spriteEm('posicao-inicial', 3, 2)?.style.color).toBe(cor(corJogador(1)));
    const numeros = [...container.querySelectorAll('[data-tipo="posicao-inicial"] .numero')].map((t) => t.textContent);
    expect(numeros).toEqual(['1', '2']);
    unmount();
  });

  test('CA-10 piso em todas as casas, sem data-tipo; casas clicáveis por cima de tudo (D9)', () => {
    const m = mapa(['bloco-fixo', p(1, 1)]);
    const { container, unmount } = render(TabuleiroEditor, { mapa: m, onCasa: () => {} });
    const pisos = [...container.querySelectorAll('use[data-sprite="piso"]')];
    expect(pisos).toHaveLength(12);
    expect(pisos.every((u) => !u.closest('[data-tipo]'))).toBe(true);
    const filhos = [...container.querySelector('svg')!.children];
    const ultimoItem = Math.max(...filhos.map((e, i) => (e.matches('[data-tipo], use') ? i : -1)));
    const primeiraCasa = filhos.findIndex((e) => e.hasAttribute('data-casa'));
    expect(primeiraCasa).toBeGreaterThan(ultimoItem);
    expect(container.querySelectorAll('[data-casa]')).toHaveLength(12);
    unmount();
  });
});
