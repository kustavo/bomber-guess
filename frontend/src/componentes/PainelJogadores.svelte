<script lang="ts">
  // Painel dos jogadores: id, bot, status, morte e a ação da etapa exibida (spec 06, CA-03, CA-04).
  import type { JogadorExibido } from '../lib/tabuleiro';
  import { corJogador, textoAcao, textoResultado } from '../lib/textos';

  let { jogadores }: { jogadores: JogadorExibido[] } = $props();
</script>

<ul class="painel">
  {#each jogadores as j (j.id)}
    <li data-jogador={j.id} class:morto={j.status === 'MORTO'}>
      <span class="cor" style="background: {corJogador(j.indice)}">{j.indice + 1}</span>
      <span class="id">{j.id}</span>
      <span class="bot">{j.bot_versao}</span>
      {#if j.status === 'MORTO'}
        <span class="status">Morto{j.morte ? ` no turno ${j.morte.turno}, etapa ${j.morte.etapa}` : ''}</span>
      {:else}
        <span class="status">Vivo</span>
      {/if}
      {#if j.acao && j.resultado}
        <span class="acao" class:bloqueada={j.bloqueado}>{textoAcao(j.acao)} · {textoResultado(j.resultado)}</span>
      {/if}
    </li>
  {/each}
</ul>

<style>
  .painel {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    gap: 0.4rem;
  }
  li {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: baseline;
  }
  .morto {
    opacity: 0.55;
  }
  .cor {
    display: inline-grid;
    place-items: center;
    width: 1.4rem;
    height: 1.4rem;
    border-radius: 50%;
    color: #fff;
    font-size: 0.8rem;
  }
  .id {
    font-weight: 600;
  }
  .bot,
  .acao {
    color: #888;
  }
  .acao.bloqueada {
    color: #e5484d;
    font-weight: 600;
  }
</style>
