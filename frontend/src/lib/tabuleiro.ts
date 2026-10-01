// Tabuleiro exibido: o que a tela mostra num instante, derivado do estado do início
// do turno e dos relatórios de etapa já liberados (spec 06, CA-01 a CA-04; plano, D3).
import type { Acao, Bomba, Estado, Explosao, Morte, Posicao, RelatorioEtapa, ResultadoAcao, Status } from './tipos';

export interface JogadorExibido {
  id: string;
  indice: number; // ordem em estado.jogadores; define a cor
  posicao: Posicao;
  status: Status;
  morte?: Morte;
  bot_versao: string;
  acao?: Acao; // ausente no início do turno
  resultado?: ResultadoAcao;
  bloqueado: boolean;
}

export interface TabuleiroExibido {
  largura: number;
  altura: number;
  turno: number;
  etapa: number; // 0 = início do turno
  blocosFixos: Posicao[];
  blocosDestrutiveis: Posicao[];
  bombas: Bomba[];
  chamas: Posicao[];
  explosoes: Explosao[]; // só as da etapa exibida, como as chamas (spec 14, D7)
  jogadores: JogadorExibido[]; // todos; só os VIVO ocupam casa (FIM-01)
}

const chave = (p: Posicao) => `${p.x},${p.y}`;

// calcularTabuleiro devolve o tabuleiro depois da etapa `ate` (0 = início do turno),
// limitada aos relatórios recebidos. Recalcula do zero; não altera as entradas.
export function calcularTabuleiro(estado: Estado, etapas: RelatorioEtapa[], ate: number): TabuleiroExibido {
  const k = Math.max(0, Math.min(ate, etapas.length));
  const liberadas = etapas.slice(0, k);
  const atual = liberadas[k - 1];

  const destruidos = new Set(liberadas.flatMap((r) => r.blocos_destruidos.map(chave)));
  const fechados = liberadas.flatMap((r) => r.blocos_fechados); // FEC-03, D13
  const comFechamento = new Set(fechados.map(chave));
  const bloqueados = new Set(atual?.movimentos_bloqueados ?? []);

  const jogadores = estado.jogadores.map((j, indice): JogadorExibido => {
    const naEtapa = atual?.jogadores.find((je) => je.id === j.id);
    let morte = j.morte ? { ...j.morte } : undefined;
    if (!morte) {
      const r = liberadas.find((r) => r.mortes.includes(j.id));
      if (r) morte = { turno: r.turno, etapa: r.etapa };
    }
    return {
      id: j.id,
      indice,
      posicao: { ...(naEtapa?.posicao ?? j.posicao) },
      status: naEtapa?.status ?? j.status,
      morte,
      bot_versao: j.bot_versao,
      acao: naEtapa ? { ...naEtapa.acao } : undefined,
      resultado: naEtapa?.resultado,
      bloqueado: bloqueados.has(j.id),
    };
  });

  return {
    largura: estado.config.largura,
    altura: estado.config.altura,
    turno: estado.turno,
    etapa: k,
    blocosFixos: [...estado.blocos_fixos, ...fechados].map((p) => ({ ...p })),
    blocosDestrutiveis: estado.blocos_destrutiveis
      .filter((p) => !destruidos.has(chave(p)) && !comFechamento.has(chave(p))) // FEC-05
      .map((p) => ({ ...p })),
    bombas: (atual?.bombas ?? estado.bombas).map((b) => ({ ...b, posicao: { ...b.posicao } })),
    chamas: (atual?.chamas ?? []).map((p) => ({ ...p })),
    explosoes: (atual?.explosoes ?? []).map((e) => ({
      origem: { ...e.origem },
      potencia: e.potencia,
      chamas: e.chamas.map((p) => ({ ...p })),
    })),
    jogadores,
  };
}
