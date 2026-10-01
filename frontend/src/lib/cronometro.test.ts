import { describe, expect, test } from 'vitest';
import { calcularDefasagem, tempoRestante } from './cronometro';

const ms = (iso: string) => Date.parse(iso);
const HORA = 3_600_000;

describe('cronômetro', () => {
  test('API-02 CA-05 defasagem é horario_servidor menos o instante local de chegada', () => {
    expect(calcularDefasagem('2026-10-01T12:00:00.250Z', ms('2026-10-01T12:00:00.000Z'))).toBe(250);
    expect(calcularDefasagem('2026-10-01T12:00:00.000Z', ms('2026-10-01T12:00:00.400Z'))).toBe(-400);
  });

  test.each([
    { nome: 'API-02 CA-05 no instante da chegada', depois: 0, restante: 750 },
    { nome: 'API-02 CA-05 500 ms depois da chegada', depois: 500, restante: 250 },
    { nome: 'API-02 CA-05 exatamente no fim', depois: 750, restante: 0 },
    { nome: 'API-02 CA-05 depois do fim: nunca negativo', depois: 2000, restante: 0 },
  ])('$nome', ({ depois, restante }) => {
    const local = ms('2026-10-01T12:00:00.000Z');
    const defasagem = calcularDefasagem('2026-10-01T12:00:00.250Z', local);
    expect(tempoRestante('2026-10-01T12:00:01.000Z', defasagem, local + depois)).toBe(restante);
  });

  test.each([
    { nome: 'API-02 CA-06 relógios iguais', desvio: 0 },
    { nome: 'API-02 CA-06 navegador 1 h adiantado', desvio: HORA },
    { nome: 'API-02 CA-06 navegador 1 h atrasado', desvio: -HORA },
  ])('$nome', ({ desvio }) => {
    const servidor = '2026-10-01T12:00:00.000Z';
    const local = ms(servidor) + desvio;
    const defasagem = calcularDefasagem(servidor, local);
    expect(tempoRestante('2026-10-01T12:00:03.000Z', defasagem, local + 1200)).toBe(1800);
  });

  test('API-02 CA-07 sem fim de fase (ENCERRADA): sem cronômetro', () => {
    expect(tempoRestante(undefined, 0, 0)).toBeNull();
  });
});
