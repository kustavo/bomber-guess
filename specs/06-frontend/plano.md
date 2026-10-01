# Marco 06: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

Um app Svelte 5 + Vite em TypeScript, sem SvelteKit, em `frontend/`. O backend não muda. A lógica fica em módulos TypeScript puros, testados sem DOM: o tabuleiro exibido, o cronômetro, o cliente da API e o ritmo da animação. Um **acompanhamento** por partida junta essas peças. Ele consulta o estado periodicamente, mantém a fila de etapas a exibir e publica uma `VisaoTela` numa store do Svelte. Os componentes só desenham essa visão. O servidor de desenvolvimento do Vite encaminha `/partidas` e `/bots` para o servidor Go (`localhost:8080`). As rotas da tela usam hash (`#/`, `#/partidas/<nome>`), sem biblioteca.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `frontend/package.json` | dependências e scripts `dev`, `build`, `check` (svelte-check + tsc), `test` (vitest run) |
| `frontend/vite.config.ts` | plugin Svelte, proxy de `/partidas` e `/bots` para `http://localhost:8080` (alvo trocável por `BOMBER_API`), configuração do Vitest (jsdom, `resolve.conditions: ['browser']` nos testes) |
| `frontend/tsconfig.json`, `frontend/tsconfig.node.json`, `frontend/svelte.config.js` | TypeScript estrito (app e `vite.config.ts`, como no template oficial); `vitePreprocess` |
| `frontend/index.html`, `frontend/src/main.ts` | ponto de entrada; monta `App.svelte` |
| `frontend/src/lib/tipos.ts` | tipos do JSON da API (espelho de `jogo` e `api`, nomes do glossário, campos em `snake_case`) |
| `frontend/src/lib/api.ts` | cliente HTTP: `buscarPartidas`, `buscarEstado`, `ErroApi` |
| `frontend/src/lib/tabuleiro.ts` | `calcularTabuleiro`: tabuleiro exibido a partir do `estado` e dos relatórios (CA-01 a CA-04) |
| `frontend/src/lib/cronometro.ts` | defasagem e tempo restante (CA-05, CA-06) |
| `frontend/src/lib/ritmo.ts` | intervalo de consulta (decisão 2) e intervalo entre etapas exibidas (decisão 5) |
| `frontend/src/lib/acompanhamento.ts` | consulta periódica, fila de etapas, erros, parada (CA-08 a CA-12, CA-14, CA-15) |
| `frontend/src/lib/textos.ts` | textos em português: fase, desfecho, resultado da ação, tempo restante (`m:ss.d`) |
| `frontend/src/lib/rota.ts` | `lerRota(hash)` e `enderecoPartida(nome)` |
| `frontend/src/App.svelte` | escolhe a tela pela rota; reage a `hashchange` |
| `frontend/src/telas/ListaPartidas.svelte` | lista (API-10), consulta a cada 2 s (CA-13) |
| `frontend/src/telas/TelaPartida.svelte` | cabeçalho (turno, etapa, fase), tabuleiro, painel, cronômetro ou desfecho, avisos (CA-07, CA-14, CA-15) |
| `frontend/src/componentes/Tabuleiro.svelte` | SVG do tabuleiro exibido (CA-01 a CA-04) |
| `frontend/src/componentes/PainelJogadores.svelte` | id, bot, status, morte, resultado da ação na etapa (CA-03) |
| `frontend/src/componentes/Cronometro.svelte` | atualiza o tempo restante a cada 100 ms (CA-05) |
| `frontend/src/testes/fabricas.ts` | construtores de `Estado`, `RelatorioEtapa`, `RespostaEstado` e da API falsa para os testes |
| `frontend/src/**/*.test.ts` | testes Vitest, ao lado de cada módulo ou componente |
| `.gitignore` | acrescentar `frontend/node_modules/` e `frontend/dist/` |

## Tipos e assinaturas

```ts
// tipos.ts: espelho exato do JSON (docs/REGRAS.md, ARQUITETURA 1.2, spec 05 decisões 5 e 9)
export type Fase = 'PLANEJAMENTO' | 'EXECUCAO' | 'ENCERRADA';
export type Status = 'VIVO' | 'MORTO';
export type TipoAcao = 'MOVER' | 'PLANTAR' | 'ESPERAR';
export type Direcao = 'CIMA' | 'BAIXO' | 'ESQUERDA' | 'DIREITA';
export type ResultadoAcao = 'EXECUTADA' | 'BLOQUEADA' | 'ABORTADA' | 'DESCARTADA' | 'IGNORADA';
export interface Posicao { x: number; y: number }
export interface Area { largura: number; altura: number }
export interface Config { largura: number; altura: number; limite_turnos: number; turno_fechamento?: number; area_minima?: Area; prazo_planejamento_ms: number; duracao_etapa_ms: number }
export interface Bomba { posicao: Posicao; jogador_id: string; potencia: number; pavio_restante: number }
export interface Morte { turno: number; etapa: number }
export interface Jogador { id: string; posicao: Posicao; status: Status; morte?: Morte; bombas_por_turno: number; potencia: number; pavio_padrao: number; acoes_por_turno: number; bot_versao: string }
export interface Estado { turno: number; config: Config; etapas_neste_turno: number; blocos_fixos: Posicao[]; blocos_destrutiveis: Posicao[]; bombas: Bomba[]; jogadores: Jogador[] }
export interface Acao { etapa: number; tipo: TipoAcao; direcao?: Direcao }
export interface JogadorEtapa { id: string; posicao: Posicao; status: Status; acao: Acao; resultado: ResultadoAcao }
export interface Explosao { origem: Posicao; potencia: number; chamas: Posicao[] }
export interface RelatorioEtapa { turno: number; etapa: number; jogadores: JogadorEtapa[]; bombas: Bomba[]; explosoes: Explosao[]; chamas: Posicao[]; mortes: string[]; blocos_destruidos: Posicao[]; movimentos_bloqueados: string[]; blocos_fechados: Posicao[] }
export interface Desfecho { terminada: boolean; empate: boolean; vencedor?: string; sobreviventes: string[] }
export interface RespostaEstado { nome: string; fase: Fase; turno: number; etapa: number; horario_servidor: string; fim_da_fase?: string; estado: Estado; etapas: RelatorioEtapa[]; desfecho: Desfecho }
export interface ResumoPartida { nome: string; fase: Fase; turno: number; fim_da_fase?: string; desfecho: Desfecho }
export interface RespostaPartidas { partidas: ResumoPartida[]; horario_servidor: string }

// api.ts
export class ErroApi extends Error { readonly status: number /* 0 = falha de rede */ }
export interface ClienteApi {
  buscarPartidas(): Promise<RespostaPartidas>;
  buscarEstado(nome: string): Promise<RespostaEstado>; // rejeita com ErroApi
}
export function criarCliente(buscar?: typeof fetch, base?: string): ClienteApi;

// tabuleiro.ts
export interface JogadorExibido { id: string; indice: number /* ordem em estado.jogadores: cor */; posicao: Posicao; status: Status; morte?: Morte; bot_versao: string; acao?: Acao; resultado?: ResultadoAcao; bloqueado: boolean }
export interface TabuleiroExibido {
  largura: number; altura: number;
  turno: number; etapa: number; // 0 = início do turno
  blocosFixos: Posicao[]; blocosDestrutiveis: Posicao[]; bombas: Bomba[]; chamas: Posicao[];
  jogadores: JogadorExibido[]; // todos; quem ocupa casa são só os VIVO (CA-03)
}
export function calcularTabuleiro(estado: Estado, etapas: RelatorioEtapa[], ate: number): TabuleiroExibido;

// cronometro.ts (tudo em ms desde a época)
export function calcularDefasagem(horarioServidor: string, recebidaEmLocal: number): number;
export function tempoRestante(fimDaFase: string | undefined, defasagem: number, agoraLocal: number): number | null; // null sem fim de fase; nunca negativo

// ritmo.ts
export const CONSULTA_MAXIMA_MS = 250;
export const CONSULTA_LISTA_MS = 2000;
export function intervaloConsulta(duracaoEtapaMs?: number): number;              // min(250, d/2)
export function intervaloEntreEtapas(atrasadas: number, duracaoEtapaMs: number): number; // d se em dia; d/4 se atrasado

// acompanhamento.ts
export interface Relogio { agora(): number; esperar(ms: number, f: () => void): () => void /* cancela */ }
export const relogioReal: Relogio;
export interface VisaoTela {
  carregando: boolean;
  naoEncontrada: boolean;  // CA-14
  semConexao: boolean;     // CA-15
  erro?: string;
  resposta?: RespostaEstado; // a última válida
  tabuleiro?: TabuleiroExibido;
  defasagem: number;
}
export interface Acompanhamento { visao: Readable<VisaoTela>; parar(): void }
export function acompanhar(nome: string, cliente: ClienteApi, relogio?: Relogio): Acompanhamento;

// rota.ts
export type Rota = { tela: 'lista' } | { tela: 'partida'; nome: string };
export function lerRota(hash: string): Rota;
export function enderecoPartida(nome: string): string; // '#/partidas/<nome>'
```

## Decisões

- **D1**: toda a lógica sai dos componentes e fica em módulos `.ts` puros. Os componentes só leem a `VisaoTela`. **Motivo**: CA-01 a CA-12 viram testes simples, sem DOM e sem esperas reais. Os testes de componente ficam só para o que é visual (marca de bloqueio, mensagens, lista).
- **D2**: o tempo é injetado por `Relogio` em `acompanhar`, e não por `vi.useFakeTimers`. **Motivo**: a spec pede relógio de teste. Um relógio próprio deixa explícito o instante em que cada resposta chega (CA-05 e CA-06) e não depende de detalhes do Vitest.
- **D3**: o tabuleiro exibido é sempre recalculado do zero, a partir de `estado` e `etapas[0..ate)`. Não há estado incremental. **Motivo**: mesma entrada dá a mesma saída (convenção do projeto), e recarregar a página não precisa de respostas anteriores (CA-12). Os blocos destrutíveis exibidos são os do `estado` menos a união dos `blocos_destruidos` das etapas 1…k (CA-02). Bombas, chamas e jogadores vêm só do relatório k.
- **D4**: fila de etapas no acompanhamento.
  - A tela guarda `(turno, etapaExibida)`. A cada resposta, `etapaLiberada = resposta.etapas.length`.
  - Enquanto `etapaExibida < etapaLiberada`, avança uma etapa a cada `intervaloEntreEtapas`, que vale `d` se só falta uma e `d/4` se faltam mais (decisão 5, CA-10).
  - Turno novo na resposta: descarta a fila e mostra o novo `estado` (CA-11).
  - `ENCERRADA` (CA-09): a consulta para, mas a fila termina de exibir o que falta, e o desfecho aparece por cima.
- **D5**: a defasagem é `horario_servidor − instante local de chegada da resposta` e é recalculada a cada resposta. Não se usa o ponto médio da ida e volta. **Motivo**: é o que a CA-05 especifica. Com a API no mesmo host, o erro é de poucos milissegundos.
- **D6**: tratamento de erro do cliente:
  - 404 vira `naoEncontrada` e a consulta para (CA-14);
  - falha de rede (`status` 0) ou 5xx vira `semConexao`, mantendo a última `resposta` e o tabuleiro, e a consulta continua no mesmo intervalo (CA-15);
  - outros 4xx mostram `erro` e a consulta continua.
- **D7**: só uma consulta de cada vez. A próxima é agendada quando a anterior termina, com `intervaloConsulta(d)`; antes da primeira resposta, `d` é desconhecido e o intervalo é 250 ms. **Motivo**: evita respostas fora de ordem e acúmulo de requisições com o servidor lento. A CA-08 é verificada contando as chamadas ao `ClienteApi` falso conforme o relógio avança.
- **D8**: tabuleiro em SVG, com uma casa = 1 unidade de `viewBox`. Cada elemento tem `data-*` (`data-tipo="bloco-fixo"`, `data-x`, `data-y`, `data-jogador`) para os testes acharem pelo DOM. Movimento por `transform` com `transition` CSS de `min(d/2, 300 ms)`. Bloqueio: contorno vermelho no jogador e `title` "BLOQUEADA" (CA-04). Bomba: círculo com o `pavio_restante`. As cores dos 4 jogadores são fixas por índice.
- **D9**: os nomes dos testes começam pelo ID da regra e citam o CA, por exemplo `"API-05 CA-02 blocos destruídos acumulam até a etapa k"`, como em `specs/README.md`.
- **D10**: `fabricas.ts` monta respostas pequenas à mão, tipadas por `tipos.ts`. Não há geração automática a partir do Go. **Motivo**: o backend fica intocado neste marco. Uma divergência de formato aparece na verificação manual (CA-16).
- **D11**: não se usa o template do `npm create vite`: o scaffold é escrito à mão, mínimo. As versões são fixadas no `package-lock.json`. **Motivo**: nada de arquivos de exemplo sobrando, e a instalação fica reprodutível.
- **D12** (registrada na implementação): ajustes pequenos, sem mudar a spec.
  - `corJogador` ficou em `textos.ts`, com os outros auxiliares de exibição; não ganhou arquivo próprio.
  - `fabricas.ts` também tem o relógio de teste (`relogioFalso`) e a API falsa (`clienteFalso`).
  - `App.svelte` recebe `cliente` e `relogio` opcionais, para os testes.
  - `.claude/launch.json` sobe o servidor Go e o Vite para a verificação manual (CA-16).
  - Um componente `.svelte` sem `<script lang="ts">` não ganha tipos no `svelte-check`; todos os componentes têm o bloco.
- **D13** (2026-10-01, marco 12): `calcularTabuleiro` acumula os `blocos_fechados` das etapas 1…k. Eles entram em `blocosFixos` e saem de `blocosDestrutiveis` (CA-17). `fabricas.ts` passa a preencher `blocos_fechados: []`. O SVG não muda, porque bloco fixo já é desenhado. A área mínima (`area_minima`, EST-10 e FEC-08) só entra nos tipos: a tela não precisa dela, porque as casas que fecham já chegam em `blocos_fechados`. **Motivo**: segue a D3 (recalcular do zero) e o mesmo padrão dos `blocos_destruidos`.

## Riscos

- Node 26 e npm 12 instalados pelo usuário (pacman) em 2026-10-01. O plano usa só npm; `deno` não entra.
- `@testing-library/svelte` com Svelte 5 exige `resolve.conditions: ['browser']` no Vitest; sem isso, `mount` falha no SSR. Já está previsto no `vite.config.ts`.
- `/partidas` é ao mesmo tempo rota da API (proxy) e prefixo da rota da tela. Não há conflito porque a tela usa hash (`#/partidas/...`), que nunca chega ao servidor.
- O `AGENTS.md` só fala de verificação em Go (`gofmt`, `go vet`, `go test`). A verificação final deste marco acrescenta `npm run check` e `npm test` em `frontend/`. Proposta: acrescentar essa linha às convenções do `AGENTS.md` no fim do marco, com aprovação do usuário.
- Fixtures escritas à mão podem divergir do JSON real (D10). A CA-16 manual é a rede de segurança.
