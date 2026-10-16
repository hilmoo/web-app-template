#!/usr/bin/env bash
# Kill any process listening on the given ports.
# Usage: .vscode/kill.sh            (defaults: 8080 3000)
#        .vscode/kill.sh 3000 5000  (custom ports)

PORTS=("${@:-8080 3000}")
# Split the default string into separate ports if no args were given
if [ $# -eq 0 ]; then
  PORTS=(8080 3000)
fi

for port in "${PORTS[@]}"; do
  # Find PIDs listening on the port (lsof first, fall back to ss)
  if command -v lsof >/dev/null 2>&1; then
    pids=$(lsof -t -i :"$port" -sTCP:LISTEN 2>/dev/null)
  else
    pids=$(ss -ltnp "sport = :$port" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | sort -u)
  fi

  if [ -z "$pids" ]; then
    echo "Port $port: nothing is listening."
    continue
  fi

  for pid in $pids; do
    echo "Port $port: killing PID $pid ($(ps -p "$pid" -o comm= 2>/dev/null))"
    kill "$pid" 2>/dev/null

    # Give it a moment to exit gracefully, then force kill if still alive
    sleep 1
    if kill -0 "$pid" 2>/dev/null; then
      echo "  PID $pid still running, sending SIGKILL"
      kill -9 "$pid" 2>/dev/null
    fi
  done
done