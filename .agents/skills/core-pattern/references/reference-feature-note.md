# Reference Feature: `note` (plain CRUD)

The smallest complete feature. Read this first; read `features/auth` second (it adds cookies,
sessions, a transaction and a `Bundle` provider on top of the same shape). For the full
scaffold with pagination, partial updates and transactions, see
[feature-module-template.md](feature-module-template.md).

`note` = a user's private notes: create, get, list, delete. Every layer is as small as it can
be while still being the real layer.

**Every feature has every layer.** `service.go` stays even when its methods only forward to the
repository — see [Pass-through service methods](../SKILL.md#pass-through-service-methods-are-intentional).

## Order to write the files

Dependencies point down; write top to bottom and each file compiles against the previous ones.

| # | File | One-line job |
|---|---|---|
| 1 | `openapi/` + regenerate | Contract: RPCs `NoteCreate`, `NoteGet`, `NoteList`, `NoteDelete` |
| 2 | `internal/db/queries/note.sql` + `task gen` | sqlc queries; ownership is a `WHERE user_id` |
| 3 | `types.go` | Domain model + service inputs |
| 4 | `errors.go` | Sentinels + `toAppError` / `toConnectError` |
| 5 | `repository.go` | `Repository` + `postgresRepo`, sqlc → domain, `helper.MapDBError` |
| 6 | `service.go` | `Service` + `service` |
| 7 | `mapper.go` | domain → proto |
| 8 | `handler.go` | `connectHandler`: validate, user from ctx, proto → input, error dispatch |
| 9 | `access.go` | Only if some procedure is not "authenticated" (omit for `note`) |
| 10 | `provider.go` | `Deps` + `Provide()` |
| 11 | `internal/app/handler.go` | One `register(...)` line |
| 12 | `service_test.go` | Service against a fake `Repository` |

## 2. `internal/db/queries/note.sql`

```sql
-- name: CreateNote :one
INSERT INTO notes (user_id, title, body) VALUES ($1, $2, $3) RETURNING *;

-- name: GetNote :one
SELECT * FROM notes WHERE id = $1 AND user_id = $2;

-- name: ListNotes :many
SELECT * FROM notes WHERE user_id = $1 ORDER BY created_at DESC;

-- name: DeleteNote :execrows
DELETE FROM notes WHERE id = $1 AND user_id = $2;
```

Ownership lives in the `WHERE`: someone else's note is indistinguishable from a missing one,
so it surfaces as not-found and never leaks existence.

## 3. `types.go`

```go
package note

import (
	"time"

	"github.com/google/uuid"
)

// Note is the domain model for the notes table.
type Note struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Title     string
	Body      string
	CreatedAt time.Time
}

// CreateInput is what the service needs to create a note.
type CreateInput struct {
	Title string
	Body  string
}
```

No proto, sqlc, pgx or herodot imports (depguard enforces it). IDs are `uuid.UUID`,
nullable columns are pointers.

## 4. `errors.go`

```go
package note

import (
	"context"
	"errors"

	errort "github.com/hilmoo/gomox/transport/error"
	"github.com/ory/herodot"
)

var ErrNoteNotFound = errors.New("note not found")

func toConnectError(ctx context.Context, err error) error {
	return errort.ConnectError(ctx, toAppError(err))
}

func toAppError(err error) *herodot.DefaultError {
	switch {
	case errors.Is(err, ErrNoteNotFound):
		return herodot.ErrNotFound.WithID("NOTE_NOT_FOUND").
			WithReason("note not found").WithWrap(err).WithDebug(err.Error())
	}
	return herodot.ErrInternalServerError.WithReason("internal error").WithWrap(err).WithDebug(err.Error())
}
```

One sentinel per outcome the client must tell apart; everything else is the internal-error
fallthrough. Add a `case` per new sentinel.

## 5. `repository.go`

```go
package note

import (
	"context"
	"fmt"

	db "web-app-template-core/internal/db/sqlc"
	"web-app-template-core/internal/helper"
	"web-app-template-core/internal/x/typex"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data-access contract for notes.
type Repository interface {
	Create(ctx context.Context, userID uuid.UUID, in CreateInput) (Note, error)
	Get(ctx context.Context, userID, id uuid.UUID) (Note, error)
	List(ctx context.Context, userID uuid.UUID) ([]Note, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type postgresRepo struct {
	q *db.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepo{q: db.New(pool)}
}

func (r *postgresRepo) Create(ctx context.Context, userID uuid.UUID, in CreateInput) (Note, error) {
	row, err := r.q.CreateNote(ctx, db.CreateNoteParams{UserID: userID, Title: in.Title, Body: in.Body})
	if err != nil {
		return Note{}, helper.MapDBError(err, "note: create", nil, nil)
	}
	return noteFromRow(row), nil
}

func (r *postgresRepo) Get(ctx context.Context, userID, id uuid.UUID) (Note, error) {
	row, err := r.q.GetNote(ctx, db.GetNoteParams{ID: id, UserID: userID})
	if err != nil {
		return Note{}, helper.MapDBError(err, "note: get", ErrNoteNotFound, nil)
	}
	return noteFromRow(row), nil
}

func (r *postgresRepo) List(ctx context.Context, userID uuid.UUID) ([]Note, error) {
	rows, err := r.q.ListNotes(ctx, userID)
	if err != nil {
		return nil, helper.MapDBError(err, "note: list", nil, nil)
	}
	return typex.MapSlice(rows, noteFromRow), nil
}

func (r *postgresRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	rows, err := r.q.DeleteNote(ctx, db.DeleteNoteParams{ID: id, UserID: userID})
	if err != nil {
		return helper.MapDBError(err, "note: delete", nil, nil)
	}
	if rows == 0 {
		return fmt.Errorf("note: delete: %w", ErrNoteNotFound)
	}
	return nil
}

func noteFromRow(r db.Note) Note {
	return Note{ID: r.ID, UserID: r.UserID, Title: r.Title, Body: r.Body, CreatedAt: r.CreatedAt}
}
```

- Every sqlc error goes through `helper.MapDBError`.
- `:execrows` deletes check `rows == 0` themselves.
- No `pool` field: it is only needed for `helper.WithTx`.

## 6. `service.go`

```go
package note

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the note business operations.
type Service interface {
	Create(ctx context.Context, userID uuid.UUID, in CreateInput) (Note, error)
	Get(ctx context.Context, userID, id uuid.UUID) (Note, error)
	List(ctx context.Context, userID uuid.UUID) ([]Note, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, in CreateInput) (Note, error) {
	return s.repo.Create(ctx, userID, in)
}

func (s *service) Get(ctx context.Context, userID, id uuid.UUID) (Note, error) {
	return s.repo.Get(ctx, userID, id)
}

func (s *service) List(ctx context.Context, userID uuid.UUID) ([]Note, error) {
	return s.repo.List(ctx, userID)
}

func (s *service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.Delete(ctx, userID, id)
}
```

Every method is a pass-through here, and that is correct: the service is where rules go when
they exist (quotas, ownership that SQL cannot express, hashing, orchestration of several repo
calls). Do not delete the layer because a feature has no rules yet.

## 7. `mapper.go`

```go
package note

import (
	modelsv1 "proto/models/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoNote(n Note) *modelsv1.Note {
	return &modelsv1.Note{
		Id:        n.ID.String(),
		Title:     n.Title,
		Body:      n.Body,
		CreatedAt: timestamppb.New(n.CreatedAt),
	}
}
```

## 8. `handler.go`

```go
package note

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
	servicesv1connect.UnimplementedNoteServiceHandler

	svc      Service
	validate *validation.Vld
}

var _ servicesv1connect.NoteServiceHandler = (*connectHandler)(nil)

func NewHandler(svc Service, validate *validation.Vld) servicesv1connect.NoteServiceHandler {
	return &connectHandler{svc: svc, validate: validate}
}

func (h *connectHandler) NoteCreate(
	ctx context.Context,
	req *servicesv1.NoteCreateRequest,
) (*servicesv1.NoteCreateResponse, error) {
	userID, errH := helper.UserIDFromContext(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}
	if errH = h.validate.ValidatePayload(req); errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	n, err := h.svc.Create(ctx, userID, CreateInput{Title: req.GetTitle(), Body: req.GetBody()})
	if err != nil {
		return nil, toConnectError(ctx, err)
	}
	return &servicesv1.NoteCreateResponse{Data: toProtoNote(n)}, nil
}

func (h *connectHandler) NoteGet(
	ctx context.Context,
	req *servicesv1.NoteGetRequest,
) (*servicesv1.NoteGetResponse, error) {
	userID, errH := helper.UserIDFromContext(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}
	id, errH := helper.StrToUUIDHTTP(req.GetId())
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	n, err := h.svc.Get(ctx, userID, id)
	if err != nil {
		return nil, toConnectError(ctx, err)
	}
	return &servicesv1.NoteGetResponse{Data: toProtoNote(n)}, nil
}

func (h *connectHandler) NoteList(
	ctx context.Context,
	_ *servicesv1.NoteListRequest,
) (*servicesv1.NoteListResponse, error) {
	userID, errH := helper.UserIDFromContext(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	notes, err := h.svc.List(ctx, userID)
	if err != nil {
		return nil, toConnectError(ctx, err)
	}
	return &servicesv1.NoteListResponse{Data: typex.MapSlice(notes, toProtoNote)}, nil
}

func (h *connectHandler) NoteDelete(
	ctx context.Context,
	req *servicesv1.NoteDeleteRequest,
) (*servicesv1.NoteDeleteResponse, error) {
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
	return &servicesv1.NoteDeleteResponse{}, nil
}
```

Handler shape, always in this order: user from ctx → syntactic validation / ID parsing →
proto → input → `svc` → `toConnectError` once → domain → proto. The handler never calls
`errors.Is`, never touches the repository, sqlc or pgx.

## 9. `access.go`

Omitted: the default is "authenticated", which is what every `note` procedure wants. Add the
file only to list public or service-only procedures (see `features/auth/access.go`).

## 10. `provider.go`

```go
package note

import (
	"proto/services/v1/servicesv1connect"

	"github.com/hilmoo/gomox/transport/validation"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Pool     *pgxpool.Pool
	Validate *validation.Vld
}

func Provide(d Deps) servicesv1connect.NoteServiceHandler {
	return NewHandler(NewService(NewPostgresRepository(d.Pool)), d.Validate)
}
```

`Deps` lists only what this feature needs. Return a `Bundle` (as `auth` does) only when
another part of the app needs the service itself.

## 11. Wiring: `internal/app/handler.go`

```go
register(servicesv1connect.NewNoteServiceHandler(
	note.Provide(note.Deps{Pool: p.Pool, Validate: p.Validate}), opts...))
```

`TestAllServicesWired` fails if this line is forgotten.

## 12. `service_test.go`

```go
package note_test

type fakeRepo struct{ notes map[uuid.UUID]note.Note }

// implement note.Repository over the map, returning note.ErrNoteNotFound for misses,
// then assert on the service: create-then-get round trips, get of another user's id
// is ErrNoteNotFound, delete of a missing id is ErrNoteNotFound.
```

With a pass-through service the interesting tests are the repository's ownership `WHERE`
(integration, behind `//go:build integration`) and the handler's error mapping
(`toAppError` table test). Service unit tests grow with the rules the service gains.

## Checklist before opening a PR

- [ ] Files 3–10 from the table exist (minus `access.go` if nothing is non-default).
- [ ] `task gen && task test && task lint:check && task fmt:check` pass — depguard rejects
      layer-boundary violations.
- [ ] No proto in `types.go`/`service*.go`/`repository*.go`; no sqlc/pgx in `handler*.go`/`service*.go`.
