<script lang="ts">
  // SVG do tabuleiro exibido (spec 06, CA-01 a CA-04; plano, D8). Uma casa = 1 unidade.
  // Desenhado com os sprites do marco 14 (plano 14, D4 e D5).
  import { formasDasChamas } from '../lib/chamas';
  import type { TabuleiroExibido } from '../lib/tabuleiro';
  import { corJogador } from '../lib/textos';
  import Sprite from './Sprite.svelte';

  interface Props {
    tabuleiro: TabuleiroExibido;
    duracaoEtapaMs?: number;
  }

  let { tabuleiro, duracaoEtapaMs = 1000 }: Props = $props();

  const transicao = $derived(Math.min(duracaoEtapaMs / 2, 300));
  const vivos = $derived(tabuleiro.jogadores.filter((j) => j.status === 'VIVO'));
  const casas = $derived(
    Array.from({ length: tabuleiro.altura }, (_, y) => Array.from({ length: tabuleiro.largura }, (_, x) => ({ x, y }))).flat(),
  );
  const formas = $derived(formasDasChamas(tabuleiro.chamas, tabuleiro.explosoes));
</script>

<svg
  class="tabuleiro"
  viewBox="0 0 {tabuleiro.largura} {tabuleiro.altura}"
  role="img"
  aria-label="Tabuleiro, turno {tabuleiro.turno}, etapa {tabuleiro.etapa}"
>
  <!-- fundo na cor do piso: esconde as frestas de subpixel entre as casas (plano 14, Riscos) -->
  <rect class="fundo" width={tabuleiro.largura} height={tabuleiro.altura} />
  {#each casas as c (`${c.x},${c.y}`)}
    <Sprite nome="piso" x={c.x} y={c.y} />
  {/each}
  {#each tabuleiro.blocosFixos as b (`${b.x},${b.y}`)}
    <g data-tipo="bloco-fixo" data-x={b.x} data-y={b.y}>
      <Sprite nome="bloco-fixo" x={b.x} y={b.y} />
    </g>
  {/each}
  {#each tabuleiro.blocosDestrutiveis as b (`${b.x},${b.y}`)}
    <g data-tipo="bloco-destrutivel" data-x={b.x} data-y={b.y}>
      <Sprite nome="bloco-destrutivel" x={b.x} y={b.y} />
    </g>
  {/each}
  {#each tabuleiro.chamas as c (`${c.x},${c.y}`)}
    <g class="chama" data-tipo="chama" data-x={c.x} data-y={c.y}>
      <Sprite nome="chama-{formas.get(`${c.x},${c.y}`) ?? 'centro'}" x={c.x} y={c.y} classe="animada-chama" />
    </g>
  {/each}
  {#each tabuleiro.bombas as b, i (`${b.posicao.x},${b.posicao.y},${i}`)}
    <g class="bomba" data-tipo="bomba" data-x={b.posicao.x} data-y={b.posicao.y} data-pavio={b.pavio_restante}>
      <Sprite nome="bomba" x={b.posicao.x} y={b.posicao.y} classe="animada-bomba" />
      <text x={b.posicao.x + 0.44} y={b.posicao.y + 0.62}>{b.pavio_restante}</text>
    </g>
  {/each}
  {#each vivos as j (j.id)}
    <g
      class="jogador deslize"
      class:bloqueado={j.bloqueado}
      data-tipo="jogador"
      data-jogador={j.id}
      data-x={j.posicao.x}
      data-y={j.posicao.y}
      data-bloqueado={j.bloqueado}
      style="transform: translate({j.posicao.x}px, {j.posicao.y}px); --transicao: {transicao}ms"
    >
      <title>{j.id}{j.bloqueado ? ': BLOQUEADA' : ''}</title>
      <Sprite nome="jogador" x={0} y={0} cor={corJogador(j.indice)} />
      {#if j.bloqueado}
        <rect class="marca-bloqueio" x="0.04" y="0.04" width="0.92" height="0.92" />
      {/if}
      <text class="numero" x="0.86" y="0.2">{j.indice + 1}</text>
    </g>
  {/each}
</svg>

<style>
  .tabuleiro {
    width: 100%;
    max-width: 640px;
    display: block;
    image-rendering: pixelated;
  }
  .fundo {
    fill: #4a8a3f; /* PALETA.G */
  }
  .bomba text,
  .numero {
    fill: #fff;
    font-family: system-ui, sans-serif;
    font-weight: 700;
    text-anchor: middle;
    dominant-baseline: central;
    paint-order: stroke;
    stroke: #1b1b22;
  }
  .bomba text {
    font-size: 0.36px;
    stroke-width: 0.06px;
  }
  .numero {
    font-size: 0.3px;
    stroke-width: 0.08px;
  }
  .marca-bloqueio {
    fill: none;
    stroke: #ff1f1f;
    stroke-width: 0.08;
  }
</style>
