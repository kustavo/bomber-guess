<svelte:options namespace="svg" />

<script lang="ts">
  // Um sprite numa casa do tabuleiro (spec 14; plano, D1). `cor` vira currentColor.
  import { idSprite, type NomeSprite } from '../lib/sprites';

  interface Props {
    nome: NomeSprite;
    x: number;
    y: number;
    cor?: string;
    classe?: string;
    centrado?: boolean; // transform-origin no centro da casa (spec 14, D11)
  }

  let { nome, x, y, cor, classe, centrado = false }: Props = $props();
</script>

<use
  href="#{idSprite(nome)}"
  data-sprite={nome}
  {x}
  {y}
  width="1"
  height="1"
  class={classe}
  style:color={cor}
  style:transform-origin={centrado ? `${x + 0.5}px ${y + 0.5}px` : undefined}
/>
