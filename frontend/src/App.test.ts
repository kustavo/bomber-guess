import { render } from '@testing-library/svelte';
import { afterEach, describe, expect, test } from 'vitest';
import App from './App.svelte';
import { clienteFalso, descarregar, relogioFalso, resposta, respostaPartidas } from './testes/fabricas';
import type { Desfecho } from './lib/tipos';

const vitoria: Desfecho = { terminada: true, empate: false, vencedor: 'jogador_1', sobreviventes: ['jogador_1'] };
const andamento: Desfecho = { terminada: false, empate: false, sobreviventes: ['jogador_1', 'jogador_2'] };

function iniciar() {
  const relogio = relogioFalso();
  const cliente = clienteFalso(
    () => resposta(),
    () =>
      respostaPartidas({
        partidas: [
          { nome: 'zeta', fase: 'ENCERRADA', turno: 12, desfecho: vitoria },
          { nome: 'alfa', fase: 'PLANEJAMENTO', turno: 3, fim_da_fase: '2026-10-01T12:00:01.000Z', desfecho: andamento },
        ],
      }),
  );
  return { relogio, cliente };
}

afterEach(() => {
  window.location.hash = '';
});

describe('App', () => {
  test('API-10 CA-13 lista as partidas na ordem recebida, com fase, turno e desfecho', async () => {
    const { relogio, cliente } = iniciar();
    const { container, unmount } = render(App, { cliente, relogio });
    await relogio.avancar(0);
    const linhas = [...container.querySelectorAll('tbody tr')].map((tr) =>
      [...tr.querySelectorAll('td')].map((td) => td.textContent),
    );
    expect(linhas).toEqual([
      ['zeta', 'Encerrada', '12', 'Vitória de jogador_1'],
      ['alfa', 'Planejamento', '3', ''],
    ]);
    expect(container.querySelector('[data-partida="alfa"] a')?.getAttribute('href')).toBe('#/partidas/alfa');
    await relogio.avancar(2000);
    expect(cliente.chamadas).toEqual(['partidas', 'partidas']); // a cada 2 s
    unmount();
  });

  test('API-10 CA-13 escolher uma partida abre a tela dela', async () => {
    const { relogio, cliente } = iniciar();
    const { container, unmount } = render(App, { cliente, relogio });
    await relogio.avancar(0);
    window.location.hash = '#/partidas/alfa';
    window.dispatchEvent(new HashChangeEvent('hashchange'));
    await descarregar();
    await relogio.avancar(0);
    expect(container.querySelector('h1')?.textContent).toBe('alfa');
    expect(cliente.chamadas).toContain('estado alfa');
    unmount();
  });

  test('API-10 CA-13 endereço da partida aberto diretamente', async () => {
    window.location.hash = '#/partidas/final-1';
    const { relogio, cliente } = iniciar();
    const { container, unmount } = render(App, { cliente, relogio });
    await relogio.avancar(0);
    expect(container.querySelector('h1')?.textContent).toBe('final-1');
    expect(cliente.chamadas).toEqual(['estado final-1']);
    expect(container.querySelector('svg')).not.toBeNull();
    unmount();
  });

  test('API-10 CA-13 lista vazia', async () => {
    const relogio = relogioFalso();
    const cliente = clienteFalso(() => resposta(), () => respostaPartidas());
    const { container, unmount } = render(App, { cliente, relogio });
    await relogio.avancar(0);
    expect(container.textContent).toContain('Nenhuma partida');
    unmount();
  });

  test('API-04 CA-17 o link "Criar partida" da lista abre o editor', async () => {
    const { relogio, cliente } = iniciar();
    const { container, unmount } = render(App, { cliente, relogio });
    await relogio.avancar(0);
    const link = container.querySelector<HTMLAnchorElement>('[data-acao="criar-partida"]');
    expect(link?.getAttribute('href')).toBe('#/criar');
    window.location.hash = link!.getAttribute('href')!;
    window.dispatchEvent(new HashChangeEvent('hashchange'));
    await descarregar();
    expect(container.querySelector('h1')?.textContent).toBe('Criar partida');
    expect(container.querySelector('[data-casa]')).not.toBeNull();
    unmount();
  });

  test('API-04 CA-17 #/criar aberto diretamente mostra o editor vazio', async () => {
    window.location.hash = '#/criar';
    const { relogio, cliente } = iniciar();
    const { container, unmount } = render(App, { cliente, relogio });
    await descarregar();
    expect(container.querySelector('h1')?.textContent).toBe('Criar partida');
    expect(container.querySelector('[data-tipo]')).toBeNull();
    expect(cliente.chamadas).toEqual(['bots']);
    unmount();
  });
});
