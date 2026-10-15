# web-app-template

Full-stack template: **Go** (ConnectRPC) backend, **React** frontend (ConnectRPC + TanStack Query + TanStack Router), Postgres.

The API contract is derived from the database schema, in one pipeline driven by
[protoforge](https://github.com/hilmoo/protoforge):

```
migration/src/schema (drizzle)
   ├─► core/internal/db/migration/migrations   (SQL migrations, embedded in core)
   └─► openapi/src/_generated/db.tsp ─► openapi/openapi.json ─► proto/registry/*.proto
                                                                  └─► proto/gen/{go,typescript}  (buf generate)
core/internal/db/sql ─► core/internal/gen/sqlc   (sqlc generate)
openapi/openapi.json ─► fe/src/api/schemas       (orval, zod schemas for forms)
```

## Layout

| Dir          | What                                                                 |
| ------------ | -------------------------------------------------------------------- |
| `core/`      | Go server: ConnectRPC handlers, sqlc queries, embedded SPA            |
| `fe/`        | Vite + React + Mantine, `@connectrpc/connect-query`, TanStack Router |
| `migration/` | Drizzle schema (source of truth for tables)                          |
| `openapi/`   | TypeSpec API definitions (`openapi.json` is generated)               |
| `proto/`     | Generated `.proto` registry + generated Go / TypeScript clients      |

## Getting started

1. Open in the devcontainer (installs devbox, pnpm deps, protoforge). Or run `devbox shell` + `pnpm install`.
2. `scripts/rename.sh <your-app-name>` — replaces the `web-app-template` placeholder everywhere (Go module, DB name, package names).
3. `cp .env.example .env`
4. `task gen-all` — generates migrations, OpenAPI, proto and clients. Needs Postgres up (`push` step) and creates the first migration.
5. `task core:gen-sql` — sqlc models. `task fe:orval` — zod schemas.
6. `task core:migrate-up`
7. `task dev` — zellij session with core (`:8080`, serves `/api` and proxies Vite) and fe (`:3000`).

Open <http://localhost:8080>.

## Adding a feature

1. Add a table in `migration/src/schema/<name>.ts`.
2. Add `openapi/src/<name>/{main,model}.tsp` (copy `todos/`) and import it in `openapi/src/main.tsp`.
3. `task gen-all`, then write `core/internal/db/sql/<name>.sql` and `task core:gen-sql`.
4. Implement `core/internal/features/<name>` (handler/service/helper, see `todo`) and register it in `core/internal/app/handler.go`.
5. Add `fe/src/features/<name>` hooks using `useQuery(Service.method.x)` / `useMutation(...)` and a route in `fe/src/routes`.

Procedures require authentication by default; opt out per procedure via an `AccessPolicy` (see `features/auth/access.go`).
