// Cliente HTTP da API do marco 5 (API-05, API-10). Só leitura.
import type { RespostaEstado, RespostaPartidas } from './tipos';

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
}

// criarCliente usa `buscar` (fetch) contra `base`; sem base, usa caminhos relativos,
// que o proxy do servidor de desenvolvimento encaminha ao servidor Go.
export function criarCliente(buscar: typeof fetch = (...a) => fetch(...a), base = ''): ClienteApi {
  async function obter<T>(caminho: string): Promise<T> {
    let r: Response;
    try {
      r = await buscar(base + caminho);
    } catch (e) {
      throw new ErroApi(0, e instanceof Error ? e.message : String(e));
    }
    let corpo: unknown;
    try {
      corpo = await r.json();
    } catch {
      corpo = undefined;
    }
    if (!r.ok) {
      const erro = (corpo as { erro?: unknown } | undefined)?.erro;
      throw new ErroApi(r.status, typeof erro === 'string' ? erro : `HTTP ${r.status}`);
    }
    if (corpo === undefined) throw new ErroApi(0, 'resposta sem JSON');
    return corpo as T;
  }

  return {
    buscarPartidas: () => obter<RespostaPartidas>('/partidas'),
    buscarEstado: (nome) => obter<RespostaEstado>(`/partidas/${encodeURIComponent(nome)}/estado`),
  };
}
