<script lang="ts">
  // SVG do tabuleiro exibido (spec 06, CA-01 a CA-04; plano, D8). Uma casa = 1 unidade.
  import type { TabuleiroExibido } from '../lib/tabuleiro';
  import { corJogador } from '../lib/textos';

  interface Props {
    tabuleiro: TabuleiroExibido;
    duracaoEtapaMs?: number;
  }

  let { tabuleiro, duracaoEtapaMs = 1000 }: Props = $props();

  const transicao = $derived(Math.min(duracaoEtapaMs / 2, 300));
  const vivos = $derived(tabuleiro.jogadores.filter((j) => j.status === 'VIVO'));
</script>

<svg
  class="tabuleiro"
  viewBox="0 0 {tabuleiro.largura} {tabuleiro.altura}"
  role="img"
  aria-label="Tabuleiro, turno {tabuleiro.turno}, etapa {tabuleiro.etapa}"
>
  <rect class="fundo" width={tabuleiro.largura} height={tabuleiro.altura} />
  {#each tabuleiro.blocosFixos as b (`${b.x},${b.y}`)}
    <rect class="bloco-fixo" data-tipo="bloco-fixo" data-x={b.x} data-y={b.y} x={b.x} y={b.y} width="1" height="1" />
  {/each}
  {#each tabuleiro.blocosDestrutiveis as b (`${b.x},${b.y}`)}
    <rect
      class="bloco-destrutivel"
      data-tipo="bloco-destrutivel"
      data-x={b.x}
      data-y={b.y}
      x={b.x + 0.05}
      y={b.y + 0.05}
      width="0.9"
      height="0.9"
      rx="0.1"
    />
  {/each}
  {#each tabuleiro.chamas as c (`${c.x},${c.y}`)}
    <rect class="chama" data-tipo="chama" data-x={c.x} data-y={c.y} x={c.x} y={c.y} width="1" height="1" />
  {/each}
  {#each tabuleiro.bombas as b, i (`${b.posicao.x},${b.posicao.y},${i}`)}
    <g class="bomba" data-tipo="bomba" data-x={b.posicao.x} data-y={b.posicao.y} data-pavio={b.pavio_restante}>
      <circle cx={b.posicao.x + 0.5} cy={b.posicao.y + 0.5} r="0.32" />
      <text x={b.posicao.x + 0.5} y={b.posicao.y + 0.5}>{b.pavio_restante}</text>
    </g>
  {/each}
  {#each vivos as j (j.id)}
    <g
      class="jogador"
      class:bloqueado={j.bloqueado}
      data-tipo="jogador"
      data-jogador={j.id}
      data-x={j.posicao.x}
      data-y={j.posicao.y}
      data-bloqueado={j.bloqueado}
      style="transform: translate({j.posicao.x}px, {j.posicao.y}px); transition: transform {transicao}ms ease-out"
    >
      <title>{j.id}{j.bloqueado ? ': BLOQUEADA' : ''}</title>
      <circle cx="0.5" cy="0.5" r="0.38" fill={corJogador(j.indice)} />
      <text x="0.5" y="0.5">{j.indice + 1}</text>
    </g>
  {/each}
</svg>

<style>
  .tabuleiro {
    width: 100%;
    max-width: 640px;
    display: block;
  }
  .fundo {
    fill: #2b3a2e;
  }
  .bloco-fixo {
    fill: #5b6470;
    stroke: #3b424b;
    stroke-width: 0.04;
  }
  .bloco-destrutivel {
    fill: #a07a4f;
    stroke: #6d5235;
    stroke-width: 0.04;
  }
  .chama {
    fill: #ff8a2a;
    opacity: 0.85;
  }
  .bomba circle {
    fill: #111;
  }
  .bomba text,
  .jogador text {
    fill: #fff;
    font-size: 0.4px;
    font-family: system-ui, sans-serif;
    text-anchor: middle;
    dominant-baseline: central;
  }
  .jogador circle {
    stroke: #fff;
    stroke-width: 0.05;
  }
  .jogador.bloqueado circle {
    stroke: #ff1f1f;
    stroke-width: 0.12;
  }
</style>
