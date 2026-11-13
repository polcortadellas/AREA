# Spike: Go ping service

## Goal

Build a minimal Go service that answers on `/ping`, to check that the environment works and that the approach is viable.

## Result

Works. Tested in this environment (Go 1.22.2, linux/amd64):

- `go vet` and `go build` pass with no errors.
- `GET /ping` returns `200 OK` with body `{"status":"ok"}` (`Content-Type: application/json`).
- `POST /ping` returns `405 Method Not Allowed`, because the `"GET /ping"` pattern restricts the method.

## Approach

- Standard library `net/http`, no external dependencies.
- Method-qualified route pattern (`"GET /ping"`), available since Go 1.22.
- Port read from the `PORT` environment variable, default `:8080`.

## Difficulties

### Encountered during the spike

- None blocked the build or the run. The code compiled and the server responded on the first attempt.

### Encountered in the research before the spike (documentation)

1. **Finding how-to material.** Official docs split into tutorials (how to do things) and reference (pkg.go.dev, what each function does). Usable code examples live in the reference, at the bottom of each function page, not in the tutorials.
2. **Go version dependency.** The method-qualified route pattern needs Go 1.22 
   or newer. Older toolchains will not accept it as written.

## Open items

- Add tests with `go test` (for example using `httptest`).

## How to reproduce

```bash
cd spikes/go
go run .
curl -i localhost:8080/ping
curl -i -X POST localhost:8080/ping   # 405
```