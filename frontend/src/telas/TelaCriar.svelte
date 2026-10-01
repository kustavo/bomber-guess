<script lang="ts">
  // Tela Criar partida: editor de mapas e escolha dos bots (spec 07, CA-07 a CA-16;
  // plano, D9 a D12).
  import type { ClienteApi } from '../lib/api';
  import { iniciarPartida } from '../lib/criacao';
  import {
    LADO_MAXIMO,
    aplicarFerramenta,
    botsDasPosicoes,
    escolherBot,
    novoMapaEmEdicao,
    paraMapa,
    pendencias,
    preencherBots,
    redimensionar,
    sugerirNomeMapa,
    type Ferramenta,
    type MapaEmEdicao,
  } from '../lib/editor';
  import { enderecoPartida } from '../lib/rota';
  import { corJogador } from '../lib/textos';
  import type { Posicao } from '../lib/tipos';
  import type { NomeSprite } from '../lib/sprites';
  import Sprite from '../componentes/Sprite.svelte';
  import TabuleiroEditor from '../componentes/TabuleiroEditor.svelte';

  interface Props {
    cliente: ClienteApi;
  }

  let { cliente }: Props = $props();

  let mapa: MapaEmEdicao = $state(novoMapaEmEdicao());
  let ferramenta: Ferramenta = $state('bloco-fixo');
  let nomePartida = $state('');
  let nomeMapaDigitado: string | undefined = $state(); // undefined: acompanha a sugestão (D11)
  let semente = $state(1);
  let catalogo: string[] = $state([]);
  let erroBots = $state(false);
  let erro: string | undefined = $state();
  let mapaSalvo: string | undefined = $state();
  let enviando = $state(false);

  const nomeMapa = $derived(nomeMapaDigitado ?? sugerirNomeMapa(nomePartida));
  const faltas = $derived(pendencias(mapa, nomePartida, nomeMapa, catalogo));

  $effect(() => {
    let parado = false;
    cliente
      .buscarBots()
      .then((r) => {
        if (parado) return;
        catalogo = r.bots;
        if (r.bots.length > 0) mapa = preencherBots(mapa, r.bots[0]);
      })
      .catch(() => {
        if (!parado) erroBots = true;
      });
    return () => {
      parado = true;
    };
  });

  // Cada ferramenta com o sprite do seu ícone (spec 14, CA-11).
  const ferramentas: { f: Ferramenta; rotulo: string; sprite: NomeSprite }[] = [
    { f: 'bloco-fixo', rotulo: 'Bloco fixo', sprite: 'bloco-fixo' },
    { f: 'bloco-destrutivel', rotulo: 'Bloco destrutível', sprite: 'bloco-destrutivel' },
    { f: 'posicao-inicial', rotulo: 'Posição inicial', sprite: 'jogador' },
    { f: 'borracha', rotulo: 'Borracha', sprite: 'piso' },
  ];

  // Campos numéricos além de largura e altura; vazio vira 0 e o servidor recusa (D9).
  const campos: { rotulo: string; ler: (m: MapaEmEdicao) => number; gravar: (m: MapaEmEdicao, v: number) => void }[] = [
    { rotulo: 'Limite de turnos', ler: (m) => m.config.limite_turnos, gravar: (m, v) => (m.config.limite_turnos = v) },
    { rotulo: 'Turno do fechamento', ler: (m) => m.config.turno_fechamento ?? 0, gravar: (m, v) => (m.config.turno_fechamento = v) },
    { rotulo: 'Área mínima: largura', ler: (m) => m.config.area_minima?.largura ?? 0, gravar: (m, v) => (m.config.area_minima = { largura: v, altura: m.config.area_minima?.altura ?? 0 }) },
    { rotulo: 'Área mínima: altura', ler: (m) => m.config.area_minima?.altura ?? 0, gravar: (m, v) => (m.config.area_minima = { largura: m.config.area_minima?.largura ?? 0, altura: v }) },
    { rotulo: 'Prazo de planejamento (ms)', ler: (m) => m.config.prazo_planejamento_ms, gravar: (m, v) => (m.config.prazo_planejamento_ms = v) },
    { rotulo: 'Duração da etapa (ms)', ler: (m) => m.config.duracao_etapa_ms, gravar: (m, v) => (m.config.duracao_etapa_ms = v) },
    { rotulo: 'Bombas por turno', ler: (m) => m.jogador_padrao.bombas_por_turno, gravar: (m, v) => (m.jogador_padrao.bombas_por_turno = v) },
    { rotulo: 'Potência', ler: (m) => m.jogador_padrao.potencia, gravar: (m, v) => (m.jogador_padrao.potencia = v) },
    { rotulo: 'Pavio padrão', ler: (m) => m.jogador_padrao.pavio_padrao, gravar: (m, v) => (m.jogador_padrao.pavio_padrao = v) },
    { rotulo: 'Ações por turno', ler: (m) => m.jogador_padrao.acoes_por_turno, gravar: (m, v) => (m.jogador_padrao.acoes_por_turno = v) },
  ];

  const numero = (e: Event) => {
    const v = (e.currentTarget as HTMLInputElement).valueAsNumber;
    return Number.isNaN(v) ? 0 : v;
  };

  function clicarCasa(p: Posicao) {
    mapa = aplicarFerramenta(mapa, ferramenta, p, catalogo[0] ?? '');
  }

  function mudarLado(e: Event, lado: 'largura' | 'altura') {
    const v = (e.currentTarget as HTMLInputElement).valueAsNumber;
    mapa =
      lado === 'largura'
        ? redimensionar(mapa, v, mapa.config.altura)
        : redimensionar(mapa, mapa.config.largura, v);
    (e.currentTarget as HTMLInputElement).value = String(mapa.config[lado]);
  }

  async function iniciar() {
    if (faltas.length > 0 || enviando) return;
    enviando = true;
    erro = undefined;
    const atual = $state.snapshot(mapa);
    const r = await iniciarPartida(
      cliente,
      paraMapa(atual, nomeMapa),
      { nome: nomePartida, mapa: nomeMapa, bots: botsDasPosicoes(atual), semente },
      mapaSalvo,
    );
    enviando = false;
    if (r.ok) {
      window.location.hash = enderecoPartida(nomePartida); // D12
    } else {
      erro = r.erro;
      mapaSalvo = r.mapaSalvo;
    }
  }
</script>

<section class="criar">
  <p><a href="#/">← Partidas</a></p>
  <h1>Criar partida</h1>

  <div class="colunas">
    <div class="painel">
      <fieldset>
        <legend>Partida</legend>
        <label>Nome da partida <input type="text" bind:value={nomePartida} autocomplete="off" /></label>
        <label>
          Nome do mapa
          <input
            type="text"
            value={nomeMapa}
            autocomplete="off"
            oninput={(e) => (nomeMapaDigitado = e.currentTarget.value)}
          />
        </label>
        <label>
          Semente
          <input
            type="number"
            min="0"
            max={Number.MAX_SAFE_INTEGER}
            step="1"
            value={semente}
            oninput={(e) => (semente = numero(e))}
          />
        </label>
      </fieldset>

      <fieldset>
        <legend>Tabuleiro</legend>
        <label>
          Largura
          <input type="number" min="1" max={LADO_MAXIMO} value={mapa.config.largura} onchange={(e) => mudarLado(e, 'largura')} />
        </label>
        <label>
          Altura
          <input type="number" min="1" max={LADO_MAXIMO} value={mapa.config.altura} onchange={(e) => mudarLado(e, 'altura')} />
        </label>
        {#each campos as c (c.rotulo)}
          <label>
            {c.rotulo}
            <input type="number" value={c.ler(mapa)} oninput={(e) => c.gravar(mapa, numero(e))} />
          </label>
        {/each}
      </fieldset>
    </div>

    <div class="editor">
      <div class="paleta" role="toolbar" aria-label="Ferramentas">
        {#each ferramentas as t (t.f)}
          <button
            type="button"
            data-ferramenta={t.f}
            class:ativa={ferramenta === t.f}
            aria-pressed={ferramenta === t.f}
            onclick={() => (ferramenta = t.f)}
          >
            <svg class="amostra" viewBox="0 0 1 1" aria-hidden="true">
              <Sprite nome={t.sprite} x={0} y={0} cor={corJogador(0)} />
            </svg>{t.rotulo}
          </button>
        {/each}
      </div>

      <TabuleiroEditor {mapa} onCasa={clicarCasa} />

      <fieldset>
        <legend>Bots</legend>
        {#if erroBots}
          <p class="aviso" role="alert" data-aviso="bots">Não foi possível carregar o catálogo de bots.</p>
        {/if}
        {#if mapa.posicoes_iniciais.length === 0}
          <p>Coloque posições iniciais no tabuleiro para escolher os bots.</p>
        {/if}
        {#each mapa.posicoes_iniciais as pi, i (`${pi.posicao.x},${pi.posicao.y}`)}
          <label class="bot">
            <span class="numero" style="background: {corJogador(i)}">{i + 1}</span>
            Posição ({pi.posicao.x}, {pi.posicao.y})
            <select
              data-posicao={i}
              value={pi.bot}
              onchange={(e) => (mapa = escolherBot(mapa, i, e.currentTarget.value))}
            >
              {#each catalogo as v (v)}
                <option value={v}>{v}</option>
              {/each}
            </select>
          </label>
        {/each}
      </fieldset>

      {#if faltas.length > 0}
        <ul class="aviso" role="status" data-aviso="pendencias">
          {#each faltas as f (f)}<li>{f}</li>{/each}
        </ul>
      {/if}
      {#if erro}
        <p class="aviso erro" role="alert" data-aviso="erro">{erro}</p>
      {/if}
      <button type="button" class="iniciar" data-acao="iniciar" disabled={faltas.length > 0 || enviando} onclick={iniciar}>
        {enviando ? 'Iniciando…' : 'Iniciar'}
      </button>
    </div>
  </div>
</section>

<style>
  .colunas {
    display: flex;
    flex-wrap: wrap;
    gap: 1.5rem;
    align-items: flex-start;
  }
  .painel {
    flex: 0 1 18rem;
  }
  .editor {
    flex: 1 1 24rem;
    min-width: 0;
  }
  fieldset {
    border: 1px solid color-mix(in srgb, CanvasText 20%, transparent);
    border-radius: 0.4rem;
    margin: 0 0 1rem;
    display: grid;
    gap: 0.5rem;
  }
  label {
    display: grid;
    grid-template-columns: 1fr 7rem;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9rem;
  }
  label.bot {
    grid-template-columns: auto 1fr 10rem;
  }
  input,
  select {
    font: inherit;
    min-width: 0;
  }
  .paleta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    margin-bottom: 0.75rem;
  }
  .paleta button {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font: inherit;
    padding: 0.3rem 0.6rem;
    border-radius: 0.4rem;
    border: 1px solid color-mix(in srgb, CanvasText 30%, transparent);
    background: Canvas;
    color: CanvasText;
    cursor: pointer;
  }
  .paleta button.ativa {
    outline: 2px solid #3a7bd5;
  }
  .amostra {
    width: 1.1rem;
    height: 1.1rem;
    display: inline-block;
    image-rendering: pixelated;
  }
  .numero {
    display: inline-grid;
    place-items: center;
    width: 1.4rem;
    height: 1.4rem;
    border-radius: 50%;
    color: #fff;
    font-size: 0.8rem;
  }
  .aviso {
    padding: 0.5rem 0.75rem;
    border-radius: 0.4rem;
    background: #fff4d6;
    color: #6b4a00;
  }
  ul.aviso {
    padding-left: 1.75rem;
  }
  .aviso.erro {
    background: #ffe0e0;
    color: #8a1010;
  }
  .iniciar {
    font: inherit;
    font-size: 1.05rem;
    padding: 0.5rem 1.5rem;
    margin-top: 0.5rem;
    cursor: pointer;
  }
</style>
