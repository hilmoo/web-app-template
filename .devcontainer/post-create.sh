#!/usr/bin/env bash
set -e

# host nix (daemon mode) is bind-mounted at /nix: make nix visible
[ -e /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh ] && . /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh

direnv allow
devbox install

[ -f .env ] || cp .env.example .env

# go stuff
devbox run go work sync

# node stuff
devbox run pnpm install

# lefthook
if [ -d .git ]; then
  devbox run lefthook install
fi

# protoforge (drizzle -> typespec -> openapi -> proto)
PROTOFORGE_VERSION=0.0.6
curl -L "https://github.com/hilmoo/protoforge/releases/download/v${PROTOFORGE_VERSION}/protoforge_${PROTOFORGE_VERSION}_linux_amd64.tar.gz" | tar -xz -C /tmp && sudo install -m 755 /tmp/protoforge /usr/local/bin/protoforge

direnv reload

perl -0pi -e 's/,(\s*[}\]])/$1/g' devbox.jsonc # vscode complains about trailing commas in devbox.jsonc, so remove them
