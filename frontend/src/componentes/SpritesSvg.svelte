<script lang="ts">
  // Definições dos sprites, desenhadas uma vez para a página toda (spec 14; plano, D1 e D10).
  // Os tabuleiros e a paleta só usam <use href="#sprite-…">.
  import { LADO_SPRITE, NOMES_SPRITES, SPRITES, idSprite, retangulos } from '../lib/sprites';

  const simbolos = NOMES_SPRITES.map((nome) => ({ id: idSprite(nome), retangulos: retangulos(SPRITES[nome]) }));
</script>

<svg class="sprites" aria-hidden="true" width="0" height="0">
  <defs>
    {#each simbolos as s (s.id)}
      <symbol id={s.id} viewBox="0 0 {LADO_SPRITE} {LADO_SPRITE}" shape-rendering="crispEdges">
        {#each s.retangulos as r (`${r.x},${r.y}`)}
          <rect x={r.x} y={r.y} width={r.largura} height="1" fill={r.cor} />
        {/each}
      </symbol>
    {/each}
  </defs>
</svg>

<style>
  .sprites {
    position: absolute;
    width: 0;
    height: 0;
    overflow: hidden;
  }
</style>
