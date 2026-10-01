# Bots (interface `Bot`)

Leia junto com `docs/REGRAS.md`, que define o formato do estado (`Estado`) e das ações. Para criar um bot não é preciso saber nada sobre API, fila ou frontend.

## Interface

Cada bot é código Go dentro do backend que implementa esta interface:

```go
type Bot interface {
    // Identificador único da versão, ex.: "claude-v1"
    Versao() string

    // Recebe uma cópia do estado e o id do próprio jogador.
    // Retorna o plano do turno: no máximo acoes_por_turno ações.
    Planejar(estado Estado, jogadorID string) []Acao
}
```

## Contrato

- **BOT-01** Cada bot recebe uma **cópia** do estado, nunca o original (`Estado.Copiar()`, cópia profunda), e não tem acesso às ações dos outros bots no mesmo turno.
- **BOT-02** Cada chamada roda com limite de tempo (`prazo_planejamento_ms`). Se o bot estourar o tempo ou entrar em `panic`, todas as suas ações do turno viram `ESPERAR`.
- **BOT-03** A saída bruta de cada bot é gravada antes de qualquer alteração, para análise estatística. Ações inválidas (infrações) contam contra o bot no ranking.
- **BOT-04** Ações inválidas: a primeira ação inválida e todas as seguintes viram `ESPERAR` (ver `Validar` em `docs/ARQUITETURA.md`).

## Onde colocar

- Um subpacote por bot: `backend/internal/bots/<nome>/`.
- O bot importa apenas `internal/jogo` e a biblioteca padrão.
- Registrar a versão no catálogo de bots (lista usada pelo `GET /bots` e pelo editor).
- Incluir testes do bot (ao menos: não gera ações inválidas em um mapa de exemplo e não entra em pânico com estado vazio ou jogador morto).

## Dicas para quem cria um bot

- O coração do jogo é **prever**: onde os adversários vão estar e onde vão plantar bombas.
- O movimento para um bloco destrutível ainda intacto na hora da execução é bloqueado e aborta o resto do plano. Planejar através de um bloco é uma aposta.
- O movimento é resolvido antes da explosão na mesma etapa: sair do alcance na etapa em que a bomba explode salva o jogador.
- Bombas empilhadas somam alcance; reações em cadeia acontecem na mesma etapa.

## Versão 2 (futuro)

`Planejar` receberá um terceiro parâmetro, `Historico` (os `Estado` dos turnos anteriores), para estudar padrões dos adversários.
