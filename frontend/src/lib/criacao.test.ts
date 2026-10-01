import { describe, expect, test } from 'vitest';
import { ErroApi } from './api';
import { chaveMapa, iniciarPartida } from './criacao';
import { clienteFalso, p } from '../testes/fabricas';
import type { Mapa, PedidoPartida } from './tipos';

const mapa: Mapa = {
  nome: 'novo',
  config: { largura: 3, altura: 1, limite_turnos: 5, prazo_planejamento_ms: 1000, duracao_etapa_ms: 1000 },
  jogador_padrao: { bombas_por_turno: 1, potencia: 1, pavio_padrao: 3, acoes_por_turno: 3 },
  blocos_fixos: [],
  blocos_destrutiveis: [],
  posicoes_iniciais: [p(0, 0), p(2, 0)],
};
const pedido: PedidoPartida = { nome: 'final-1', mapa: 'novo', bots: ['v1', 'v2'], semente: 1 };

const falha = (status: number, mensagem: string) => async (): Promise<never> => {
  throw new ErroApi(status, mensagem);
};

describe('iniciarPartida', () => {
  test('API-03 API-04 MAP-03 CA-14 grava o mapa e depois cria a partida', async () => {
    const cliente = clienteFalso();
    expect(await iniciarPartida(cliente, mapa, pedido, undefined)).toEqual({ ok: true });
    expect(cliente.chamadas).toEqual(['salvar mapa novo', 'criar partida final-1']);
    expect(cliente.mapas).toEqual([mapa]);
    expect(cliente.pedidos).toEqual([pedido]);
  });

  test.each([
    { nome: 'API-03 CA-15 400 no POST /mapas', status: 400, erro: 'pedido inválido: mapa inválido: largura 0' },
    { nome: 'API-03 CA-15 409 no POST /mapas', status: 409, erro: 'nome já usado: mapa "novo"' },
    { nome: 'API-03 CA-15 servidor fora do ar no POST /mapas', status: 0, erro: 'servidor indisponível' },
  ])('$nome: mostra o erro e não cria a partida', async ({ status, erro }) => {
    const cliente = clienteFalso(undefined, undefined, {
      salvarMapa: falha(status, status === 0 ? 'Failed to fetch' : erro),
    });
    const r = await iniciarPartida(cliente, mapa, pedido, undefined);
    expect(r).toEqual({ ok: false, erro, mapaSalvo: undefined });
    expect(cliente.chamadas).toEqual(['salvar mapa novo']);
  });

  test.each([
    { nome: 'API-04 CA-15 409 no POST /partidas', status: 409, erro: 'nome já usado: "final-1"' },
    { nome: 'API-04 CA-15 servidor fora do ar no POST /partidas', status: 0, erro: 'servidor indisponível' },
  ])('$nome: devolve o mapa já salvo', async ({ status, erro }) => {
    const cliente = clienteFalso(undefined, undefined, {
      criarPartida: falha(status, status === 0 ? 'Failed to fetch' : erro),
    });
    const r = await iniciarPartida(cliente, mapa, pedido, undefined);
    expect(r).toEqual({ ok: false, erro, mapaSalvo: chaveMapa(mapa) });
  });

  test('API-03 API-04 CA-16 nova tentativa com o mesmo mapa não grava de novo (decisão 5)', async () => {
    const cliente = clienteFalso();
    const r = await iniciarPartida(cliente, mapa, { ...pedido, nome: 'final-2' }, chaveMapa(mapa));
    expect(r).toEqual({ ok: true });
    expect(cliente.chamadas).toEqual(['criar partida final-2']);
  });

  test('API-03 CA-16 mapa alterado depois de salvo é gravado de novo', async () => {
    const cliente = clienteFalso();
    const alterado = { ...mapa, blocos_fixos: [p(1, 0)] };
    await iniciarPartida(cliente, alterado, pedido, chaveMapa(mapa));
    expect(cliente.chamadas).toEqual(['salvar mapa novo', 'criar partida final-1']);
  });

  test('API-03 CA-16 erro na nova tentativa da partida mantém o mapa salvo', async () => {
    const cliente = clienteFalso(undefined, undefined, { criarPartida: falha(409, 'repetido') });
    const r = await iniciarPartida(cliente, mapa, pedido, chaveMapa(mapa));
    expect(r).toEqual({ ok: false, erro: 'repetido', mapaSalvo: chaveMapa(mapa) });
  });

  test('CA-16 chaveMapa muda com qualquer campo do mapa', () => {
    expect(chaveMapa({ ...mapa })).toBe(chaveMapa(mapa));
    expect(chaveMapa({ ...mapa, nome: 'outro' })).not.toBe(chaveMapa(mapa));
    expect(chaveMapa({ ...mapa, config: { ...mapa.config, limite_turnos: 6 } })).not.toBe(chaveMapa(mapa));
  });
});
