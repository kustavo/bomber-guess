import { describe, expect, test } from 'vitest';
import { LADO_SPRITE, NOMES_SPRITES, PALETA, SPRITES, idSprite, retangulos } from './sprites';

describe('sprites', () => {
  test('CA-01 existe um sprite para cada nome da spec', () => {
    expect([...NOMES_SPRITES].sort()).toEqual(
      [
        'piso',
        'bloco-fixo',
        'bloco-destrutivel',
        'bomba',
        'jogador',
        'chama-centro',
        'chama-horizontal',
        'chama-vertical',
        'chama-ponta-cima',
        'chama-ponta-baixo',
        'chama-ponta-esquerda',
        'chama-ponta-direita',
      ].sort(),
    );
    expect(Object.keys(SPRITES).sort()).toEqual([...NOMES_SPRITES].sort());
  });

  test.each(NOMES_SPRITES.map((nome) => ({ nome })))('CA-01 $nome: grade 16 × 16 só com cores da paleta', ({ nome }) => {
    const desenho = SPRITES[nome];
    expect(LADO_SPRITE).toBe(16);
    expect(desenho).toHaveLength(LADO_SPRITE);
    for (const linha of desenho) {
      expect(linha).toHaveLength(LADO_SPRITE);
      for (const letra of linha) expect(letra === '.' || letra in PALETA, `letra ${letra} fora da paleta`).toBe(true);
    }
    expect(desenho.join('').replaceAll('.', '')).not.toBe(''); // tem algum pixel pintado
  });

  test('CA-01 piso, blocos e chama horizontal cobrem as bordas que encostam na vizinha', () => {
    for (const nome of ['piso', 'bloco-destrutivel'] as const) {
      expect(SPRITES[nome].join('')).not.toContain('.');
    }
    const h = SPRITES['chama-horizontal'];
    const v = SPRITES['chama-vertical'];
    // o braço horizontal encosta nas bordas esquerda e direita na mesma faixa de linhas
    expect(h.map((l) => l[0] !== '.')).toEqual(h.map((l) => l[LADO_SPRITE - 1] !== '.'));
    // o vertical é o horizontal transposto
    expect(v.map((_, y) => h.map((l) => l[y]).join(''))).toEqual([...v]);
  });

  test('CA-01 só o jogador usa currentColor (a cor de cada um)', () => {
    const usam = NOMES_SPRITES.filter((n) => SPRITES[n].join('').includes('c'));
    expect(usam).toEqual(['jogador']);
    expect(PALETA.c).toBe('currentColor');
  });

  test('CA-01 retangulos junta pixels vizinhos de mesma cor e pula os transparentes', () => {
    const desenho = ['kk.kw', '.....', 'wwwww'];
    expect(retangulos(desenho)).toEqual([
      { x: 0, y: 0, largura: 2, cor: PALETA.k },
      { x: 3, y: 0, largura: 1, cor: PALETA.k },
      { x: 4, y: 0, largura: 1, cor: PALETA.w },
      { x: 0, y: 2, largura: 5, cor: PALETA.w },
    ]);
  });

  test.each(NOMES_SPRITES.map((nome) => ({ nome })))('CA-01 $nome: retangulos cobre exatamente os pixels pintados', ({ nome }) => {
    const desenho = SPRITES[nome];
    const pintados = new Set<string>();
    for (const r of retangulos(desenho)) {
      for (let x = r.x; x < r.x + r.largura; x++) {
        const chave = `${x},${r.y}`;
        expect(pintados.has(chave)).toBe(false);
        pintados.add(chave);
        expect(PALETA[desenho[r.y][x]]).toBe(r.cor);
      }
    }
    const esperado = desenho.flatMap((l, y) => [...l].flatMap((c, x) => (c === '.' ? [] : [`${x},${y}`])));
    expect([...pintados].sort()).toEqual(esperado.sort());
  });

  test('CA-01 idSprite', () => {
    expect(idSprite('bomba')).toBe('sprite-bomba');
  });
});
