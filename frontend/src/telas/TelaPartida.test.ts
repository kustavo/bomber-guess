import { render } from '@testing-library/svelte';
import { describe, expect, test } from 'vitest';
import TelaPartida from './TelaPartida.svelte';
import { ErroApi } from '../lib/api';
import type { RespostaEstado } from '../lib/tipos';
import { clienteFalso, relatorio, relogioFalso, resposta } from '../testes/fabricas';

async function abrir(respostas: (RespostaEstado | Error)[]) {
  const relogio = relogioFalso();
  const cliente = clienteFalso((n) => {
    const r = respostas[Math.min(n, respostas.length) - 1];
    if (r instanceof Error) throw r;
    return r;
  });
  const tela = render(TelaPartida, { nome: 'final-1', cliente, relogio });
  await relogio.avancar(0);
  const campo = (c: string) => tela.container.querySelector(`[data-campo="${c}"]`)?.textContent;
  const aviso = (a: string) => tela.container.querySelector(`[data-aviso="${a}"]`);
  return { tela, relogio, cliente, campo, aviso };
}

describe('TelaPartida', () => {
  test('API-02 CA-12 primeira resposta: turno, etapa, fase e cronômetro', async () => {
    // Servidor 12:00:00.250 quando o navegador marca 12:00:00.000; fase até 12:00:01.000.
    const r = resposta({
      fase: 'EXECUCAO',
      etapa: 2,
      etapas: [relatorio(1), relatorio(2)],
      horario_servidor: '2026-10-01T12:00:00.250Z',
    });
    const { tela, relogio, campo } = await abrir([r]);
    expect(campo('turno')).toBe('Turno 1');
    expect(campo('etapa')).toBe('Etapa 2 de 3');
    expect(campo('fase')).toBe('Execução');
    expect(tela.container.querySelector('.cronometro')?.textContent).toBe('0:00.7');
    await relogio.avancar(200); // antes da 2ª consulta, que recalcula a defasagem
    expect(tela.container.querySelector('.cronometro')?.textContent).toBe('0:00.5');
    expect(tela.container.querySelectorAll('[data-tipo="jogador"]')).toHaveLength(2);
    tela.unmount();
  });

  test('API-02 CA-07 ENCERRADA: sem cronômetro, com o desfecho', async () => {
    const r = resposta({
      fase: 'ENCERRADA',
      fim_da_fase: undefined,
      desfecho: { terminada: true, empate: false, vencedor: 'jogador_2', sobreviventes: ['jogador_2'] },
    });
    const { tela, campo } = await abrir([r]);
    expect(tela.container.querySelector('.cronometro')).toBeNull();
    expect(campo('fase')).toBe('Encerrada');
    expect(campo('desfecho')).toBe('Vitória de jogador_2');
    tela.unmount();
  });

  test('API-05 CA-14 partida inexistente: mensagem e nada de tabuleiro', async () => {
    const { tela, relogio, cliente, aviso } = await abrir([new ErroApi(404, 'partida "final-1" não encontrada')]);
    expect(aviso('nao-encontrada')?.textContent).toBe('Partida não encontrada.');
    expect(tela.container.querySelector('svg')).toBeNull();
    await relogio.avancar(2000);
    expect(cliente.chamadas).toHaveLength(1);
    tela.unmount();
  });

  test('API-01 CA-15 sem conexão: mantém o tabuleiro, avisa e o aviso some', async () => {
    const r = resposta();
    const { tela, relogio, aviso } = await abrir([r, new ErroApi(0, 'Failed to fetch'), r]);
    expect(aviso('sem-conexao')).toBeNull();
    await relogio.avancar(250);
    expect(aviso('sem-conexao')).not.toBeNull();
    expect(tela.container.querySelector('svg')).not.toBeNull();
    await relogio.avancar(250);
    expect(aviso('sem-conexao')).toBeNull();
    tela.unmount();
  });

  test('API-01 CA-08 ao sair da tela, a consulta para', async () => {
    const { tela, relogio, cliente } = await abrir([resposta()]);
    tela.unmount();
    await relogio.avancar(2000);
    expect(cliente.chamadas).toHaveLength(1);
  });
});
