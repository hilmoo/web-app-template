# web-app-template

Full-stack template: **Go** (ConnectRPC) backend, **React** frontend (TanStack Query + Router), **Postgres**.
Ships with email/password auth (register, login, sessions, edit profile) and nothing else.

## Project map

```
core/         Go server: ConnectRPC handlers, sqlc queries, embeds the built SPA
frontend/     Vite + React + Mantine, connect-query, TanStack Router
migration/    Drizzle schema: source of truth for tables
openapi/      TypeSpec API definitions (openapi.json is generated)
proto/        Generated .proto registry + Go / TypeScript clients
scripts/      rename.sh (one-off placeholder rename), kill-orphans.sh
```

Where things live:

| Path                                  | Purpose                                      |
| ------------------------------------- | -------------------------------------------- |
| `migration/src/schema/<name>.ts`      | One Drizzle table file per feature           |
| `openapi/src/<name>/`                 | TypeSpec API per feature (`main.tsp`, ...)   |
| `openapi/src/main.tsp`                | Imports every feature's API                  |
| `core/internal/features/<name>/`      | Handlers, services, helpers per feature      |
| `core/internal/app/handler.go`        | Registers feature handlers                   |
| `core/internal/db/sql/`               | Hand-written SQL queries (input to sqlc)     |
| `core/internal/config/`               | Env config                                   |
| `frontend/src/features/<name>/`       | API hooks, components, pages per feature     |
| `frontend/src/routes/`                | File-based routes                            |

Root tooling: `Taskfile.yml` (run `task --list`; each package has its own Taskfile), `devbox.jsonc` (pinned toolchain), `.devcontainer/`, `protoforge.yaml` (generation pipeline), `.lefthook.yml` (git hooks), `.github/workflows/ci.yml`, `Dockerfile`, `zellij-dev.kdl` (layout for `task dev`), `renovate.json`.

## Code generation

One pipeline, driven by [protoforge](https://github.com/hilmoo/protoforge):

```
Drizzle schema → SQL migrations + TypeSpec → openapi.json → .proto → Go / TS clients
```

Generated files are committed and marked `linguist-generated`. Edit the **source**, never the output.

| Source                  | Generated                                                                                  |
| ----------------------- | ------------------------------------------------------------------------------------------ |
| `migration/src/schema`  | `core/internal/db/migration/migrations`, `migration/schema.dbml`, `openapi/src/_generated` |
| `openapi/src`           | `openapi/openapi.json`, `proto/registry`, `proto/gen` (Go and TypeScript)                  |
| `core/internal/db/sql`  | `core/internal/gen/sqlc`                                                                   |
| `openapi/openapi.json`  | `frontend/src/api/schemas` (zod, via orval)                                                |
| `frontend/src/routes`   | `frontend/src/routeTree.gen.ts`                                                            |

| Command         | Does                                                                          |
| --------------- | ----------------------------------------------------------------------------- |
| `task gen-all`  | Everything below (needs Postgres up, since Drizzle `push` runs)               |
| `task gen-sql`  | Drizzle → SQL migrations, then sqlc                                           |
| `task gen-api`  | TypeSpec → OpenAPI → proto → buf (also builds the TS proto package) → orval   |

## Getting started

1. Scaffold (also replaces the `web-app-template` placeholder everywhere):
   ```bash
   npx degit hilmoo/web-app-template my-new-app
   cd my-new-app
   scripts/rename.sh my-new-app && rm scripts/rename.sh
   git init && git add -A && git commit -m "init"
   ```
2. Open in the devcontainer (installs devbox, pnpm deps, git hooks, protoforge and creates `.env`). Without it: `devbox shell`, `pnpm install`, `cp .env.example .env`.
3. `task proto:build-typescript` once, so the frontend can import the TS proto package (`task gen-all` does this too).
4. `task core:migrate-up` applies migrations.
5. `task dev` starts core (`:8080`, serves `/api` and proxies Vite) and frontend (`:3000`) in zellij.

Open <http://localhost:8080> and register an account.

## Adding a feature

Use `auth` as the reference (`core/internal/features/auth`, `frontend/src/features/auth`).

1. Add a table: `migration/src/schema/<name>.ts`.
2. Add `openapi/src/<name>/` (copy the shape of `openapi/src/auth/`) and import it in `openapi/src/main.tsp`.
3. `task gen-all`, write `core/internal/db/sql/<name>.sql`, then `task core:gen-sql`.
4. Implement `core/internal/features/<name>` and register it in `core/internal/app/handler.go`.
5. Add `frontend/src/features/<name>` hooks (`useQuery` / `useMutation`) and a route in `frontend/src/routes`.

Procedures require authentication by default. Opt out per procedure with an `AccessPolicy` (see `core/internal/features/auth/access.go`).

Cross-feature calls, events and transactions follow the rules in [docs/architecture.md](docs/architecture.md).

## Configuration

Set in `.env` (git-ignored; see `.env.example`):

| Variable                   | Notes                                                                 |
| -------------------------- | --------------------------------------------------------------------- |
| `DATABASE_URL`             | Postgres connection string (required)                                 |
| `SECRET`                   | Hashes session tokens (required)                                      |
| `SERVICE_SECRET`           | Bearer token for trusted internal services (required)                 |
| `APP_ENV`                  | `production` enables secure cookies and panic recovery, and requires `SECRET` and `SERVICE_SECRET` to be distinct, 32+ chars, and not the example placeholder |
| `LISTEN_ADDR`, `LOG_LEVEL` | Optional. Defaults: `127.0.0.1:8080`, `INFO`                          |

## Tests and checks

Run per package: `task core:<x>` and `task frontend:<x>`.

| Task                     | core | frontend |
| ------------------------ | :--: | :------: |
| `test`                   | yes  | yes      |
| `lint` / `lint:check`    | yes  | yes      |
| `fmt` / `fmt:check`      | yes  | yes      |
| `compile`                | yes  |          |
| `build`                  | yes  | yes      |

Root shortcuts: `task lint`, `task build`. Frontend typecheck: `pnpm --filter frontend typecheck`.

- Go tests sit next to the code (`*_test.go`). Tests needing Postgres use `//go:build integration`; run with `go test -tags=integration ./...`.
- Frontend tests are `src/**/*.test.ts`.
- CI (`.github/workflows/ci.yml`) runs lint, format check, `sqlc diff`, tests, typecheck and builds.
- [lefthook](https://github.com/evilmartians/lefthook) pre-commit lints and format-checks staged Go and TypeScript files. Skip once with `--no-verify`.

## Production

```bash
docker build -t my-new-app .
docker run --env-file .env -e APP_ENV=production -p 8080:8080 my-new-app
```

The image embeds the built frontend, runs as non-root, and has a healthcheck (`/app/main healthcheck`). It needs an external Postgres. Apply migrations with `/app/main migrate up`.
