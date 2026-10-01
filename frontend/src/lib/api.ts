// Cliente HTTP da API do marco 5 (API-05, API-10) e da criação de partidas
// do marco 7 (API-03, API-04, API-09).
import type {
  Mapa,
  PedidoPartida,
  RespostaBots,
  RespostaEstado,
  RespostaMapa,
  RespostaPartidas,
} from './tipos';

// ErroApi é uma resposta de erro da API (corpo {"erro": ...}) ou uma falha de rede (status 0).
export class ErroApi extends Error {
  readonly status: number;

  constructor(status: number, mensagem: string) {
    super(mensagem);
    this.name = 'ErroApi';
    this.status = status;
  }
}

export interface ClienteApi {
  buscarPartidas(): Promise<RespostaPartidas>;
  buscarEstado(nome: string): Promise<RespostaEstado>;
  buscarBots(): Promise<RespostaBots>;
  salvarMapa(mapa: Mapa): Promise<RespostaMapa>;
  criarPartida(pedido: PedidoPartida): Promise<RespostaEstado>;
}

// criarCliente usa `buscar` (fetch) contra `base`; sem base, usa caminhos relativos,
// que o proxy do servidor de desenvolvimento encaminha ao servidor Go.
export function criarCliente(buscar: typeof fetch = (...a) => fetch(...a), base = ''): ClienteApi {
  // pedir faz GET, ou POST com `corpo` em JSON.
  async function pedir<T>(caminho: string, corpo?: unknown): Promise<T> {
    const init: RequestInit | undefined =
      corpo === undefined
        ? undefined
        : { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(corpo) };
    let r: Response;
    try {
      r = await buscar(base + caminho, init);
    } catch (e) {
      throw new ErroApi(0, e instanceof Error ? e.message : String(e));
    }
    let resposta: unknown;
    try {
      resposta = await r.json();
    } catch {
      resposta = undefined;
    }
    if (!r.ok) {
      const erro = (resposta as { erro?: unknown } | undefined)?.erro;
      throw new ErroApi(r.status, typeof erro === 'string' ? erro : `HTTP ${r.status}`);
    }
    if (resposta === undefined) throw new ErroApi(0, 'resposta sem JSON');
    return resposta as T;
  }

  return {
    buscarPartidas: () => pedir<RespostaPartidas>('/partidas'),
    buscarEstado: (nome) => pedir<RespostaEstado>(`/partidas/${encodeURIComponent(nome)}/estado`),
    buscarBots: () => pedir<RespostaBots>('/bots'),
    salvarMapa: (mapa) => pedir<RespostaMapa>('/mapas', mapa),
    criarPartida: (pedido) => pedir<RespostaEstado>('/partidas', pedido),
  };
}
