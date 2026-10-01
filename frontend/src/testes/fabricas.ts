// Construtores de dados para os testes (plano, D10). Valores pequenos e legíveis;
// cada teste sobrescreve só o que importa.
import type {
  Desfecho,
  Estado,
  Jogador,
  JogadorEtapa,
  Posicao,
  RelatorioEtapa,
  RespostaEstado,
  RespostaPartidas,
} from '../lib/tipos';

export const p = (x: number, y: number): Posicao => ({ x, y });

export function jogador(id: string, posicao: Posicao, extra: Partial<Jogador> = {}): Jogador {
  return {
    id,
    posicao,
    status: 'VIVO',
    bombas_por_turno: 1,
    potencia: 2,
    pavio_padrao: 3,
    acoes_por_turno: 3,
    bot_versao: 'aleatorio-v1',
    ...extra,
  };
}

// estado: tabuleiro 5×5 com dois jogadores nos cantos, um bloco fixo no centro
// e um bloco destrutível em (2,0).
export function estado(extra: Partial<Estado> = {}): Estado {
  return {
    turno: 1,
    config: {
      largura: 5,
      altura: 5,
      limite_turnos: 10,
      prazo_planejamento_ms: 1000,
      duracao_etapa_ms: 1000,
    },
    etapas_neste_turno: 3,
    blocos_fixos: [p(2, 2)],
    blocos_destrutiveis: [p(2, 0)],
    bombas: [],
    jogadores: [jogador('jogador_1', p(0, 0)), jogador('jogador_2', p(4, 4))],
    ...extra,
  };
}

export function jogadorEtapa(id: string, posicao: Posicao, extra: Partial<JogadorEtapa> = {}): JogadorEtapa {
  return {
    id,
    posicao,
    status: 'VIVO',
    acao: { etapa: 1, tipo: 'ESPERAR' },
    resultado: 'EXECUTADA',
    ...extra,
  };
}

// relatorio: etapa sem acontecimentos, com os jogadores do estado() parados.
export function relatorio(etapa: number, extra: Partial<RelatorioEtapa> = {}): RelatorioEtapa {
  return {
    turno: 1,
    etapa,
    jogadores: [
      jogadorEtapa('jogador_1', p(0, 0), { acao: { etapa, tipo: 'ESPERAR' } }),
      jogadorEtapa('jogador_2', p(4, 4), { acao: { etapa, tipo: 'ESPERAR' } }),
    ],
    bombas: [],
    explosoes: [],
    chamas: [],
    mortes: [],
    blocos_destruidos: [],
    movimentos_bloqueados: [],
    blocos_fechados: [],
    ...extra,
  };
}

export const emAndamento: Desfecho = { terminada: false, empate: false, sobreviventes: ['jogador_1', 'jogador_2'] };

// resposta: turno 1 em PLANEJAMENTO, servidor em 12:00:00.000, fase até 12:00:01.000.
export function resposta(extra: Partial<RespostaEstado> = {}): RespostaEstado {
  return {
    nome: 'final-1',
    fase: 'PLANEJAMENTO',
    turno: 1,
    etapa: 0,
    horario_servidor: '2026-10-01T12:00:00.000Z',
    fim_da_fase: '2026-10-01T12:00:01.000Z',
    estado: estado(),
    etapas: [],
    desfecho: emAndamento,
    ...extra,
  };
}

export function respostaPartidas(extra: Partial<RespostaPartidas> = {}): RespostaPartidas {
  return { partidas: [], horario_servidor: '2026-10-01T12:00:00.000Z', ...extra };
}

// relogioFalso: relógio de teste (plano, D2). O tempo só anda em avancar(), que dispara
// as esperas vencidas em ordem e deixa as promessas pendentes resolverem entre elas.
export function relogioFalso(inicio = Date.parse('2026-10-01T12:00:00.000Z')) {
  let agora = inicio;
  let seq = 0;
  let esperas: { quando: number; seq: number; f: () => void }[] = [];
  const relogio = {
    agora: () => agora,
    esperar(ms: number, f: () => void) {
      const e = { quando: agora + ms, seq: seq++, f };
      esperas.push(e);
      return () => {
        esperas = esperas.filter((x) => x !== e);
      };
    },
    async avancar(ms: number) {
      const fim = agora + ms;
      await descarregar();
      for (;;) {
        const proxima = esperas
          .filter((e) => e.quando <= fim)
          .sort((a, b) => a.quando - b.quando || a.seq - b.seq)[0];
        if (!proxima) break;
        esperas = esperas.filter((e) => e !== proxima);
        agora = proxima.quando;
        proxima.f();
        await descarregar();
      }
      agora = fim;
    },
  };
  return relogio;
}

// descarregar deixa rodar as promessas pendentes (microtarefas).
export const descarregar = () => new Promise<void>((r) => setTimeout(r, 0));

// clienteFalso: API falsa. `responder` recebe o número da chamada (1, 2, ...) e devolve
// a resposta, lança um erro, ou devolve uma promessa (para consultas que demoram).
export function clienteFalso(
  responder: (n: number) => RespostaEstado | Promise<RespostaEstado>,
  partidas: () => RespostaPartidas | Promise<RespostaPartidas> = () => respostaPartidas(),
) {
  const chamadas: string[] = [];
  return {
    chamadas,
    buscarEstado: async (nome: string) => {
      chamadas.push(`estado ${nome}`);
      return responder(chamadas.length);
    },
    buscarPartidas: async () => {
      chamadas.push('partidas');
      return partidas();
    },
  };
}
