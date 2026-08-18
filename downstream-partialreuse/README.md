# downstream-partialreuse

A downstream BFF that reuses most of upstream's code, but demonstrates two
ways to diverge from it: decorating a service, and attaching a brand new
handler upstream has no concept of.

## Run

```
PORT=8083 go run ./cmd/api
curl -H "Authorization: Bearer secret-token" -X POST localhost:8083/api/sandbox/ -d '{"name":"my-box"}'
curl -H "Authorization: Bearer secret-token" localhost:8083/api/audit/sandboxes
curl -H "Authorization: Bearer secret-token" localhost:8083/api/sandbox-uptime/sandbox-1
curl -H "X-Downstream-Key: downstream-secret" localhost:8083/debug/ping
```

Create response — compare to upstream's `{"name":"my-box","status":"running"}`:
`{"name":"my-box (downstream-managed)","status":"provisioning"}`.

`/api/audit/sandboxes` returns `{"total":N,"byStatus":{...}}` — 404 on
upstream, since it doesn't exist there.

`/debug/ping` returns `pong`, but only with `X-Downstream-Key:
downstream-secret` — upstream's `Authorization: Bearer secret-token`
does **not** work here, on purpose (see below).

## How it works

`pkg/services/sandbox.go` embeds `upstreamservices.SandboxService` and
overrides only `CreateSandbox`:

```go
type Service struct {
    upstreamservices.SandboxService // embedded — inherits Get/List/Delete as-is
    logger *slog.Logger
}

func (s *Service) CreateSandbox(ctx context.Context, name string) (*upstreamservices.Sandbox, error) {
    sb, err := s.SandboxService.CreateSandbox(ctx, name) // delegate to upstream
    // ...tag it, change status...
    return sb, err
}
```

Note this repo's own `pkg/services` package is unrelated to (and imported
alongside, under a different alias, in) upstream's `pkg/services` — same
package name, different module. See "Package naming" below.

`GetSandbox`, `ListSandboxes`, and `DeleteSandbox` are never redefined —
calls fall straight through to the embedded upstream implementation. Try
`GET /api/sandbox/sandbox-1` after creating one: it returns the decorated
object, proving the override's effect persists through an inherited method
that downstream never touched.

`cmd/api/main.go` reuses upstream's `server.NewServer`, `Services` struct,
and `SandboxHandler` completely unmodified — it only swaps in the
decorated service:

```go
base := upstreamservices.NewDefaultSandboxService()
decorated := downstreamservices.NewService(base)
svcs := server.Services{Gateway: upstreamservices.NewDefaultGatewayService(), Sandbox: decorated}
```

Gateway isn't customized at all here, same as `../downstream-reuse`.

## Extending vs. overriding via embedding

`pkg/services/sandbox_extended.go` shows the other side of interface
embedding: adding a capability instead of changing one.

```go
type SandboxUptimeService interface {
    upstreamservices.SandboxService
    SandboxUptime(ctx context.Context, id string) (time.Duration, error)
}

type ExtendedService struct {
    upstreamservices.SandboxService // embedded -- all 4 methods inherited untouched
}

func (s *ExtendedService) SandboxUptime(ctx context.Context, id string) (time.Duration, error) {
    sb, err := s.GetSandbox(ctx, id) // inherited call, not overridden
    // ...
}
```

Unlike `Service` above, `ExtendedService` implements none of
`SandboxService`'s methods itself -- it only adds `SandboxUptime`, built
out of the embedded interface's existing `GetSandbox`.

`SandboxUptime` has no upstream equivalent, so there's no upstream
interface to reuse for it -- `SandboxUptimeService` is downstream's own
interface, embedding `upstreamservices.SandboxService` so it's a superset
of (not a parallel contract to) the original. `pkg/handlers.UptimeHandler`
depends on `SandboxUptimeService`, never on the concrete `*ExtendedService`
-- same "accept interfaces" discipline as upstream's own handlers, just
with an interface downstream had to define itself because the capability
is downstream's:

```go
extendedSandboxSvc := downstreamservices.NewExtendedService(decoratedSandboxSvc)
uptimeHandler := handlers.NewUptimeHandler(extendedSandboxSvc)
```

`extendedSandboxSvc` wraps `decoratedSandboxSvc` (itself already wrapping
upstream's base service), so `GET /api/sandbox-uptime/{id}` reflects
downstream's `CreateSandbox` decoration too -- three layers of embedding
composing cleanly because each only depends on the interface below it.

## New handlers, each with their own middleware stack

`server.Option` is handed the raw `chi.Router` — upstream imposes no
required auth, route grouping, or flags. `main.go` uses this three ways:

```go
srv := server.NewServer(cfg, svcs,
    // pkg/handlers.AuditHandler: a new capability (sandbox stats) with no
    // upstream equivalent, built purely by composing upstream's existing
    // SandboxService.ListSandboxes. Mounted at /api/audit reusing
    // upstream's own middleware.RequireAuth.
    server.WithRoutes("/api/audit", auditHandler, upstreammiddleware.RequireAuth),

    // pkg/handlers.UptimeHandler: another new capability, depending on
    // downstream's own SandboxUptimeService interface. Also reuses
    // upstream's middleware.RequireAuth.
    server.WithRoutes("/api/sandbox-uptime", uptimeHandler, upstreammiddleware.RequireAuth),

    // /debug/ping: a route with a completely separate, downstream-only
    // auth scheme (pkg/middleware.RequireDownstreamKey, a plain header
    // check unrelated to upstream's Bearer token). Takes the router
    // directly -- no upstream helper involved at all.
    func(r chi.Router) {
        r.Route("/debug", func(r chi.Router) {
            r.Use(downstreammiddleware.RequireDownstreamKey)
            r.Get("/ping", pingHandler)
        })
    },
)
```

All three are equally valid `Option`s; upstream's `NewServer` doesn't
special-case any of them. `AuditHandler` and `UptimeHandler` each only need
a `RegisterRoutes(chi.Router)` method — no upstream interface to
implement, no field on `server.Services`. Both live in `pkg/handlers`,
mirroring how upstream keeps `GatewayHandler`/`SandboxHandler` together in
its own `pkg/handlers` — one package per BFF for its HTTP layer, one file
per handler within it. `AuditHandler` reads through the same decorated
sandbox service as everything else, so its counts reflect downstream's
`CreateSandbox` behavior too.

## When to use which pattern

| Change needed | Pattern | Handler / middleware |
|---|---|---|
| None | Reuse upstream service directly | Reuse upstream handler directly (`../downstream-reuse`) |
| Behavior change, same method signature | Embed + override (`pkg/services/sandbox.go`) | Reuse upstream handler unmodified |
| New method alongside unchanged existing ones | Embed + extend, define your own interface for the new method (`pkg/services/sandbox_extended.go`) | New handler depending on the new interface, attached via `server.WithRoutes` (`pkg/handlers.UptimeHandler`) |
| A genuinely new route/capability upstream doesn't have | Compose existing service methods (or add a new interface) | New handler, attached via `server.WithRoutes` (`pkg/handlers.AuditHandler`) |
| A route needing different auth/logging than upstream's | N/A | `server.Option` as a raw `func(chi.Router)` with its own middleware (`/debug/ping`) |

## Package naming

Both this module and upstream have a package literally named `services`
(`pkg/services`, one flat package per module, one file per domain) — they
just live in different Go modules, so there's no collision. Every import
site in this repo aliases one or both explicitly (`upstreamservices`,
`downstreamservices`) purely for readability, not because Go requires it.

## Local dev without a published upstream version

`go.mod` uses a `replace` directive pointing at `../upstream` (see the
comment on that line). Once upstream tags a release, drop the `replace`
and pin the `require` to the real version.
