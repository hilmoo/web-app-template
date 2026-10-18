---
name: project-pattern
description: Map of this monorepo (Go ConnectRPC core, React frontend, Drizzle migration, TypeSpec openapi, generated proto) and the end-to-end workflow for changing it. Use whenever adding a new feature or endpoint, changing a table/column, editing an API contract, regenerating code, running `task` commands (gen, migrate, lint, test, build, dev), or when unsure where a file belongs or whether a file is generated, even if the user only names one layer ("add a field to users", "new page for X", "run codegen"). Read it before touching migration/, openapi/, proto/, or any *_generated / gen / sqlc directory.
---

# Project Pattern

Full-stack template: **Go** (ConnectRPC) backend in `core/`, **React** (Vite, Mantine,
TanStack Query + Router, connect-query) in `frontend/`, **Postgres**. Ships with
email/password auth only; `auth` is the reference feature for every layer.

For the Go layering (handler → service → repository, errors, Deps/Provide wiring), load
the **core-pattern** skill. This skill covers the parts that cross packages.

## Project map

```
core/        Go server: ConnectRPC handlers, sqlc queries, embeds the built SPA
frontend/    Vite + React app (pnpm workspace package "frontend")
migration/   Drizzle schema: source of truth for tables
openapi/     TypeSpec API definitions -> openapi.json
proto/       Generated .proto registry + Go module ("proto") + TS package ("proto")
scripts/     rename.sh (one-off placeholder rename), kill-orphans.sh
```

| Path | Purpose |
|---|---|
| `migration/src/schema/<name>.ts` | One Drizzle table file per table (comment each column; comments flow into docs) |
| `openapi/src/<feature>/main.tsp` | Feature `interface` with one op per RPC; models in sibling `.tsp` files |
| `openapi/src/main.tsp` | Imports every feature folder |
| `openapi/src/common/` | Shared `CommonError.*`, `CommonConnect.Empty`, response helpers |
| `openapi/src/_generated/db.augmented.tsp` | **Hand-written** formats/visibility on top of generated `db.tsp` |
| `core/internal/db/queries/<entity>.sql` | Hand-written sqlc queries, singular (`user.sql`, `session.sql`) |
| `core/internal/features/<feature>/` | Handler, service, repository, errors, access policy |
| `core/internal/app/handler.go` | `RegisterConnect`: Provide each feature and `register(...)` its handler |
| `core/internal/config/` | Env config |
| `frontend/src/features/<feature>/` | `api/` hooks, `components/`, `pages/`, `schemas.ts`, `index.ts` (public API) |
| `frontend/src/routes/` | TanStack file routes (thin: import from `#/features/<x>`) |
| `frontend/src/shared/` | Cross-feature components and lib |

### Naming

Pick one singular feature name (e.g. `note`) and use it in every layer, so a feature is
easy to find across packages:

| Layer | Convention | `auth` / `users` example | New `note` feature |
|---|---|---|---|
| Table / Drizzle file | plural snake_case | `users`, `schema/users.ts` | `notes`, `schema/notes.ts` |
| Query file | singular entity | `queries/user.sql` | `queries/note.sql` |
| TypeSpec folder, Go package, frontend folder | feature name | `openapi/src/auth`, `features/auth` | `openapi/src/note`, `features/note` |
| Connect service | `<Feature>Service` | `AuthService` | `NoteService` |
| RPC / route | `<Feature><Op>` (feature first) | `AuthLogin`, `AuthUpdateMe` | `NoteCreate`, `NoteList` |

Follow the `<Feature><Op>` order even when the user names it differently ("GetStats" becomes
`StatsGet`). Query names in `.sql` are verb-first (`GetUser`, `ListSessionsByUser`).

Root tooling: `Taskfile.yml` (+ one per package), `protoforge.yaml` (generation pipeline),
`devbox.jsonc` (pinned toolchain), `.lefthook.yml`, `.github/workflows/ci.yml`, `Dockerfile`,
`zellij-dev.kdl`.

## Generated code: edit the source, never the output

Pipeline (driven by protoforge, configured in `protoforge.yaml`):

```
Drizzle schema ─► SQL migrations ─► sqlc Go
      └────────► TypeSpec db.tsp ─► openapi.json ─► .proto ─► Go + TS clients
                                         └────────► zod schemas (orval)
```

| Source (edit this) | Generated (do not edit) |
|---|---|
| `migration/src/schema/` | `migration/drizzle/`, `migration/schema.dbml`, `core/internal/db/migration/migrations/`, `openapi/src/_generated/db.tsp` |
| `openapi/src/` | `openapi/openapi.json`, `proto/registry/`, `proto/gen/` |
| `core/internal/db/queries/` + migrations | `core/internal/db/sqlc/` |
| `openapi/openapi.json` | `frontend/src/api/schemas/` |
| `frontend/src/routes/` | `frontend/src/routeTree.gen.ts` (Vite plugin / `pnpm --filter frontend generate-routes`) |

Generated files are committed. If generated output looks wrong, fix the source or the
generator config and regenerate; a hand edit is overwritten next run and fails CI drift checks.
Never hand-write a migration SQL file: change the Drizzle schema and generate.

## `task` commands

Run `task --list` for the full set. Namespaces: `gen:`, `core:`, `frontend:`, `migration:`,
`openapi:`, `proto:`. The ones you will actually use:

**Codegen** (`gen:*` comes from the remote protoforge Taskfile):

| Command | Does | Needs DB? |
|---|---|---|
| `task gen:all` | Full pipeline: `gen:sql` then `gen:api` | yes |
| `task gen:sql` | Drizzle generate + **push to the dev DB**, sync migrations into core, then `task core:gen-sql` | yes |
| `task gen:api` | TypeSpec → openapi.json → .proto → buf (Go + TS, tag injection, TS build) → orval | no |
| `task gen:preview` | Drizzle → DBML + OpenAPI only, never touches the DB (`:watch` variant too) | no |
| `task gen:proto` | Schema → OpenAPI → .proto, no codegen | no |
| `task gen:plan` | Dry run; prints what would happen | no |
| `task gen:verify` | CI-style drift check on `openapi/` and `proto/` | no |
| `task gen:doctor` | Checks required external tools are installed | no |
| `task core:gen-sql` | pg_format the queries, then sqlc only (use after editing `.sql` files) | no |
| `task frontend:orval` | Regenerate zod schemas only | no |
| `task proto:build-typescript` | Build the TS `proto` package (needed before frontend typecheck/lint on a fresh clone) | no |

`gen:sql` / `gen:all` push schema to the database in `DATABASE_URL`, so Postgres must be
running (devcontainer provides it). If the DB is unavailable, say so rather than skipping
the step silently; `gen:api` alone is safe when only `openapi/src` changed.

**Dev:** `task dev` (core on `:8080` serving `/api` and proxying Vite, frontend on
`:3000`, in zellij), `task dev:kill`.

**Dev database:** Drizzle `push` keeps it in sync. `task gen:sql` / `task gen:all` push
automatically; `task migration:push` pushes without regenerating (e.g. a fresh dev DB).

**`task core:migrate-*` is for production databases only** (`migrate-up`, `migrate-down`,
`migrate-status`, `migrate-up-to:<v>`, `migrate-down-to:<v>`; in the image:
`/app/main migrate up`). Never run these as part of the dev workflow: the dev DB already has
the change from `push`, and these commands act on whatever `DATABASE_URL` points to. Run them
only when the user explicitly asks to migrate a production database.

**Checks:** per package `task core:<x>` / `task frontend:<x>` / `task migration:<x>`:

| Task | core | frontend | migration |
|---|:-:|:-:|:-:|
| `lint` (fixes) / `lint:check` | ✓ | ✓ | ✓ |
| `fmt` / `fmt:check` | ✓ | ✓ | ✓ |
| `test` | ✓ | ✓ | |
| `compile` | ✓ | | |
| `build` | ✓ | ✓ | |

Root shortcuts: `task lint`, `task build` (frontend build then embed into core binary).
Frontend typecheck: `pnpm --filter frontend typecheck`. `task openapi:lint` runs spectral.

## Adding a feature (end to end)

Work top-down along the pipeline; each step consumes the previous step's output. Use
`auth` as the copy source at every layer.

1. **Table**: add `migration/src/schema/<table>.ts` (uuid `id` with `defaultRandom()`,
   `timestamp(..., { withTimezone: true })`, snake_case column names, comment per column,
   explicit indexes). Skip if the feature needs no new table.
2. **API contract**: create `openapi/src/<feature>/main.tsp` with a `@tag`'d `interface`;
   each op is `@post @route("/services.v1.<Feature>Service/<Feature><Op>")` and returns
   `Model | CommonError.<...>`. Build models from `DbSchema.<Table>Row` (e.g.
   `OmitProperties`, `FilterVisibility`) in sibling `.tsp` files; put extra formats or
   visibility in `_generated/db.augmented.tsp`. Add `import "./<feature>";` to
   `openapi/src/main.tsp`.
3. **Generate**: `task gen:all`.
4. **Queries**: write `core/internal/db/queries/<entity>.sql` (sqlc annotations; add to the
   existing file when the table already has one), then `task core:gen-sql`.
5. **Backend**: implement `core/internal/features/<feature>/` per the core-pattern skill,
   then in `core/internal/app/handler.go` call `<feature>.Provide(...)` and
   `register(servicesv1connect.New<Feature>ServiceHandler(bundle.Handler, opts...))`.
   Procedures require auth by default; public ones go in the feature's `AccessPolicy`
   (see `core/internal/features/auth/access.go`), internal-only ones in `AccessInternalOnly`.
6. **Frontend**: in `frontend/src/features/<feature>/`:
   - `api/use-<op>.ts`: `useQuery` / `useMutation` from `@connectrpc/connect-query` with
     `<Feature>Service.method.<op>` imported from `proto/servicesV1`; update or invalidate
     the query cache in `onSuccess`.
   - `schemas.ts`: extend the orval zod body from `#/api/schemas/<tag>/...` with UI messages.
   - `components/`, `pages/`, and `index.ts` exporting only what other code may use.
   - Route in `frontend/src/routes/<path>.tsx`: `createFileRoute`, `component` from
     `#/features/<feature>`, guards like `requireAuth` in `beforeLoad`.
7. **Verify**: see checklist below.

Changing an existing column or endpoint is the same flow starting at the layer you changed:
schema change → step 1 onward; contract-only change → step 2 onward with `task gen:api`;
query-only change → `task core:gen-sql`.

## Before finishing

- `task core:compile`, `task core:test`, `task core:lint:check`, `task core:fmt:check`
- `pnpm --filter frontend typecheck`, `task frontend:test`, `task frontend:lint:check`, `task frontend:fmt:check`
- `task migration:lint:check` if the schema changed
- Generated files committed alongside their sources (`git status` should show both)

CI runs the same lint/format checks plus `sqlc diff`, Go tests, frontend typecheck/test/build,
and a Docker build. Lefthook runs lint + format on staged Go/TS files at pre-commit.

Go tests live next to the code (`*_test.go`); tests needing Postgres use
`//go:build integration` and run with `go test -tags=integration ./...`. Frontend tests are
`src/**/*.test.ts`.

## Configuration

`.env` (git-ignored, copy from `.env.example`): `DATABASE_URL`, `SECRET`, `SERVICE_SECRET`
(required); `APP_ENV=production` enables secure cookies and panic recovery and enforces
strong distinct secrets; optional `LISTEN_ADDR` (default `127.0.0.1:8080`), `LOG_LEVEL`
(default `INFO`). New config goes in `core/internal/config/` and `.env.example`.
