# Specs: desenvolvimento guiado por especificação

A spec é a fonte da verdade; o código é derivado dela. Nenhum código entra sem uma spec aprovada que o justifique.

## Camadas

| Camada | Onde | Muda quando |
|---|---|---|
| Constituição | `AGENTS.md` (o `CLAUDE.md` só o importa) | Raramente. Convenções que valem para todo o projeto. |
| Spec de domínio | `docs/*.md` | Uma regra do jogo ou da arquitetura muda. Regras têm IDs estáveis (`BOM-05`, `VAL-02`...). |
| Spec de entrega | `specs/NN-nome/` | Um marco do `docs/ROADMAP.md` é executado. |

## Fluxo de um marco

Cada fase termina com a aprovação do usuário antes da próxima. Uma conversa por marco.

1. **Especificar** (`/especificar N`): copia `_modelo/` para `specs/NN-nome/` e escreve o `spec.md`: o quê e por quê, critérios de aceitação e IDs de regra cobertos. Sem falar de implementação.
2. **Planejar** (`/planejar N`): escreve `plano.md` (tipos, assinaturas, arquivos, decisões) e `tarefas.md` (checklist em que cada tarefa aponta para critérios).
3. **Implementar** (`/implementar N`): executa as tarefas em ordem, uma de cada vez, com testes, e marca o checklist.
4. **Verificar**: todo critério de aceitação tem teste passando, e `gofmt`, `go vet` e `go test ./...` estão limpos. O status do `spec.md` passa a `concluída`.

## Rastreabilidade

- Critério de aceitação: `CA-NN`, local ao marco (ex.: `02/CA-07`).
- Cada critério cita as regras que exercita: `CA-07 (BOM-05)`.
- O nome de cada caso de teste começa pelo ID da regra: `{nome: "BOM-05 pavio 3 na etapa 2 explode na etapa 5"}`.
- `grep -rn "BOM-05"` mostra a regra, a spec e os testes que a cobrem.

## Quando a realidade diverge da spec

Parar, atualizar a spec (e `docs/` se for uma regra do jogo), pedir aprovação e só então mudar o código. Decisões novas entram em `docs/REGRAS.md` (seção 8, `DEC-NN`) ou na seção "Decisões" do `plano.md` do marco.

## Status de uma spec

`rascunho` → `aprovada` → `em implementação` → `concluída`
