// Ritmo da consulta e da animação (spec 06, decisões 2 e 5).

export const CONSULTA_MAXIMA_MS = 250;
export const CONSULTA_LISTA_MS = 2000;

// intervaloConsulta: min(250 ms, duracao_etapa_ms / 2); 250 ms antes de saber a duração.
export function intervaloConsulta(duracaoEtapaMs?: number): number {
  if (duracaoEtapaMs === undefined) return CONSULTA_MAXIMA_MS;
  return Math.max(1, Math.min(CONSULTA_MAXIMA_MS, Math.floor(duracaoEtapaMs / 2)));
}

// intervaloEntreEtapas: d quando falta no máximo uma etapa para alcançar o servidor;
// d/4 quando há mais de uma atrasada (exibição acelerada, sem pular nenhuma).
export function intervaloEntreEtapas(atrasadas: number, duracaoEtapaMs: number): number {
  return atrasadas > 1 ? Math.floor(duracaoEtapaMs / 4) : duracaoEtapaMs;
}
