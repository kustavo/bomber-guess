import { render } from '@testing-library/svelte';
import { describe, expect, test } from 'vitest';
import SpritesSvg from './SpritesSvg.svelte';
import Sprite from './Sprite.svelte';
import { NOMES_SPRITES, PALETA, SPRITES, idSprite, retangulos } from '../lib/sprites';

describe('SpritesSvg', () => {
  test('CA-01 um <symbol> 16 × 16 por sprite, com os retângulos do desenho', () => {
    const { container, unmount } = render(SpritesSvg);
    const simbolos = [...container.querySelectorAll('symbol')];
    expect(simbolos.map((s) => s.id).sort()).toEqual(NOMES_SPRITES.map(idSprite).sort());
    for (const nome of NOMES_SPRITES) {
      const s = container.querySelector(`#${idSprite(nome)}`)!;
      expect(s.getAttribute('viewBox')).toBe('0 0 16 16');
      const rects = [...s.querySelectorAll('rect')].map((r) => ({
        x: Number(r.getAttribute('x')),
        y: Number(r.getAttribute('y')),
        largura: Number(r.getAttribute('width')),
        cor: r.getAttribute('fill'),
      }));
      expect(rects).toEqual(retangulos(SPRITES[nome]));
      for (const r of rects) expect(Object.values(PALETA)).toContain(r.cor);
    }
    unmount();
  });

  test('CA-02 nada vem da rede: sem <image> nem href externo', () => {
    const { container, unmount } = render(SpritesSvg);
    expect(container.querySelector('image')).toBeNull();
    expect(container.querySelector('[href]')).toBeNull();
    unmount();
  });

  test('CA-02 CA-05 Sprite aponta para o símbolo local, na casa, com a cor dada', () => {
    const { container, unmount } = render(Sprite, { nome: 'jogador', x: 3, y: 2, cor: '#123456', classe: 'x' });
    const use = container.querySelector('use')!;
    expect(use.getAttribute('href')).toBe('#sprite-jogador');
    expect(use.getAttribute('data-sprite')).toBe('jogador');
    expect([use.getAttribute('x'), use.getAttribute('y'), use.getAttribute('width'), use.getAttribute('height')]).toEqual(['3', '2', '1', '1']);
    expect(use.style.color).toBe('rgb(18, 52, 86)');
    expect(use.classList.contains('x')).toBe(true);
    unmount();
  });
});
