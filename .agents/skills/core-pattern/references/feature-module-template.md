# Feature Module Template

Copy-paste scaffolding for a new feature in `core/internal/features/<feature>/`.
Replace `<feature>` with the package name (e.g. `note`), `<Feature>` with PascalCase
(e.g. `Note`), and `<FEATURE>` with SCREAMING_SNAKE (e.g. `NOTE`).

Imports used throughout:

```go
servicesv1 "proto/services/v1"
"proto/services/v1/servicesv1connect"
modelsv1 "proto/models/v1"

db "web-app-template-core/internal/db/sqlc"
"web-app-template-core/internal/helper"
```

---

## File: types.go

```go
package <feature>

import (
	"time"

	"github.com/google/uuid"
)

// <Feature> is the domain model for the <feature>s table.
type <Feature> struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Note      *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Create<Feature>Input is what the service needs to create a <feature>.
type Create<Feature>Input struct {
	Name string
	Note *string
}

// Update<Feature>Input is a partial update; nil fields are left unchanged.
type Update<Feature>Input struct {
	Name *string
	Note *string
}
```

**Rules:**

- Domain types carry no proto types, no sqlc types, no tags — they are internal.
- IDs are `uuid.UUID` (sqlc override maps `uuid` → `github.com/google/uuid.UUID`).
- Nullable fields are pointers (`*string`, `*time.Time`) — never `sql.Null*` / `pgtype.*`.
- Input structs are plain Go; the handler maps proto → input. Services never see proto.

---

## File: repository.go

```go
package <feature>

import (
	"context"
	"fmt"

	db "web-app-template-core/internal/db/sqlc"
	"web-app-template-core/internal/helper"
	"web-app-template-core/internal/x/typex"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data-access contract for <feature>s.
type Repository interface {
	List(ctx context.Context, userID uuid.UUID, page helper.Page) ([]<Feature>, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (<Feature>, error)
	Create(ctx context.Context, userID uuid.UUID, in Create<Feature>Input) (<Feature>, error)
	Update(ctx context.Context, id uuid.UUID, in Update<Feature>Input) (<Feature>, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresRepo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepo{pool: pool, q: db.New(pool)}
}

func (r *postgresRepo) List(ctx context.Context, userID uuid.UUID, page helper.Page) ([]<Feature>, int64, error) {
	total, err := r.q.Count<Feature>s(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("<feature>: count: %w", err)
	}
	rows, err := r.q.List<Feature>s(ctx, db.List<Feature>sParams{
		UserID: userID,
		Limit:  int32(page.Size),     //nolint:gosec // bounded by helper.NewPage
		Offset: int32(page.Offset()), //nolint:gosec // bounded by helper.NewPage
	})
	if err != nil {
		return nil, 0, fmt.Errorf("<feature>: list: %w", err)
	}
	return typex.MapSlice(rows, <feature>FromRow), total, nil
}

func (r *postgresRepo) FindByID(ctx context.Context, id uuid.UUID) (<Feature>, error) {
	row, err := r.q.Get<Feature>(ctx, id)
	if err != nil {
		return <Feature>{}, helper.MapDBError(err, "<feature>: find", Err<Feature>NotFound, nil)
	}
	return <feature>FromRow(row), nil
}

func (r *postgresRepo) Create(ctx context.Context, userID uuid.UUID, in Create<Feature>Input) (<Feature>, error) {
	row, err := r.q.Create<Feature>(ctx, db.Create<Feature>Params{
		UserID: userID,
		Name:   in.Name,
		Note:   in.Note,
	})
	if err != nil {
		return <Feature>{}, helper.MapDBError(err, "<feature>: create", nil, Err<Feature>AlreadyExists)
	}
	return <feature>FromRow(row), nil
}

func (r *postgresRepo) Update(ctx context.Context, id uuid.UUID, in Update<Feature>Input) (<Feature>, error) {
	row, err := r.q.Update<Feature>(ctx, db.Update<Feature>Params{
		ID:   id,
		Name: in.Name,
		Note: in.Note,
	})
	if err != nil {
		return <Feature>{}, helper.MapDBError(err, "<feature>: update", Err<Feature>NotFound, Err<Feature>AlreadyExists)
	}
	return <feature>FromRow(row), nil
}

func (r *postgresRepo) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q.Delete<Feature>(ctx, id)
	if err != nil {
		return fmt.Errorf("<feature>: delete: %w", err)
	}
	if rows == 0 {
		return <feature>NotFound(id)
	}
	return nil
}

func <feature>FromRow(r db.<Feature>) <Feature> {
	return <Feature>{
		ID:        r.ID,
		UserID:    r.UserID,
		Name:      r.Name,
		Note:      r.Note,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
```

**Rules:**

- Repository interface speaks domain types only; sqlc types never leave this file.
- Every sqlc error goes through `helper.MapDBError(err, "<feature>: <operation>", notFound, conflict)`: `pgx.ErrNoRows` → `notFound`, unique violation → `conflict`, anything else → `fmt.Errorf("<feature>: <operation>: %w", err)`. Pass `nil` for a sentinel the call cannot produce. Do not hand-roll `errors.Is(err, pgx.ErrNoRows)` or `sqlx.ClassifyPgError` in a repository.
- FK / not-null / check violations (`sqlx.ErrConflict`, `sqlx.ErrInvalid`) are not mapped by the helper; handle one explicitly only when the client needs to see it.
- `:execrows` deletes check `rows == 0` and return the not-found sentinel themselves.
- IDs come from the DB default (`gen_random_uuid()`), so `Create` doesn't pass one.
- `<feature>FromRow()` converts sqlc row → domain; use `typex.MapSlice` for slices.
- Delete uses `:execrows` and checks `rows == 0`.
- When a query returns a sqlc `...Row` struct (column subset/join), add a `fromXRow` for it.

### Transaction Pattern

Multi-write atomicity lives in the repository and uses `helper.WithTx`:

```go
// helper.WithTx (target signature)
func WithTx[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	queries *db.Queries,
	fn func(q *db.Queries) (T, error),
) (T, error)
```

```go
func (r *postgresRepo) CreateWithChildren(ctx context.Context, userID uuid.UUID, in Create<Feature>Input) (<Feature>, error) {
	return helper.WithTx(ctx, r.pool, r.q, func(q *db.Queries) (<Feature>, error) {
		row, err := q.Create<Feature>(ctx, db.Create<Feature>Params{UserID: userID, Name: in.Name})
		if err != nil {
			return <Feature>{}, helper.MapDBError(err, "<feature>: create", nil, Err<Feature>AlreadyExists)
		}
		// ... more q.Method(ctx, ...) calls on the same tx ...
		return <feature>FromRow(row), nil
	})
}
```

- Use the `q` passed to the callback — never `r.q` — inside the transaction.
- `WithTx` rolls back on any error and wraps begin/commit failures; the callback just returns errors.
- Name tx methods after the business operation (`CreateWithSession`), not "Tx".

---

## File: service.go

```go
package <feature>

import (
	"context"
	"log/slog"

	"web-app-template-core/internal/helper"

	"github.com/google/uuid"
)

// Service defines the <feature> business operations.
type Service interface {
	List(ctx context.Context, userID uuid.UUID, page helper.Page) ([]<Feature>, int64, error)
	Get(ctx context.Context, userID, id uuid.UUID) (<Feature>, error)
	Create(ctx context.Context, userID uuid.UUID, in Create<Feature>Input) (<Feature>, error)
	Update(ctx context.Context, userID, id uuid.UUID, in Update<Feature>Input) (<Feature>, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) List(ctx context.Context, userID uuid.UUID, page helper.Page) ([]<Feature>, int64, error) {
	return s.repo.List(ctx, userID, page)
}

func (s *service) Get(ctx context.Context, userID, id uuid.UUID) (<Feature>, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return <Feature>{}, err
	}
	// Ownership check: not-found instead of forbidden (don't leak existence).
	if item.UserID != userID {
		return <Feature>{}, <feature>NotFound(id)
	}
	return item, nil
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, in Create<Feature>Input) (<Feature>, error) {
	return s.repo.Create(ctx, userID, in)
}

func (s *service) Update(ctx context.Context, userID, id uuid.UUID, in Update<Feature>Input) (<Feature>, error) {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return <Feature>{}, err
	}
	return s.repo.Update(ctx, id, in)
}

func (s *service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
```

**Rules:**

- Service methods receive `userID` from the handler (which reads it from the auth interceptor's context).
- Ownership check → `NotFound`, never `Forbidden`.
- Semantic validation (quotas, state transitions, password checks) lives here and returns sentinels.
- No proto, herodot, connect, or sqlc imports.
- The `logger` is only for best-effort side effects (cache/event failures) — never log-and-return.

---

## File: handler.go

```go
package <feature>

import (
	"context"

	servicesv1 "proto/services/v1"
	"proto/services/v1/servicesv1connect"

	"web-app-template-core/internal/helper"
	"web-app-template-core/internal/x/typex"

	errort "github.com/hilmoo/gomox/transport/error"
	"github.com/hilmoo/gomox/transport/validation"
)

type connectHandler struct {
	servicesv1connect.Unimplemented<Feature>ServiceHandler

	svc      Service
	validate *validation.Vld
}

var _ servicesv1connect.<Feature>ServiceHandler = (*connectHandler)(nil)

func NewHandler(svc Service, validate *validation.Vld) servicesv1connect.<Feature>ServiceHandler {
	return &connectHandler{svc: svc, validate: validate}
}

const maxPageSize = 100

func (h *connectHandler) <Feature>List(
	ctx context.Context,
	req *servicesv1.<Feature>ListRequest,
) (*servicesv1.<Feature>ListResponse, error) {
	userID, errH := helper.UserIDFromContext(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	page := helper.NewPage(req.GetPage(), req.GetPerPage(), maxPageSize)
	items, total, err := h.svc.List(ctx, userID, page)
	if err != nil {
		return nil, toConnectError(ctx, err)
	}

	return &servicesv1.<Feature>ListResponse{
		Data:       typex.MapSlice(items, toProto<Feature>),
		Pagination: toProtoPagination(page, total),
	}, nil
}

func (h *connectHandler) <Feature>Get(
	ctx context.Context,
	req *servicesv1.<Feature>GetRequest,
) (*servicesv1.<Feature>GetResponse, error) {
	userID, errH := helper.UserIDFromContext(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}
	id, errH := helper.StrToUUIDHTTP(req.GetId())
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	item, err := h.svc.Get(ctx, userID, id)
	if err != nil {
		return nil, toConnectError(ctx, err)
	}

	return &servicesv1.<Feature>GetResponse{Data: toProto<Feature>(item)}, nil
}

func (h *connectHandler) <Feature>Create(
	ctx context.Context,
	req *servicesv1.<Feature>CreateRequest,
) (*servicesv1.<Feature>CreateResponse, error) {
	userID, errH := helper.UserIDFromContext(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}
	if errH = h.validate.ValidatePayload(req); errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	item, err := h.svc.Create(ctx, userID, Create<Feature>Input{
		Name: req.GetName(),
		Note: req.Note,
	})
	if err != nil {
		return nil, toConnectError(ctx, err)
	}

	return &servicesv1.<Feature>CreateResponse{Data: toProto<Feature>(item)}, nil
}

func (h *connectHandler) <Feature>Delete(
	ctx context.Context,
	req *servicesv1.<Feature>DeleteRequest,
) (*servicesv1.<Feature>DeleteResponse, error) {
	userID, errH := helper.UserIDFromContext(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}
	id, errH := helper.StrToUUIDHTTP(req.GetId())
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	if err := h.svc.Delete(ctx, userID, id); err != nil {
		return nil, toConnectError(ctx, err)
	}

	return &servicesv1.<Feature>DeleteResponse{}, nil
}
```

**Rules:**

- Embed `Unimplemented<Feature>ServiceHandler` and assert the interface with `var _ ... = (*connectHandler)(nil)`.
- Handler fields: `svc Service` + `validate *validation.Vld` (+ transport-only config like `isProd` for cookies). No pool, no queries.
- Order inside a method: user from context → parse IDs → `ValidatePayload` → proto→input → service → domain→proto.
- Two error paths only:
  - syntactic (`helper.UserIDFromContext`, `helper.StrToUUIDHTTP`, `ValidatePayload`) → `errort.ConnectError(ctx, errH)`
  - domain (anything from the service) → `toConnectError(ctx, err)`
- Never `errors.Is` in a handler. Never call the repository from a handler.
- Read optional proto fields as pointers (`req.Note`), required ones with getters (`req.GetName()`).
- Transport concerns (cookies, headers via `connect.CallInfoForHandlerContext`) stay in the handler.

---

## File: mapper.go

```go
package <feature>

import (
	modelsv1 "proto/models/v1"

	"web-app-template-core/internal/helper"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProto<Feature>(t <Feature>) *modelsv1.<Feature> {
	return &modelsv1.<Feature>{
		Id:        t.ID.String(),
		Name:      t.Name,
		Note:      t.Note,
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}
}

// Shared across features once a second list endpoint exists — then move it to internal/helper.
func toProtoPagination(p helper.Page, total int64) *modelsv1.Pagination {
	return &modelsv1.Pagination{
		Total:   total,
		Page:    int32(p.Number), //nolint:gosec // bounded by helper.NewPage
		PerPage: int32(p.Size),   //nolint:gosec // bounded by helper.NewPage
	}
}
```

- One `toProto<Type>` per domain type; `uuid` → `.String()`, `time.Time` → `timestamppb.New`.
- Never expose secrets (password hashes, token hashes) — they don't exist on the proto model.

---

## File: errors.go

```go
package <feature>

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	errort "github.com/hilmoo/gomox/transport/error"
	"github.com/ory/herodot"
)

var (
	Err<Feature>NotFound      = errors.New("<feature> not found")
	Err<Feature>AlreadyExists = errors.New("<feature> already exists")
)

func <feature>NotFound(id uuid.UUID) error {
	return fmt.Errorf("%w: %s", Err<Feature>NotFound, id)
}

func toConnectError(ctx context.Context, err error) error {
	return errort.ConnectError(ctx, toAppError(err))
}

func toAppError(err error) *herodot.DefaultError {
	switch {
	case errors.Is(err, Err<Feature>NotFound):
		return herodot.ErrNotFound.WithID("<FEATURE>_NOT_FOUND").WithReason("<feature> not found").WithWrap(err).WithDebug(err.Error())
	case errors.Is(err, Err<Feature>AlreadyExists):
		return herodot.ErrConflict.WithID("<FEATURE>_ALREADY_EXISTS").WithReason("<feature> already exists").WithWrap(err).WithDebug(err.Error())
	}
	return herodot.ErrInternalServerError.WithReason("internal error").WithWrap(err).WithDebug(err.Error())
}
```

See [error-flow.md](error-flow.md) for the full mapping table and rules.

---

## File: provider.go

```go
package <feature>

import (
	"log/slog"

	"proto/services/v1/servicesv1connect"

	"github.com/hilmoo/gomox/transport/validation"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Pool     *pgxpool.Pool
	Validate *validation.Vld
	Logger   *slog.Logger
}

func Provide(d Deps) servicesv1connect.<Feature>ServiceHandler {
	repo := NewPostgresRepository(d.Pool)
	svc := NewService(repo, d.Logger)
	return NewHandler(svc, d.Validate)
}
```

### When other code needs the service (e.g. an interceptor resolver)

```go
type Bundle struct {
	Handler servicesv1connect.<Feature>ServiceHandler
	Svc     Service
}

func Provide(d Deps) Bundle {
	repo := NewPostgresRepository(d.Pool)
	svc := NewService(repo, d.Logger)
	return Bundle{Handler: NewHandler(svc, d.Validate), Svc: svc}
}
```

### With Cache + Events (see cache-event-patterns.md)

```go
type Deps struct {
	Pool     *pgxpool.Pool
	Validate *validation.Vld
	Logger   *slog.Logger
	Cache    *cache.Cache // may be nil
	EventBus *event.Bus   // may be nil
}

func Provide(d Deps) servicesv1connect.<Feature>ServiceHandler {
	eb := event.Noop()
	if d.EventBus != nil {
		eb = d.EventBus
	}

	repo := NewPostgresRepository(d.Pool)
	svc := NewService(repo, d.Cache, eb, d.Logger)
	registerListeners(d.EventBus, d.Logger)
	return NewHandler(svc, d.Validate)
}
```

**Rules:**

- `Deps` lists only what this feature uses — config values are passed as fields (`Secret string`, `IsProd bool`), not the whole `config.Config`.
- `Provide()` is the single wiring point: repo → service → handler.
- Returns the generated handler interface (or a `Bundle`); only service/repo are feature-defined interfaces.
- `internal/app` builds `Deps` from `helper.HandlerParams` (the infra container) — features never receive `HandlerParams`.

---

## File: access.go

```go
package <feature>

import (
	"proto/services/v1/servicesv1connect"

	"web-app-template-core/internal/helper"
)

// Access lists the <feature> procedures that differ from the default (authenticated).
//
//nolint:gochecknoglobals // static policy table
var Access = helper.AccessPolicy{
	servicesv1connect.<Feature>ServicePublicThingProcedure: helper.AccessPublic,
}
```

- Only list exceptions. Unlisted procedures require authentication.
- Omit the file if every procedure is authenticated (and don't pass it in `RegisterConnect`).

---

## SQL Queries: internal/db/queries/<feature>.sql

```sql
-- name: Count<Feature>s :one
SELECT count(*)
FROM <feature>s
WHERE "user_id" = $1;

-- name: List<Feature>s :many
SELECT "id", "user_id", "name", "note", "created_at", "updated_at"
FROM <feature>s
WHERE "user_id" = $1
ORDER BY "created_at" DESC
LIMIT $2 OFFSET $3;

-- name: Get<Feature> :one
SELECT "id", "user_id", "name", "note", "created_at", "updated_at"
FROM <feature>s
WHERE "id" = $1;

-- name: Create<Feature> :one
INSERT INTO <feature>s ("user_id", "name", "note")
    VALUES ($1, $2, $3)
RETURNING "id", "user_id", "name", "note", "created_at", "updated_at";

-- name: Update<Feature> :one
UPDATE <feature>s
SET "name" = COALESCE(sqlc.narg('name'), "name"),
    "note" = COALESCE(sqlc.narg('note'), "note"),
    "updated_at" = now()
WHERE "id" = $1
RETURNING "id", "user_id", "name", "note", "created_at", "updated_at";

-- name: Delete<Feature> :execrows
DELETE FROM <feature>s
WHERE "id" = $1;
```

**Rules:**

- Names are `VerbEntity`: `Get<Feature>`, `List<Feature>s`, `Count<Feature>s`, `Create<Feature>`, `Update<Feature>`, `Delete<Feature>`. Qualifiers go after: `Get<Feature>ByEmail`.
- List explicit columns — no `SELECT *`. When all table columns are selected, sqlc still returns the table model `db.<Feature>`.
- `:one` single row, `:many` lists, `:exec` fire-and-forget, `:execrows` deletes (check 0 rows).
- Always `RETURNING` the full column list on `:one` insert/update.
- Partial updates use `COALESCE(sqlc.narg('col'), "col")` so nil pointers keep the old value.
- Lists are offset-paginated: `LIMIT $n OFFSET $m` plus a `Count...` query; default sort `ORDER BY "created_at" DESC`.
- Quote identifiers (`"user_id"`), as the formatter (`pg_format`) keeps them.
- Run `task gen` in `core/` (formats SQL, wipes and regenerates `internal/db/sqlc`). Never edit `internal/db/sqlc` by hand.

---

## Pagination helper: internal/helper/page.go (create on first use)

```go
package helper

const defaultPageSize = 20

// Page is a 1-based offset page.
type Page struct {
	Number int
	Size   int
}

func (p Page) Offset() int { return (p.Number - 1) * p.Size }

// NewPage clamps the requested page/size to sane bounds.
func NewPage(number, size int32, maxSize int) Page {
	p := Page{Number: int(number), Size: int(size)}
	if p.Number < 1 {
		p.Number = 1
	}
	if p.Size < 1 {
		p.Size = defaultPageSize
	}
	if p.Size > maxSize {
		p.Size = maxSize
	}
	return p
}
```

---

## Tests

- Service: unit tests with a fake `Repository` (hand-written struct in `service_test.go`), table-driven, `t.Parallel()`.
- Helpers/mappers: same-package tests with `//nolint:testpackage // exercises unexported helpers`.
- Repository: integration tests behind `//go:build integration` (need Postgres), run with `go test -tags=integration ./...`.
- Wiring: `internal/app/handler_test.go` `TestAllServicesWired` must keep passing.
