import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
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

describe('grade do editor (marco 14)', () => {
  test('CA-16 grade clara depois dos blocos e antes das posições e das casas clicáveis', () => {
    const m = mapa(['bloco-fixo', p(1, 1)], ['posicao-inicial', p(0, 0)]);
    const { container, unmount } = render(TabuleiroEditor, { mapa: m, onCasa: () => {} });
    const grade = container.querySelector('[data-grade]')!;
    expect(grade.getAttribute('d')).toBe('M1 0V3M2 0V3M3 0V3M0 1H4M0 2H4');
    expect(grade.getAttribute('pointer-events')).toBe('none');
    const filhos = [...container.querySelector('svg')!.children];
    const pos = (sel: string) => filhos.findIndex((f) => f.matches(sel));
    expect(pos('[data-grade]')).toBeGreaterThan(pos('[data-tipo="bloco-fixo"]'));
    expect(pos('[data-grade]')).toBeLessThan(pos('[data-tipo="posicao-inicial"]'));
    expect(pos('[data-grade]')).toBeLessThan(pos('[data-casa]'));
    unmount();
  });
});

describe('foco das casas (marco 7, reabertura)', () => {
  // O jsdom não desenha o anel de foco; o teste lê o <style> do componente (D16).
  const fonte = readFileSync(resolve(process.cwd(), 'src/componentes/TabuleiroEditor.svelte'), 'utf8');
  const estilo = fonte.slice(fonte.indexOf('<style>'));
  const corpo = (seletor: string) => {
    const i = estilo.indexOf(seletor);
    return i < 0 ? undefined : estilo.slice(estilo.indexOf('{', i) + 1, estilo.indexOf('}', i));
  };

  test('CA-21 casas nunca usam o anel de foco do navegador', () => {
    expect(corpo('.casa {')).toMatch(/outline:\s*none/);
  });

  test('CA-21 o foco por teclado continua visível pelo realce da casa', () => {
    expect(estilo).toMatch(/\.casa:focus-visible[^{]*\{[^}]*stroke:\s*#fff/);
  });
});
