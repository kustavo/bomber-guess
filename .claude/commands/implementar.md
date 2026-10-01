---
description: Implementa as tarefas de um marco com plano aprovado
argument-hint: <número do marco> [tarefa, ex.: T3]
---

Fase **Implementar** do marco $ARGUMENTS, conforme `specs/README.md`.

1. Abra `spec.md`, `plano.md` e `tarefas.md` do marco. Mude o status da spec para `em implementação`.
2. Execute a próxima tarefa não marcada (ou a indicada no argumento), uma por vez:
   - escreva primeiro os testes de tabela dos CAs da tarefa, com o nome de cada caso começando pelo ID da regra (ex.: `"BOM-05 ..."`);
   - implemente até passarem;
   - rode `gofmt -l`, `go vet ./...` e `go test ./...`;
   - marque a tarefa em `tarefas.md`.
3. Se a implementação revelar que a spec ou o plano estão errados ou incompletos, **pare**. Explique a divergência, proponha a correção na spec ou no plano e espere aprovação antes de seguir.
4. Na última tarefa, confirme que todo CA tem teste e mude o status da spec para `concluída`.
