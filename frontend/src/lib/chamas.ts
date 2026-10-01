// Forma de cada casa em chamas, para escolher o sprite (spec 14, CA-03; plano, D6).
import type { Explosao, Posicao } from './tipos';

export type FormaChama =
  | 'centro'
  | 'horizontal'
  | 'vertical'
  | 'ponta-cima'
  | 'ponta-baixo'
  | 'ponta-esquerda'
  | 'ponta-direita';

const chave = (p: Posicao) => `${p.x},${p.y}`;

// formasDaExplosao: a origem é o centro; na linha dela, braço horizontal; na coluna,
// vertical; a casa mais distante de cada lado é a ponta desse lado.
function formasDaExplosao(e: Explosao): Map<string, FormaChama> {
  const formas = new Map<string, FormaChama>([[chave(e.origem), 'centro']]);
  const lados = [
    { ponta: 'ponta-esquerda', braco: 'horizontal', distancia: (p: Posicao) => (p.y === e.origem.y ? e.origem.x - p.x : 0) },
    { ponta: 'ponta-direita', braco: 'horizontal', distancia: (p: Posicao) => (p.y === e.origem.y ? p.x - e.origem.x : 0) },
    { ponta: 'ponta-cima', braco: 'vertical', distancia: (p: Posicao) => (p.x === e.origem.x ? e.origem.y - p.y : 0) },
    { ponta: 'ponta-baixo', braco: 'vertical', distancia: (p: Posicao) => (p.x === e.origem.x ? p.y - e.origem.y : 0) },
  ] as const;
  for (const lado of lados) {
    const casas = e.chamas.filter((p) => lado.distancia(p) > 0);
    const maior = Math.max(0, ...casas.map(lado.distancia));
    for (const p of casas) formas.set(chave(p), lado.distancia(p) === maior ? lado.ponta : lado.braco);
  }
  return formas;
}

// formasDasChamas devolve a forma de cada casa de `chamas`, pela chave "x,y".
// Explosões que discordam sobre uma casa a deixam como centro, e uma casa fora
// de qualquer explosão também é centro.
export function formasDasChamas(chamas: Posicao[], explosoes: Explosao[]): Map<string, FormaChama> {
  const porCasa = new Map<string, FormaChama>();
  for (const e of explosoes) {
    for (const [c, forma] of formasDaExplosao(e)) {
      const antes = porCasa.get(c);
      porCasa.set(c, antes === undefined || antes === forma ? forma : 'centro');
    }
  }
  return new Map(chamas.map((p) => [chave(p), porCasa.get(chave(p)) ?? 'centro']));
}
