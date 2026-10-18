# Response Shape

Responses are protobuf messages generated from `openapi/` → `proto/`. There is no JSON
envelope written by hand: Connect serializes the message (JSON or binary) and errors.

## Success

Every response message wraps its payload in `data`:

```protobuf
message ThingGetResponse {
  models.v1.Thing data = 1;
}
```

```go
return &servicesv1.ThingGetResponse{Data: toProtoThing(item)}, nil
```

| Operation | Response |
|---|---|
| Get / Create / Update | `{ data: <Model> }` |
| List | `{ data: [<Model>], pagination: { total, page, per_page } }` |
| Delete / Logout / fire-and-forget | empty message `{}` |

Create returns the created entity (Connect has no 201 — success is success).

## Pagination

List requests carry `page` (1-based) and `per_page`; the handler clamps them with
`helper.NewPage(req.GetPage(), req.GetPerPage(), maxPageSize)`, and the response
includes a shared `models.v1.Pagination` (define it once in `openapi/` as a common schema).

```go
return &servicesv1.ThingListResponse{
	Data:       typex.MapSlice(items, toProtoThing),
	Pagination: toProtoPagination(page, total),
}, nil
```

## Errors

`errort.ConnectError(ctx, errH)` produces a Connect error:

```json
{
  "code": "not_found",
  "message": "The requested resource could not be found",
  "details": [
    { "type": "google.rpc.ErrorInfo", "value": "...", "debug": { "reason": "thing not found" } }
  ]
}
```

| Part | Source |
|---|---|
| `code` | `herodot.DefaultError.GRPCStatus()` code (from the herodot base: `ErrNotFound` → `not_found`) |
| `message` | herodot's generic error text for that status |
| `ErrorInfo.reason` | `WithReason(...)` — the user-safe message the frontend shows |
| `BadRequest.field_violations` | herodot `Details` — filled by `ValidatePayload` for invalid fields |
| (log only) | `WithID`, `WithDebug`, wrapped cause — attached to the request log by `errort` |

## Handler Decision Tree

```
Caller/ID/payload problem (UserIDFromContext, StrToUUIDHTTP, ValidatePayload)?
  ├─ YES → return nil, errort.ConnectError(ctx, errH)
  └─ NO  → service returned error?
           ├─ YES → return nil, toConnectError(ctx, err)
           └─ NO  → return &...Response{Data: toProto...(x)} (+ Pagination for lists)
```

## Rules

- Never hand-build response JSON or `connect.NewError` in features.
- Never put internal error text in `WithReason`; use `WithDebug`.
- Proto field names come from the OpenAPI schema — change the contract in `openapi/`, regenerate, never edit generated proto Go.
