import { describe, expect, test } from 'vitest';
import { formasDasChamas, type FormaChama } from './chamas';
import { explosao, p } from '../testes/fabricas';
import type { Explosao, Posicao } from './tipos';

const formas = (chamas: Posicao[], explosoes: Explosao[]) => Object.fromEntries(formasDasChamas(chamas, explosoes));
const todas = (es: Explosao[]) => es.flatMap((e) => e.chamas);

describe('formasDasChamas', () => {
  test.each<{ nome: string; explosoes: Explosao[]; extra?: Posicao[]; esperado: Record<string, FormaChama> }>([
    {
      nome: 'CA-03 cruz completa: centro, braços e pontas',
      explosoes: [explosao(p(2, 2), [p(0, 2), p(1, 2), p(3, 2), p(4, 2), p(2, 0), p(2, 1), p(2, 3), p(2, 4)])],
      esperado: {
        '2,2': 'centro',
        '0,2': 'ponta-esquerda',
        '1,2': 'horizontal',
        '3,2': 'horizontal',
        '4,2': 'ponta-direita',
        '2,0': 'ponta-cima',
        '2,1': 'vertical',
        '2,3': 'vertical',
        '2,4': 'ponta-baixo',
      },
    },
    {
      nome: 'CA-03 braço de uma casa só é ponta',
      explosoes: [explosao(p(1, 1), [p(0, 1), p(2, 1), p(1, 0)])],
      esperado: { '1,1': 'centro', '0,1': 'ponta-esquerda', '2,1': 'ponta-direita', '1,0': 'ponta-cima' },
    },
    {
      nome: 'CA-03 explosão sem braços: só o centro',
      explosoes: [explosao(p(3, 3), [])],
      esperado: { '3,3': 'centro' },
    },
    {
      nome: 'CA-03 na borda do tabuleiro: só um lado',
      explosoes: [explosao(p(0, 0), [p(1, 0), p(2, 0)])],
      esperado: { '0,0': 'centro', '1,0': 'horizontal', '2,0': 'ponta-direita' },
    },
    {
      nome: 'CA-03 duas explosões que se cruzam: a casa em comum vira centro',
      explosoes: [explosao(p(0, 1), [p(1, 1), p(2, 1)]), explosao(p(1, 0), [p(1, 1), p(1, 2)])],
      esperado: {
        '0,1': 'centro',
        '1,1': 'centro',
        '2,1': 'ponta-direita',
        '1,0': 'centro',
        '1,2': 'ponta-baixo',
      },
    },
    {
      nome: 'CA-03 duas explosões com a mesma forma na casa em comum mantêm a forma',
      explosoes: [explosao(p(0, 0), [p(1, 0), p(2, 0), p(3, 0)]), explosao(p(4, 0), [p(3, 0), p(2, 0), p(1, 0)])],
      esperado: { '0,0': 'centro', '1,0': 'centro', '2,0': 'horizontal', '3,0': 'centro', '4,0': 'centro' },
    },
    {
      nome: 'CA-03 casa em chamas fora de qualquer explosão é centro',
      explosoes: [],
      extra: [p(5, 5)],
      esperado: { '5,5': 'centro' },
    },
  ])('$nome', ({ explosoes, extra = [], esperado }) => {
    expect(formas([...todas(explosoes), ...extra], explosoes)).toEqual(esperado);
  });

  test('CA-03 só devolve formas para as casas de chamas', () => {
    const e = explosao(p(1, 1), [p(0, 1), p(2, 1)]);
    expect(formas([p(1, 1)], [e])).toEqual({ '1,1': 'centro' });
  });
});
