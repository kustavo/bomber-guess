// Textos da tela em português, com os termos do glossário (spec 06, decisão 7).
import type { Acao, Desfecho, Fase, ResultadoAcao } from './tipos';

const fases: Record<Fase, string> = {
  PLANEJAMENTO: 'Planejamento',
  EXECUCAO: 'Execução',
  ENCERRADA: 'Encerrada',
};

export const textoFase = (f: Fase) => fases[f];

export function textoDesfecho(d: Desfecho): string {
  if (!d.terminada) return 'Em andamento';
  if (d.empate) return d.sobreviventes.length > 0 ? `Empate entre ${d.sobreviventes.join(', ')}` : 'Empate';
  return `Vitória de ${d.vencedor}`;
}

const direcoes = { CIMA: '↑', BAIXO: '↓', ESQUERDA: '←', DIREITA: '→' } as const;

export function textoAcao(a: Acao): string {
  if (a.tipo === 'MOVER') return `Mover ${a.direcao ? direcoes[a.direcao] : ''}`.trim();
  return a.tipo === 'PLANTAR' ? 'Plantar' : 'Esperar';
}

const resultados: Record<ResultadoAcao, string> = {
  EXECUTADA: 'executada',
  BLOQUEADA: 'bloqueada',
  ABORTADA: 'abortada',
  DESCARTADA: 'descartada',
  IGNORADA: 'ignorada',
};

export const textoResultado = (r: ResultadoAcao) => resultados[r];

// formatarTempo: ms → "m:ss.d" (décimos truncados).
export function formatarTempo(ms: number): string {
  const decimos = Math.floor(Math.max(0, ms) / 100);
  const s = Math.floor(decimos / 10);
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}.${decimos % 10}`;
}

// Cor de cada jogador pela ordem em estado.jogadores (plano, D8).
const cores = ['#e5484d', '#3e8ef7', '#30a46c', '#f5a524'];
export const corJogador = (indice: number) => cores[indice % cores.length];
