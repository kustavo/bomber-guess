// Sprites em pixel art do tabuleiro (spec 14, decisões 1 e 2; plano, D2 e D3).
// Cada sprite tem 16 linhas de 16 letras; cada letra é uma cor da PALETA e '.' é
// transparente. Desenhos originais, feitos para o projeto.

export const LADO_SPRITE = 16;

export const PALETA: Readonly<Record<string, string>> = Object.freeze({
  k: '#1b1b22', // contorno e corpo da bomba
  w: '#ffffff',
  A: '#b4bcc6', // bloco fixo: luz
  B: '#7c8692', // bloco fixo: face
  C: '#4c535d', // bloco fixo: sombra
  D: '#e0a86a', // tijolo: luz
  E: '#b5773d', // tijolo: face
  F: '#6b4220', // tijolo: argamassa
  G: '#4a8a3f', // piso
  H: '#3c7533', // piso: tufo escuro
  I: '#5ea052', // piso: tufo claro
  L: '#4a4a5a', // brilho da bomba, botas
  M: '#d8c08c', // pavio
  R: '#e4421b', // chama: borda
  O: '#ff9d23', // chama
  Y: '#ffe455', // chama: miolo
  s: '#f4c99e', // pele
  c: 'currentColor', // roupa do jogador: a cor de cada um
});

export type NomeSprite =
  | 'piso'
  | 'bloco-fixo'
  | 'bloco-destrutivel'
  | 'bomba'
  | 'jogador'
  | 'chama-centro'
  | 'chama-horizontal'
  | 'chama-vertical'
  | 'chama-ponta-cima'
  | 'chama-ponta-baixo'
  | 'chama-ponta-esquerda'
  | 'chama-ponta-direita';

const piso = [
  'GGGGGGGGGGGGGGGG',
  'GGGIGGGGGGGGHGGG',
  'GGGGGGGGGGGGGGGG',
  'GGGGGGHGGGGGGGGG',
  'GIGGGGGGGGGGGGIG',
  'GGGGGGGGGGIGGGGG',
  'GGGGGGGGGGGGGGGG',
  'GGGHGGGGGGGGGGGG',
  'GGGGGGGGGHGGGGGG',
  'GGGGGGGGGGGGGGGG',
  'GGGGGIGGGGGGGGHG',
  'GGGGGGGGGGGGGGGG',
  'GHGGGGGGGGGIGGGG',
  'GGGGGGGGGGGGGGGG',
  'GGGGGGHGGGGGGGGG',
  'GGGGGGGGGGGGGGGG',
];

const blocoFixo = [
  'kkkkkkkkkkkkkkkk',
  'kAAAAAAAAAAAAABk',
  'kAABBBBBBBBBBBCk',
  'kABAkBBBBBBAkBCk',
  'kABkCBBBBBBkCBCk',
  'kABBBBBBBBBBBBCk',
  'kABBBBBBBBBBBBCk',
  'kABBBBBBBBBBBBCk',
  'kABBBBBBBBBBBBCk',
  'kABBBBBBBBBBBBCk',
  'kABBBBBBBBBBBBCk',
  'kABAkBBBBBBAkBCk',
  'kABkCBBBBBBkCBCk',
  'kABBBBBBBBBBBCCk',
  'kBCCCCCCCCCCCCCk',
  'kkkkkkkkkkkkkkkk',
];

const blocoDestrutivel = [
  'DDDDDDDFDDDDDDDF',
  'EEEEEEEFEEEEEEEF',
  'EEEEEEEFEEEEEEEF',
  'FFFFFFFFFFFFFFFF',
  'DDDFDDDDDDDFDDDD',
  'EEEFEEEEEEEFEEEE',
  'EEEFEEEEEEEFEEEE',
  'FFFFFFFFFFFFFFFF',
  'DDDDDDDFDDDDDDDF',
  'EEEEEEEFEEEEEEEF',
  'EEEEEEEFEEEEEEEF',
  'FFFFFFFFFFFFFFFF',
  'DDDFDDDDDDDFDDDD',
  'EEEFEEEEEEEFEEEE',
  'EEEFEEEEEEEFEEEE',
  'FFFFFFFFFFFFFFFF',
];

const bomba = [
  '..........Y.O...',
  '.........OwY....',
  '.........M......',
  '........M.......',
  '.....kkkkk......',
  '...kkkkkkkkk....',
  '..kkLLkkkkkkk...',
  '..kLwLkkkkkkk...',
  '.kkLLkkkkkkkkk..',
  '.kkkkkkkkkkkkk..',
  '.kkkkkkkkkkkkk..',
  '.kkkkkkkkkkkkk..',
  '..kkkkkkkkkkk...',
  '..kkkkkkkkkkk...',
  '...kkkkkkkkk....',
  '.....kkkkk......',
];

const jogador = [
  '......kkkk......',
  '....kkwwwwkk....',
  '...kwwwwwwwwk...',
  '..kwwwwwwwwwwk..',
  '..kwkkkkkkkkwk..',
  '..kwksksskskwk..',
  '..kwksssssskwk..',
  '..kwwkkkkkkwwk..',
  '...kwwwwwwwwk...',
  '....kkcccckk....',
  '..kwkcccccckwk..',
  '..kkkccwwcckkk..',
  '....kcccccck....',
  '....kcckkcck....',
  '...kLLk..kLLk...',
  '...kkkk..kkkk...',
];

const chamaHorizontal = [
  '................',
  '................',
  '................',
  'RRRRRRRRRRRRRRRR',
  'OOOOOOOOOOOOOOOO',
  'YYYOYYYYYYYOYYYY',
  'YYYYYYYYYYYYYYYY',
  'wwwwwwwwwwwwwwww',
  'wwwwwwwwwwwwwwww',
  'YYYYYYYYYYYYYYYY',
  'YYYYYYOYYYYYYYOY',
  'OOOOOOOOOOOOOOOO',
  'RRRRRRRRRRRRRRRR',
  '................',
  '................',
  '................',
];

const chamaPontaDireita = [
  '................',
  '................',
  '................',
  'RRRRRRRRRR......',
  'OOOOOOOOOORR....',
  'YYYOYYYYYYOOR...',
  'YYYYYYYYYYYOR...',
  'wwwwwwwwwwYOOR..',
  'wwwwwwwwwwYOOR..',
  'YYYYYYYYYYYOR...',
  'YYYYYYOYYYOOR...',
  'OOOOOOOOOORR....',
  'RRRRRRRRRR......',
  '................',
  '................',
  '................',
];

// Rotações e espelhos, para que os braços e as pontas encaixem exatamente.
const transpor = (d: string[]) => d.map((_, y) => d.map((l) => l[y]).join(''));
const espelharHorizontal = (d: string[]) => d.map((l) => [...l].reverse().join(''));
const espelharVertical = (d: string[]) => [...d].reverse();

// combinar sobrepõe desenhos, ficando com a cor mais quente de cada pixel.
const calor = '.ROYw';
const combinar = (...ds: string[][]) =>
  ds[0].map((_, y) =>
    [...ds[0][y]].map((_, x) => ds.map((d) => d[y][x]).reduce((a, b) => (calor.indexOf(b) > calor.indexOf(a) ? b : a))).join(''),
  );

const chamaVertical = transpor(chamaHorizontal);
const chamaPontaBaixo = transpor(chamaPontaDireita);

export const SPRITES: Readonly<Record<NomeSprite, readonly string[]>> = Object.freeze({
  piso,
  'bloco-fixo': blocoFixo,
  'bloco-destrutivel': blocoDestrutivel,
  bomba,
  jogador,
  'chama-centro': combinar(chamaHorizontal, chamaVertical),
  'chama-horizontal': chamaHorizontal,
  'chama-vertical': chamaVertical,
  'chama-ponta-direita': chamaPontaDireita,
  'chama-ponta-esquerda': espelharHorizontal(chamaPontaDireita),
  'chama-ponta-baixo': chamaPontaBaixo,
  'chama-ponta-cima': espelharVertical(chamaPontaBaixo),
});

export const NOMES_SPRITES = Object.freeze(Object.keys(SPRITES) as NomeSprite[]);

export interface Retangulo {
  x: number;
  y: number;
  largura: number;
  cor: string;
}

// retangulos junta pixels consecutivos da mesma cor numa linha (D2).
export function retangulos(desenho: readonly string[]): Retangulo[] {
  const r: Retangulo[] = [];
  desenho.forEach((linha, y) => {
    let x = 0;
    while (x < linha.length) {
      const letra = linha[x];
      let fim = x + 1;
      while (fim < linha.length && linha[fim] === letra) fim++;
      if (letra !== '.') r.push({ x, y, largura: fim - x, cor: PALETA[letra] });
      x = fim;
    }
  });
  return r;
}

export const idSprite = (nome: NomeSprite) => `sprite-${nome}`;
