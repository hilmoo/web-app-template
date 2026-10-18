---
name: core-pattern
description: Architecture and coding pattern for the Go backend in core/ (Connect RPC + sqlc + pgx + herodot). Use whenever adding or changing a feature, RPC handler, service, repository, sqlc query, transaction, error mapping, access policy, interceptor, or wiring in core/internal — so every feature follows the same handler → service → repository layering, sentinel-error flow, and per-feature Deps/Provide wiring.
---

# Core Pattern (`core/`)

The Go backend in `core/` serves Connect RPC (protobuf) under `/api/`, a health
endpoint, and the embedded frontend. Every feature under `core/internal/features/<feature>/`
follows the same layering, error flow, and wiring described here.

## Stack

| Concern | Choice |
|---|---|
| Transport | Connect RPC (`connectrpc.com/connect`) on `net/http` ServeMux — generated from `proto/` |
| API contract | `openapi/` → proto (`proto/registry/...`) → generated Go in `proto/services/v1`, `proto/models/v1` |
| Validation | `validate:"..."` tags injected into proto structs; `*validation.Vld.ValidatePayload` (gomox) |
| DB | Postgres via `pgx/v5` `*pgxpool.Pool`; queries via `sqlc` (`internal/db/sqlc`, package `db`) |
| Schema | Migrations in `internal/db/migration/migrations` (authored from `migration/` with drizzle) |
| Errors | Plain Go errors + feature sentinels → `*herodot.DefaultError` → `errort.ConnectError` |
| Logging | `*slog.Logger`; request/error logging is done once by the `mlog.ConnectLogger` interceptor |
| Auth | `helper.NewAuthInterceptor` + per-procedure `helper.AccessPolicy` |

## Directory Layout

```
core/
├── cmd/main.go                    # → internal/cmd.RootCmd()
└── internal/
    ├── cmd/                       # CLI: serve, migrate, healthcheck (urfave/cli)
    ├── config/                    # env config (caarlos0/env)
    ├── app/                       # composition root: deps, server, mux, RegisterConnect
    ├── helper/                    # cross-feature infra: auth interceptor, WithTx, Page, HandlerParams
    ├── db/
    │   ├── migration/migrations/  # *.up.sql (embedded)
    │   ├── queries/<table>.sql    # sqlc input
    │   └── sqlc/                  # sqlc output — generated, never edit
    ├── features/<feature>/        # one package per feature (see below)
    ├── cache/  event/             # future: see references/cache-event-patterns.md
    └── x/typex/                   # tiny generic utilities
```

## Feature Module Layout

```
internal/features/<feature>/
├── types.go        # domain model + service input structs (no proto, no sqlc types)
├── repository.go   # Repository interface + postgresRepo (sqlc → domain, pgx errors → sentinels)
├── service.go      # Service interface + service struct (business rules, ownership, quotas)
├── handler.go      # connectHandler: proto ⇄ domain, validation, user context, error dispatch
├── mapper.go       # domain → proto converters (toProto<Feature>)
├── errors.go       # sentinels, <feature>NotFound(id), toAppError, toConnectError
├── provider.go     # Deps + Provide(): repo → service → handler
├── access.go       # AccessPolicy overrides (public / service-only procedures)
├── events.go       # optional: topic constants
└── listeners.go    # optional: registerListeners
```

## Layer Rules

```
Connect request
   │
   ▼
handler   ── validate proto (syntactic) → user from ctx → proto→input → svc → domain→proto
   │          errors: syntactic → errort.ConnectError(ctx, errH); domain → toConnectError(ctx, err)
   ▼
service   ── semantic rules, ownership (→ NotFound), quotas; returns sentinels / wrapped errors
   │          knows nothing about proto, Connect, herodot
   ▼
repository── sqlc calls, transactions (helper.WithTx), row → domain;
              pgx.ErrNoRows / unique violation → sentinels; other → fmt.Errorf("<feature>: <op>: %w")
```

| Layer | May import | Must NOT import |
|---|---|---|
| handler | service iface, proto, herodot, errort, helper | sqlc, pgx |
| service | repository iface, domain types, slog | proto, herodot, connect, sqlc |
| repository | sqlc `db`, pgx, gomox `sqlx`, helper.WithTx | proto, herodot, connect |
| feature ↔ feature | — | never import another feature; use consumer-side interfaces wired in `internal/app` |

## Adding a Feature — Checklist

1. **Contract:** add the service/RPCs in `openapi/`, regenerate proto. RPC names are `<Feature><Action>` (e.g. `AuthLogin`); responses wrap payload in `data` (+ `pagination` for lists).
2. **Schema:** add the migration; write `internal/db/queries/<table>.sql`; run `task gen` in `core/`.
3. **Feature package:** create the files above from [references/feature-module-template.md](references/feature-module-template.md).
4. **Errors:** declare sentinels and map them in `toAppError` — [references/error-flow.md](references/error-flow.md).
5. **Access:** list non-default procedures in `access.go` (default = authenticated).
6. **Wire:** in `internal/app/handler.go` `RegisterConnect`, call `register(servicesv1connect.New<Feature>ServiceHandler(<feature>.Provide(<feature>.Deps{...}), opts...))` and pass `<feature>.Access` to `NewAuthInterceptor` — [references/interceptor-wiring.md](references/interceptor-wiring.md).
7. **Test:** `task test` (`TestAllServicesWired` fails if a proto service is not registered); unit-test the service with a fake `Repository`.
8. **Lint:** `task lint:check` and `task fmt:check`.

## Core Rules (summary)

- Services and repositories return plain `error`; only the handler turns errors into herodot/Connect errors.
- `pgx.ErrNoRows` → feature sentinel in the repository, always.
- Ownership failure → `NotFound` (never leak existence).
- Never log **and** return an error. The `ConnectLogger` interceptor logs every failed call once; `errort.ConnectError` attaches the herodot reason/debug to that log line.
- Handlers never use `errors.Is` — they call `toConnectError(ctx, err)` once.
- Proto types stop at the handler; sqlc types stop at the repository. The domain struct is what flows between them.
- Each feature declares its own `Deps` with only what it needs; `Provide()` is its single wiring point.
- Multi-write atomicity uses `helper.WithTx` inside the repository.
- Cache and event bus are optional deps; degrade gracefully when nil.

## References

| File | Covers |
|---|---|
| [references/feature-module-template.md](references/feature-module-template.md) | Copy-paste scaffold for every file + SQL queries |
| [references/error-flow.md](references/error-flow.md) | Repository → service → handler → Connect error pipeline |
| [references/interceptor-wiring.md](references/interceptor-wiring.md) | Mux layout, interceptor chain, access policy, user context, consumer-side interfaces |
| [references/response-shape.md](references/response-shape.md) | Proto response shape, pagination, Connect error wire format |
| [references/cache-event-patterns.md](references/cache-event-patterns.md) | Future: Redis cache-aside and in-process event bus |

## Migration Status

The pattern above is the target. Code written before it still deviates — migrate it when you touch it:

- `features/auth` services return `*herodot.DefaultError`, take proto requests, call sqlc directly, and are wired with `auth.NewConnect(helper.HandlerParams)` instead of `Deps`/`Provide`.
- `helper.WithTx` and `helper.WithUserID` take callbacks returning `*herodot.DefaultError`; the target signatures return `error` (see error-flow.md).
- `queries/user.sql` and `queries/session.sql` use `EntityVerb` names (`UserGetById`) and `SELECT *`; new queries use `VerbEntity` names and explicit columns.
