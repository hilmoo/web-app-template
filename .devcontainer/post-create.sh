#!/usr/bin/env bash
set -e

add_to_zshrc() {
  touch ~/.zshrc
  grep -qxF "$1" ~/.zshrc || echo "$1" >> ~/.zshrc
}

# host nix (daemon mode) is bind-mounted at /nix: make nix visible and fetch devbox if it isn't
[ -e /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh ] && . /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh
command -v devbox >/dev/null || {
  nix --extra-experimental-features 'nix-command flakes' profile add github:NixOS/nixpkgs/319e4cebef10cbb636c6ea392137b52ad105336a#devbox # v0.18.4
}
add_to_zshrc 'eval "$(devbox global shellenv)"'
eval "$(devbox global shellenv)"

direnv allow
devbox install

[ -f .env ] || cp .env.example .env

# node stuff
devbox run pnpm install

# other stuff
devbox run lefthook install

# protoforge (drizzle -> typespec -> openapi -> proto)
PROTOFORGE_VERSION=0.0.1
curl -L "https://github.com/hilmoo/protoforge/releases/download/v${PROTOFORGE_VERSION}/protoforge_${PROTOFORGE_VERSION}_linux_amd64.tar.gz" | tar -xz -C /tmp && sudo install -m 755 /tmp/protoforge /usr/local/bin/protoforge

direnv reload

perl -0pi -e 's/,(\s*[}\]])/$1/g' devbox.jsonc

add_to_zshrc 'eval "$(task --completion zsh)"'