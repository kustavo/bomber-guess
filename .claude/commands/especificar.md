---
description: Escreve a spec (o quê e por quê) de um marco do roadmap
argument-hint: <número do marco>
---

Fase **Especificar** do marco $ARGUMENTS, conforme `specs/README.md`.

1. Leia `specs/README.md` e a linha do marco $ARGUMENTS em `docs/ROADMAP.md`. Leia **apenas** os documentos que o marco lista.
2. Crie `specs/NN-<nome-curto>/` (NN com dois dígitos) copiando `specs/_modelo/spec.md`. Se a pasta já existir, revise o que está lá em vez de sobrescrever.
3. Preencha o objetivo, o escopo (dentro e fora), as regras cobertas (IDs de `docs/`) e os critérios de aceitação `CA-NN` no formato Dado/Quando/Então. Cada critério cita os IDs de regra que exercita e precisa ser verificável por teste automatizado.
4. Não escreva plano técnico nem código. Ambiguidades e lacunas da spec de domínio vão em "Questões em aberto", com uma proposta de resposta.
5. Termine mostrando um resumo dos critérios e as questões em aberto, e peça aprovação. Com aprovação, mude o status para `aprovada`.
