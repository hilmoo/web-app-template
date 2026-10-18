# Error Flow

Every error travels one path: plain Go errors inside the feature, converted to a
herodot error exactly once in the handler, turned into a Connect error by
`errort.ConnectError`, and logged exactly once by the `ConnectLogger` interceptor.

## The Chain

```
Repository              Service                 Handler                         Interceptor
   │                       │                       │                                 │
   ├─ pgx.ErrNoRows        │                       │                                 │
   │  → Err<X>NotFound     │                       │                                 │
   ├─ unique violation     │                       │                                 │
   │  → Err<X>AlreadyExists│  (all via helper.MapDBError)                            │
   ├─ other DB error       │                       │                                 │
   │  → fmt.Errorf(        │                       │                                 │
   │   "<feat>: <op>: %w") │                       │                                 │
   └──────────────────────→│                       │                                 │
                           ├─ ownership → <x>NotFound(id)                            │
                           ├─ business rule → Err<Rule>                              │
                           └──────────────────────→│                                 │
                                                   │ toConnectError(ctx, err)        │
                                                   │   toAppError(err)               │
                                                   │   ├─ errors.Is(sentinel) → herodot.ErrX.WithID(CODE)
                                                   │   └─ unmapped → herodot.ErrInternalServerError
                                                   │   errort.ConnectError(ctx, errH)│
                                                   │     ├─ attach herodot attrs to request log
                                                   │     └─ connect.NewError(code, errH) + ErrorInfo
                                                   └────────────────────────────────→│
                                                                     mlog.ConnectLogger logs the call once
                                                                     (level by Connect code)
```

## Stage 1: Repository — map through `helper.MapDBError`

```go
func (r *postgresRepo) FindByID(ctx context.Context, id uuid.UUID) (Thing, error) {
	row, err := r.q.GetThing(ctx, id)
	if err != nil {
		return Thing{}, helper.MapDBError(err, "thing: find", ErrThingNotFound, nil)
	}
	return thingFromRow(row), nil
}

func (r *postgresRepo) Create(ctx context.Context, in CreateThingInput) (Thing, error) {
	row, err := r.q.CreateThing(ctx, db.CreateThingParams{Name: in.Name})
	if err != nil {
		return Thing{}, helper.MapDBError(err, "thing: create", nil, ErrThingAlreadyExists)
	}
	return thingFromRow(row), nil
}
```

`MapDBError(err, op, notFound, conflict)`: `pgx.ErrNoRows` → `notFound`, unique violation → `conflict`, anything else → `fmt.Errorf("<op>: %w", err)`; `nil` err → `nil`. Pass `nil` for a sentinel the call cannot produce. The original error stays in the chain, so logs keep the pgx detail.

**Rules:**

- Every sqlc call's error goes through `helper.MapDBError`; no hand-rolled `errors.Is(err, pgx.ErrNoRows)` or `sqlx.ClassifyPgError` in repositories.
- `op` is `"<feature>: <operation>"`.
- FK / not-null / check violations are not mapped by the helper; handle one explicitly only when the client must see it.
- Never return raw pgx errors upward without wrapping; never import herodot/connect here.
- Don't use `sqlx.MapPgxErrorHTTP` in new code — it builds herodot errors inside the data layer.

## Stage 2: Service — semantic validation + sentinels

```go
func (s *service) Get(ctx context.Context, userID, id uuid.UUID) (Thing, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Thing{}, err // sentinel or wrapped error passes through
	}
	if item.UserID != userID {
		return Thing{}, thingNotFound(id) // ownership → not-found
	}
	return item, nil
}

func (s *service) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	user, err := s.repo.FindByEmail(ctx, in.Email)
	if errors.Is(err, ErrUserNotFound) {
		return LoginResult{}, ErrInvalidCredentials // don't reveal which part was wrong
	}
	if err != nil {
		return LoginResult{}, err
	}
	// ...
}
```

**Rules:**

- Return sentinels or already-wrapped repo errors; add context with `fmt.Errorf("<feature>: <op>: %w", err)` only for non-sentinel failures produced here (hashing, token generation).
- Ownership failure → `NotFound`, never `Forbidden`.
- Services may translate one sentinel into another (`ErrUserNotFound` → `ErrInvalidCredentials`) — that is a business decision.
- Never log and return. Log only best-effort side effects you swallow (cache, events).

## Stage 3: Handler — single-line dispatch

```go
item, err := h.svc.Get(ctx, userID, id)
if err != nil {
	return nil, toConnectError(ctx, err)
}
```

| Error source | How to handle |
|---|---|
| `helper.UserIDFromContext`, `helper.StrToUUIDHTTP`, `h.validate.ValidatePayload` (already `*herodot.DefaultError`) | `return nil, errort.ConnectError(ctx, errH)` |
| Anything returned by the service | `return nil, toConnectError(ctx, err)` |

Never `errors.Is` in a handler. Never build `connect.NewError` by hand.

## Stage 4: toAppError — the only sentinel → herodot mapping

```go
func toConnectError(ctx context.Context, err error) error {
	return errort.ConnectError(ctx, toAppError(err))
}

func toAppError(err error) *herodot.DefaultError {
	switch {
	case errors.Is(err, ErrThingNotFound):
		return herodot.ErrNotFound.WithID("THING_NOT_FOUND").WithReason("thing not found").WithWrap(err).WithDebug(err.Error())
	case errors.Is(err, ErrThingQuotaExceeded):
		return herodot.ErrConflict.WithID("THING_QUOTA_EXCEEDED").WithReason("quota exceeded").WithWrap(err).WithDebug(err.Error())
	}
	return herodot.ErrInternalServerError.WithReason("internal error").WithWrap(err).WithDebug(err.Error())
}
```

**Rules:**

- One `case` per sentinel; add a case whenever you add a sentinel.
- `WithID` = stable SCREAMING_SNAKE code (`<FEATURE>_<PROBLEM>`), `WithReason` = user-safe message.
- `WithDebug(err.Error())` carries the internal detail to the log only (not to the client); `WithWrap(err)` keeps the chain.
- Unmapped errors fall through to 500 with a generic reason — internals never leak to the client.

### Sentinel → herodot mapping

| Meaning | herodot base | Connect code |
|---|---|---|
| Not found / not owned | `herodot.ErrNotFound` | `not_found` |
| Semantic bad input | `herodot.ErrBadRequest` | `invalid_argument` |
| Bad credentials / no session | `herodot.ErrUnauthorized` | `unauthenticated` |
| Authenticated but not allowed (non-ownership) | `herodot.ErrForbidden` | `permission_denied` |
| Already exists / conflict | `herodot.ErrConflict` | `failed_precondition` |
| Business rule / state / quota | `herodot.ErrConflict` (distinct `WithID`) | `failed_precondition` |
| Upstream / unexpected | `herodot.ErrInternalServerError` (distinct `WithID`, e.g. `UPSTREAM_FAILED`) | `internal` |

The Connect code comes from the herodot base's `GRPCCodeField` (herodot v0.10.8 has only
`ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrInternalServerError`, `ErrBadRequest`,
`ErrUnsupportedMediaType`, `ErrConflict`, `ErrMisconfiguration`, `ErrUpstreamError`).
Do **not** use `ErrUpstreamError`: it has no gRPC code set, so it would become Connect code 0.
The frontend tells same-code errors apart by `ErrorInfo.reason` / the logged `WithID`.

## Sentinel Declaration

```go
// errors.go
var (
	ErrThingNotFound      = errors.New("thing not found")
	ErrThingAlreadyExists = errors.New("thing already exists")
	ErrThingQuotaExceeded = errors.New("thing quota exceeded")
)

func thingNotFound(id uuid.UUID) error {
	return fmt.Errorf("%w: %s", ErrThingNotFound, id)
}
```

- Prefix `Err<Feature>...`; message lowercase, no punctuation.
- Constructor helpers (`thingNotFound(id)`) are for service-level ownership checks: they add the ID for logs while staying `errors.Is`-matchable. Repositories use `helper.MapDBError` instead.

## Shared helpers (target signatures)

`helper.WithTx` and `helper.WithUserID` predate this flow and take herodot callbacks.
New code expects:

```go
func WithTx[T any](ctx context.Context, pool *pgxpool.Pool, queries *db.Queries,
	fn func(q *db.Queries) (T, error)) (T, error)
// begin/commit failures: fmt.Errorf("begin transaction: %w", err) / ("commit transaction: %w", err)
```

Prefer explicit `helper.UserIDFromContext` + `errort.ConnectError` in handlers over `helper.WithUserID`.
Change the helper signatures in the same change that migrates the last caller.

## Anti-Patterns

| Anti-pattern | Why | Correct |
|---|---|---|
| `logger.Error(...); return err` | Logged twice (here + ConnectLogger) | `return err` |
| `if errors.Is(err, ErrX) { return nil, connect.NewError(...) }` in handler | Duplicates `toAppError`, bypasses `errort` | `return nil, toConnectError(ctx, err)` |
| `herodot.ErrNotFound...` returned from service/repo | Data/business layer knows transport | Return sentinel; map in `toAppError` |
| `sqlx.MapPgxErrorHTTP(err, "x")` in repo | Same — herodot in data layer | `fmt.Errorf("x: op: %w", err)` / sentinel |
| `WithReason(err.Error())` | Leaks internals to client | `WithReason("safe message").WithDebug(err.Error())` |
| Returning `Forbidden` for someone else's resource | Leaks existence | `NotFound` |
