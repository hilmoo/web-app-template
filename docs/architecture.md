# Backend architecture rules (`core/`)

Features live in `core/internal/features/<name>/`. `core/internal/app/` is the only
composition root: it builds every feature and wires them together. This page is the
rulebook for how features talk to each other.

## The rules

1. **Features never import other features.** Enforced by `core/internal/archtest`
   (runs in `go test ./...`).
2. **Features never import `internal/app` or `internal/cmd`.** Shared packages
   (`helper`, `gen`, `config`, `x`) never import `internal/features`. Both enforced by
   `depguard` in `core/.golangci.yml`.
3. **A feature only calls the sqlc queries of tables it owns.** `db/sql/<name>.sql` belongs
   to feature `<name>` (today `user.sql` and `session.sql` belong to `auth`). The sqlc
   package is shared, so this one is enforced by review.
4. **The consumer defines the interface, the provider satisfies it implicitly,
   `internal/app` wires them.**
5. **No DI frameworks, no globals** beyond static policy tables.

## Synchronous calls between features

Say `billing` needs a user's email and `auth` owns users.

```go
// features/billing/deps.go: the consumer declares only what it needs, in its own types.
type UserLookup interface {
	UserByID(ctx context.Context, id uuid.UUID) (BillingUser, error)
}

// features/auth/directory.go: the provider exports a purpose-built type, and knows nothing of billing.
type Directory struct{ /* ... */ }
func (d *Directory) UserByID(ctx context.Context, id uuid.UUID) (UserInfo, error)

// internal/app: the only place that sees both. Add an adapter only when the types differ.
type billingUsers struct{ d *auth.Directory }

func (a billingUsers) UserByID(ctx context.Context, id uuid.UUID) (billing.BillingUser, error) {
	u, err := a.d.UserByID(ctx, id)
	return billing.BillingUser{ID: u.ID, Email: u.Email}, err
}
```

Guidelines:

- Interfaces have 1 to 3 methods and use the consumer's own types (or primitives, `uuid.UUID`).
- Cross the boundary with plain `error` (or sentinel errors the provider exports).
  Convert to `*herodot.DefaultError` at the consumer's handler edge.
- Pass the acting user as an argument. Callees do not read identity from `ctx`.
- Wiring order is decided in `internal/app`. A wiring cycle means: merge the features,
  or turn one direction into an event.

The reference implementation is `helper.SessionResolver` (consumer: the auth interceptor
in `helper`), `auth.SessionResolver` (provider), wired in `app.RegisterConnect`.

## Events

Use events only for "react to X happened", where the caller must not wait on or fail
because of the reaction. Use a synchronous call if the caller needs an answer.

The chosen mechanism is a **transactional outbox in Postgres**: the producer inserts an
`outbox` row in the same transaction as its domain write, and a worker started from
`internal/app` (polling with `FOR UPDATE SKIP LOCKED`, optionally woken by
`LISTEN/NOTIFY`) dispatches to handlers registered in `internal/app`. Handlers must be
idempotent (delivery is at-least-once). An in-process bus is not used, since events would
be lost on a crash. **None of this is built yet**; build it with the first real use case.

## Transactions across features

Planned design (build it with the first multi-feature write): a tiny runner in
`internal/platform/pgtx` that carries the `pgx.Tx` in `ctx`.

```go
func (r *Runner) InTx(ctx context.Context, fn func(ctx context.Context) error) error // joins a tx already in ctx
func (r *Runner) Q(ctx context.Context) *db.Queries                                    // queries bound to the ctx tx, else the pool
```

- Feature code gets its queries from `Q(ctx)`, never from a captured `*db.Queries`.
- The consumer that needs atomicity declares `type TxRunner interface { InTx(...) error }`;
  `internal/app` injects the runner. Test fakes just call `fn(ctx)`.
- Only the outermost caller starts a transaction. A provider method never starts its own;
  it joins the caller's, or runs on the pool when called alone.
- Provider interfaces stay free of `pgx`/sqlc types, which is why the tx travels in `ctx`
  rather than as a parameter.

## Shared types

Create `internal/shared/` only when a second feature needs a common **value type** (IDs,
`Money`, pagination). It may contain immutable types and pure methods. It must not hold
services or I/O interfaces, and must not import features, `gen`, `helper`, `config` or
`pgx`. Types that only exist because two features are adjacent belong in the consumer's
interface file instead.

## Testing a feature in isolation

Give the consumer its dependency as an interface field and use a hand-written fake in the
test (no mock generators):

```go
type fakeUsers struct {
	u   billing.BillingUser
	err error
}

func (f fakeUsers) UserByID(context.Context, uuid.UUID) (billing.BillingUser, error) { return f.u, f.err }
```

See `core/internal/helper/auth_test.go` for the pattern. Providers are tested against a
real database with `//go:build integration` tests.

## When to merge two features

- They need each other (A calls B and B calls A).
- They write the same tables, or one invariant spans both and needs several shared-tx calls.
- One is never used without the other.
- The adapter in `internal/app` is growing to mirror the provider's whole API.

## Checklist for a new feature

The steps in the README ("Adding a feature") still apply, plus:

1. Register its handler with `register(...)` in `app.RegisterConnect`.
2. Pass its `Access` policy to `helper.NewAuthInterceptor` there (not covered by
   `TestAllServicesWired`).
3. Need something from another feature? Define the interface in your package and add the
   wiring in `internal/app`. Do not import the other feature.
