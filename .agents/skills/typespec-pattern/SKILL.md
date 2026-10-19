---
name: typespec-pattern
description: Conventions for writing the TypeSpec API contract in openapi/src for this Connect RPC project — feature folders with main.tsp (routes) and model.tsp (models), POST-only operations, `/<package>.<Feature>Service/<Feature><Op>` route paths, and the Id/Base/CreateRequest/UpdateRequest model naming. Use whenever adding or changing an endpoint, RPC, interface, operation, request/response model, or feature folder under openapi/src, even if the user only says "add an API for X", "expose a list endpoint", or "add a field to the request".
---

# TypeSpec Pattern (`openapi/src/`)

The TypeSpec in `openapi/src/` is the source of the API contract. It is compiled to
`openapi/openapi.json`, which is then converted to `.proto` and generated into Go and
TS Connect RPC clients (`task gen:api`, see the `project-pattern` skill). Because the
consumer is **Connect RPC**, not a REST client, the HTTP shape is fixed: every
operation is a POST to a path that encodes `package.Service/Method`. Getting the path
or verb wrong produces a proto service that doesn't match what the Go handler and the
frontend expect, so follow these rules exactly.

## Layout

```
openapi/src/
├── main.tsp                  # @service root; imports every feature folder
├── common/
│   ├── error.tsp             # CommonError.* error models + aliases
│   ├── connect.tsp           # CommonConnect.Empty
│   └── response.tsp          # misc response helpers (rarely needed for RPC)
├── _generated/
│   ├── db.tsp                # GENERATED from drizzle — never edit
│   └── db.augmented.tsp      # hand-written @@format / visibility on DbSchema.*
└── <feature>/
    ├── main.tsp              # the interface: routes / operations
    └── model.tsp             # the models used by those operations
```

- One folder per feature, lowercase, singular, same name as the Go feature package
  and frontend folder (`note`, `auth`).
- `main.tsp` defines routing only; `model.tsp` defines models only. If a feature's
  models grow large you may split them into more files next to `model.tsp`, but
  `model.tsp` stays the entry point that `main.tsp` imports.
- Register a new feature by adding `import "./<feature>";` to `openapi/src/main.tsp`
  (importing a folder loads its `main.tsp`).

## Rule 1: every operation is a POST with a body (or nothing)

Connect unary RPCs are `POST` with a JSON body. So:

- Always `@post`. Never `@get`, `@put`, `@patch`, `@delete`.
- No `@path`, `@query`, or `@header` parameters. Anything the RPC needs goes in
  the body model — including IDs (`NoteModel.Id`).
- Input is a single `@body body: <Model>`, or no parameters at all when the RPC
  takes nothing (e.g. `list()`, `me()`).
- An RPC with nothing to return responds with `CommonConnect.Empty`.

## Rule 2: route path is `/<package>.<Feature>Service/<Feature><Op>`

Connect's route format is `/<Package>.<Service>/<Method>`. In TypeSpec this maps to:

| Part | Value | Casing | Example |
|---|---|---|---|
| Package | the proto package of the service | as given by the user; default `services.v1` | `services.v1` |
| Service | `<InterfaceName>Service` | PascalCase | `NoteService` |
| Method | `<InterfaceName><OperationName>` | PascalCase | `NoteCreate`, `AuthUpdateMe` |

So `interface Note { create(...) }` → `@route("/services.v1.NoteService/NoteCreate")`.

- The package is the only part the user decides. Use `services.v1` (what every
  existing feature uses) unless the user names a different package; if they do,
  use theirs verbatim on every route in that interface.
- The interface name is the PascalCase feature name; the TypeSpec operation name is
  camelCase (`create`, `updateMe`), and the method segment is the interface name
  followed by that operation name in PascalCase (`updateMe` → `AuthUpdateMe`).
  The feature comes first so methods group together in generated code.
- Every route in one interface uses the same service segment.

## Reference feature: `note`

The examples use the same `note` feature as `core-pattern` and `frontend-pattern`
(a signed-in user's private notes: `NoteCreate`, `NoteGet`, `NoteList`, `NoteDelete` on
`NoteService`), so this is the contract those skills' Go handler and React hooks
consume. The `notes` table has `id`, `userId`, `title`, `body`, `createdAt`, generated
as `DbSchema.NoteRow`.

### `openapi/src/note/main.tsp`

```tsp
import "@typespec/http";
import "@typespec/openapi";
import "@typespec/openapi3";

import "../common/error.tsp";
import "../common/connect.tsp";
import "./model.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace CoreServices;

@tag("Note")
interface Note {
  @doc("Create a new note for the current user.")
  @post
  @route("/services.v1.NoteService/NoteCreate")
  create(@body body: NoteModel.CreateRequest):
    | NoteModel.Base
    | CommonError.CreateCommonErrorAuth;

  @doc("Get one of the current user's notes by its ID.")
  @post
  @route("/services.v1.NoteService/NoteGet")
  get(@body body: NoteModel.Id):
    | NoteModel.Base
    | CommonError.GetCommonErrorAuth;

  @doc("List the current user's notes, newest first.")
  @post
  @route("/services.v1.NoteService/NoteList")
  list(): NoteModel.Base[] | CommonError.ListCommonErrorAuth;

  @doc("Delete one of the current user's notes by its ID.")
  @post
  @route("/services.v1.NoteService/NoteDelete")
  delete(@body body: NoteModel.Id):
    | CommonConnect.Empty
    | CommonError.DeleteCommonErrorAuth;
}
```

Notes:
- The interface lives in `namespace CoreServices;` (the `@service` namespace from the
  root `main.tsp`), not in the model namespace.
- `@tag("<Feature>")` matches the interface name.
- Every operation has a `@doc` — one sentence, ends with a period.
- Pick the error alias by operation kind (from `common/error.tsp`):

| Operation kind | Error alias |
|---|---|
| create / other writes that make something | `CommonError.CreateCommonErrorAuth` |
| get one | `CommonError.GetCommonErrorAuth` |
| list | `CommonError.ListCommonErrorAuth` |
| update | `CommonError.UpdateCommonErrorAuth` |
| delete | `CommonError.DeleteCommonErrorAuth` |

  If none fit, add a new alias in `common/error.tsp` following the same naming
  (`<Kind>CommonError<Variant>`) rather than inlining a union in the feature.

### `openapi/src/note/model.tsp`

```tsp
import "@typespec/openapi";
import "@typespec/openapi3";
import "../_generated/db.augmented.tsp";

namespace NoteModel;

model Id {
  @doc("The ID of the note.")
  @example("123e4567-e89b-12d3-a456-426614174000")
  noteId: string;
}

@doc("A user's private note.")
model Base is OmitProperties<DbSchema.NoteRow, "userId">;

@doc("A request to create a new note.")
model CreateRequest is FilterVisibility<
  Base,
  #{ all: #[Lifecycle.Create] },
  "CreateRequest"
>;
```

`userId` is omitted from `Base` because the server takes it from the session and never
returns it. `id` and `createdAt` drop out of `CreateRequest` through their read-only
visibility in `db.tsp`, leaving `title` and `body`.

### Adding update (`NoteUpdate`)

The reference feature has no update. When one is needed, add an `UpdateRequest` that
spreads `Id` and makes the create fields optional, and route it like the others:

```tsp
// model.tsp
@doc("A request to update a note.")
model UpdateRequest {
  ...Id;
  ...FilterVisibility<
    OptionalProperties<CreateRequest>,
    #{ all: #[Lifecycle.Update] },
    "UpdateRequest"
  >;
}

// main.tsp, inside interface Note
  @doc("Update one of the current user's notes by its ID.")
  @post
  @route("/services.v1.NoteService/NoteUpdate")
  update(@body body: NoteModel.UpdateRequest):
    | NoteModel.Base
    | CommonError.UpdateCommonErrorAuth;
```

If some create fields are immutable afterwards, wrap `CreateRequest` in
`OmitProperties<CreateRequest, "...">` before `OptionalProperties`.

## Model naming convention

Models are namespaced, so names inside the namespace are short and generic. Use these
exact names so every feature reads the same and generated proto/TS types are
predictable (`NoteModel.Base`, `NoteModel.CreateRequest`, ...).

| Name | Purpose | Shape |
|---|---|---|
| namespace `<Feature>Model` | Container for all of the feature's models | `NoteModel` |
| `Id` | Identifies one entity; body of get / delete; spread into update | single field `<feature>Id: string` (camelCase, e.g. `noteId`) with `@doc` + uuid `@example` |
| `Base` | The full entity as returned to the client | `is DbSchema.<Entity>Row` (from `_generated`), plus any computed/extra fields; use `OmitProperties` to hide secrets or server-only fields (e.g. `passwordHash`, `userId`) |
| `CreateRequest` | Body of create | `FilterVisibility<OmitProperties<Base, server-owned fields>, #{ all: #[Lifecycle.Create] }, "CreateRequest">` |
| `UpdateRequest` | Body of update | `...Id` + `FilterVisibility<OptionalProperties<...CreateRequest...>, #{ all: #[Lifecycle.Update] }, "UpdateRequest">` |
| `<Op>Request` | Body of any other operation | e.g. `ListRequest` (filters/pagination), `ArchiveRequest`, `LoginRequest` |
| `<Op>Response` | Return type when an op doesn't return `Base` / `Base[]` / `Empty` | e.g. `ListResponse`, `StatsResponse` |

Guidelines behind these:
- **Derive, don't redeclare.** `Base` builds on the generated `DbSchema.*Row`, and
  the request models build on `Base`/`CreateRequest`. That way a column added in
  `migration/` flows to every model after `task gen:api`, and the `@visibility`
  rules in `db.tsp` / `db.augmented.tsp` decide which fields appear in create vs
  update (read-only `id`, `createdAt` drop out automatically).
- The string third argument of `FilterVisibility` is the model's name; keep it equal
  to the model name so the emitted schema name is stable.
- Server-owned fields (owner, foreign keys set from the session, computed ids) are
  removed with `OmitProperties` in `Base` (if never returned) or `CreateRequest` (if
  returned but not client-settable); fields that are immutable after creation are
  additionally omitted in `UpdateRequest`.
- Every field you declare by hand gets `@doc` and an `@example`.
- Every model gets a `@doc` describing it ("A user's private note.", "A request to ...").
- Need a new format or visibility on a DB column? Put an `@@format(...)` /
  `@@visibility(...)` augmentation in `_generated/db.augmented.tsp` — never edit
  `_generated/db.tsp`, it is regenerated.
- Shared models used across features belong in `common/`, in a `Common<Thing>`
  namespace, not in another feature's `model.tsp`.

## Checklist for a new feature

1. Make sure the table exists in `migration/src/schema` and `DbSchema.<Entity>Row` is
   in `_generated/db.tsp` (run `task gen:preview` if not).
2. Create `openapi/src/<feature>/model.tsp` with `namespace <Feature>Model;` and the
   models you need (`Id`, `Base`, `CreateRequest`, `UpdateRequest`, ...).
3. Create `openapi/src/<feature>/main.tsp` with `interface <Feature>` in
   `namespace CoreServices;`, one `@post` + `@route` + `@doc` per operation.
4. Check every route: `/<package>.<Feature>Service/<Feature><Op>`, same service
   segment throughout, no GET/path/query params.
5. Add `import "./<feature>";` to `openapi/src/main.tsp`.
6. Compile: `pnpm --dir openapi exec tsp compile src` (or `task gen:api` to run the
   full pipeline to proto + clients). Fix any diagnostics before moving on to the
   Go handler (`core-pattern`) and frontend (`frontend-pattern`).
