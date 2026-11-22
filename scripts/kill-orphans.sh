#!/usr/bin/env bash
# Kill leftover dev processes: the watcher, the core binary it started, the Vite dev server and anything on the dev ports.
# Usage: scripts/kill-orphans.sh            (default ports: 8080 3000)
#        scripts/kill-orphans.sh 3000 5000  (custom ports)

set -u

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
GRACE_SECONDS=5

# Watchers first so they cannot respawn what we kill next.
PATTERNS=(
  "watchexec .*go build -tags dev"
  "$ROOT/core/.bin/main.bin"
  "$ROOT/frontend/node_modules/.*vite"
)

PORTS=("$@")
if [ ${#PORTS[@]} -eq 0 ]; then
  PORTS=(8080 3000)
fi

port_pids() {
  if command -v lsof >/dev/null 2>&1; then
    lsof -t -i :"$1" -sTCP:LISTEN 2>/dev/null
  else
    ss -ltnp "sport = :$1" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | sort -u
  fi
}

terminate() {
  local pids=("$@")
  [ ${#pids[@]} -eq 0 ] && return

  for pid in "${pids[@]}"; do
    echo "killing PID $pid ($(ps -p "$pid" -o args= 2>/dev/null | cut -c1-100))"
    kill "$pid" 2>/dev/null
  done

  for _ in $(seq "$GRACE_SECONDS"); do
    local alive=0
    for pid in "${pids[@]}"; do
      kill -0 "$pid" 2>/dev/null && alive=1
    done
    [ "$alive" -eq 0 ] && return
    sleep 1
  done

  for pid in "${pids[@]}"; do
    if kill -0 "$pid" 2>/dev/null; then
      echo "  PID $pid still running, sending SIGKILL"
      kill -9 "$pid" 2>/dev/null
    fi
  done
}

for pattern in "${PATTERNS[@]}"; do
  mapfile -t pids < <(pgrep -f "$pattern" | grep -vxE "$$|$PPID")
  terminate "${pids[@]}"
done

for port in "${PORTS[@]}"; do
  mapfile -t pids < <(port_pids "$port")
  if [ ${#pids[@]} -eq 0 ]; then
    echo "Port $port: nothing is listening."
    continue
  fi
  terminate "${pids[@]}"
done
