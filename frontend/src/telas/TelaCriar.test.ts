import { cleanup, fireEvent, render } from '@testing-library/svelte';
import { afterEach, describe, expect, test } from 'vitest';
import TelaCriar from './TelaCriar.svelte';
import { ErroApi } from '../lib/api';
import { clienteFalso, descarregar, resposta, respostaBots } from '../testes/fabricas';
import type { ClienteApi } from '../lib/api';
import type { Ferramenta } from '../lib/editor';

type Criacao = Parameters<typeof clienteFalso>[2];

async function abrir(criacao: Criacao = {}) {
  const cliente = clienteFalso(undefined, undefined, criacao);
  const tela = render(TelaCriar, { cliente: cliente as ClienteApi });
  await descarregar();
  const c = tela.container;
  const q = <T extends Element>(sel: string) => c.querySelector<T>(sel);
  const ferramenta = (f: Ferramenta) => fireEvent.click(q(`[data-ferramenta="${f}"]`)!);
  const casa = (x: number, y: number) => fireEvent.click(q(`[data-casa="${x},${y}"]`)!);
  const escrever = (rotulo: string, valor: string) => fireEvent.input(tela.getByLabelText(rotulo), { target: { value: valor } });
  const numeros = () => [...c.querySelectorAll('[data-tipo="posicao-inicial"]')].map((g) => [g.getAttribute('data-x'), g.getAttribute('data-y'), g.getAttribute('data-numero')]);
  const seletores = () => [...c.querySelectorAll<HTMLSelectElement>('select[data-posicao]')];
  const iniciar = () => q<HTMLButtonElement>('[data-acao="iniciar"]')!;
  const aviso = (a: string) => q(`[data-aviso="${a}"]`);
  // preencher: nomes e duas posições iniciais em (0,0) e (2,0).
  async function preencher() {
    await escrever('Nome da partida', 'final-1');
    await ferramenta('posicao-inicial');
    await casa(0, 0);
    await casa(2, 0);
  }
  return { tela, cliente, q, ferramenta, casa, escrever, numeros, seletores, iniciar, aviso, preencher };
}

afterEach(() => {
  cleanup();
  window.location.hash = '';
});

describe('TelaCriar', () => {
  test('MAP-01 MAP-02 CA-07 abre vazia, com os padrões da decisão 3', async () => {
    const { tela, q } = await abrir();
    expect(q('[data-tipo]')).toBeNull();
    expect(tela.container.querySelectorAll('[data-casa]')).toHaveLength(15 * 13);
    expect((tela.getByLabelText('Largura') as HTMLInputElement).value).toBe('15');
    expect((tela.getByLabelText('Limite de turnos') as HTMLInputElement).value).toBe('50');
    expect((tela.getByLabelText('Ações por turno') as HTMLInputElement).value).toBe('7');
    expect((tela.getByLabelText('Semente') as HTMLInputElement).value).toBe('1');
  });

  test('CA-08 clicar numa casa aplica a ferramenta; a borracha esvazia', async () => {
    const { q, ferramenta, casa } = await abrir();
    await ferramenta('bloco-fixo');
    await casa(1, 1);
    expect(q('[data-tipo="bloco-fixo"][data-x="1"][data-y="1"]')).not.toBeNull();
    await ferramenta('bloco-destrutivel');
    await casa(1, 1);
    expect(q('[data-tipo="bloco-fixo"]')).toBeNull();
    expect(q('[data-tipo="bloco-destrutivel"][data-x="1"][data-y="1"]')).not.toBeNull();
    await ferramenta('borracha');
    await casa(1, 1);
    expect(q('[data-tipo]')).toBeNull();
  });

  test('CA-08 teclado: Enter na casa aplica a ferramenta (D14)', async () => {
    const { q, ferramenta } = await abrir();
    await ferramenta('bloco-fixo');
    await fireEvent.keyDown(q('[data-casa="3,2"]')!, { key: 'Enter' });
    expect(q('[data-tipo="bloco-fixo"][data-x="3"][data-y="2"]')).not.toBeNull();
  });

  test('MAP-03 CA-09 posições numeradas; remover B renumera e C mantém o bot', async () => {
    const { ferramenta, casa, numeros, seletores } = await abrir();
    await ferramenta('posicao-inicial');
    await casa(0, 0);
    await casa(5, 5);
    await casa(2, 0);
    expect(numeros()).toEqual([['0', '0', '1'], ['5', '5', '2'], ['2', '0', '3']]);
    await fireEvent.change(seletores()[2], { target: { value: 'aleatorio-v2' } });
    await ferramenta('borracha');
    await casa(5, 5);
    expect(numeros()).toEqual([['0', '0', '1'], ['2', '0', '2']]);
    expect(seletores().map((s) => s.value)).toEqual(['aleatorio-v1', 'aleatorio-v2']);
  });

  test('MAP-05 CA-10 diminuir a largura remove os itens de fora', async () => {
    const { tela, q, ferramenta, casa } = await abrir();
    await ferramenta('bloco-fixo');
    await casa(14, 0);
    await casa(1, 0);
    await fireEvent.change(tela.getByLabelText('Largura'), { target: { value: '10' } });
    expect(tela.container.querySelectorAll('[data-casa]')).toHaveLength(10 * 13);
    expect(q('[data-tipo="bloco-fixo"][data-x="14"]')).toBeNull();
    expect(q('[data-tipo="bloco-fixo"][data-x="1"]')).not.toBeNull();
  });

  test('API-09 CA-12 cada posição tem um seletor com o catálogo; a nova vem com a primeira versão', async () => {
    const { ferramenta, casa, seletores, cliente } = await abrir({ buscarBots: async () => respostaBots(['b-v1', 'a-v9', 'c-v2']) });
    expect(cliente.chamadas).toEqual(['bots']);
    await ferramenta('posicao-inicial');
    await casa(0, 0);
    await casa(1, 0);
    expect(seletores()).toHaveLength(2);
    for (const s of seletores()) {
      expect([...s.options].map((o) => o.value)).toEqual(['b-v1', 'a-v9', 'c-v2']);
      expect(s.value).toBe('b-v1');
    }
  });

  test('API-09 CA-12 falha em GET /bots: aviso e Iniciar desabilitado (decisão 4)', async () => {
    const { aviso, iniciar, preencher } = await abrir({
      buscarBots: async () => {
        throw new ErroApi(0, 'Failed to fetch');
      },
    });
    await preencher();
    expect(aviso('bots')).not.toBeNull();
    expect(iniciar().disabled).toBe(true);
  });

  test('MAP-05 MAP-06 CA-13 falta algo: Iniciar desabilitado, aviso e nenhuma requisição', async () => {
    const { escrever, ferramenta, casa, iniciar, aviso, cliente } = await abrir();
    expect(iniciar().disabled).toBe(true);
    expect(aviso('pendencias')?.textContent).toContain('Dê um nome à partida.');
    expect(aviso('pendencias')?.textContent).toContain('Coloque pelo menos 2 posições iniciais.');
    await escrever('Nome da partida', 'final-1');
    await ferramenta('posicao-inicial');
    await casa(0, 0);
    expect(iniciar().disabled).toBe(true);
    expect(aviso('pendencias')?.textContent).not.toContain('nome à partida');
    await fireEvent.click(iniciar());
    await descarregar();
    expect(cliente.chamadas).toEqual(['bots']);
    await casa(1, 0);
    expect(iniciar().disabled).toBe(false);
    expect(aviso('pendencias')).toBeNull();
  });

  test('CA-13 nome do mapa acompanha o da partida até ser editado (D11)', async () => {
    const { tela, escrever } = await abrir();
    const nomeMapa = () => (tela.getByLabelText('Nome do mapa') as HTMLInputElement).value;
    await escrever('Nome da partida', 'final-1');
    expect(nomeMapa()).toBe('final-1');
    await escrever('Nome da partida', 'Final 1');
    expect(nomeMapa()).toBe('');
    await escrever('Nome do mapa', 'meu-mapa');
    await escrever('Nome da partida', 'final-2');
    expect(nomeMapa()).toBe('meu-mapa');
  });

  test('API-03 API-04 MAP-03 MAP-06 CA-14 Iniciar grava o mapa, cria a partida e abre a tela dela', async () => {
    const { tela, q, ferramenta, casa, seletores, iniciar, preencher, cliente } = await abrir();
    await ferramenta('bloco-fixo');
    await casa(1, 1);
    await preencher();
    await fireEvent.change(seletores()[1], { target: { value: 'aleatorio-v2' } });
    await fireEvent.input(tela.getByLabelText('Semente'), { target: { value: '7' } });
    await fireEvent.click(iniciar());
    await descarregar();
    expect(cliente.chamadas).toEqual(['bots', 'salvar mapa final-1', 'criar partida final-1']);
    expect(cliente.mapas[0]).toEqual({
      nome: 'final-1',
      config: {
        largura: 15,
        altura: 13,
        limite_turnos: 50,
        turno_fechamento: 30,
        area_minima: { largura: 5, altura: 5 },
        prazo_planejamento_ms: 1000,
        duracao_etapa_ms: 1000,
      },
      jogador_padrao: { bombas_por_turno: 2, potencia: 2, pavio_padrao: 3, acoes_por_turno: 7 },
      blocos_fixos: [{ x: 1, y: 1 }],
      blocos_destrutiveis: [],
      posicoes_iniciais: [{ x: 0, y: 0 }, { x: 2, y: 0 }],
    });
    expect(cliente.pedidos).toEqual([
      { nome: 'final-1', mapa: 'final-1', bots: ['aleatorio-v1', 'aleatorio-v2'], semente: 7 },
    ]);
    expect(window.location.hash).toBe('#/partidas/final-1');
    expect(q('[data-aviso="erro"]')).toBeNull();
  });

  test.each([
    { nome: 'API-03 CA-15 400 no POST /mapas', em: 'salvarMapa' as const, status: 400, texto: 'mapa inválido: limite_turnos 0' },
    { nome: 'API-03 CA-15 409 no POST /mapas', em: 'salvarMapa' as const, status: 409, texto: 'nome já usado: mapa "final-1"' },
    { nome: 'API-03 CA-15 servidor fora do ar', em: 'salvarMapa' as const, status: 0, texto: 'servidor indisponível' },
    { nome: 'API-04 CA-15 409 no POST /partidas', em: 'criarPartida' as const, status: 409, texto: 'nome já usado: "final-1"' },
  ])('$nome: mostra o erro e continua no editor com o mapa', async ({ em, status, texto }) => {
    const { q, numeros, iniciar, preencher, aviso, cliente } = await abrir({
      [em]: async () => {
        throw new ErroApi(status, status === 0 ? 'Failed to fetch' : texto);
      },
    });
    await preencher();
    await fireEvent.click(iniciar());
    await descarregar();
    expect(aviso('erro')?.textContent).toContain(texto);
    expect(window.location.hash).toBe('');
    expect(numeros()).toEqual([['0', '0', '1'], ['2', '0', '2']]);
    expect(q('[data-casa]')).not.toBeNull();
    const criou = cliente.chamadas.includes('criar partida final-1');
    expect(criou).toBe(em === 'criarPartida');
  });

  test('API-03 API-04 CA-16 após 409 da partida, nova tentativa sem gravar o mapa de novo', async () => {
    let tentativas = 0;
    const { escrever, iniciar, preencher, aviso, cliente } = await abrir({
      criarPartida: async (pedido) => {
        if (tentativas++ === 0) throw new ErroApi(409, 'nome já usado');
        return resposta({ nome: pedido.nome });
      },
    });
    await preencher();
    await escrever('Nome do mapa', 'final-1'); // fixa o nome do mapa
    await fireEvent.click(iniciar());
    await descarregar();
    expect(aviso('erro')).not.toBeNull();
    await escrever('Nome da partida', 'final-2');
    await fireEvent.click(iniciar());
    await descarregar();
    expect(cliente.chamadas).toEqual(['bots', 'salvar mapa final-1', 'criar partida final-1', 'criar partida final-2']);
    expect(window.location.hash).toBe('#/partidas/final-2');
  });

  test('API-03 CA-16 mapa alterado depois do 409 é gravado de novo', async () => {
    const { ferramenta, casa, iniciar, preencher, cliente } = await abrir({
      criarPartida: async () => {
        throw new ErroApi(409, 'nome já usado');
      },
    });
    await preencher();
    await fireEvent.click(iniciar());
    await descarregar();
    await ferramenta('bloco-fixo');
    await casa(1, 1);
    await fireEvent.click(iniciar());
    await descarregar();
    expect(cliente.chamadas.filter((c) => c.startsWith('salvar mapa'))).toHaveLength(2);
  });

  test('CA-11 botões da paleta mostram o sprite do item; a borracha mostra o piso', async () => {
    const { q } = await abrir();
    const icone = (f: Ferramenta) => q(`[data-ferramenta="${f}"] use`)?.getAttribute('data-sprite');
    expect(icone('bloco-fixo')).toBe('bloco-fixo');
    expect(icone('bloco-destrutivel')).toBe('bloco-destrutivel');
    expect(icone('posicao-inicial')).toBe('jogador');
    expect(icone('borracha')).toBe('piso');
    expect(q('[data-ferramenta] [data-tipo]')).toBeNull();
  });
});
