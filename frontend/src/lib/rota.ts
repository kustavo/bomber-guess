// Rotas por hash (spec 06, decisão 1): '#/' é a lista; '#/partidas/<nome>' é uma partida.

export type Rota = { tela: 'lista' } | { tela: 'partida'; nome: string };

const prefixo = '#/partidas/';

export function lerRota(hash: string): Rota {
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
