# downstream-partialreuse

A downstream BFF that reuses most of upstream's code, but demonstrates two
ways to diverge from it: decorating a service, and attaching a brand new
handler upstream has no concept of.

## Run

```
PORT=8083 go run ./cmd/api
curl -H "Authorization: Bearer secret-token" -X POST localhost:8083/api/sandbox/ -d '{"name":"my-box"}'
curl -H "Authorization: Bearer secret-token" localhost:8083/api/audit/sandboxes
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

## New handlers, each with their own middleware stack

`server.Option` is handed the raw `chi.Router` — upstream imposes no
required auth, route grouping, or flags. `main.go` uses this two ways:

```go
srv := server.NewServer(cfg, svcs,
    // pkg/audit.Handler: a new capability (sandbox stats) with no
    // upstream equivalent, built purely by composing upstream's existing
    // SandboxService.ListSandboxes. Mounted at /api/audit reusing
    // upstream's own middleware.RequireAuth.
    server.WithRoutes("/api/audit", auditHandler, upstreammiddleware.RequireAuth),

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

Both are equally valid `Option`s; upstream's `NewServer` doesn't special-case
either. `audit.Handler` only needs a `RegisterRoutes(chi.Router)` method —
no upstream interface to implement, no field on `server.Services`. It
reads through the same decorated sandbox service as everything else, so
its counts reflect downstream's `CreateSandbox` behavior too.

## When to use which pattern

| Change needed | Pattern | Handler / middleware |
|---|---|---|
| None | Reuse upstream service directly | Reuse upstream handler directly (`../downstream-reuse`) |
| Behavior change, same method signature | Embed + override (`pkg/services/sandbox.go`) | Reuse upstream handler unmodified |
| A genuinely new route/capability upstream doesn't have | Compose existing service methods (or add a new interface) | New handler, attached via `server.WithRoutes` (`pkg/audit`) |
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
