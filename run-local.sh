#!/usr/bin/env bash

set -euo pipefail

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$project_dir"

if [[ ! -f .env ]]; then
  echo "Falta .env. Copiá .env.example y completá sus valores." >&2
  exit 1
fi

set -a
source .env
set +a

: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD es obligatorio en .env}"
: "${JWT_SECRET:?JWT_SECRET es obligatorio en .env}"

postgres_user="${POSTGRES_USER:-app}"
postgres_db="${POSTGRES_DB:-inge_soft_3}"
export DATABASE_URL="postgres://${postgres_user}:${POSTGRES_PASSWORD}@127.0.0.1:5432/${postgres_db}?sslmode=disable"
export HTTP_ADDR="${HTTP_ADDR:-127.0.0.1:8080}"

backend_pid=""
frontend_pid=""
backend_binary=""
log_dir="$project_dir/local-run-logs"
backend_log="$log_dir/$(date '+%Y%m%d-%H%M%S').log"

cleanup() {
  trap - EXIT INT TERM
  [[ -n "$backend_pid" ]] && kill "$backend_pid" 2>/dev/null || true
  [[ -n "$frontend_pid" ]] && kill "$frontend_pid" 2>/dev/null || true
  [[ -n "$backend_pid" ]] && wait "$backend_pid" 2>/dev/null || true
  [[ -n "$frontend_pid" ]] && wait "$frontend_pid" 2>/dev/null || true
  [[ -n "$backend_binary" ]] && rm -f "$backend_binary"
  docker compose stop db >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

docker compose stop backend frontend >/dev/null 2>&1 || true
docker compose up -d db

if [[ ! -d frontend/node_modules ]]; then
  (cd frontend && npm ci)
fi

backend_binary="$(mktemp "${TMPDIR:-/tmp}/fitpro-api.XXXXXX")"
(cd backend && go build -o "$backend_binary" ./cmd/api)

mkdir -p "$log_dir"
"$backend_binary" > >(tee "$backend_log") 2>&1 &
backend_pid=$!

(cd frontend && exec npm run dev) &
frontend_pid=$!

echo "Backend: http://127.0.0.1:8080"
echo "Frontend: revisá la URL informada por Vite"
echo "Logs backend: $backend_log"
echo "Presioná Ctrl+C para detener todo."

wait -n "$backend_pid" "$frontend_pid"
