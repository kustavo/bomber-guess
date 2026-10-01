# Marco 14: plano técnico

**Spec**: `spec.md` (aprovada)
**Status**: aprovado em 2026-10-01

## Visão geral

Só o frontend muda. Os sprites são **dados** num módulo TypeScript (`sprites.ts`): cada um tem 16 linhas de 16 letras, e cada letra é uma cor de uma paleta comum (decisões 1 e 2). Uma função pura converte cada sprite em retângulos, juntando os pixels vizinhos da mesma cor numa linha. O componente `SpritesSvg.svelte` desenha todos os sprites **uma vez**, como `<symbol id="sprite-<nome>">` num SVG escondido montado pelo `App`. Os tabuleiros e a paleta só fazem `<use href="#sprite-<nome>">`.

A cor do jogador entra por `currentColor`: no sprite `jogador`, a letra da roupa tem a cor `currentColor`, e cada `<use>` recebe `style="color: <corJogador>"`. A forma da chama é calculada por uma função pura (`chamas.ts`) a partir das `explosoes` da etapa, que passam a fazer parte do tabuleiro exibido. As animações ficam numa folha de estilos global (`sprites.css`), com `prefers-reduced-motion`.

Os atributos `data-*` que os testes dos marcos 6 e 7 usam continuam nos mesmos elementos (CA-04).

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| `frontend/src/lib/sprites.ts` (+ `.test.ts`) | paleta, os 12 sprites em texto, `NOMES_SPRITES`, `retangulos` (CA-01) |
| `frontend/src/lib/chamas.ts` (+ `.test.ts`) | `FormaChama`, `formasDasChamas` (CA-03) |
| `frontend/src/lib/tabuleiro.ts` (+ `.test.ts`) | `TabuleiroExibido` ganha `explosoes` (as da etapa exibida; vazio no início do turno) |
| `frontend/src/componentes/SpritesSvg.svelte` (+ `.test.ts`) | SVG escondido com um `<symbol>` por sprite (CA-01, CA-02) |
| `frontend/src/componentes/Sprite.svelte` | `<use>` de um sprite numa casa, com cor e classe opcionais |
| `frontend/src/componentes/Tabuleiro.svelte` (+ `.test.ts`) | piso, blocos, bombas, chamas e jogadores com sprites; marca de bloqueio (CA-04 a CA-09, CA-12) |
| `frontend/src/componentes/TabuleiroEditor.svelte` | piso, blocos e posições iniciais com sprites; casas clicáveis transparentes por cima (CA-10) |
| `frontend/src/telas/TelaCriar.svelte` (+ `.test.ts`) | ícones da paleta com sprites (CA-11) |
| `frontend/src/estilos/sprites.css` (+ `frontend/src/lib/estilos.test.ts`) | `crispEdges`, animações da bomba e da chama, `prefers-reduced-motion` (CA-12) |
| `frontend/src/App.svelte` | monta `SpritesSvg` uma vez e importa `sprites.css` |
| `frontend/src/testes/fabricas.ts` | `explosao(origem, chamas)` para os testes |

## Tipos e assinaturas

```ts
// sprites.ts
export const LADO_SPRITE = 16;                                // decisão 1
export const PALETA: Readonly<Record<string, string>>;       // letra → cor; '.' é transparente; 'c' é currentColor
export const NOMES_SPRITES: readonly NomeSprite[];
export type NomeSprite =
  | 'piso' | 'bloco-fixo' | 'bloco-destrutivel' | 'bomba' | 'jogador'
  | 'chama-centro' | 'chama-horizontal' | 'chama-vertical'
  | 'chama-ponta-cima' | 'chama-ponta-baixo' | 'chama-ponta-esquerda' | 'chama-ponta-direita';
export const SPRITES: Readonly<Record<NomeSprite, readonly string[]>>; // 16 linhas de 16 letras
export interface Retangulo { x: number; y: number; largura: number; cor: string }
export function retangulos(desenho: readonly string[]): Retangulo[]; // junta pixels vizinhos da mesma cor na linha
export const idSprite = (nome: NomeSprite) => `sprite-${nome}`;

// chamas.ts
export type FormaChama =
  | 'centro' | 'horizontal' | 'vertical'
  | 'ponta-cima' | 'ponta-baixo' | 'ponta-esquerda' | 'ponta-direita';
// formasDasChamas devolve a forma de cada casa de `chamas`, pela chave "x,y" (CA-03).
export function formasDasChamas(chamas: Posicao[], explosoes: Explosao[]): Map<string, FormaChama>;

// tabuleiro.ts (TabuleiroExibido ganha)
explosoes: Explosao[]; // as da etapa exibida (CA-09)
```

```svelte
<!-- Sprite.svelte -->
<script lang="ts">
  interface Props { nome: NomeSprite; x: number; y: number; cor?: string; classe?: string }
</script>
```

## Decisões

- **D1**: Sprites como `<symbol>` definidos uma vez (`SpritesSvg` no `App`) e usados com `<use href="#sprite-…">`. **Motivo**: um tabuleiro de 30 × 30 teria 900 pisos. Com `<use>`, cada casa custa um elemento, e não centenas de retângulos. Os componentes testados isoladamente continuam renderizando (o `<use>` aponta para um id que não existe no teste), e os testes conferem o `href`.
- **D2**: `retangulos` junta pixels consecutivos da mesma cor numa linha. **Motivo**: um sprite de 256 pixels vira algumas dezenas de retângulos. O desenho fica fiel e o DOM, leve.
- **D3**: Paleta pequena e comum a todos os sprites (por volta de 14 cores), com letras fixas: contornos, cinzas do bloco fixo, marrons do bloco destrutível, verdes do piso, pretos e branco da bomba, amarelo/laranja/vermelho da chama, pele e a letra `c` = `currentColor` para a roupa do jogador. **Motivo**: o CA-01 exige só cores da paleta, e a coerência visual sai de graça.
- **D4**: Os elementos com `data-tipo` continuam sendo os mesmos `<g>` ou `<rect>` de hoje: o `<use>` do sprite entra **dentro** deles (ou ao lado, no caso dos blocos, que viram `<g data-tipo=…>`). O piso e os ícones da paleta **não** têm `data-tipo`. **Motivo**: CA-04. Os testes dos marcos 6 e 7 contam `[data-tipo]` e conferem `data-x`, `data-y` e `textContent` (o pavio da bomba é o único texto dentro dela).
- **D5**: O número do jogador continua como `<text>` pequeno no canto da casa, com contorno escuro, e não dentro do sprite. A marca de bloqueio vira um quadrado vermelho em volta da casa (`.marca-bloqueio`), e o `<title>` com `BLOQUEADA` continua. **Motivo**: CA-05 e CA-07. O número muda por jogador; o sprite é um só.
- **D6**: `formasDasChamas` olha cada explosão: a origem é `centro`; as outras casas da linha da origem são `horizontal` e as da coluna, `vertical`; a mais distante em cada direção é a `ponta`. Se explosões diferentes dão formas diferentes à mesma casa, fica `centro`. Uma casa de `chamas` fora de qualquer explosão também é `centro`. Braço de uma casa só é `ponta`. **Motivo**: decisão 3 e CA-03. No backend, `Explosao.chamas` já inclui a origem.
- **D7**: `TabuleiroExibido.explosoes` vem do relatório da etapa exibida (`atual.explosoes`), como as `chamas`, e não acumula. **Motivo**: segue a regra do CA-02 do marco 6 para chamas ("só da etapa k").
- **D8**: As animações ficam em `src/estilos/sprites.css`, importado pelo `App`, com classes `animada-bomba` (pulso de escala) e `animada-chama` (tremor de opacidade). Com `@media (prefers-reduced-motion: reduce)`, elas ficam `animation: none`, e o deslize do jogador (`transition`) vira `none`. O teste do CA-12 lê o arquivo CSS com `fs` e confere as regras. **Motivo**: o jsdom não avalia `@media` nem animações; ler a folha de estilos é o "teste da folha de estilos" que a spec permite. A transição do jogador hoje está em `style=` inline; passa a ser uma variável CSS (`--transicao`) usada pela classe, para que a media query consiga desligá-la.
- **D9**: No editor, a ordem das camadas é: piso, itens e, por cima, os `<rect data-casa>` transparentes, que recebem clique, teclado e o realce de foco. **Motivo**: o CA-08 do marco 7 clica em `[data-casa]`. Assim o item nunca rouba o clique, e o foco do teclado continua visível.
- **D10**: `shape-rendering: crispEdges` no `<symbol>` e `image-rendering: pixelated` no SVG do tabuleiro. **Motivo**: decisão 1, sprites nítidos ao ampliar (CA-14).

## Reabertura (2026-10-01): CA-15 a CA-17

- **D11** (CA-15): `Sprite` ganha a prop `centrado`, que põe `transform-origin` no centro da própria casa (`x + 0.5`, `y + 0.5`, em unidades do SVG). Em `sprites.css`, `.animada-bomba` usa `transform-box: view-box` e o pulso vai de `scale(1)` a `scale(1.04)` em 1,2 s. **Motivo**: com `fill-box`, o centro era o da caixa do desenho da bomba, e não o da casa; por isso ela parecia andar.
- **D12** (CA-16): componente `Grade.svelte` (`largura`, `altura`) com um único `<path>` de linhas internas, `pointer-events="none"`, traço branco com opacidade 0,35 e espessura 0,03. No `Tabuleiro` ele entra depois dos blocos e antes das chamas; no `TabuleiroEditor`, depois dos blocos e antes das posições iniciais. As casas clicáveis do editor perdem o contorno escuro, mas mantêm o realce de hover e foco.
- **D13** (CA-17): `JogadorExibido` ganha `bombas_por_turno` e `bombas_usadas`. `calcularTabuleiro` conta, nos relatórios das etapas 1…k, as ações `PLANTAR` com resultado `EXECUTADA` do jogador. `textos.ts` ganha `textoBombasUsadas(usadas, total)` → `"usou u de n (u/n)"`, e as restantes são `max(0, n − u)`.
- **D14** (CA-17): o `Tabuleiro` passa a ficar dentro de um `<div class="caixa-tabuleiro">` com `position: relative`. O jogador sob o mouse (ou com foco) fica num estado local, e o cartão é um `<div data-cartao>` posicionado em porcentagem da casa, à direita do jogador (ou à esquerda, nas duas últimas colunas). O `<g>` do jogador ganha `tabindex="0"`, `role="button"`, `aria-label` com o resumo e `outline: none` (a lição do CA-21 do marco 7).

## Riscos

- **Linhas finas entre casas** em alguns zooms, por arredondamento de subpixel. Mitigação: o piso de cada casa sobra 0,01 para os lados, ou os pixels de borda usam a mesma cor do vizinho. Confirmar na verificação manual (CA-14).
- **Desenho em texto dá trabalho de ajustar.** Os sprites são originais e feitos à mão; a primeira versão pode precisar de retoques depois de ver na tela. Isso fica na T10 (verificação manual), sem mudar a spec.
- **`<use>` para um id inexistente** nos testes isolados não quebra nada, mas também não prova o desenho. O desenho é provado pelos testes de `sprites.ts` e de `SpritesSvg`, e pela verificação manual.
