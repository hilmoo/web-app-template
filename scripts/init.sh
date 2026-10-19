#!/usr/bin/env bash
# Interactive setup: replace the `web-app-template` placeholder with your app
# name and configure the devcontainer's nix (host or shared docker volume).
# Usage: scripts/rename.sh   (deletes itself when done)
set -euo pipefail

self="$(realpath "$0")"
cd "$(dirname "$self")/.."

ask() { # ask <prompt> <default> -> sets REPLY
  read -r -p "$1 [$2]: " REPLY
  REPLY="${REPLY:-$2}"
}

confirm() { # confirm <prompt> <y|n default>
  local hint="y/N"
  [[ "$2" == y ]] && hint="Y/n"
  read -r -p "$1 [$hint]: " REPLY
  REPLY="${REPLY:-$2}"
  [[ "$REPLY" =~ ^[Yy] ]]
}

# --- project name ---
while true; do
  ask "Project name (lowercase letters, digits, dashes)" "$(basename "$PWD")"
  name="$REPLY"
  [[ "$name" =~ ^[a-z][a-z0-9-]*$ ]] && break
  echo "Invalid name: use lowercase letters, digits and dashes, starting with a letter." >&2
done

grep -rIl "web-app-template" . \
  --exclude-dir=node_modules --exclude-dir=.git --exclude-dir=.devbox --exclude=rename.sh \
  | xargs -r sed -i "s/web-app-template/${name}/g"
echo "Renamed to ${name}."

# --- devcontainer nix ---
env_file=.devcontainer/.env
write_env=true
if [[ -f "$env_file" ]] && ! confirm "$env_file already exists. Overwrite?" n; then
  write_env=false
fi

if $write_env; then
  if confirm "Use the host's nix (/nix, daemon mode) in the devcontainer?" n; then
    printf 'NIX_MOUNT_SOURCE=/nix\nNIX_REMOTE=daemon\n' > "$env_file"
    echo "Wrote $env_file (host nix)."
  else
    : > "$env_file"
    echo "Wrote $env_file (shared named volume)."
  fi
fi

# The compose volume is `external: true`, so it must exist before the container starts.
if ! grep -q '^NIX_MOUNT_SOURCE=' "$env_file" 2>/dev/null; then
  volume="$(sed -n 's/^NIX_VOLUME_NAME=//p' "$env_file" 2>/dev/null | tail -n1)"
  volume="${volume:-$(sed -n 's/.*NIX_VOLUME_NAME:-\([^}]*\)}.*/\1/p' .devcontainer/compose.yaml | head -n1)}"
  if ! command -v docker >/dev/null 2>&1; then
    echo "docker not found; create the volume manually: docker volume create ${volume}" >&2
  elif docker volume inspect "$volume" >/dev/null 2>&1; then
    echo "Docker volume ${volume} already exists."
  else
    docker volume create "$volume" >/dev/null
    echo "Created docker volume ${volume}."
  fi
fi

rm -- "$self"
echo "Done (removed scripts/rename.sh). Regenerate with: task gen-all"
