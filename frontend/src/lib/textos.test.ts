import { describe, expect, test } from 'vitest';
import { formatarTempo, textoAcao, textoDesfecho, textoFase } from './textos';

describe('textos', () => {
  test.each([
    { nome: 'FIM-02 CA-07 vitória', d: { terminada: true, empate: false, vencedor: 'jogador_3', sobreviventes: ['jogador_3'] }, texto: 'Vitória de jogador_3' },
    { nome: 'FIM-03 CA-07 empate sem sobreviventes', d: { terminada: true, empate: true, sobreviventes: [] }, texto: 'Empate' },
    { nome: 'FIM-04 CA-07 empate por limite de turnos', d: { terminada: true, empate: true, sobreviventes: ['jogador_1', 'jogador_2'] }, texto: 'Empate entre jogador_1, jogador_2' },
    { nome: 'DEC-06 CA-13 em andamento', d: { terminada: false, empate: false, sobreviventes: ['jogador_1', 'jogador_2'] }, texto: 'Em andamento' },
  ])('$nome', ({ d, texto }) => {
    expect(textoDesfecho(d)).toBe(texto);
  });

  test.each([
    { nome: 'API-02 CA-05 zero', ms: 0, texto: '0:00.0' },
    { nome: 'API-02 CA-05 décimos truncados', ms: 1234, texto: '0:01.2' },
    { nome: 'API-02 CA-05 minutos', ms: 65_900, texto: '1:05.9' },
  ])('$nome', ({ ms, texto }) => {
    expect(formatarTempo(ms)).toBe(texto);
  });

  test('API-05 CA-12 fases e ações em português', () => {
    expect(textoFase('EXECUCAO')).toBe('Execução');
    expect(textoAcao({ etapa: 1, tipo: 'MOVER', direcao: 'CIMA' })).toBe('Mover ↑');
    expect(textoAcao({ etapa: 1, tipo: 'PLANTAR' })).toBe('Plantar');
  });
});
