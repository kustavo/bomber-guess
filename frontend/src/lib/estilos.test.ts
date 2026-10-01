import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, test } from 'vitest';

// O jsdom não avalia @media nem animações; o teste lê a folha de estilos (plano 14, D8).
// Os testes rodam a partir de frontend/ (npm test).
const css = readFileSync(resolve(process.cwd(), 'src/estilos/sprites.css'), 'utf8');

// regra devolve o corpo do primeiro bloco `seletor { ... }` dentro de `trecho`.
function regra(trecho: string, seletor: string): string | undefined {
  const i = trecho.indexOf(seletor);
  if (i < 0) return undefined;
  const abre = trecho.indexOf('{', i);
  return trecho.slice(abre + 1, trecho.indexOf('}', abre));
}

describe('sprites.css', () => {
  const inicioReduzido = css.indexOf('@media (prefers-reduced-motion: reduce)');
  const reduzido = css.slice(inicioReduzido);

  test('CA-12 bomba e chama têm animação', () => {
    expect(regra(css, '.animada-bomba {')).toMatch(/animation:\s*pulso-bomba/);
    expect(regra(css, '.animada-chama {')).toMatch(/animation:\s*tremor-chama/);
    expect(css).toContain('@keyframes pulso-bomba');
    expect(css).toContain('@keyframes tremor-chama');
  });

  test('CA-12 movimento reduzido desliga as animações e o deslize (decisão 6)', () => {
    expect(inicioReduzido).toBeGreaterThan(-1);
    expect(regra(reduzido, '.animada-chama')).toMatch(/animation:\s*none/);
    expect(reduzido).toMatch(/\.animada-bomba,\s*\.animada-chama\s*\{\s*animation:\s*none/);
    expect(regra(reduzido, '.deslize')).toMatch(/transition:\s*none/);
  });

  test('CA-14 sprites nítidos: crispEdges nos símbolos (D10)', () => {
    expect(regra(css, 'symbol {')).toMatch(/shape-rendering:\s*crispEdges/);
  });
});
