# Marco 07: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

No backend, o `Gerenciador` (pacote `partida`) ganha `SalvarMapa`, porque já é dono do diretório de mapas e da regra de nomes (marco 5, D8). A API troca o 501 de `POST /mapas` por um handler que decodifica o mapa, chama `SalvarMapa` e traduz os erros em 400 ou 409. O pacote `jogo` não muda, porque `VerificarMapa` já cobre o MAP-05.

No frontend, a lógica fica em dois módulos TypeScript puros, testados sem DOM:
- `editor.ts`: o **mapa em edição** e as operações sobre ele (ferramenta, redimensionar, converter para o JSON do mapa, pendências);
- `criacao.ts`: a sequência `POST /mapas` → `POST /partidas`, que lembra o mapa já salvo.

A nova tela `TelaCriar.svelte` (rota `#/criar`) junta tudo isso com o formulário, a paleta e um tabuleiro clicável (`TabuleiroEditor.svelte`). O cliente da API ganha `buscarBots`, `salvarMapa` e `criarPartida`, e o proxy do Vite passa a encaminhar `/mapas`.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `backend/internal/partida/gerenciador.go` | `SalvarMapa(m jogo.Mapa) error` |
| `backend/internal/partida/gerenciador_test.go` | casos de `SalvarMapa`: nome, MAP-05, nome repetido, arquivo gravado (CA-01, CA-03 a CA-05) |
| `backend/internal/api/api.go` | rota `/mapas` com `POST: a.salvarMapa`; handler `salvarMapa` |
| `backend/internal/api/respostas.go` | `respostaMapa` |
| `backend/internal/api/api_test.go` | `POST /mapas` ponta a ponta no `httptest` (CA-01 a CA-06); a tabela do CA-24 do marco 5 perde o caso `POST /mapas` 501 e ganha `GET /mapas` 405 |
| `frontend/vite.config.ts` | proxy de `/mapas` |
| `frontend/src/lib/tipos.ts` | `Atributos`, `Mapa`, `RespostaBots`, `RespostaMapa`, `PedidoPartida` |
| `frontend/src/lib/api.ts` | `buscarBots`, `salvarMapa`, `criarPartida`; `obter` vira `pedir(metodo, caminho, corpo?)` |
| `frontend/src/lib/editor.ts` (+ `.test.ts`) | mapa em edição: padrões, ferramenta, redimensionar, conversão, pendências (CA-07 a CA-11, CA-13) |
| `frontend/src/lib/criacao.ts` (+ `.test.ts`) | `iniciarPartida` (CA-14 a CA-16) |
| `frontend/src/lib/rota.ts` (+ `.test.ts`) | rota `{ tela: 'criar' }` e `ENDERECO_CRIAR` (CA-17) |
| `frontend/src/componentes/TabuleiroEditor.svelte` | SVG com uma casa clicável por posição; itens e números das posições iniciais (CA-08, CA-09) |
| `frontend/src/telas/TelaCriar.svelte` (+ `.test.ts`) | formulário, paleta, tabuleiro, seletores de bot, Iniciar, avisos e erros (CA-12 a CA-16) |
| `frontend/src/telas/ListaPartidas.svelte` | link "Criar partida"; o texto de lista vazia passa a apontar para ele (CA-17) |
| `frontend/src/App.svelte` (+ `App.test.ts`) | despacha `criar` para `TelaCriar` (CA-17) |
| `frontend/src/testes/fabricas.ts` | `clienteFalso` com os três métodos novos, que registram as chamadas e têm respostas configuráveis |

## Tipos e assinaturas

```go
// partida/gerenciador.go

// SalvarMapa grava m em <DirMapas>/<m.Nome>.json (API-03). Os erros embrulham
// ErrPedidoInvalido (nome fora de nomeValido, MAP-05 via jogo.VerificarMapa)
// ou ErrNomeRepetido (o arquivo já existe; nada é sobrescrito).
func (g *Gerenciador) SalvarMapa(m jogo.Mapa) error
```

```go
// api/respostas.go
type respostaMapa struct {
	Nome            string  `json:"nome"`
	HorarioServidor horario `json:"horario_servidor"`
}

// api/api.go
func (a *api) salvarMapa(w http.ResponseWriter, r *http.Request)
```

```ts
// tipos.ts
export interface Atributos { bombas_por_turno: number; potencia: number; pavio_padrao: number; acoes_por_turno: number }
export interface Mapa {
  nome: string;
  config: Config;
  jogador_padrao: Atributos;
  blocos_fixos: Posicao[];
  blocos_destrutiveis: Posicao[];
  posicoes_iniciais: Posicao[];
}
export interface RespostaBots { bots: string[]; horario_servidor: string }
export interface RespostaMapa { nome: string; horario_servidor: string }
export interface PedidoPartida { nome: string; mapa: string; bots: string[]; semente: number }

// api.ts (ClienteApi ganha)
buscarBots(): Promise<RespostaBots>;
salvarMapa(mapa: Mapa): Promise<RespostaMapa>;
criarPartida(pedido: PedidoPartida): Promise<RespostaEstado>;

// editor.ts
export type Item = 'bloco-fixo' | 'bloco-destrutivel' | 'posicao-inicial';
export type Ferramenta = Item | 'borracha';
export interface PosicaoInicial { posicao: Posicao; bot: string }
export interface MapaEmEdicao {
  config: Config;                    // sempre com turno_fechamento e area_minima
  jogador_padrao: Atributos;
  blocos_fixos: Posicao[];
  blocos_destrutiveis: Posicao[];
  posicoes_iniciais: PosicaoInicial[]; // na ordem em que foram colocadas (MAP-03)
}
export const CONFIG_PADRAO: Config;          // decisão 3 (= mapas/exemplo.json)
export const JOGADOR_PADRAO: Atributos;      // decisão 3
export const LADO_MAXIMO = 30;               // decisão 3, só na tela
export const NOME_VALIDO: RegExp;            // /^[a-z0-9-]{1,64}$/ (decisão 1)
export function novoMapaEmEdicao(): MapaEmEdicao;                                         // CA-07
export function itemEm(m: MapaEmEdicao, p: Posicao): Item | undefined;
export function aplicarFerramenta(m: MapaEmEdicao, f: Ferramenta, p: Posicao, botPadrao: string): MapaEmEdicao; // CA-08, CA-09
export function redimensionar(m: MapaEmEdicao, largura: number, altura: number): MapaEmEdicao;                  // CA-10
export function paraMapa(m: MapaEmEdicao, nome: string): Mapa;                            // CA-11
export function botsDasPosicoes(m: MapaEmEdicao): string[];                               // MAP-03
export function escolherBot(m: MapaEmEdicao, i: number, bot: string): MapaEmEdicao;       // MAP-03, CA-09
export function preencherBots(m: MapaEmEdicao, bot: string): MapaEmEdicao;               // CA-12 (posições sem bot quando o catálogo chega)
export function pendencias(m: MapaEmEdicao, nomePartida: string, nomeMapa: string, bots: string[]): string[]; // CA-13
export function sugerirNomeMapa(nomePartida: string): string;                             // decisão 1

// criacao.ts
export type ResultadoInicio =
  | { ok: true }
  | { ok: false; erro: string; mapaSalvo: string | undefined };
// mapaSalvo: chave do último mapa gravado com sucesso nesta tela (decisão 5).
export function chaveMapa(mapa: Mapa): string;
export function iniciarPartida(cliente: ClienteApi, mapa: Mapa, pedido: PedidoPartida, mapaSalvo: string | undefined): Promise<ResultadoInicio>;

// rota.ts
export type Rota = { tela: 'lista' } | { tela: 'partida'; nome: string } | { tela: 'criar' };
export const ENDERECO_CRIAR = '#/criar';
```

## Decisões

- **D1**: `SalvarMapa` fica no `Gerenciador` (pacote `partida`), e não no `api`. **Motivo**: o gerenciador já lê `<DirMapas>/<mapa>.json` com `nomeValido`. A regra de nome e o caminho ficam num só lugar, e o CA-02 (salvar e usar) passa a ser garantido por construção.
- **D2**: Gravação exclusiva e atômica: escrever num arquivo temporário do próprio diretório (`.<nome>-*.tmp`), depois `os.Link` para `<nome>.json` (falha se já existe → `ErrNomeRepetido`) e apagar o temporário. **Motivo**: dois `POST /mapas` simultâneos com o mesmo nome não se sobrescrevem (CA-05), e um `POST /partidas` concorrente nunca lê um arquivo pela metade. O temporário não termina em `.json` e nunca é lido como mapa.
- **D3**: O arquivo é o JSON do mapa com `json.MarshalIndent` (2 espaços) e quebra de linha final. **Motivo**: legível e versionável como `mapas/exemplo.json`. Ida e volta por `jogo.LerMapa` dá o mesmo `Mapa` (CA-01).
- **D4**: `POST /mapas` decodifica com `DisallowUnknownFields` e o mesmo limite de corpo de `POST /partidas`. A ordem dos erros é: corpo inválido (400) → nome (400) → MAP-05 (400) → já existe (409). **Motivo**: um corpo de outro formato, como um pedido de partida, é recusado como "não é mapa" (CA-04), sem passar por uma validação confusa de MAP-05.
- **D5**: No mapa em edição, cada posição inicial guarda o bot escolhido (`PosicaoInicial`). **Motivo**: remover a posição B não troca o bot de C (CA-09). A lista de `bots` do pedido é derivada da ordem das posições (MAP-03).
- **D6**: `aplicarFerramenta` com um item numa casa que já tem o mesmo item não faz nada. Uma posição inicial não vai para o fim da fila nem perde o bot. Com outro item ou com a borracha, o que havia na casa sai, e a posição inicial sai da lista. **Motivo**: é o comportamento esperado de um "maker" e mantém o CA-08 e o CA-09 previsíveis.
- **D7**: `redimensionar` limita os lados a 1…`LADO_MAXIMO` e remove os itens de fora. Também ajusta `area_minima` para caber (mínimo entre ela e o novo lado). **Motivo**: o CA-10 e a decisão 3. Sem o ajuste, diminuir o tabuleiro abaixo de 5 × 5 deixaria o mapa inválido por um campo que o usuário não tocou.
- **D8**: A sequência de criação fica em `criacao.ts`, fora do componente. A tela guarda só a `mapaSalvo` devolvida. A chave é `JSON.stringify(paraMapa(...))`, então qualquer mudança no mapa (itens, `config`, nome) obriga a gravar de novo. **Motivo**: CA-14 a CA-16 testáveis sem DOM, e a decisão 5.
- **D9**: Campos numéricos com `<input type="number">`. Valor vazio ou não inteiro vai como está (vazio vira 0), e o servidor recusa com o `erro` de MAP-05 (decisão 8). A exceção são largura e altura, que passam por `redimensionar`. **Motivo**: uma só implementação de MAP-05.
- **D10**: Pendências do CA-13: nome da partida vazio, nome do mapa vazio, menos de 2 posições iniciais e catálogo de bots não carregado (decisão 4). Cada uma gera uma mensagem curta em português, na ordem em que aparecem no formulário. **Motivo**: o aviso diz o que falta; o resto é com o servidor.
- **D11**: O nome do mapa começa vazio e acompanha a sugestão (`sugerirNomeMapa(nomePartida)`: o nome da partida, se casar com `NOME_VALIDO`, senão vazio) até o usuário editá-lo à mão. **Motivo**: decisão 1, sem sobrescrever o que o usuário digitou.
- **D12**: Ao terminar com sucesso, a tela faz `window.location.hash = enderecoPartida(nome)`. O `App` já reage a `hashchange`. **Motivo**: segue o roteamento do marco 6, sem estado global.
- **D13**: `TabuleiroEditor` é um componente novo, sem reaproveitar `Tabuleiro.svelte`, mas com as mesmas cores e formas de bloco. As posições iniciais aparecem com o número e a cor de `corJogador(i)`. **Motivo**: o `Tabuleiro` desenha um `TabuleiroExibido` (jogadores, bombas, chamas) e não tem casas clicáveis. Adaptá-lo misturaria as duas telas.
- **D15** (CA-20): `normalizarNome(texto)` em `editor.ts` devolve o texto em minúsculas, e os dois campos de nome o aplicam no `input`. `pendencias` ganha a mensagem "O nome da partida só aceita letras minúsculas, números e -." (e a mesma para o mapa) quando o nome não vazio não casa com `NOME_VALIDO`. **Motivo**: decisão 9; a regra continua com uma só expressão (`NOME_VALIDO`).
- **D16** (CA-21): `.casa { outline: none }` em todos os estados no `TabuleiroEditor`; o realce de foco por teclado continua em `.casa:focus-visible` (preenchimento claro e contorno branco). O teste lê o `<style>` do componente, como no CA-12 do marco 14. **Motivo**: decisão 10.
- **D14**: O teclado também funciona: cada casa é um `<rect>` com `role="button"` e `tabindex`, e Enter ou espaço aplicam a ferramenta. **Motivo**: acessibilidade básica, e o `svelte-check` reclama de clique sem teclado.

## Riscos

- **`os.Link` sem suporte** (sistemas de arquivos sem hard link). É improvável no diretório `mapas/` local; se acontecer, o erro vira 500 com a mensagem. Alternativa, se preciso: `O_CREATE|O_EXCL` direto no arquivo final.
- **Mapas de teste no repositório**: o CA-19 manual grava em `mapas/` de verdade. É preciso apagar os arquivos criados ou não commitá-los (fica anotado no `tarefas.md`).
- **`ClienteApi` cresce**: todo cliente falso dos testes do marco 6 precisa dos métodos novos. A mitigação é concentrar tudo no `clienteFalso` de `fabricas.ts`.
- **Semente acima de 2^53** perde precisão no JavaScript. O campo limita a `Number.MAX_SAFE_INTEGER`. O servidor aceita `uint64`, então não há erro, só um teto menor pela tela.
