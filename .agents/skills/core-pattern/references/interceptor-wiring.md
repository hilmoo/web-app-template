# Mux, Interceptors & Wiring

All composition happens in `core/internal/app`. It is the only package that imports
every feature, and the only place features are connected to each other.

## Startup Flow

```
cmd/main.go → internal/cmd.RootCmd()          # load config, build slog logger, urfave/cli
  serve → app.Main(cfg, logger)
    setupDependencies()  (deps.go)            # pgxpool, validator, sqlc Queries → helper.HandlerParams
    startServer()        (app.go)             # http.Server, HTTP/1.1 + h2c
    awaitShutdown()                            # SIGINT/SIGTERM → graceful Shutdown(10s)
    cleanup()                                  # close infra (event bus first, then Redis, then DB)
```

`helper.HandlerParams` is the **infra container** (Validate, Pool, Queries, Config, Logger).
Only `internal/app` (and `helper` itself) touch it; features get their own `Deps`.

## Mux Layout (`app/handler.go`)

```go
mux.Handle("/api/", http.StripPrefix("/api", api)) // Connect services + gRPC reflection
mux.Handle("GET /api/health", healthMux)           // 204, used by `healthcheck` command
mux.Handle("/", web)                               // embedded SPA (prod) / Vite proxy (-tags dev)
```

New RPC services go on `api` via `RegisterConnect`. Plain HTTP endpoints are rare —
follow `features/health` (`NewHTTP().Register(mux)`) and mount them explicitly before `/`.

## Interceptor Chain (ordered)

```go
opts := []connect.HandlerOption{
	connect.WithInterceptors(
		mlog.ConnectLogger(arg.Logger),      // 1. request log + error log (once per call)
		helper.NewAuthInterceptor(           // 2. authenticate + enforce AccessPolicy
			arg,
			authBundle.Svc,                  //    consumer-side helper.SessionResolver
			auth.Access,                     //    per-feature policy tables...
			thing.Access,
			AccessInternalOnly,
		),
	),
}
if arg.Config.IsProd {
	opts = append(opts, connect.WithRecover(...)) // panic → CodeInternal, logged with procedure
}
```

Order matters: the logger wraps auth so rejected calls are logged too.

## Registering a Feature

```go
func RegisterConnect(arg helper.HandlerParams, mux *http.ServeMux) []string {
	authBundle := auth.Provide(auth.Deps{
		Pool:     arg.Pool,
		Validate: arg.Validate,
		Logger:   arg.Logger,
		Secret:   arg.Config.Secret,
		IsProd:   arg.Config.IsProd,
	})
	thingHandler := thing.Provide(thing.Deps{
		Pool:     arg.Pool,
		Validate: arg.Validate,
		Logger:   arg.Logger,
	})

	opts := ... // interceptor chain above, with thing.Access added

	register(servicesv1connect.NewAuthServiceHandler(authBundle.Handler, opts...))
	register(servicesv1connect.NewThingServiceHandler(thingHandler, opts...))
	// reflection is built from the registered names
}
```

**Rules:**

- Every proto service in `services.v1` must be registered — `TestAllServicesWired` enforces it.
- Every procedure in `AccessInternalOnly` must belong to a registered service (same test).
- Build `Deps` from `arg`; pass config **values**, not `arg.Config`.
- Same `opts` for every service — don't give one service a different interceptor chain.

## Access Policy

```go
// helper
AccessAuthenticated // default: user session cookie OR service secret
AccessPublic        // no credentials; stale credentials are ignored
AccessService       // service secret only; user sessions rejected (PermissionDenied)
```

| Procedure needs… | Declare in |
|---|---|
| Logged-in user (default) | nothing |
| No auth (login, register, public read) | `<feature>/access.go` → `helper.AccessPublic` |
| Service-to-service only | `app.AccessInternalOnly` (or feature `access.go`) → `helper.AccessService` |

New procedures are protected by default; only list exceptions.

### Credentials

- **User:** `session` cookie (`helper.SessionCookieName`), resolved via `helper.SessionResolver`.
- **Service:** `Authorization: Bearer <SERVICE_SECRET>` (constant-time compare); may add `X-User-Id: <uuid>` to act as a user.

## Reading the Caller in Handlers

```go
userID, errH := helper.UserIDFromContext(ctx) // session user, or X-User-Id for service callers
if errH != nil {
	return nil, errort.ConnectError(ctx, errH)
}

principal, ok := helper.PrincipalFromContext(ctx) // when you need to know it's a service call
```

- Read the caller only from context — never re-parse cookies/headers for identity in a feature.
- Exception: transport work on the raw token (logout deletes the session) uses `helper.SessionTokenFromHeader`.
- Request metadata (peer IP, User-Agent, response cookies) comes from `connect.CallInfoForHandlerContext(ctx)` in the handler, passed to the service as plain values.

## Consumer-Side Interface Pattern

When infra (interceptor) or one feature needs something from another feature, the
**consumer** declares a narrow interface; the provider satisfies it implicitly; `internal/app` wires them.

```go
// internal/helper/auth.go — consumer declares
type SessionResolver interface {
	UserIDByToken(ctx context.Context, token string) (uuid.UUID, error)
}
```

```go
// internal/features/auth/service.go — provider satisfies implicitly
func (s *service) UserIDByToken(ctx context.Context, token string) (uuid.UUID, error) { ... }
```

```go
// internal/app/handler.go — the only place that imports both
helper.NewAuthInterceptor(arg, authBundle.Svc, ...)
```

Feature-to-feature works the same way: feature B declares `type userLookup interface { ... }`
in its own package and takes it in `Deps`; `internal/app` passes feature A's `Bundle.Svc`.

**Rules:**

- Interface lives in the consumer package, with only the methods it calls.
- Features never import other features; `helper` never imports features.
- Cross-feature calls get a timeout if they hit the network (`context.WithTimeout`).

## Anti-Patterns

| Anti-pattern | Fix |
|---|---|
| Feature handler parses the session cookie to find the user | `helper.UserIDFromContext(ctx)` |
| Feature imports another feature | Consumer-side interface in `Deps`, wired in `internal/app` |
| `helper` imports a feature | Declare the interface in `helper`, pass the implementation from `app` |
| Listing an authenticated procedure in `Access` "to be explicit" | Only list exceptions |
| Feature receives `helper.HandlerParams` / full `config.Config` | Feature `Deps` with only the fields used |
| New service missing from `RegisterConnect` | Register it; `TestAllServicesWired` will fail otherwise |
