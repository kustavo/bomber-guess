# Marco 01: tipos em Go e mapa de exemplo

**Status**: concluída
**Roadmap**: `docs/ROADMAP.md`, marco 1
**Documentos de referência**: `docs/REGRAS.md` (seções 1 a 3), `docs/BOTS.md` (interface), `docs/EDITOR.md` (formato do mapa)

## Objetivo

Criar o vocabulário do jogo em código: os tipos do pacote `internal/jogo` que representam o `Estado`, o `Plano`, as ações e a interface `Bot`, com JSON idêntico ao de `docs/REGRAS.md`. Também definir o formato de mapa e um mapa de exemplo que vira o estado inicial de uma partida. Todos os marcos seguintes, e os bots das outras IAs, dependem desses tipos.

## Escopo

**Dentro**
- Módulo Go em `backend/` (`go.mod`) e pacote `backend/internal/jogo`.
- Tipos: `Posicao`, `Config`, `Bomba`, `Jogador`, `Estado`, `Acao`, `Plano`, `Mapa`.
- Enums como constantes string: `TipoAcao` (`MOVER`, `PLANTAR`, `ESPERAR`), `Direcao` (`CIMA`, `BAIXO`, `ESQUERDA`, `DIREITA`), `Status` (`VIVO`, `MORTO`).
- Interface `Bot` (`Versao`, `Planejar`).
- Utilitário de coordenadas: a vizinha de uma posição em uma direção.
- Formato JSON de mapa, `mapas/exemplo.json` e a conversão de mapa em `Estado` inicial.

**Fora** (fica para outro marco)
- Validação de planos (`Validar`) e resolução do turno (`ResolverTurno`): marco 2.
- Leitura de arquivo do disco: o pacote `jogo` é puro e recebe `[]byte`. Quem lê o arquivo é o `cmd/` (marco 4).
- Bots: marco 3.

## Regras cobertas

- `TAB-01` a `TAB-04`: grade, formato de posição, origem e eixos, vizinhança ortogonal.
- `EST-01` a `EST-08`: campos do `Estado`, inclusive `morte`.
- `MAP-01` a `MAP-06`: formato do mapa e estado inicial.
- `ACA-01`: tipos de ação; `MOVER` exige `direcao`.
- `BOT-01`: o bot recebe uma cópia do estado (`Estado.Copiar()`).

## Critérios de aceitação

**Serialização**
- **CA-01** (EST-01 a EST-08, TAB-02): **Dado** o JSON de exemplo do `Estado` em `docs/REGRAS.md` seção 2, **quando** ele é decodificado e codificado de novo, **então** o resultado é semanticamente igual ao original (mesmas chaves e valores; ordem e espaços não importam).
- **CA-02** (ACA-01): **Dado** o JSON de exemplo do plano em `docs/REGRAS.md` seção 3, **quando** ele faz a ida e volta, **então** o resultado é semanticamente igual ao original.
- **CA-03** (ACA-01, EST-07): **Dada** uma ação `PLANTAR` ou `ESPERAR`, ou um jogador vivo, **quando** codificados, **então** as chaves `direcao` e `morte`, respectivamente, não aparecem.
- **CA-04**: **Dadas** as constantes dos enums, **então** os valores são exatamente `"MOVER"`, `"PLANTAR"`, `"ESPERAR"`, `"CIMA"`, `"BAIXO"`, `"ESQUERDA"`, `"DIREITA"`, `"VIVO"` e `"MORTO"`.

**Coordenadas**
- **CA-05** (TAB-03, TAB-04): **Dada** a posição `{x: 5, y: 5}`, **quando** se pede a vizinha em cada direção, **então** `CIMA` → `{5, 4}`, `BAIXO` → `{5, 6}`, `ESQUERDA` → `{4, 5}` e `DIREITA` → `{6, 5}`.
- **CA-06** (TAB-01): **Dado** um tabuleiro 15×13, **então** `{0,0}` e `{14,12}` estão dentro dele, e `{-1,0}`, `{15,0}` e `{0,13}` estão fora.

**Interface `Bot`**
- **CA-07** (BOT-01): **Dado** um bot de teste que implementa `Versao() string` e `Planejar(estado Estado, jogadorID string) []Acao`, **então** ele satisfaz `jogo.Bot` em tempo de compilação, e alterar o `Estado` recebido dentro de `Planejar` não altera o estado de quem chamou (o teste usa `Estado.Copiar()`).

**Mapa e estado inicial**
- **CA-08** (MAP-01 a MAP-04): **Dado** `mapas/exemplo.json`, **quando** ele é convertido em `Estado` inicial com N versões de bot (N = número de posições iniciais), **então**:
  - `turno` = 1;
  - `config` e as dimensões vêm do mapa;
  - os jogadores são `jogador_1` … `jogador_N`, na ordem das posições iniciais, cada um em sua posição, com status `VIVO`, os atributos padrão do mapa e a `bot_versao` correspondente;
  - não há bombas;
  - os blocos são os do mapa.
- **CA-09** (EST-03): **Dado** o estado inicial, **então** `etapas_neste_turno` é igual ao maior `acoes_por_turno` entre os jogadores vivos.
- **CA-10** (MAP-05, MAP-06): **Dado** um mapa inválido, **quando** ele é convertido, **então** a conversão devolve um erro descritivo e nenhum estado. São inválidos:
  - JSON malformado;
  - largura ou altura ≤ 0;
  - bloco ou posição inicial fora do tabuleiro;
  - posição inicial sobre bloco;
  - mesma casa com bloco fixo e destrutível;
  - menos de 2 posições iniciais;
  - quantidade de versões de bot diferente da quantidade de posições iniciais.
- **CA-11**: **Dado** `mapas/exemplo.json`, **então** ele é um mapa válido de 15×13 com 4 posições iniciais nos cantos, blocos fixos em todas as casas com `x` e `y` ímpares, e cada canto livre em "L" (o canto e suas duas vizinhas sem bloco).
- **CA-13**: **Dado** um `Estado` com bombas e jogadores, **quando** se chama `Copiar()` e a cópia é alterada (slices e campos), **então** o original não muda.

**Pureza**
- **CA-12**: **Dado** o pacote `internal/jogo`, **então** ele importa apenas a biblioteca padrão (`go list -deps` sem pacotes do projeto), e a conversão de mapa não altera o `Mapa` recebido.

## Decisões

Questões levantadas no rascunho e aprovadas pelo usuário em 2026-09-30:

- **Q1** Formato do mapa → documentado em `docs/EDITOR.md` (`MAP-01` a `MAP-06`).
- **Q2** Bots atribuídos por `EstadoInicial(mapa Mapa, versoes []string) (Estado, error)` (MAP-03).
- **Q3** Campo `morte` no `Jogador` (`EST-07`, `EST-08`; exemplo de `docs/REGRAS.md` atualizado).
- **Q4** `Estado.Copiar()` com cópia profunda (CA-13; `BOT-01` atualizado).
- **Q5** Campo `etapa` mantido em cada ação; a consistência fica para o marco 2.
- **Q6** Mapa de exemplo no layout clássico (CA-11).
