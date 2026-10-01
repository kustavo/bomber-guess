<script lang="ts">
  // Lista de partidas (API-10; spec 06, CA-13), consultada a cada 2 s (decisão 2).
  import { relogioReal, type Relogio } from '../lib/acompanhamento';
  import type { ClienteApi } from '../lib/api';
  import { CONSULTA_LISTA_MS } from '../lib/ritmo';
  import { enderecoPartida } from '../lib/rota';
  import { textoDesfecho, textoFase } from '../lib/textos';
  import type { ResumoPartida } from '../lib/tipos';

  interface Props {
    cliente: ClienteApi;
    relogio?: Relogio;
  }

  let { cliente, relogio = relogioReal }: Props = $props();

  let partidas: ResumoPartida[] | undefined = $state();
  let semConexao = $state(false);

  $effect(() => {
    let parado = false;
    let cancelar = () => {};
    const consultar = async () => {
      try {
        const r = await cliente.buscarPartidas();
        if (parado) return;
        partidas = r.partidas;
        semConexao = false;
      } catch {
        if (parado) return;
        semConexao = true;
      }
      cancelar = relogio.esperar(CONSULTA_LISTA_MS, consultar);
    };
    void consultar();
    return () => {
      parado = true;
      cancelar();
    };
  });
</script>

<section class="lista">
  <h1>Partidas</h1>
  {#if semConexao}
    <p class="aviso" role="status" data-aviso="sem-conexao">Sem conexão com o servidor. Tentando de novo…</p>
  {/if}
  {#if partidas === undefined}
    {#if !semConexao}<p>Carregando…</p>{/if}
  {:else if partidas.length === 0}
    <p>Nenhuma partida. Crie uma com <code>POST /partidas</code>.</p>
  {:else}
    <table>
      <thead>
        <tr><th>Partida</th><th>Fase</th><th>Turno</th><th>Desfecho</th></tr>
      </thead>
      <tbody>
        {#each partidas as p (p.nome)}
          <tr data-partida={p.nome}>
            <td><a href={enderecoPartida(p.nome)}>{p.nome}</a></td>
            <td>{textoFase(p.fase)}</td>
            <td>{p.turno}</td>
            <td>{p.fase === 'ENCERRADA' ? textoDesfecho(p.desfecho) : ''}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</section>

<style>
  table {
    border-collapse: collapse;
    width: 100%;
    max-width: 48rem;
  }
  th,
  td {
    text-align: left;
    padding: 0.4rem 0.75rem;
    border-bottom: 1px solid #8884;
  }
  .aviso {
    padding: 0.5rem 0.75rem;
    border-radius: 0.4rem;
    background: #fff4d6;
    color: #6b4a00;
  }
</style>
