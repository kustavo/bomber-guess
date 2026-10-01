import { get } from 'svelte/store';
import { describe, expect, test } from 'vitest';
import { acompanhar } from './acompanhamento';
import { ErroApi } from './api';
import { clienteFalso, estado, relatorio, relogioFalso, resposta } from '../testes/fabricas';
import type { RespostaEstado } from './tipos';

const D = 1000; // duracao_etapa_ms do estado() das fábricas

function execucao(liberadas: number, extra: Partial<RespostaEstado> = {}): RespostaEstado {
  return resposta({
    fase: 'EXECUCAO',
    etapa: liberadas,
    etapas: Array.from({ length: liberadas }, (_, i) => relatorio(i + 1)),
    ...extra,
  });
}

// iniciar: acompanha 'final-1' com uma sequência de respostas (a última se repete).
async function iniciar(respostas: (RespostaEstado | Error)[]) {
  const relogio = relogioFalso();
  const cliente = clienteFalso((n) => {
    const r = respostas[Math.min(n, respostas.length) - 1];
    if (r instanceof Error) throw r;
    return r;
  });
  const a = acompanhar('final-1', cliente, relogio);
  await relogio.avancar(0);
  return { a, relogio, cliente, visao: () => get(a.visao) };
}

describe('acompanhamento', () => {
  test('API-01 CA-08 consulta o estado a cada intervalo, só por GET de estado', async () => {
    const { a, relogio, cliente } = await iniciar([resposta()]);
    expect(cliente.chamadas).toEqual(['estado final-1']);
    await relogio.avancar(1000);
    expect(cliente.chamadas).toEqual(Array(5).fill('estado final-1')); // 1 + 1000/250
    a.parar();
  });

  test('API-01 CA-08 uma consulta por vez: a próxima só depois da resposta', async () => {
    const relogio = relogioFalso();
    let responder: (r: RespostaEstado) => void = () => {};
    const cliente = clienteFalso(() => new Promise((ok) => (responder = ok)));
    const a = acompanhar('final-1', cliente, relogio);
    await relogio.avancar(2000);
    expect(cliente.chamadas).toHaveLength(1);
    responder(resposta());
    await relogio.avancar(250);
    expect(cliente.chamadas).toHaveLength(2);
    a.parar();
  });

  test('API-01 CA-08 parar() encerra a consulta', async () => {
    const { a, relogio, cliente } = await iniciar([resposta()]);
    a.parar();
    await relogio.avancar(5000);
    expect(cliente.chamadas).toHaveLength(1);
  });

  test('API-01 CA-09 partida ENCERRADA: a consulta para', async () => {
    const encerrada = resposta({
      fase: 'ENCERRADA',
      fim_da_fase: undefined,
      desfecho: { terminada: true, empate: false, vencedor: 'jogador_1', sobreviventes: ['jogador_1'] },
    });
    const { relogio, cliente, visao } = await iniciar([resposta(), encerrada]);
    await relogio.avancar(5000);
    expect(cliente.chamadas).toHaveLength(2);
    expect(visao().resposta?.fase).toBe('ENCERRADA');
  });

  test('API-02 CA-12 primeira resposta já sincroniza: etapa, turno, fase e defasagem', async () => {
    const r = execucao(3, { turno: 4, estado: estado({ turno: 4 }), horario_servidor: '2026-10-01T12:00:00.500Z' });
    const { a, visao } = await iniciar([r]);
    const v = visao();
    expect(v.carregando).toBe(false);
    expect(v.resposta).toEqual(r);
    expect(v.tabuleiro?.turno).toBe(4);
    expect(v.tabuleiro?.etapa).toBe(3);
    expect(v.defasagem).toBe(500);
    a.parar();
  });

  test('API-02 CA-12 antes da primeira resposta: carregando', () => {
    const relogio = relogioFalso();
    const a = acompanhar('final-1', clienteFalso(() => new Promise(() => {})), relogio);
    expect(get(a.visao).carregando).toBe(true);
    expect(get(a.visao).tabuleiro).toBeUndefined();
    a.parar();
  });

  test('EST-02 CA-10 em dia: cada etapa aparece quando é liberada', async () => {
    // O servidor libera uma etapa por segundo; a tela acompanha uma a uma.
    const relogio = relogioFalso();
    const inicio = relogio.agora();
    const cliente = clienteFalso(() => {
      const k = Math.min(3, Math.floor((relogio.agora() - inicio) / D));
      return k === 0 ? resposta() : execucao(k);
    });
    const a = acompanhar('final-1', cliente, relogio);
    const vistas: number[] = [];
    const parar = a.visao.subscribe((v) => {
      const e = v.tabuleiro?.etapa;
      if (e !== undefined && vistas.at(-1) !== e) vistas.push(e);
    });
    await relogio.avancar(4000);
    expect(vistas).toEqual([0, 1, 2, 3]);
    parar();
    a.parar();
  });

  test('EST-02 CA-10 consulta atrasada: etapas em ordem, sem pular, mais rápido que d', async () => {
    const { a, relogio, visao } = await iniciar([execucao(1), execucao(4)]);
    const vistas: [number, number][] = [];
    const inicio = relogio.agora();
    const parar = a.visao.subscribe((v) => {
      const e = v.tabuleiro!.etapa;
      if (vistas.at(-1)?.[0] !== e) vistas.push([e, relogio.agora() - inicio]);
    });
    await relogio.avancar(3000);
    expect(vistas.map(([e]) => e)).toEqual([1, 2, 3, 4]);
    for (let i = 1; i < vistas.length; i++) {
      expect(vistas[i][1] - vistas[i - 1][1]).toBeLessThanOrEqual(D);
    }
    expect(vistas[2][1] - vistas[1][1]).toBe(D / 4); // atrasada: acelerada
    expect(visao().tabuleiro?.etapa).toBe(4);
    parar();
    a.parar();
  });

  test('API-05 CA-11 turno seguinte: tabuleiro do novo estado, sem chamas nem marcas', async () => {
    const comChamas = execucao(1, {
      etapas: [relatorio(1, { chamas: [{ x: 1, y: 0 }], movimentos_bloqueados: ['jogador_1'] })],
    });
    const novo = resposta({ turno: 2, estado: estado({ turno: 2 }) });
    const { a, relogio, visao } = await iniciar([comChamas, novo]);
    expect(visao().tabuleiro?.chamas).toHaveLength(1);
    await relogio.avancar(250);
    const t = visao().tabuleiro!;
    expect(t.turno).toBe(2);
    expect(t.etapa).toBe(0);
    expect(t.chamas).toEqual([]);
    expect(t.jogadores.some((j) => j.bloqueado)).toBe(false);
    a.parar();
  });

  test('API-05 CA-14 partida inexistente: naoEncontrada e a consulta para', async () => {
    const { relogio, cliente, visao } = await iniciar([new ErroApi(404, 'partida "final-1" não encontrada')]);
    expect(visao().naoEncontrada).toBe(true);
    expect(visao().carregando).toBe(false);
    await relogio.avancar(5000);
    expect(cliente.chamadas).toHaveLength(1);
  });

  test.each([
    { nome: 'API-01 CA-15 falha de rede', erro: new ErroApi(0, 'Failed to fetch') },
    { nome: 'API-01 CA-15 resposta 5xx', erro: new ErroApi(502, 'HTTP 502') },
  ])('$nome: mantém o tabuleiro, avisa, continua e o aviso some', async ({ erro }) => {
    const r = execucao(2);
    const { a, relogio, cliente, visao } = await iniciar([r, erro, erro, r]);
    await relogio.avancar(250);
    expect(visao().semConexao).toBe(true);
    expect(visao().resposta).toEqual(r);
    expect(visao().tabuleiro?.etapa).toBe(2);
    await relogio.avancar(250);
    expect(cliente.chamadas).toHaveLength(3);
    expect(visao().semConexao).toBe(true);
    await relogio.avancar(250);
    expect(visao().semConexao).toBe(false);
    a.parar();
  });

  test('API-05 outro erro 4xx: mostra o erro e continua consultando', async () => {
    const { a, relogio, cliente, visao } = await iniciar([new ErroApi(400, 'pedido inválido'), resposta()]);
    expect(visao().erro).toBe('pedido inválido');
    await relogio.avancar(250);
    expect(cliente.chamadas).toHaveLength(2);
    expect(visao().erro).toBeUndefined();
    a.parar();
  });
});
