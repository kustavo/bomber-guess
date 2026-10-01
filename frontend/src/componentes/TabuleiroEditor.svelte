<script lang="ts">
  // Tabuleiro clicável do editor de mapas (spec 07, CA-08 e CA-09; plano, D13 e D14).
  // Desenhado com os sprites do marco 14; cada casa é um botão transparente por cima
  // dos itens (plano 14, D9).
  import { itemEm, type Item, type MapaEmEdicao } from '../lib/editor';
  import { corJogador } from '../lib/textos';
  import type { Posicao } from '../lib/tipos';
  import Sprite from './Sprite.svelte';

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
  {#each casas as c (`piso ${c.x},${c.y}`)}
    <Sprite nome="piso" x={c.x} y={c.y} />
  {/each}
  {#each mapa.blocos_fixos as b (`${b.x},${b.y}`)}
    <g data-tipo="bloco-fixo" data-x={b.x} data-y={b.y}>
      <Sprite nome="bloco-fixo" x={b.x} y={b.y} />
    </g>
  {/each}
  {#each mapa.blocos_destrutiveis as b (`${b.x},${b.y}`)}
    <g data-tipo="bloco-destrutivel" data-x={b.x} data-y={b.y}>
      <Sprite nome="bloco-destrutivel" x={b.x} y={b.y} />
    </g>
  {/each}
  {#each mapa.posicoes_iniciais as pi, i (`${pi.posicao.x},${pi.posicao.y}`)}
    <g data-tipo="posicao-inicial" data-x={pi.posicao.x} data-y={pi.posicao.y} data-numero={i + 1}>
      <Sprite nome="jogador" x={pi.posicao.x} y={pi.posicao.y} cor={corJogador(i)} />
      <text class="numero" x={pi.posicao.x + 0.86} y={pi.posicao.y + 0.2}>{i + 1}</text>
    </g>
  {/each}
  <!-- D9: casas clicáveis por cima dos itens, transparentes -->
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
</svg>

<style>
  .tabuleiro-editor {
    width: 100%;
    max-width: 640px;
    display: block;
    user-select: none;
    image-rendering: pixelated;
  }
  .casa {
    fill: transparent;
    stroke: rgb(0 0 0 / 0.12);
    stroke-width: 0.03;
    cursor: pointer;
  }
  .casa:hover,
  .casa:focus-visible {
    fill: rgb(255 255 255 / 0.18);
    stroke: #fff;
    stroke-width: 0.06;
    outline: none;
  }
  .numero {
    fill: #fff;
    font-family: system-ui, sans-serif;
    font-weight: 700;
    font-size: 0.3px;
    text-anchor: middle;
    dominant-baseline: central;
    paint-order: stroke;
    stroke: #1b1b22;
    stroke-width: 0.08px;
    pointer-events: none;
  }
</style>
