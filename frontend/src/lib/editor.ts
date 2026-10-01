// Mapa em edição do editor de mapas (spec 07, CA-07 a CA-13). Funções puras:
// recebem um mapa e devolvem outro, sem alterar o recebido.
import type { Atributos, Config, Mapa, Posicao } from './tipos';

// Item é o que o editor coloca numa casa; uma casa tem no máximo um.
export type Item = 'bloco-fixo' | 'bloco-destrutivel' | 'posicao-inicial';
export type Ferramenta = Item | 'borracha';

// PosicaoInicial guarda o bot escolhido junto com a casa (D5).
export interface PosicaoInicial {
  posicao: Posicao;
  bot: string;
}

export interface MapaEmEdicao {
  config: Config; // sempre com turno_fechamento e area_minima
  jogador_padrao: Atributos;
  blocos_fixos: Posicao[];
  blocos_destrutiveis: Posicao[];
  posicoes_iniciais: PosicaoInicial[]; // na ordem em que foram colocadas (MAP-03)
}

// Padrões da decisão 3: os de mapas/exemplo.json.
export const CONFIG_PADRAO: Readonly<Config> = Object.freeze({
  largura: 15,
  altura: 13,
  limite_turnos: 50,
  turno_fechamento: 30,
  area_minima: Object.freeze({ largura: 5, altura: 5 }),
  prazo_planejamento_ms: 1000,
  duracao_etapa_ms: 1000,
});

export const JOGADOR_PADRAO: Readonly<Atributos> = Object.freeze({
  bombas_por_turno: 2,
  potencia: 2,
  pavio_padrao: 3,
  acoes_por_turno: 7,
});

// LADO_MAXIMO limita largura e altura só na tela (decisão 3).
export const LADO_MAXIMO = 30;

// NOME_VALIDO é o formato de nomes de mapa e de partida (decisão 1; spec 05, D8).
export const NOME_VALIDO = /^[a-z0-9-]{1,64}$/;

const mesma = (a: Posicao) => (b: Posicao) => a.x === b.x && a.y === b.y;
const outra = (a: Posicao) => (b: Posicao) => a.x !== b.x || a.y !== b.y;

function copiar(m: MapaEmEdicao): MapaEmEdicao {
  return {
    config: { ...m.config, area_minima: m.config.area_minima && { ...m.config.area_minima } },
    jogador_padrao: { ...m.jogador_padrao },
    blocos_fixos: m.blocos_fixos.map((b) => ({ ...b })),
    blocos_destrutiveis: m.blocos_destrutiveis.map((b) => ({ ...b })),
    posicoes_iniciais: m.posicoes_iniciais.map((pi) => ({ posicao: { ...pi.posicao }, bot: pi.bot })),
  };
}

export function novoMapaEmEdicao(): MapaEmEdicao {
  return copiar({
    config: CONFIG_PADRAO,
    jogador_padrao: JOGADOR_PADRAO,
    blocos_fixos: [],
    blocos_destrutiveis: [],
    posicoes_iniciais: [],
  });
}

export function itemEm(m: MapaEmEdicao, p: Posicao): Item | undefined {
  if (m.blocos_fixos.some(mesma(p))) return 'bloco-fixo';
  if (m.blocos_destrutiveis.some(mesma(p))) return 'bloco-destrutivel';
  if (m.posicoes_iniciais.some((pi) => mesma(p)(pi.posicao))) return 'posicao-inicial';
  return undefined;
}

const dentro = (m: MapaEmEdicao) => (p: Posicao) =>
  p.x >= 0 && p.y >= 0 && p.x < m.config.largura && p.y < m.config.altura;

// aplicarFerramenta põe o item da ferramenta na casa, no lugar do que havia nela,
// ou a esvazia com a borracha (CA-08). O mesmo item na mesma casa não muda nada (D6).
// Uma posição inicial nova recebe botPadrao (CA-12).
export function aplicarFerramenta(m: MapaEmEdicao, f: Ferramenta, p: Posicao, botPadrao: string): MapaEmEdicao {
  if (!dentro(m)(p) || itemEm(m, p) === f) return m;
  const r = copiar(m);
  r.blocos_fixos = r.blocos_fixos.filter(outra(p));
  r.blocos_destrutiveis = r.blocos_destrutiveis.filter(outra(p));
  r.posicoes_iniciais = r.posicoes_iniciais.filter((pi) => outra(p)(pi.posicao));
  const casa = { x: p.x, y: p.y };
  if (f === 'bloco-fixo') r.blocos_fixos.push(casa);
  if (f === 'bloco-destrutivel') r.blocos_destrutiveis.push(casa);
  if (f === 'posicao-inicial') r.posicoes_iniciais.push({ posicao: casa, bot: botPadrao });
  return r;
}

// escolherBot troca o bot da posição inicial i (MAP-03).
export function escolherBot(m: MapaEmEdicao, i: number, bot: string): MapaEmEdicao {
  const r = copiar(m);
  if (r.posicoes_iniciais[i]) r.posicoes_iniciais[i].bot = bot;
  return r;
}

// preencherBots dá o bot às posições colocadas antes de o catálogo chegar (CA-12).
export function preencherBots(m: MapaEmEdicao, bot: string): MapaEmEdicao {
  const r = copiar(m);
  for (const pi of r.posicoes_iniciais) if (pi.bot === '') pi.bot = bot;
  return r;
}

function lado(valor: number, atual: number): number {
  if (!Number.isFinite(valor)) return atual;
  return Math.min(LADO_MAXIMO, Math.max(1, Math.trunc(valor)));
}

// redimensionar muda o tamanho do tabuleiro (1…LADO_MAXIMO), remove os itens que
// ficam de fora (CA-10) e encolhe area_minima para caber (D7).
export function redimensionar(m: MapaEmEdicao, largura: number, altura: number): MapaEmEdicao {
  const r = copiar(m);
  r.config.largura = lado(largura, m.config.largura);
  r.config.altura = lado(altura, m.config.altura);
  if (r.config.area_minima) {
    r.config.area_minima.largura = Math.min(r.config.area_minima.largura, r.config.largura);
    r.config.area_minima.altura = Math.min(r.config.area_minima.altura, r.config.altura);
  }
  const cabe = dentro(r);
  r.blocos_fixos = r.blocos_fixos.filter(cabe);
  r.blocos_destrutiveis = r.blocos_destrutiveis.filter(cabe);
  r.posicoes_iniciais = r.posicoes_iniciais.filter((pi) => cabe(pi.posicao));
  return r;
}

// paraMapa converte para o corpo de POST /mapas, no formato de docs/EDITOR.md (CA-11).
export function paraMapa(m: MapaEmEdicao, nome: string): Mapa {
  const c = copiar(m);
  return {
    nome,
    config: c.config,
    jogador_padrao: c.jogador_padrao,
    blocos_fixos: c.blocos_fixos,
    blocos_destrutiveis: c.blocos_destrutiveis,
    posicoes_iniciais: c.posicoes_iniciais.map((pi) => pi.posicao),
  };
}

// botsDasPosicoes: a versão i joga na posição inicial i (MAP-03).
export function botsDasPosicoes(m: MapaEmEdicao): string[] {
  return m.posicoes_iniciais.map((pi) => pi.bot);
}

// pendencias diz o que falta para iniciar, na ordem do formulário (CA-13, D10).
// O resto de MAP-05 fica com o servidor (decisão 8).
export function pendencias(m: MapaEmEdicao, nomePartida: string, nomeMapa: string, bots: string[]): string[] {
  const faltas: string[] = [];
  if (nomePartida.trim() === '') faltas.push('Dê um nome à partida.');
  if (nomeMapa.trim() === '') faltas.push('Dê um nome ao mapa.');
  if (m.posicoes_iniciais.length < 2) faltas.push('Coloque pelo menos 2 posições iniciais.');
  if (bots.length === 0) faltas.push('O catálogo de bots não foi carregado.');
  return faltas;
}

// sugerirNomeMapa: o nome da partida, se ele servir de nome de mapa (decisão 1, D11).
export function sugerirNomeMapa(nomePartida: string): string {
  return NOME_VALIDO.test(nomePartida) ? nomePartida : '';
}
