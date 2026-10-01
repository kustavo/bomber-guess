#!/usr/bin/env bash
# Sobe o servidor Go (API, porta 8080) e o frontend Svelte (porta 5173) juntos.
# Uso: ./iniciar.sh        Ctrl+C encerra os dois.
set -euo pipefail

raiz="$(cd "$(dirname "$0")" && pwd)"

for porta in 8080 5173; do
  if (exec 3<>"/dev/tcp/127.0.0.1/$porta") 2>/dev/null || (exec 3<>"/dev/tcp/::1/$porta") 2>/dev/null; then
    echo "A porta $porta já está em uso: provavelmente o servidor ou o frontend já está rodando."
    echo "Feche o que está aberto (Ctrl+C no terminal dele) e rode de novo."
    exit 1
  fi
done

if [ ! -d "$raiz/frontend/node_modules" ]; then
  echo "Instalando as dependências do frontend..."
  npm --prefix "$raiz/frontend" install
fi

pids=()
encerrar() {
  trap - INT TERM EXIT
  echo
  echo "Encerrando..."
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  wait 2>/dev/null || true
}
trap encerrar INT TERM EXIT

go -C "$raiz/backend" run ./cmd/servidor -mapas "$raiz/mapas" &
pids+=($!)

npm --prefix "$raiz/frontend" run dev -- --port 5173 --strictPort &
pids+=($!)

echo
echo "API:      http://localhost:8080"
echo "Frontend: http://localhost:5173   (Ctrl+C encerra os dois)"
echo

# Se um dos dois cair, encerra o outro também.
wait -n
