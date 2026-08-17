# upstream

The "core" BFF. Chi router, contract-based interfaces (`pkg/services/*`),
handlers that depend only on those interfaces (`pkg/handlers`). Everything
here is exported so downstream can reuse it directly or decorate it.

## Run

```
go run ./cmd/api
```

Defaults to `:8080`. Every route under `/api` requires
`Authorization: Bearer secret-token`.

```
curl -H "Authorization: Bearer secret-token" -X POST localhost:8080/api/sandbox/ -d '{"name":"my-box"}'
curl -H "Authorization: Bearer secret-token" localhost:8080/api/sandbox/sandbox-1
curl -H "Authorization: Bearer secret-token" localhost:8080/api/gateway/
```

## What's here

- `pkg/services` — one flat package, one file per domain
  (`gateway.go`, `sandbox.go`). Each defines a public interface + a
  default in-memory implementation. Sandbox is plain CRUD
  (Create/Get/List/Delete). No per-domain subpackages, so consumers import
  it once as `services` instead of aliasing multiple same-named packages.
- `pkg/handlers` — HTTP handlers depending only on the service interfaces,
  never the concrete structs.
- `pkg/server` — `NewServer(cfg, Services, ...Option)` wires handlers to
  routes. `Services` is the injection point for swapping a decorated
  service in; `Option` is the injection point for attaching handlers
  upstream doesn't know about.
- `pkg/middleware` — `RequireAuth`.

## Attaching custom handlers

`Option` is handed the raw `chi.Router` — no flags, no hidden mounting
rules. It decides its own route grouping and its own middleware, which may
or may not have anything to do with upstream's:

```go
srv := server.NewServer(cfg, svcs,
    // Convenience helper: mount a RegisterRoutes(chi.Router)-shaped
    // handler at a pattern, reusing upstream's own auth middleware.
    server.WithRoutes("/api/audit", auditHandler, middleware.RequireAuth),

    // Or take the router directly for full control -- own route group,
    // own middleware stack, no relation to upstream's at all.
    func(r chi.Router) {
        r.Route("/debug", func(r chi.Router) {
            r.Use(myOwnAuth)
            r.Get("/ping", pingHandler)
        })
    },
)
```

See `../downstream-partialreuse` for both patterns working side by side.

See `../downstream-reuse` and `../downstream-partialreuse` for what
downstream does with this.
