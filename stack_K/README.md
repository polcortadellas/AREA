# AREA Stack (Flutter + Hono/Bun + PostgreSQL)

## Start Everything with Docker Compose

```bash
cd stack
docker compose up --build
```

### Services
- **Backend (Hono + Bun)**: `http://localhost:8080` (`GET /ping` -> `{"status":"ok"}`)
- **Frontend (Flutter Web)**: `http://localhost:8081` (connected to `http://localhost:8080/ping`)
- **Database (PostgreSQL 16)**: `localhost:5432`

Stop everything:
```bash
docker compose down
```
