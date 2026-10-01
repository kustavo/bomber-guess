import { describe, expect, test } from 'vitest';
import { CONSULTA_MAXIMA_MS, intervaloConsulta, intervaloEntreEtapas } from './ritmo';

describe('ritmo', () => {
  test.each([
    { nome: 'API-01 CA-08 antes da primeira resposta: 250 ms', d: undefined, esperado: 250 },
    { nome: 'API-01 CA-08 etapa de 1 s: 250 ms', d: 1000, esperado: 250 },
    { nome: 'API-01 CA-08 etapa de 200 ms: metade, 100 ms', d: 200, esperado: 100 },
    { nome: 'API-01 CA-08 etapa de 0 ms: ainda consulta (mínimo 1 ms)', d: 0, esperado: 1 },
  ])('$nome', ({ d, esperado }) => {
    expect(intervaloConsulta(d)).toBe(esperado);
    expect(intervaloConsulta(d)).toBeLessThanOrEqual(CONSULTA_MAXIMA_MS);
  });

  test.each([
    { nome: 'EST-02 CA-10 em dia (1 etapa a exibir): d', atrasadas: 1, d: 1000, esperado: 1000 },
    { nome: 'EST-02 CA-10 atrasado (3 etapas a exibir): d/4', atrasadas: 3, d: 1000, esperado: 250 },
    { nome: 'EST-02 CA-10 nada a exibir: d', atrasadas: 0, d: 800, esperado: 800 },
  ])('$nome', ({ atrasadas, d, esperado }) => {
    const i = intervaloEntreEtapas(atrasadas, d);
    expect(i).toBe(esperado);
    expect(i).toBeLessThanOrEqual(d); // nunca mais lento que duracao_etapa_ms
  });
});
