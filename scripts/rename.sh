#!/usr/bin/env bash
# Replace the `web-app-template` placeholder with your app name.
# Usage: scripts/rename.sh my-app
set -euo pipefail

name="${1:?usage: scripts/rename.sh <app-name>}"
[[ "$name" =~ ^[a-z][a-z0-9-]*$ ]] || { echo "name must be lowercase letters, digits and dashes" >&2; exit 1; }

cd "$(dirname "$0")/.."

grep -rIl "web-app-template" . \
  --exclude-dir=node_modules --exclude-dir=.git --exclude-dir=.devbox --exclude=rename.sh \
  | xargs sed -i "s/web-app-template/${name}/g"

echo "Renamed to ${name}. Regenerate with: task gen-all"
