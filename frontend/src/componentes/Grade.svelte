<svelte:options namespace="svg" />

<script lang="ts">
  // Linhas claras entre as casas, desenhadas de uma vez e sem receber cliques
  // (spec 14, CA-16; plano, D12).
  interface Props {
    largura: number;
    altura: number;
  }

  let { largura, altura }: Props = $props();

  const caminho = $derived(
    [
      ...Array.from({ length: largura - 1 }, (_, i) => `M${i + 1} 0V${altura}`),
      ...Array.from({ length: altura - 1 }, (_, i) => `M0 ${i + 1}H${largura}`),
    ].join(''),
  );
</script>

<path class="grade" data-grade d={caminho} pointer-events="none" />

<style>
  .grade {
    fill: none;
    stroke: rgb(255 255 255 / 0.35);
    stroke-width: 0.03;
  }
</style>
