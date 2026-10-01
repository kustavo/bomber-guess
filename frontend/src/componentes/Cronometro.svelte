<script lang="ts">
  // Tempo até o fim da fase no relógio do servidor (API-02; spec 06, CA-05 a CA-07).
  import { relogioReal, type Relogio } from '../lib/acompanhamento';
  import { tempoRestante } from '../lib/cronometro';
  import { formatarTempo } from '../lib/textos';

  interface Props {
    fimDaFase?: string;
    defasagem: number;
    relogio?: Relogio;
  }

  let { fimDaFase, defasagem, relogio = relogioReal }: Props = $props();

  let agora = $state(0);

  $effect(() => {
    let cancelar = () => {};
    const tique = () => {
      agora = relogio.agora();
      cancelar = relogio.esperar(100, tique);
    };
    tique();
    return () => cancelar();
  });

  // Antes do primeiro tique (agora = 0) não há o que mostrar.
  const restante = $derived(agora === 0 ? null : tempoRestante(fimDaFase, defasagem, agora));
</script>

{#if restante !== null}
  <span class="cronometro" data-restante={restante}>{formatarTempo(restante)}</span>
{/if}

<style>
  .cronometro {
    font-variant-numeric: tabular-nums;
    font-weight: 600;
  }
</style>
