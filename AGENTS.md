# Bomber Guess: contexto do projeto

Jogo inspirado em Bomberman, por turnos com **resolução simultânea**: no início do turno cada jogador define secretamente todas as ações do turno, e depois todas são executadas ao mesmo tempo, etapa por etapa. Todos os jogadores são **bots** (implementações em Go da interface `Bot`), cada um criado por uma IA diferente. O objetivo é colocar as IAs para competir e medir a melhor por meio de um ranking.

## Stack

- Backend: Go (somente biblioteca padrão no núcleo do jogo).
- Frontend: Svelte (só para assistir às partidas e editar mapas).
- Comunicação: API HTTP (sem WebSockets). Fila de ações atrás de uma interface: primeiro em memória, depois Kafka.

## Estrutura de pastas

```
bomberai/
├── AGENTS.md              este arquivo: instruções para qualquer agente de IA (vai em toda conversa)
├── CLAUDE.md              só importa o AGENTS.md (Claude Code não lê AGENTS.md sozinho)
├── docs/
│   ├── REGRAS.md          regras do jogo, formato do estado e das ações
│   ├── BOTS.md            interface Bot e contrato dos bots
│   ├── ARQUITETURA.md     validador, resolução do turno, registro, loop, API, fila
│   ├── EDITOR.md          editor de mapas e criação de partidas
│   ├── RANKING.md         pontuação e estatísticas
│   └── ROADMAP.md         marcos de implementação e o que cada um precisa
├── specs/                 specs de entrega, uma pasta por marco (ver specs/README.md)
│   ├── README.md          processo spec-driven
│   └── _modelo/           modelos de spec.md, plano.md e tarefas.md
├── .claude/commands/      atalhos do Claude Code para as fases: /especificar, /planejar, /implementar
├── mapas/                 mapas em JSON (ex.: exemplo.json)
├── backend/
│   ├── go.mod
│   ├── cmd/
│   │   ├── terminal/      partida completa no terminal, tabuleiro em ASCII
│   │   └── servidor/      servidor HTTP
│   └── internal/
│       ├── jogo/          tipos, validador e resolução do turno (puro: sem I/O, sem rede, sem banco)
│       ├── bots/          um subpacote por bot (ex.: bots/aleatorio, bots/claude)
│       ├── partida/       loop de turnos em tempo real
│       ├── fila/          interface da fila + implementações (memoria, kafka)
│       ├── api/           handlers HTTP
│       └── ranking/       pontuação e estatísticas
└── frontend/              app Svelte
```

## Convenções

- Nomes do domínio em português, seguindo a linguagem ubíqua do glossário de `docs/REGRAS.md` (código, docs e JSON usam os mesmos termos). Pode sugerir nomes melhores.
- Tags JSON em `snake_case`, exatamente como nos exemplos de `docs/REGRAS.md`.
- Enums como constantes string: `"MOVER"`, `"PLANTAR"`, `"ESPERAR"`, `"CIMA"`, `"BAIXO"`, `"ESQUERDA"`, `"DIREITA"`, `"VIVO"`, `"MORTO"`.
- O pacote `internal/jogo` não importa nenhum outro pacote do projeto. Tudo depende dele, ele não depende de nada.
- Resolução determinística: mesma entrada gera sempre a mesma saída. Aleatoriedade (quando existir) entra por uma semente explícita.
- Funções recebem e devolvem cópias do estado; nunca alteram o estado recebido.
- Testes de tabela (`[]struct{ nome string; ... }` + `t.Run`) para cada regra. Todo código novo vem com testes.
- `gofmt` e `go vet` sem avisos. `go test ./...` precisa passar antes de encerrar uma tarefa.
- Frontend (quando a tarefa mexer em `frontend/`): `npm run check` sem erros e `npm test` passando, ambos dentro de `frontend/`.
- Alterações pontuais: não reescrever arquivos inteiros sem necessidade.

## Como trabalhar (spec-driven)

- A spec é a fonte da verdade. Nenhum código sem spec aprovada; ao divergir, atualizar a spec primeiro. Processo completo em `specs/README.md`.
- Um marco do `docs/ROADMAP.md` por conversa, em três fases com aprovação entre elas: especificar → planejar → implementar. No Claude Code há os atalhos `/especificar N`, `/planejar N` e `/implementar N`. Em outras ferramentas, siga os passos descritos em `.claude/commands/*.md`, que valem para qualquer agente.
- Ler apenas os documentos que o marco lista como necessários.
- Regras têm IDs estáveis (`BOM-05`, `VAL-02`...). Critérios de aceitação e nomes de casos de teste citam esses IDs. Nunca renumerar.
