import { describe, expect, test } from 'vitest';
import {
  CONFIG_PADRAO,
  JOGADOR_PADRAO,
  LADO_MAXIMO,
  aplicarFerramenta,
  botsDasPosicoes,
  escolherBot,
  itemEm,
  normalizarNome,
  novoMapaEmEdicao,
  paraMapa,
  pendencias,
  preencherBots,
  redimensionar,
  sugerirNomeMapa,
  type Ferramenta,
  type MapaEmEdicao,
} from './editor';
import { p } from '../testes/fabricas';
import type { Posicao } from './tipos';

// aplicar aplica as ferramentas em sequência, com o bot padrão 'v1'.
function aplicar(m: MapaEmEdicao, ...passos: [Ferramenta, Posicao][]): MapaEmEdicao {
  return passos.reduce((acc, [f, pos]) => aplicarFerramenta(acc, f, pos, 'v1'), m);
}

describe('mapa em edição', () => {
  test('MAP-01 MAP-02 CA-07 editor recém-aberto: vazio, com os padrões da decisão 3', () => {
    const m = novoMapaEmEdicao();
    expect(m.blocos_fixos).toEqual([]);
    expect(m.blocos_destrutiveis).toEqual([]);
    expect(m.posicoes_iniciais).toEqual([]);
    expect(m.config).toEqual({
      largura: 15,
      altura: 13,
      limite_turnos: 50,
      turno_fechamento: 30,
      area_minima: { largura: 5, altura: 5 },
      prazo_planejamento_ms: 1000,
      duracao_etapa_ms: 1000,
    });
    expect(m.jogador_padrao).toEqual({ bombas_por_turno: 2, potencia: 2, pavio_padrao: 3, acoes_por_turno: 7 });
  });

  test('MAP-01 CA-07 cada editor tem a sua cópia dos padrões', () => {
    const a = novoMapaEmEdicao();
    a.config.area_minima!.largura = 1;
    a.jogador_padrao.potencia = 9;
    expect(novoMapaEmEdicao().config).toEqual(CONFIG_PADRAO);
    expect(novoMapaEmEdicao().jogador_padrao).toEqual(JOGADOR_PADRAO);
    expect(CONFIG_PADRAO.area_minima).toEqual({ largura: 5, altura: 5 });
  });

  test.each<{ nome: string; passos: [Ferramenta, Posicao][]; esperado: ReturnType<typeof itemEm> }>([
    { nome: 'CA-08 bloco fixo numa casa vazia', passos: [['bloco-fixo', p(1, 1)]], esperado: 'bloco-fixo' },
    {
      nome: 'CA-08 bloco destrutível substitui bloco fixo',
      passos: [['bloco-fixo', p(1, 1)], ['bloco-destrutivel', p(1, 1)]],
      esperado: 'bloco-destrutivel',
    },
    {
      nome: 'CA-08 posição inicial substitui bloco destrutível',
      passos: [['bloco-destrutivel', p(1, 1)], ['posicao-inicial', p(1, 1)]],
      esperado: 'posicao-inicial',
    },
    {
      nome: 'CA-08 bloco fixo substitui posição inicial',
      passos: [['posicao-inicial', p(1, 1)], ['bloco-fixo', p(1, 1)]],
      esperado: 'bloco-fixo',
    },
    { nome: 'CA-08 borracha esvazia a casa', passos: [['bloco-fixo', p(1, 1)], ['borracha', p(1, 1)]], esperado: undefined },
    { nome: 'CA-08 borracha numa casa vazia', passos: [['borracha', p(1, 1)]], esperado: undefined },
  ])('$nome', ({ passos, esperado }) => {
    const m = aplicar(novoMapaEmEdicao(), ...passos);
    expect(itemEm(m, p(1, 1))).toBe(esperado);
    const ocupacoes = m.blocos_fixos.length + m.blocos_destrutiveis.length + m.posicoes_iniciais.length;
    expect(ocupacoes).toBe(esperado === undefined ? 0 : 1);
  });

  test('CA-08 casa fora do tabuleiro não muda nada', () => {
    const m = novoMapaEmEdicao();
    expect(aplicarFerramenta(m, 'bloco-fixo', p(15, 0), 'v1')).toEqual(m);
    expect(aplicarFerramenta(m, 'bloco-fixo', p(-1, 0), 'v1')).toEqual(m);
  });

  test('CA-08 não altera o mapa recebido', () => {
    const m = novoMapaEmEdicao();
    aplicarFerramenta(m, 'bloco-fixo', p(1, 1), 'v1');
    expect(m.blocos_fixos).toEqual([]);
  });

  test('MAP-03 CA-09 posições numeradas na ordem em que foram colocadas', () => {
    const [a, b, c] = [p(0, 0), p(5, 5), p(2, 0)];
    let m = aplicar(novoMapaEmEdicao(), ['posicao-inicial', a], ['posicao-inicial', b], ['posicao-inicial', c]);
    expect(m.posicoes_iniciais.map((x) => x.posicao)).toEqual([a, b, c]);
    m = escolherBot(m, 2, 'v2');
    expect(botsDasPosicoes(m)).toEqual(['v1', 'v1', 'v2']);

    const semB = aplicar(m, ['borracha', b]);
    expect(semB.posicoes_iniciais).toEqual([
      { posicao: a, bot: 'v1' },
      { posicao: c, bot: 'v2' },
    ]);
    const blocoEmB = aplicar(m, ['bloco-destrutivel', b]);
    expect(blocoEmB.posicoes_iniciais.map((x) => x.posicao)).toEqual([a, c]);
    expect(botsDasPosicoes(blocoEmB)).toEqual(['v1', 'v2']);
  });

  test('MAP-03 CA-09 posição inicial sobre posição inicial não muda a ordem nem o bot (D6)', () => {
    let m = aplicar(novoMapaEmEdicao(), ['posicao-inicial', p(0, 0)], ['posicao-inicial', p(1, 0)]);
    m = escolherBot(m, 0, 'v2');
    expect(aplicar(m, ['posicao-inicial', p(0, 0)])).toEqual(m);
  });

  test('API-09 CA-12 preencherBots completa só as posições sem bot', () => {
    let m = novoMapaEmEdicao();
    m = aplicarFerramenta(m, 'posicao-inicial', p(0, 0), '');
    m = aplicarFerramenta(m, 'posicao-inicial', p(1, 0), 'v2');
    expect(botsDasPosicoes(preencherBots(m, 'v1'))).toEqual(['v1', 'v2']);
  });

  test('MAP-05 CA-10 diminuir o tabuleiro remove os itens de fora', () => {
    const m = aplicar(
      novoMapaEmEdicao(),
      ['bloco-fixo', p(1, 1)],
      ['bloco-fixo', p(14, 0)],
      ['bloco-destrutivel', p(0, 12)],
      ['bloco-destrutivel', p(2, 2)],
      ['posicao-inicial', p(0, 0)],
      ['posicao-inicial', p(9, 9)],
      ['posicao-inicial', p(3, 3)],
    );
    const r = redimensionar(m, 10, 9);
    expect(r.config.largura).toBe(10);
    expect(r.config.altura).toBe(9);
    expect(r.blocos_fixos).toEqual([p(1, 1)]);
    expect(r.blocos_destrutiveis).toEqual([p(2, 2)]);
    expect(r.posicoes_iniciais.map((x) => x.posicao)).toEqual([p(0, 0), p(3, 3)]);
  });

  test.each([
    { nome: 'MAP-05 CA-10 lado abaixo de 1 vira 1', l: 0, a: -3, esperado: [1, 1] },
    { nome: 'MAP-05 CA-10 lado acima do máximo vira o máximo (decisão 3)', l: 99, a: 31, esperado: [LADO_MAXIMO, LADO_MAXIMO] },
    { nome: 'MAP-05 CA-10 lado não inteiro é truncado', l: 7.8, a: 6.2, esperado: [7, 6] },
    { nome: 'MAP-05 CA-10 lado inválido mantém o atual', l: NaN, a: Infinity, esperado: [15, 13] },
  ])('$nome', ({ l, a, esperado }) => {
    const r = redimensionar(novoMapaEmEdicao(), l, a);
    expect([r.config.largura, r.config.altura]).toEqual(esperado);
  });

  test('FEC-08 CA-10 area_minima encolhe para caber no tabuleiro (D7)', () => {
    const r = redimensionar(novoMapaEmEdicao(), 3, 8);
    expect(r.config.area_minima).toEqual({ largura: 3, altura: 5 });
  });

  test('MAP-01 MAP-02 MAP-03 CA-11 conversão para o corpo de POST /mapas', () => {
    let m = aplicar(
      novoMapaEmEdicao(),
      ['bloco-fixo', p(1, 1)],
      ['bloco-destrutivel', p(2, 1)],
      ['posicao-inicial', p(0, 0)],
      ['posicao-inicial', p(4, 4)],
    );
    m = escolherBot(m, 1, 'v2');
    const mapa = paraMapa(m, 'novo');
    expect(Object.keys(mapa)).toEqual([
      'nome',
      'config',
      'jogador_padrao',
      'blocos_fixos',
      'blocos_destrutiveis',
      'posicoes_iniciais',
    ]);
    expect(JSON.parse(JSON.stringify(mapa))).toEqual({
      nome: 'novo',
      config: CONFIG_PADRAO,
      jogador_padrao: JOGADOR_PADRAO,
      blocos_fixos: [p(1, 1)],
      blocos_destrutiveis: [p(2, 1)],
      posicoes_iniciais: [p(0, 0), p(4, 4)],
    });
    // o mapa convertido não compartilha nada com o mapa em edição
    mapa.config.largura = 1;
    mapa.blocos_fixos.push(p(0, 1));
    expect(m.config.largura).toBe(15);
    expect(m.blocos_fixos).toEqual([p(1, 1)]);
  });
});

describe('pendências e nome do mapa', () => {
  const duas = aplicar(novoMapaEmEdicao(), ['posicao-inicial', p(0, 0)], ['posicao-inicial', p(1, 0)]);
  const uma = aplicar(novoMapaEmEdicao(), ['posicao-inicial', p(0, 0)]);
  const catalogo = ['v1'];

  test.each([
    { nome: 'MAP-05 MAP-06 CA-13 tudo preenchido: nada falta', m: duas, partida: 'final-1', mapa: 'novo', bots: catalogo, esperado: [] },
    {
      nome: 'MAP-05 CA-13 sem posições iniciais',
      m: novoMapaEmEdicao(),
      partida: 'final-1',
      mapa: 'novo',
      bots: catalogo,
      esperado: ['Coloque pelo menos 2 posições iniciais.'],
    },
    { nome: 'MAP-05 CA-13 uma posição inicial', m: uma, partida: 'final-1', mapa: 'novo', bots: catalogo, esperado: ['Coloque pelo menos 2 posições iniciais.'] },
    { nome: 'CA-13 sem nome da partida', m: duas, partida: '', mapa: 'novo', bots: catalogo, esperado: ['Dê um nome à partida.'] },
    { nome: 'CA-13 nome da partida só com espaços', m: duas, partida: '  ', mapa: 'novo', bots: catalogo, esperado: ['Dê um nome à partida.'] },
    { nome: 'CA-13 sem nome do mapa', m: duas, partida: 'final-1', mapa: '', bots: catalogo, esperado: ['Dê um nome ao mapa.'] },
    {
      nome: 'API-09 CA-13 catálogo de bots não carregado (decisão 4)',
      m: duas,
      partida: 'final-1',
      mapa: 'novo',
      bots: [],
      esperado: ['O catálogo de bots não foi carregado.'],
    },
    {
      nome: 'MAP-05 MAP-06 CA-13 várias pendências, na ordem do formulário',
      m: uma,
      partida: '',
      mapa: '',
      bots: [],
      esperado: [
        'Dê um nome à partida.',
        'Dê um nome ao mapa.',
        'Coloque pelo menos 2 posições iniciais.',
        'O catálogo de bots não foi carregado.',
      ],
    },
  ])('$nome', ({ m, partida, mapa, bots, esperado }) => {
    expect(pendencias(m, partida, mapa, bots)).toEqual(esperado);
  });

  test.each([
    { nome: 'CA-13 nome válido vira o nome do mapa (decisão 1)', partida: 'final-1', esperado: 'final-1' },
    { nome: 'CA-13 nome fora do padrão não é sugerido', partida: 'Final 1', esperado: '' },
    { nome: 'CA-13 sublinhado não é sugerido', partida: 'a_b', esperado: '' },
    { nome: 'CA-13 vazio não é sugerido', partida: '', esperado: '' },
  ])('$nome', ({ partida, esperado }) => {
    expect(sugerirNomeMapa(partida)).toBe(esperado);
  });
});

describe('nomes em minúsculas (reabertura)', () => {
  const duas = aplicar(novoMapaEmEdicao(), ['posicao-inicial', p(0, 0)], ['posicao-inicial', p(1, 0)]);

  test.each([
    { nome: 'CA-20 maiúsculas viram minúsculas', texto: 'Teste1', esperado: 'teste1' },
    { nome: 'CA-20 já em minúsculas não muda', texto: 'final-1', esperado: 'final-1' },
    { nome: 'CA-20 outros caracteres ficam como estão', texto: 'Meu Mapa_2', esperado: 'meu mapa_2' },
    { nome: 'CA-20 vazio continua vazio', texto: '', esperado: '' },
  ])('$nome', ({ texto, esperado }) => {
    expect(normalizarNome(texto)).toBe(esperado);
  });

  test.each([
    { nome: 'CA-20 nome da partida com espaço', partida: 'final 1', mapa: 'novo', esperado: ['O nome da partida só aceita letras minúsculas, números e -.'] },
    { nome: 'CA-20 nome do mapa com sublinhado', partida: 'final-1', mapa: 'meu_mapa', esperado: ['O nome do mapa só aceita letras minúsculas, números e -.'] },
    { nome: 'CA-20 nome com acento', partida: 'partida-ação', mapa: 'novo', esperado: ['O nome da partida só aceita letras minúsculas, números e -.'] },
    { nome: 'CA-20 nome com mais de 64 caracteres', partida: 'a'.repeat(65), mapa: 'novo', esperado: ['O nome da partida só aceita letras minúsculas, números e -.'] },
    { nome: 'CA-20 nomes válidos: nada falta', partida: 'final-1', mapa: 'novo-2', esperado: [] },
  ])('$nome', ({ partida, mapa, esperado }) => {
    expect(pendencias(duas, partida, mapa, ['v1'])).toEqual(esperado);
  });
});
