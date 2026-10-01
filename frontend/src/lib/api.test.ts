import { describe, expect, test } from 'vitest';
import { ErroApi, criarCliente } from './api';
import { resposta, respostaPartidas } from '../testes/fabricas';

function buscarFalso(status: number, corpo: unknown) {
  const pedidos: string[] = [];
  const buscar = (async (url: string | URL | Request) => {
    pedidos.push(String(url));
    return new Response(JSON.stringify(corpo), { status, headers: { 'Content-Type': 'application/json' } });
  }) as typeof fetch;
  return { buscar, pedidos };
}

async function erroDe(p: Promise<unknown>): Promise<ErroApi> {
  try {
    await p;
  } catch (e) {
    if (e instanceof ErroApi) return e;
    throw e;
  }
  throw new Error('esperava ErroApi');
}

describe('cliente da API', () => {
  test('API-05 CA-12 buscarEstado: GET /partidas/{nome}/estado devolve a resposta', async () => {
    const r = resposta();
    const { buscar, pedidos } = buscarFalso(200, r);
    expect(await criarCliente(buscar).buscarEstado('final-1')).toEqual(r);
    expect(pedidos).toEqual(['/partidas/final-1/estado']);
  });

  test('API-05 CA-12 buscarEstado escapa o nome no caminho', async () => {
    const { buscar, pedidos } = buscarFalso(200, resposta());
    await criarCliente(buscar, 'http://servidor').buscarEstado('a/b');
    expect(pedidos).toEqual(['http://servidor/partidas/a%2Fb/estado']);
  });

  test('API-10 CA-13 buscarPartidas: GET /partidas', async () => {
    const r = respostaPartidas();
    const { buscar, pedidos } = buscarFalso(200, r);
    expect(await criarCliente(buscar).buscarPartidas()).toEqual(r);
    expect(pedidos).toEqual(['/partidas']);
  });

  test.each([
    { nome: 'API-05 CA-14 404 com erro', status: 404, corpo: { erro: 'partida "x" não encontrada' }, mensagem: 'partida "x" não encontrada' },
    { nome: 'API-01 CA-15 500 com erro', status: 500, corpo: { erro: 'falhou' }, mensagem: 'falhou' },
    { nome: 'API-01 CA-15 502 sem JSON', status: 502, corpo: 'proxy', mensagem: 'HTTP 502' },
  ])('$nome', async ({ status, corpo, mensagem }) => {
    const buscar = (async () =>
      new Response(typeof corpo === 'string' ? corpo : JSON.stringify(corpo), { status })) as typeof fetch;
    const e = await erroDe(criarCliente(buscar).buscarEstado('x'));
    expect(e.status).toBe(status);
    expect(e.message).toBe(mensagem);
  });

  test('API-01 CA-15 falha de rede vira ErroApi com status 0', async () => {
    const buscar = (async () => {
      throw new TypeError('Failed to fetch');
    }) as typeof fetch;
    const e = await erroDe(criarCliente(buscar).buscarEstado('x'));
    expect(e.status).toBe(0);
  });
});
