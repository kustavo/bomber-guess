<script lang="ts">
  // Tela de uma partida: cabeçalho, tabuleiro, painel, cronômetro ou desfecho, avisos
  // (spec 06, CA-01 a CA-15).
  import { acompanhar, relogioReal, type Relogio, type VisaoTela } from '../lib/acompanhamento';
  import type { ClienteApi } from '../lib/api';
  import { textoDesfecho, textoFase } from '../lib/textos';
  import Cronometro from '../componentes/Cronometro.svelte';
  import PainelJogadores from '../componentes/PainelJogadores.svelte';
  import Tabuleiro from '../componentes/Tabuleiro.svelte';

  interface Props {
    nome: string;
    cliente: ClienteApi;
    relogio?: Relogio;
  }

  let { nome, cliente, relogio = relogioReal }: Props = $props();

  let visao: VisaoTela = $state({ carregando: true, naoEncontrada: false, semConexao: false, defasagem: 0 });

  $effect(() => {
    const a = acompanhar(nome, cliente, relogio);
    const cancelar = a.visao.subscribe((v) => (visao = v));
    return () => {
      cancelar();
      a.parar();
    };
  });

  const r = $derived(visao.resposta);
  const t = $derived(visao.tabuleiro);
</script>

<section class="partida">
  <p><a href="#/">← Partidas</a></p>
  <h1>{nome}</h1>

  {#if visao.naoEncontrada}
    <p class="aviso" role="alert" data-aviso="nao-encontrada">Partida não encontrada.</p>
  {:else}
    {#if visao.semConexao}
      <p class="aviso" role="status" data-aviso="sem-conexao">Sem conexão com o servidor. Tentando de novo…</p>
    {/if}
    {#if visao.erro}
      <p class="aviso" role="status" data-aviso="erro">{visao.erro}</p>
    {/if}

    {#if visao.carregando}
      <p>Carregando…</p>
    {:else if r && t}
      <div class="cabecalho">
        <span data-campo="turno">Turno {t.turno}</span>
        <span data-campo="etapa">Etapa {t.etapa} de {r.estado.etapas_neste_turno}</span>
        <span data-campo="fase">{textoFase(r.fase)}</span>
        {#if r.fase === 'ENCERRADA'}
          <strong data-campo="desfecho">{textoDesfecho(r.desfecho)}</strong>
        {:else}
          <Cronometro fimDaFase={r.fim_da_fase} defasagem={visao.defasagem} {relogio} />
        {/if}
      </div>
      <div class="jogo">
        <Tabuleiro tabuleiro={t} duracaoEtapaMs={r.estado.config.duracao_etapa_ms} />
        <PainelJogadores jogadores={t.jogadores} />
      </div>
    {/if}
  {/if}
</section>

<style>
  .cabecalho {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-items: baseline;
    margin-bottom: 0.75rem;
  }
  .jogo {
    display: grid;
    grid-template-columns: minmax(0, 640px) minmax(14rem, 1fr);
    gap: 1.5rem;
    align-items: start;
  }
  @media (max-width: 720px) {
    .jogo {
      grid-template-columns: 1fr;
    }
  }
  .aviso {
    padding: 0.5rem 0.75rem;
    border-radius: 0.4rem;
    background: #fff4d6;
    color: #6b4a00;
  }
</style>
