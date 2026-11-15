# Spike: FastAPI ping service

## Goal

Build a minimal FastAPI service that answers on `/ping`, to compare it with the Go, NestJS and Hono/Bun spikes on the same target.

Prior experience: some Python, never used FastAPI before.

## Timebox

**60 minutes**, fixed before starting (20:49). The bootstrap sizes its exercises at 40 to 90 minutes and the target is minimal, so 60 minutes leaves room for installing the toolchain without letting the spike grow into a feature.

## Where I was when the time ran out

Done at 20:52, after **3 minutes**, so the timebox never rang. The code was written with AI assistance (Claude), so the minutes measure friction (install, run, test, Docker), not typing speed. Every line was then run and reviewed by me.

| Checkpoint | Reached |
|---|---|
| `GET /ping` returns exactly `{"status":"ok"}` | Yes |
| Port read from `PORT`, no literal in the source (missing variable = explicit error) | Yes |
| Automated test (`pytest`, 2 tests) | Yes |
| Docker image builds and answers | Yes, **214 MB** (`python:3.13-slim`) |
| OpenAPI served and browsable | Yes, `/openapi.json` (OpenAPI 3.1) and `/docs`, **with zero extra code** |

## Result

Tested in this environment (Python 3.13, FastAPI 0.163, Uvicorn 0.54, Windows 11):

- `GET /ping` returns `200 OK` with body `{"status":"ok"}` (`Content-Type: application/json`).
- `POST /ping` returns `405 Method Not Allowed` (like Go, unlike NestJS which answers `404`).
- Without `PORT`, the server refuses to start: `PORT environment variable is required`.

## What consumed the most minutes

1. **Docker check.** The first `curl` against the container got no answer because it hit the port before Uvicorn was ready. Retrying fixed it. The real stack will need a healthcheck rather than relying on start order.
2. **Ecosystem churn.** Starlette 1.10 already warns that `TestClient` with `httpx` is deprecated in favour of `httpx2`. Versions must be pinned from day 1 (see `requirements.txt`).

## What I would need to go further

- OAuth2 / OIDC: FastAPI ships helpers for *being* an OAuth2 server, not for *consuming* Google or GitHub. A client library such as Authlib is needed.
- Hooks: FastAPI has no scheduler. Polling would need APScheduler, or a job queue (arq, Celery) for ENG-ASYNC.
- Typing is optional in Python. `mypy` in CI would be needed to catch errors before runtime.
- The bigger image (214 MB vs a few MB for a static Go binary) matters little locally, but more for SYS-RELEASE or SYS-K8S.

## How to reproduce

```bash
cd spikes/fastapi
python -m venv .venv
.venv/Scripts/python -m pip install -r requirements-dev.txt   # Linux/macOS: .venv/bin/python
PORT=8080 .venv/Scripts/python main.py
curl -i localhost:8080/ping
curl -i -X POST localhost:8080/ping        # 405
.venv/Scripts/python -m pytest -q          # 2 passed
# open http://localhost:8080/docs for the generated OpenAPI UI

# Docker
docker build -t area-spike-fastapi .
docker run --rm -e PORT=8080 -p 8080:8080 area-spike-fastapi
```
