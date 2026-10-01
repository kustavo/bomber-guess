// Acompanhamento de uma partida: consulta periódica do estado (API-01, API-05), fila de
// etapas a exibir e erros (spec 06, CA-08 a CA-12, CA-14, CA-15; plano, D2, D4, D6, D7).
import { writable, type Readable } from 'svelte/store';
import { ErroApi, type ClienteApi } from './api';
import { calcularDefasagem } from './cronometro';
import { intervaloConsulta, intervaloEntreEtapas } from './ritmo';
import { calcularTabuleiro, type TabuleiroExibido } from './tabuleiro';
import type { RespostaEstado } from './tipos';

// Relogio é a fonte de tempo; os testes usam um relógio controlado (plano, D2).
export interface Relogio {
  agora(): number;
  esperar(ms: number, f: () => void): () => void; // devolve a função que cancela
}

export const relogioReal: Relogio = {
  agora: () => Date.now(),
  esperar(ms, f) {
    const id = setTimeout(f, ms);
    return () => clearTimeout(id);
  },
};

export interface VisaoTela {
  carregando: boolean; // nenhuma resposta ainda
  naoEncontrada: boolean; // CA-14
  semConexao: boolean; // CA-15
  erro?: string;
  resposta?: RespostaEstado; // a última válida
  tabuleiro?: TabuleiroExibido; // o da etapa exibida, que pode estar atrás da resposta
  defasagem: number; // servidor − navegador, em ms (CA-05)
}

export interface Acompanhamento {
  visao: Readable<VisaoTela>;
  parar(): void;
}

export function acompanhar(nome: string, cliente: ClienteApi, relogio: Relogio = relogioReal): Acompanhamento {
  let visao: VisaoTela = { carregando: true, naoEncontrada: false, semConexao: false, defasagem: 0 };
  const loja = writable(visao);
  let parado = false;
  let cancelarConsulta: (() => void) | undefined;
  let cancelarEtapa: (() => void) | undefined;
  let etapaExibida = 0;

  const publicar = (mudancas: Partial<VisaoTela>) => {
    visao = { ...visao, ...mudancas };
    loja.set(visao);
  };

  const duracaoEtapa = () => visao.resposta?.estado.config.duracao_etapa_ms;

  const exibir = (etapa: number) => {
    const r = visao.resposta!;
    etapaExibida = etapa;
    publicar({ tabuleiro: calcularTabuleiro(r.estado, r.etapas, etapa) });
  };

  // avancarEtapa exibe a próxima etapa liberada, se houver, e agenda a seguinte (D4).
  const avancarEtapa = () => {
    cancelarEtapa = undefined;
    const liberadas = visao.resposta?.etapas.length ?? 0;
    if (parado || etapaExibida >= liberadas) return;
    exibir(etapaExibida + 1);
    cancelarEtapa = relogio.esperar(intervaloEntreEtapas(liberadas - etapaExibida, duracaoEtapa() ?? 0), avancarEtapa);
  };

  const recebida = (r: RespostaEstado, chegada: number) => {
    const anterior = visao.resposta;
    publicar({
      carregando: false,
      semConexao: false,
      erro: undefined,
      resposta: r,
      defasagem: calcularDefasagem(r.horario_servidor, chegada),
    });
    if (!anterior) {
      exibir(r.etapas.length); // CA-12: a primeira resposta basta
    } else if (r.turno !== anterior.turno) {
      cancelarEtapa?.(); // CA-11: turno novo descarta o que faltava do anterior
      cancelarEtapa = undefined;
      exibir(0);
      avancarEtapa();
    } else if (!cancelarEtapa) {
      avancarEtapa();
    }
  };

  const falhou = (e: unknown) => {
    const erro = e instanceof ErroApi ? e : new ErroApi(0, String(e));
    if (erro.status === 404) {
      publicar({ carregando: false, naoEncontrada: true, erro: erro.message });
    } else if (erro.status === 0 || erro.status >= 500) {
      publicar({ carregando: visao.resposta === undefined, semConexao: true });
    } else {
      publicar({ carregando: visao.resposta === undefined, erro: erro.message });
    }
  };

  const consultar = async () => {
    cancelarConsulta = undefined;
    try {
      const r = await cliente.buscarEstado(nome);
      if (parado) return;
      recebida(r, relogio.agora());
    } catch (e) {
      if (parado) return;
      falhou(e);
    }
    if (visao.naoEncontrada || visao.resposta?.fase === 'ENCERRADA') return; // CA-09, CA-14
    cancelarConsulta = relogio.esperar(intervaloConsulta(duracaoEtapa()), consultar); // D7
  };

  void consultar();

  return {
    visao: { subscribe: loja.subscribe },
    parar() {
      parado = true;
      cancelarConsulta?.();
      cancelarEtapa?.();
    },
  };
}
