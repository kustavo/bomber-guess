// Tipos do JSON da API, espelho exato do backend: jogo.Estado e jogo.RelatorioEtapa
// (docs/REGRAS.md, docs/ARQUITETURA.md 1.2) e as respostas do marco 5 (spec 05, decisões 5 e 9).

export type Fase = 'PLANEJAMENTO' | 'EXECUCAO' | 'ENCERRADA';
export type Status = 'VIVO' | 'MORTO';
export type TipoAcao = 'MOVER' | 'PLANTAR' | 'ESPERAR';
export type Direcao = 'CIMA' | 'BAIXO' | 'ESQUERDA' | 'DIREITA';
export type ResultadoAcao = 'EXECUTADA' | 'BLOQUEADA' | 'ABORTADA' | 'DESCARTADA' | 'IGNORADA';

export interface Posicao {
  x: number;
  y: number;
}

export interface Config {
  largura: number;
  altura: number;
  limite_turnos: number;
  turno_fechamento?: number; // EST-09: ausente ou 0 = sem fechamento
  prazo_planejamento_ms: number; // EST-01
  duracao_etapa_ms: number; // EST-02
}

export interface Bomba {
  posicao: Posicao;
  jogador_id: string;
  potencia: number;
  pavio_restante: number;
}

export interface Morte {
  turno: number;
  etapa: number;
}

export interface Jogador {
  id: string;
  posicao: Posicao; // EST-08: se morto, a casa onde morreu
  status: Status;
  morte?: Morte; // EST-07: ausente enquanto vivo
  bombas_por_turno: number;
  potencia: number;
  pavio_padrao: number;
  acoes_por_turno: number;
  bot_versao: string;
}

export interface Estado {
  turno: number;
  config: Config;
  etapas_neste_turno: number;
  blocos_fixos: Posicao[];
  blocos_destrutiveis: Posicao[];
  bombas: Bomba[];
  jogadores: Jogador[];
}

export interface Acao {
  etapa: number;
  tipo: TipoAcao;
  direcao?: Direcao;
}

export interface JogadorEtapa {
  id: string;
  posicao: Posicao;
  status: Status;
  acao: Acao;
  resultado: ResultadoAcao; // DEC-09
}

export interface Explosao {
  origem: Posicao;
  potencia: number;
  chamas: Posicao[];
}

export interface RelatorioEtapa {
  turno: number;
  etapa: number;
  jogadores: JogadorEtapa[]; // todos, na ordem do Estado
  bombas: Bomba[]; // no tabuleiro ao fim da etapa
  explosoes: Explosao[];
  chamas: Posicao[];
  mortes: string[];
  blocos_destruidos: Posicao[];
  movimentos_bloqueados: string[];
  blocos_fechados: Posicao[]; // FEC-03: só na última etapa do turno
}

export interface Desfecho {
  terminada: boolean;
  empate: boolean;
  vencedor?: string; // ausente em empate ou partida em andamento
  sobreviventes: string[];
}

// RespostaEstado é a resposta de GET /partidas/{nome}/estado (API-05).
export interface RespostaEstado {
  nome: string;
  fase: Fase;
  turno: number;
  etapa: number; // 0 no planejamento
  horario_servidor: string; // RFC 3339, UTC, com milissegundos
  fim_da_fase?: string; // ausente em ENCERRADA
  estado: Estado; // o do início do turno
  etapas: RelatorioEtapa[]; // só as já liberadas
  desfecho: Desfecho;
}

// ResumoPartida é um item de GET /partidas (API-10).
export interface ResumoPartida {
  nome: string;
  fase: Fase;
  turno: number;
  fim_da_fase?: string;
  desfecho: Desfecho;
}

export interface RespostaPartidas {
  partidas: ResumoPartida[];
  horario_servidor: string;
}
