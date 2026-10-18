# Cache-Aside & Event Bus Patterns with NATS (future)

Neither `internal/cache` nor `internal/event` exists yet. When a feature first needs
caching or pub/sub, create the package with the API below and follow these patterns —
don't invent a different shape. Both are **optional** dependencies: every feature must
keep working when they are nil.

Both packages sit on a single shared NATS connection (`*nats.Conn`) created in
`internal/app`:

- **Cache** → JetStream **Key-Value** store (`jetstream.KeyValue`).
- **Events** → NATS **subjects** (core NATS for fire-and-forget, JetStream for acked publish).

Libraries: `github.com/nats-io/nats.go` and `github.com/nats-io/nats.go/jetstream`
(use the new `jetstream` package, not the legacy `nc.JetStream()` API).

## Cache Layer (`internal/cache/`)

Typed, generic KV operations with JSON serialization and singleflight deduplication.

### NATS KV constraints (read first)

- **Keys may not contain `:`.** Allowed characters are `A-Za-z0-9`, `-`, `_`, `/`, `=`, `.`.
  Use `.` as the separator: `thing.<uuid>`, never `thing:<uuid>`.
- **Namespace = bucket.** Use one bucket per app (`webapp`); feature prefixes stay in the key.
- **No native batch operations.** Batch reads are concurrent `Get`s (bounded `errgroup`).
  Batch writes are concurrent `Put`s. There is no pipelining.
- **TTL** is set on the bucket (`KeyValueConfig.TTL`) as the default. Per-key TTL
  (`jetstream.KeyTTL(d)`) requires the bucket to be created with `LimitMarkerTTL`.
- A miss is `jetstream.ErrKeyNotFound`; the cache package maps it to `(zero, nil)`.

### API

```go
kv, err := cache.NewBucket(ctx, js, cache.BucketConfig{
	Name:            "webapp",
	TTL:             time.Hour,       // default TTL for the bucket
	LimitMarkerTTL:  time.Second,     // enables per-key TTL
	History:         1,
	Storage:         jetstream.MemoryStorage, // or FileStorage
})

c := cache.New(kv)

cache.Set(ctx, c, "user."+id, user, time.Hour)  // ttl <= 0 → bucket default
u, err := cache.Get[User](ctx, c, "user."+id)   // (zero, nil) on miss
err := cache.Delete(ctx, c, "user."+id)         // invalidate (idempotent)
cache.Forget(c, "user."+id)                     // clear singleflight entry
cache.MSetWithTTL(ctx, c, items, time.Hour)     // concurrent batch write
```

`cache.NewBucket` uses `CreateOrUpdateKeyValue` so startup is idempotent.

### Cache-aside: FetchOne

```go
const (
	cacheKeyPrefix = "thing."
	cacheTTL       = 15 * time.Minute
)

func cacheKey(id uuid.UUID) string { return cacheKeyPrefix + id.String() }

func (s *service) Get(ctx context.Context, userID, id uuid.UUID) (Thing, error) {
	item, err := s.find(ctx, id)
	if err != nil {
		return Thing{}, err
	}
	if item.UserID != userID {
		return Thing{}, thingNotFound(id)
	}
	return item, nil
}

func (s *service) find(ctx context.Context, id uuid.UUID) (Thing, error) {
	if s.cache == nil {
		return s.repo.FindByID(ctx, id)
	}
	return cache.FetchOne(ctx, s.cache, cacheKey(id), func(ctx context.Context) (Thing, error) {
		return s.repo.FindByID(ctx, id)
	}, cacheTTL)
}
```

Flow: `kv.Get` → `ErrKeyNotFound` → singleflight-guarded fetcher → `kv.Put` → return.
Ownership checks run **after** the cache, on the domain value — never cache per-user
"allowed" results.

### Cache-aside: FetchWithSingleflight (batch)

```go
func (s *service) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]Thing, error) {
	if s.cache == nil {
		return s.repo.ListByIDs(ctx, ids)
	}
	return cache.FetchWithSingleflight(
		ctx, s.cache, ids,
		cacheKey,                                // func(id) string
		func(t Thing) uuid.UUID { return t.ID }, // extract ID
		"thing.list-by-ids",                     // singleflight group key
		s.repo.ListByIDs,                        // batch fetcher
		cacheTTL,
	)
}
```

Flow: build keys → concurrent `kv.Get` (bounded, e.g. `errgroup.SetLimit(16)`) → collect
misses → singleflight-guarded DB fetch → concurrent `kv.Put` of misses → merge.
For very large ID lists, skip the cache and go straight to the DB rather than fanning
out hundreds of KV round-trips.

### Invalidation on write

```go
func (s *service) Update(ctx context.Context, userID, id uuid.UUID, in UpdateThingInput) (Thing, error) {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return Thing{}, err
	}
	item, err := s.repo.Update(ctx, id, in)
	if err != nil {
		return Thing{}, err
	}
	s.invalidate(ctx, id)
	return item, nil
}

func (s *service) invalidate(ctx context.Context, id uuid.UUID) {
	if s.cache == nil {
		return
	}
	if err := cache.Delete(ctx, s.cache, cacheKey(id)); err != nil {
		s.logger.WarnContext(ctx, "cache delete failed", "key", cacheKey(id), "error", err)
	}
	cache.Forget(s.cache, cacheKey(id))
}
```

`cache.Delete` writes a delete marker (`kv.Delete`), not a purge; with `History: 1` the
old value is not readable afterwards. Use `kv.Purge` only if you need to drop history.

**Rules:**

- Always `Delete` + `Forget` on create/update/delete affecting a cached key.
- Invalidate **after** the write succeeds (after commit when in a transaction).
- Cache failures are best-effort: `WarnContext` and continue — never fail the request.
- Cache domain types only (never proto, never sqlc rows).
- Keys use `.` separators only (no `:`).
- `Deps.Cache` may be nil → DB-only path.
- Multi-instance note: KV is shared across instances, so invalidation is global. An
  in-process singleflight only dedupes within one instance — that is fine.

---

## Event Bus (`internal/event/`)

Pub/sub for side effects and cross-feature reactions, backed by the NATS server already
used for the cache. Unlike an in-process bus, events are **serialized (JSON)** and can
reach subscribers in other instances.

### Delivery semantics

| Method | Transport | Guarantee | Use for |
|---|---|---|---|
| `PublishAsync` | core NATS `nc.Publish` | at-most-once, non-blocking | default side effects (logs, stats, cache warm) |
| `Publish` | JetStream `js.Publish` (waits for ack) | persisted, at-least-once | events that must not be lost |

JetStream publishing requires a stream covering the subject space. `event.New` creates/updates
one stream (default name `EVENTS`, subjects `events.>`). Handlers for JetStream events may
see duplicates — make them idempotent (`Event.ID` is sent as the `Nats-Msg-Id` header, so
the server also dedupes within the stream's duplicate window).

### Types

```go
// Publisher is what services depend on (not *Bus).
type Publisher interface {
	Publish(ctx context.Context, e Event) error      // durable: JetStream publish, waits for ack
	PublishAsync(ctx context.Context, e Event) error // fire-and-forget: core NATS publish
}

func Noop() Publisher // discards everything

type Event struct {
	ID      string          // generated UUID
	Topic   string
	Payload json.RawMessage // JSON-encoded domain value
}

func NewEvent(topic string, payload any) (Event, error) // marshals payload to JSON
func PayloadAs[T any](e Event) (T, error)               // unmarshals Payload into T

func New(nc *nats.Conn, logger *slog.Logger, opts ...Option) (*Bus, error)
func (b *Bus) Subscribe(topic, subscriberID string, fn func(context.Context, Event) error) error
func (b *Bus) Close() error // nc.Drain-style: stop subscriptions, finish in-flight handlers
```

Wire format: subject `events.<topic>`, headers `Nats-Msg-Id` (= `Event.ID`) and
`Event-Topic`, body is the JSON payload.

### Topics (`<feature>/events.go`)

```go
const (
	TopicThingCreated = "thing.created"
	TopicThingUpdated = "thing.updated"
	TopicThingDeleted = "thing.deleted"
)
```

Format `<feature>.<action>`, lowercase, dot-separated (dots are NATS subject tokens, so
`events.thing.*` and `events.>` wildcards work). Only declare topics the feature actually
publishes. Never put `*`, `>`, spaces, or `:` in a topic.

### Publishing (service)

```go
func (s *service) Create(ctx context.Context, userID uuid.UUID, in CreateThingInput) (Thing, error) {
	item, err := s.repo.Create(ctx, userID, in)
	if err != nil {
		return Thing{}, err
	}
	s.publish(ctx, TopicThingCreated, item)
	return item, nil
}

func (s *service) publish(ctx context.Context, topic string, payload any) {
	e, err := event.NewEvent(topic, payload)
	if err == nil {
		err = s.events.PublishAsync(ctx, e)
	}
	if err != nil {
		s.logger.WarnContext(ctx, "publish event failed", "topic", topic, "error", err)
	}
}
```

**Rules:**

- `PublishAsync` by default — handlers must not block the RPC. Use `Publish` only when the
  event must survive a subscriber or server restart.
- Publish **after** the write (after commit), never inside a transaction callback.
- Publish failure is a warning, not an error.
- Payload is the domain value (or its ID for deletes) — never proto. It must be
  JSON-serializable, and its JSON shape is now a **cross-instance contract**: add fields,
  don't rename or remove them.

### Subscribing (`<feature>/listeners.go`)

```go
func registerListeners(bus *event.Bus, logger *slog.Logger) error {
	if bus == nil {
		return nil
	}
	return bus.Subscribe(TopicThingCreated, "thing.on-created", func(ctx context.Context, e event.Event) error {
		created, err := event.PayloadAs[Thing](e)
		if err != nil {
			return err
		}
		logger.InfoContext(ctx, "thing created", "id", created.ID)
		return nil
	})
}
```

- Guard `bus == nil`.
- Subscriber ID `<feature>.<handler-name>`. It is used as the **NATS queue group**, so when
  the app runs N instances, exactly one instance handles each event per subscriber ID.
  Two different subscriber IDs on the same topic each receive every event.
- For JetStream-published events, the subscriber ID is also the **durable consumer name**
  (dots replaced with `-`), so the handler resumes after restarts.
- Return errors (the bus logs them; JetStream events are `Nak`ed with backoff, core events
  are dropped); handlers are panic-isolated by the bus.

### Wiring

Service gets `event.Publisher` (interface, `event.Noop()` when nil); listeners get `*event.Bus`.
See the "With Cache + Events" `Provide()` in feature-module-template.md.

Cross-feature subscriptions (A publishes, B reacts) are registered in `internal/app`,
not inside either feature, so neither imports the other:

```go
// internal/app
err := eventBus.Subscribe(thing.TopicThingCreated, "stats.on-thing-created", func(ctx context.Context, e event.Event) error {
	t, err := event.PayloadAs[thing.Thing](e)
	if err != nil {
		return err
	}
	return statsBundle.Svc.TrackCreation(ctx, t.ID)
})
```

Because payloads cross a serialization boundary, the subscriber decodes into its own view
of the type; for loose coupling, define a small local struct with only the fields it needs
instead of importing the publisher's domain type.

### Lifecycle (`app/deps.go`)

```go
nc, err := nats.Connect(cfg.NATSURL,
	nats.Name("webapp"),
	nats.MaxReconnects(-1),
	nats.ReconnectWait(2*time.Second),
)
if err != nil { /* NATS is optional: log, leave nc/js nil, continue */ }

js, _ := jetstream.New(nc)
kv, _ := cache.NewBucket(ctx, js, cacheCfg)
appCache := cache.New(kv)
eventBus, _ := event.New(nc, logger)

cleanup := func() {
	_ = eventBus.Close() // FIRST: drain subscriptions/async handlers while NATS and DB are still open
	_ = nc.Drain()       // flush pending publishes, then close the connection
	dbPool.Close()
}
```

If `nats.Connect` fails or `cfg.NATSURL` is empty, pass `nil` for the cache and
`event.Noop()` for the publisher — features must keep working.