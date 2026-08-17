# downstream-reuse

Baseline example: a downstream BFF that reuses **100% of upstream's code**,
unmodified. No decoration, no overrides, no local packages.

## Run

```
go run ./cmd/api
```

Defaults to `:8080` (set `PORT` to change; upstream also defaults to
`:8080`, so run them on different ports if testing side by side).

```
PORT=8082 go run ./cmd/api &
curl -H "Authorization: Bearer secret-token" -X POST localhost:8082/api/sandbox/ -d '{"name":"my-box"}'
```

Response is identical in shape and values to hitting upstream directly —
`"status":"running"`, no name mangling. That's the point: this module adds
nothing.

## How it works

`cmd/api/main.go` imports `upstream/pkg/server`, `pkg/services/gateway`,
and `pkg/services/sandbox` directly and wires them exactly like upstream's
own `main.go` does. There is no `pkg/` directory in this module.

`go.mod` uses a `replace` directive pointing at `../upstream` since
upstream hasn't tagged a release yet (see that line's comment for how to
remove it once it has).

## vs. downstream-partialreuse

This module answers "does upstream work as-is with zero changes?".
`../downstream-partialreuse` answers "what does it look like when
downstream needs to change one behavior?" — see that README for the
interface-decoration example.
