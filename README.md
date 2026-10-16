# web-app-template

Full-stack template: **Go** (ConnectRPC) backend, **React** frontend (ConnectRPC + TanStack Query + TanStack Router), Postgres. It ships with email/password auth (register, login, sessions, edit profile) and nothing else, so you can start adding your own features right away.

The API contract is derived from the database schema, in one pipeline driven by
[protoforge](https://github.com/hilmoo/protoforge):

```
migration/src/schema (drizzle)
   ├─► core/internal/db/migration/migrations   (SQL migrations, embedded in core)
   └─► openapi/src/_generated/db.tsp ─► openapi/openapi.json ─► proto/registry/*.proto
                                                                  └─► proto/gen/{go,typescript}  (buf generate)
core/internal/db/sql ─► core/internal/gen/sqlc   (sqlc generate)
openapi/openapi.json ─► frontend/src/api/schemas       (orval, zod schemas for forms)
```

## Layout

| Dir          | What                                                                         |
| ------------ | ---------------------------------------------------------------------------- |
| `core/`      | Go server: ConnectRPC handlers, sqlc queries, embedded SPA                   |
| `frontend/`  | Vite + React + Mantine, `@connectrpc/connect-query`, TanStack Router         |
| `migration/` | Drizzle schema (source of truth for tables; not where the SQL migrations live) |
| `openapi/`   | TypeSpec API definitions (`openapi.json` is generated)                       |
| `proto/`     | Generated `.proto` registry + generated Go / TypeScript clients              |

Tooling files at the root:

| File                                         | What                                                                    |
| -------------------------------------------- | ----------------------------------------------------------------------- |
| `Taskfile.yml` (+ one per package)           | All commands, run with [`task`](https://taskfile.dev). `task --list`    |
| `devbox.json`, `.envrc`, `init-hook.sh`      | Pinned toolchain (Go, Node, pnpm, sqlc, buf, ...) via [devbox](https://www.jetify.com/devbox) |
| `.devcontainer/`                             | Devcontainer with Postgres; installs devbox, deps and protoforge        |
| `protoforge.yaml`                            | The schema -> API -> client generation pipeline                         |
| `lefthook.yml`                               | Local git hooks (see [Checks](#checks-and-git-hooks))                   |
| `.github/workflows/ci.yml`                   | CI, mirrors the git hooks                                               |
| `zellij-dev.kdl`                             | Layout for `task dev` (core + frontend side by side)                    |
| `Dockerfile`                                 | Production image (frontend is embedded in the Go binary)                |
| `renovate.json`                              | Dependency updates                                                      |
| `scripts/rename.sh`                          | One-off placeholder rename, removed after scaffolding                   |

## Getting started

0. Scaffold a new project from this template. This also replaces the `web-app-template` placeholder everywhere (Go module, DB name, package names) and removes the rename script:
   ```bash
   npx degit hilmoo/web-app-template my-new-app
   cd my-new-app
   scripts/rename.sh my-new-app && rm scripts/rename.sh
   git init && git add -A && git commit -m "init"
   ```
1. Open in the devcontainer (installs devbox, pnpm deps, builds the TS proto package, installs protoforge). Or run `devbox shell`, then `task setup`.
2. `cp .env.example .env` (the devcontainer does this for you). `.env` is git-ignored.
3. `task core:migrate-up` applies the committed initial migration (users, sessions).
4. `task dev` starts a zellij session with core (`:8080`, serves `/api` and proxies Vite) and frontend (`:3000`).

Open <http://localhost:8080> and register an account.

After changing the schema or API definitions, regenerate with `task gen-all` (needs Postgres up, since drizzle `push` runs), then `task core:gen-sql` / `task frontend:orval` as needed.

## Generated code

Generated files are committed so a plain checkout builds without protoforge, except the TypeScript proto build output, which is built by `task setup` (and in CI and the Dockerfile). They are marked `linguist-generated` in `.gitattributes`, so GitHub collapses them in pull request diffs. When reviewing a change, read the sources first and check the generated files only for surprises:

| Source                              | Generated                                                                                  |
| ----------------------------------- | ------------------------------------------------------------------------------------------ |
| `migration/src/schema`              | `core/internal/db/migration/migrations`, `migration/schema.dbml`, `openapi/src/_generated` |
| `openapi/src`                       | `openapi/openapi.json`, `proto/registry`, `proto/gen`                                      |
| `core/internal/db/sql`              | `core/internal/gen/sqlc` (CI runs `sqlc diff` to catch drift)                              |
| `openapi/openapi.json`              | `frontend/src/api/schemas`                                                                 |
| `frontend/src/routes`               | `frontend/src/routeTree.gen.ts`                                                            |

## Adding a feature

Use `auth` as the reference implementation (`core/internal/features/auth`, `frontend/src/features/auth`).

1. Add a table in `migration/src/schema/<name>.ts`.
2. Add `openapi/src/<name>/{main,model}.tsp` (copy the shape of `openapi/src/auth/`) and import it in `openapi/src/main.tsp`.
3. `task gen-all`, then write `core/internal/db/sql/<name>.sql` and `task core:gen-sql`.
4. Implement `core/internal/features/<name>` (handler/service/helper) and register it in `core/internal/app/handler.go`.
5. Add `frontend/src/features/<name>` hooks using `useQuery(Service.method.x)` / `useMutation(...)` and a route in `frontend/src/routes`.

Procedures require authentication by default; opt out per procedure via an `AccessPolicy` (see `features/auth/access.go`).

## Configuration

Set in `.env` (see `.env.example`):

| Variable                    | Notes                                                                                           |
| --------------------------- | ----------------------------------------------------------------------------------------------- |
| `DATABASE_URL`              | Postgres connection string (required)                                                           |
| `SECRET`                    | Hashes session tokens (required)                                                                |
| `SERVICE_SECRET`            | Bearer token for trusted internal services (required)                                           |
| `APP_ENV`                   | `production` enables secure cookies and panic recovery, and **requires** `SECRET` and `SERVICE_SECRET` to be distinct, at least 32 characters and not the `.env.example` placeholder |
| `LISTEN_ADDR`, `LOG_LEVEL`  | Optional, defaults `127.0.0.1:8080` and `INFO`                                                  |

## Tests

```bash
task test            # core (go test) + frontend (vitest)
task core:test
task frontend:test
```

Go tests live next to the code (`*_test.go`; see `core/internal/config` and `core/internal/features/auth`). Put tests that need Postgres behind `//go:build integration` and run them with `go test -tags=integration ./...`. Frontend tests are `src/**/*.test.ts` (see `frontend/src/features/auth/schemas.test.ts`).

## Checks and git hooks

[lefthook](https://github.com/evilmartians/lefthook) installs on `pnpm install` (or run `pnpm exec lefthook install`):

- **pre-commit**: lints and format-checks staged Go and TypeScript files.
- **pre-push**: runs `task check`, which is the same set of checks as CI (lint, format check, typecheck, unit tests, compile).

Skip a hook once with `--no-verify`. If a project has no GitHub Actions access, delete `.github/workflows/ci.yml` and rely on the hooks.

Other useful tasks: `task fmt`, `task lint` (auto-fixes), `task typecheck`, `task build`.

## Production

```bash
docker build -t my-new-app .
docker run --env-file .env -e APP_ENV=production -p 8080:8080 my-new-app
```

The image embeds the built frontend, runs as non-root and has a healthcheck (`/app/main healthcheck`). It expects an external Postgres and runs `serve` by default; apply migrations with `/app/main migrate up`.
