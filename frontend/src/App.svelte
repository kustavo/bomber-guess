<script lang="ts">
  // Escolhe a tela pela rota em hash (spec 06, decisão 1; CA-13; spec 07, CA-17).
  import { relogioReal, type Relogio } from './lib/acompanhamento';
  import { criarCliente, type ClienteApi } from './lib/api';
  import { lerRota } from './lib/rota';
  import './estilos/sprites.css';
  import SpritesSvg from './componentes/SpritesSvg.svelte';
  import ListaPartidas from './telas/ListaPartidas.svelte';
  import TelaCriar from './telas/TelaCriar.svelte';
  import TelaPartida from './telas/TelaPartida.svelte';

  interface Props {
    cliente?: ClienteApi;
    relogio?: Relogio;
  }

  let { cliente = criarCliente(), relogio = relogioReal }: Props = $props();

  let hash = $state(window.location.hash);
  const rota = $derived(lerRota(hash));

  $effect(() => {
    const mudou = () => (hash = window.location.hash);
    window.addEventListener('hashchange', mudou);
    return () => window.removeEventListener('hashchange', mudou);
  });
</script>

<SpritesSvg />

<main>
  {#if rota.tela === 'partida'}
    {#key rota.nome}
      <TelaPartida nome={rota.nome} {cliente} {relogio} />
    {/key}
  {:else if rota.tela === 'criar'}
    <TelaCriar {cliente} />
  {:else}
    <ListaPartidas {cliente} {relogio} />
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: system-ui, sans-serif;
    color-scheme: light dark;
    background: Canvas;
    color: CanvasText;
  }
  main {
    padding: 1rem;
    max-width: 72rem;
    margin: 0 auto;
  }
</style>
