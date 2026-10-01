import { describe, expect, test } from 'vitest';
import { ENDERECO_CRIAR, enderecoPartida, lerRota } from './rota';

describe('rota', () => {
  test.each([
    { nome: 'API-10 CA-13 vazio: lista', hash: '', rota: { tela: 'lista' } },
    { nome: 'API-10 CA-13 #/: lista', hash: '#/', rota: { tela: 'lista' } },
    { nome: 'API-10 CA-13 partida', hash: '#/partidas/final-1', rota: { tela: 'partida', nome: 'final-1' } },
    { nome: 'API-10 CA-13 partida com barra final', hash: '#/partidas/final-1/', rota: { tela: 'partida', nome: 'final-1' } },
    { nome: 'API-10 CA-13 nome escapado', hash: '#/partidas/a%20b', rota: { tela: 'partida', nome: 'a b' } },
    { nome: 'API-10 CA-13 sem nome: lista', hash: '#/partidas/', rota: { tela: 'lista' } },
    { nome: 'API-10 CA-13 desconhecida: lista', hash: '#/outra', rota: { tela: 'lista' } },
    { nome: 'API-04 CA-17 criar partida', hash: '#/criar', rota: { tela: 'criar' } },
    { nome: 'API-04 CA-17 criar partida com barra final', hash: '#/criar/', rota: { tela: 'criar' } },
    { nome: 'API-04 CA-17 ENDERECO_CRIAR', hash: ENDERECO_CRIAR, rota: { tela: 'criar' } },
  ])('$nome', ({ hash, rota }) => {
    expect(lerRota(hash)).toEqual(rota);
  });

  test('API-10 CA-13 enderecoPartida e lerRota são inversos', () => {
    for (const nome of ['final-1', 'a b', 'x%y']) {
      expect(lerRota(enderecoPartida(nome))).toEqual({ tela: 'partida', nome });
    }
  });
});
