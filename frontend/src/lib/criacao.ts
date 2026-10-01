// Criação de partida pelo editor (spec 07, CA-14 a CA-16): POST /mapas e depois
// POST /partidas, lembrando o mapa já gravado (decisão 5, D8).
import { ErroApi, type ClienteApi } from './api';
import type { Mapa, PedidoPartida } from './tipos';

// mapaSalvo é a chave do último mapa gravado com sucesso nesta tela.
export type ResultadoInicio = { ok: true } | { ok: false; erro: string; mapaSalvo: string | undefined };

export const chaveMapa = (mapa: Mapa): string => JSON.stringify(mapa);

function mensagem(e: unknown): string {
  if (e instanceof ErroApi && e.status !== 0) return e.message;
  return 'servidor indisponível';
}

export async function iniciarPartida(
  cliente: ClienteApi,
  mapa: Mapa,
  pedido: PedidoPartida,
  mapaSalvo: string | undefined,
): Promise<ResultadoInicio> {
  const chave = chaveMapa(mapa);
  if (mapaSalvo !== chave) {
    try {
      await cliente.salvarMapa(mapa);
    } catch (e) {
      return { ok: false, erro: mensagem(e), mapaSalvo };
    }
  }
  try {
    await cliente.criarPartida(pedido);
  } catch (e) {
    return { ok: false, erro: mensagem(e), mapaSalvo: chave };
  }
  return { ok: true };
}
