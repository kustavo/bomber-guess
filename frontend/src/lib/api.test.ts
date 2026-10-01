import { describe, expect, test } from 'vitest';
import { ErroApi, criarCliente } from './api';
import { resposta, respostaPartidas } from '../testes/fabricas';
import type { Mapa } from './tipos';

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

  test('API-09 CA-12 buscarBots: GET /bots', async () => {
    const r = { bots: ['aleatorio-v1', 'aleatorio-v2'], horario_servidor: '2026-10-01T12:00:00.000Z' };
    const { buscar, pedidos } = buscarFalso(200, r);
    expect(await criarCliente(buscar).buscarBots()).toEqual(r);
    expect(pedidos).toEqual(['/bots']);
  });

  test('API-03 CA-14 salvarMapa: POST /mapas com o mapa em JSON', async () => {
    const mapa: Mapa = {
      nome: 'novo',
      config: { largura: 3, altura: 1, limite_turnos: 5, prazo_planejamento_ms: 1000, duracao_etapa_ms: 1000 },
      jogador_padrao: { bombas_por_turno: 1, potencia: 1, pavio_padrao: 3, acoes_por_turno: 3 },
      blocos_fixos: [],
      blocos_destrutiveis: [],
      posicoes_iniciais: [{ x: 0, y: 0 }, { x: 2, y: 0 }],
    };
    const { buscar, envios } = enviarFalso(201, { nome: 'novo', horario_servidor: '2026-10-01T12:00:00.000Z' });
    expect((await criarCliente(buscar).salvarMapa(mapa)).nome).toBe('novo');
    expect(envios).toEqual([{ url: '/mapas', metodo: 'POST', tipo: 'application/json', corpo: mapa }]);
  });

  test('API-04 CA-14 criarPartida: POST /partidas com o pedido em JSON', async () => {
    const pedido = { nome: 'final-1', mapa: 'novo', bots: ['a', 'b'], semente: 1 };
    const { buscar, envios } = enviarFalso(201, resposta());
    expect(await criarCliente(buscar).criarPartida(pedido)).toEqual(resposta());
    expect(envios).toEqual([{ url: '/partidas', metodo: 'POST', tipo: 'application/json', corpo: pedido }]);
  });

  test('API-04 CA-15 409 em POST /partidas vira ErroApi com o erro do servidor', async () => {
    const { buscar } = enviarFalso(409, { erro: 'nome já usado: "x"' });
    const e = await erroDe(criarCliente(buscar).criarPartida({ nome: 'x', mapa: 'y', bots: [], semente: 1 }));
    expect(e.status).toBe(409);
    expect(e.message).toBe('nome já usado: "x"');
  });
});

// enviarFalso registra método, Content-Type e corpo de cada requisição.
function enviarFalso(status: number, corpo: unknown) {
  const envios: { url: string; metodo: string; tipo: string | null; corpo: unknown }[] = [];
  const buscar = (async (url: string | URL | Request, init?: RequestInit) => {
    envios.push({
      url: String(url),
      metodo: init?.method ?? 'GET',
      tipo: new Headers(init?.headers).get('Content-Type'),
      corpo: JSON.parse(String(init?.body)),
    });
    return new Response(JSON.stringify(corpo), { status });
  }) as typeof fetch;
  return { buscar, envios };
}
