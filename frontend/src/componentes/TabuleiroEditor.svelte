<script lang="ts">
  // Tabuleiro clicável do editor de mapas (spec 07, CA-08 e CA-09; plano, D13 e D14).
  // Mesmas cores do Tabuleiro; cada casa é um botão que aplica a ferramenta.
  import { itemEm, type Item, type MapaEmEdicao } from '../lib/editor';
  import { corJogador } from '../lib/textos';
  import type { Posicao } from '../lib/tipos';

  interface Props {
    mapa: MapaEmEdicao;
    onCasa: (p: Posicao) => void;
  }

  let { mapa, onCasa }: Props = $props();

  const casas = $derived(
    Array.from({ length: mapa.config.altura }, (_, y) =>
      Array.from({ length: mapa.config.largura }, (_, x) => ({ x, y })),
    ).flat(),
  );

  const nomes: Record<Item, string> = {
    'bloco-fixo': 'bloco fixo',
    'bloco-destrutivel': 'bloco destrutível',
    'posicao-inicial': 'posição inicial',
  };

  function rotulo(p: Posicao): string {
    const item = itemEm(mapa, p);
    return `Casa ${p.x},${p.y}: ${item ? nomes[item] : 'vazia'}`;
  }

  function tecla(e: KeyboardEvent, p: Posicao) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onCasa(p);
    }
  }
</script>

<svg
  class="tabuleiro-editor"
  viewBox="0 0 {mapa.config.largura} {mapa.config.altura}"
  aria-label="Tabuleiro do editor, {mapa.config.largura} × {mapa.config.altura}"
  role="group"
>
  {#each casas as c (`${c.x},${c.y}`)}
    <rect
      class="casa"
      data-casa="{c.x},{c.y}"
      x={c.x}
      y={c.y}
      width="1"
      height="1"
      role="button"
      tabindex="0"
      aria-label={rotulo(c)}
      onclick={() => onCasa(c)}
      onkeydown={(e) => tecla(e, c)}
    />
  {/each}
  {#each mapa.blocos_fixos as b (`${b.x},${b.y}`)}
    <rect class="bloco-fixo" data-tipo="bloco-fixo" data-x={b.x} data-y={b.y} x={b.x} y={b.y} width="1" height="1" />
  {/each}
  {#each mapa.blocos_destrutiveis as b (`${b.x},${b.y}`)}
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
  {#each mapa.posicoes_iniciais as pi, i (`${pi.posicao.x},${pi.posicao.y}`)}
    <g
      class="posicao-inicial"
      data-tipo="posicao-inicial"
      data-x={pi.posicao.x}
      data-y={pi.posicao.y}
      data-numero={i + 1}
    >
      <circle cx={pi.posicao.x + 0.5} cy={pi.posicao.y + 0.5} r="0.38" fill={corJogador(i)} />
      <text x={pi.posicao.x + 0.5} y={pi.posicao.y + 0.5}>{i + 1}</text>
    </g>
  {/each}
</svg>

<style>
  .tabuleiro-editor {
    width: 100%;
    max-width: 640px;
    display: block;
    user-select: none;
  }
  .casa {
    fill: #2b3a2e;
    stroke: #3d5141;
    stroke-width: 0.03;
    cursor: pointer;
  }
  .casa:hover,
  .casa:focus-visible {
    fill: #3f5543;
    outline: none;
  }
  .bloco-fixo,
  .bloco-destrutivel,
  .posicao-inicial {
    pointer-events: none;
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
  .posicao-inicial circle {
    stroke: #fff;
    stroke-width: 0.05;
  }
  .posicao-inicial text {
    fill: #fff;
    font-size: 0.4px;
    font-family: system-ui, sans-serif;
    text-anchor: middle;
    dominant-baseline: central;
  }
</style>
