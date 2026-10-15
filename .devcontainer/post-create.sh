direnv allow
devbox install

[ -f .env ] || cp .env.example .env

# golang stuff
devbox run go install -v golang.org/x/tools/gopls@latest

# node stuff
devbox run pnpm install

# protoforge (drizzle -> typespec -> openapi -> proto)
curl -L https://github.com/hilmoo/protoforge/releases/download/v0.0.1/protoforge_0.0.1_linux_amd64.tar.gz | tar -xz -C /tmp && sudo install -m 755 /tmp/protoforge /usr/local/bin/protoforge

direnv reload
