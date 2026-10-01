// Rotas por hash (spec 06, decisão 1): '#/' é a lista; '#/partidas/<nome>' é uma partida;
// '#/criar' é o editor de mapas (spec 07, CA-17).

export type Rota = { tela: 'lista' } | { tela: 'partida'; nome: string } | { tela: 'criar' };

const prefixo = '#/partidas/';

export const ENDERECO_CRIAR = '#/criar';

export function lerRota(hash: string): Rota {
  if (hash === ENDERECO_CRIAR || hash === ENDERECO_CRIAR + '/') return { tela: 'criar' };
  if (hash.startsWith(prefixo)) {
    const resto = hash.slice(prefixo.length).replace(/\/$/, '');
    if (resto !== '' && !resto.includes('/')) {
      try {
        return { tela: 'partida', nome: decodeURIComponent(resto) };
      } catch {
        // escape inválido: cai na lista
      }
    }
  }
  return { tela: 'lista' };
}

export const enderecoPartida = (nome: string) => prefixo + encodeURIComponent(nome);
