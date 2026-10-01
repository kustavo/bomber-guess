---
description: Escreve o plano técnico e as tarefas de um marco com spec aprovada
argument-hint: <número do marco>
---

Fase **Planejar** do marco $ARGUMENTS, conforme `specs/README.md`.

1. Abra `specs/NN-*/spec.md` do marco $ARGUMENTS. Se o status não for `aprovada`, pare e avise.
2. Leia os documentos de referência listados na spec e o código existente que o marco toca.
3. Escreva `plano.md` a partir de `specs/_modelo/plano.md`: arquivos, tipos e assinaturas públicas (sem corpo), decisões com motivo e riscos. Siga as convenções do `AGENTS.md` e a linguagem ubíqua do glossário de `docs/REGRAS.md`.
4. Escreva `tarefas.md` a partir de `specs/_modelo/tarefas.md`: tarefas pequenas e ordenadas. Cada uma cita os `CA-NN` que atende, e todo CA aparece em ao menos uma tarefa.
5. Se o plano exigir mudar uma regra de `docs/`, não mude sozinho: proponha a mudança e peça aprovação.
6. Termine com um resumo do plano e peça aprovação.
