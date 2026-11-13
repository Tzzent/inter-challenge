# QR → Stats pipeline

Coding challenge solution: two REST APIs that work in a chain.

- **go-api** (Go + Fiber v3): receives a rectangular matrix, computes its full **QR factorization** with hand-written Householder reflections, sends Q and R to the Node API, and returns Q, R and the stats.
- **node-api** (Node + Express 5, TypeScript): computes max, min, average, sum, and whether any matrix is diagonal, over all values of all matrices in a single pass.

Both are protected with JWT (HS256), containerized, and tested inside Docker. A minimal frontend is served by the Go API at `/`.

## Architecture

```
Client ──POST /api/v1/qr {matrix}──▶ go-api :3000 ──POST /api/v1/stats {matrices:[Q,R]}──▶ node-api :4000
   ▲                                    │  (forwards Authorization)                         │
   └──────────── {q, r, stats} ◀────────┴──────────────── {max,min,average,sum,...} ◀────────┘
```

```
go-api/
  cmd/server/main.go        env wiring, -healthcheck flag
  internal/qr               Householder QR (pure, no HTTP); gonum used only in tests as an oracle
  internal/handler          routes, input validation, error mapping
  internal/statsclient      HTTP client for node-api (5 s timeout)
  internal/auth             JWT issue/verify, constant-time credential check
node-api/
  src/stats.ts              computeStats, isDiagonal, validateMatrices (pure)
  src/app.ts / server.ts    Express app (no listen) / entrypoint
  src/auth.ts               JWT middleware
web/index.html              login → matrix → Q, R, stats tables
scripts/integration-test.mjs  end-to-end test against the running stack
```

## Quick start

Requires Docker with Compose v2+. Nothing else is needed on the host.

```sh
cp .env.example .env    # edit JWT_SECRET and the demo credentials
docker compose up --build
```

- Frontend: http://localhost:3000
- Go API: http://localhost:3000, Node API: http://localhost:4000

## curl examples

```sh
TOKEN=$(curl -s -X POST localhost:3000/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | sed 's/.*"token":"\([^"]*\)".*/\1/')

curl -s -X POST localhost:3000/api/v1/qr \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"matrix":[[3,0],[4,5]]}'
# {"q":[[-0.6,-0.8],[-0.8,0.6]],"r":[[-5,-4],[0,3]],
#  "stats":{"max":3,"min":-5,"average":-0.95,"sum":-7.6,"anyDiagonal":false,"diagonal":[false,false]}}
```

(Values shown rounded; floating-point output may carry ~1e-16 noise.)

## Tests

All tests run in containers:

```sh
docker compose --profile test run --rm --build go-api-test     # go test ./...
docker compose --profile test run --rm --build node-api-test   # node:test + supertest

# end-to-end, against the running stack
docker compose up -d --build --wait go-api node-api
docker compose --profile test run --rm --build integration-test
docker compose down
```

## API reference

Errors are always JSON: `{"error": "..."}`.

### go-api (port 3000)

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | public | `{"status":"ok","statsApi":"ok"\|"down"}`, always 200 while Go is up |
| POST | `/auth/token` | public | body `{"username","password"}` → `{"token"}` (HS256, 1 h). 401 on bad credentials |
| POST | `/api/v1/qr` | Bearer | body `{"matrix": number[][]}` → `{"q","r","stats"}` |
| GET | `/` | public | frontend |

`/api/v1/qr` returns **400** for malformed JSON, empty matrix or rows, ragged rows, non-numeric/non-finite values, or more than 100 rows/columns (body limit 1 MiB); **401** without a valid token; **502** if node-api fails, times out (5 s) or rejects the request.

Q is m×m orthogonal and R is m×n upper triangular (full QR) for any shape. Householder signs may differ from a hand computation: `(-Q)(-R) = QR`.

### node-api (port 4000)

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | public | `{"status":"ok"}` |
| POST | `/api/v1/stats` | Bearer | body `{"matrices": number[][][]}` → `{"max","min","average","sum","anyDiagonal","diagonal": boolean[]}` |

Returns **400** for empty `matrices`, an empty matrix, ragged rows or non-finite values. A matrix is diagonal when every element with i ≠ j satisfies |x| ≤ 1e-9; non-square ("rectangular diagonal") matrices are allowed.

### Environment

| Variable | Used by | Description |
|---|---|---|
| `JWT_SECRET` | both | shared HS256 secret (required) |
| `AUTH_USERNAME`, `AUTH_PASSWORD` | go-api | demo credentials for `/auth/token` |
| `STATS_API_URL` | go-api | node-api base URL (compose sets `http://node-api:4000`) |
| `PORT` | both | 3000 / 4000 |

## Docs (Spanish)

- [docs/01-problem.md](docs/01-problem.md): the problem explained without jargon, with a worked example.
- [docs/02-decisions.md](docs/02-decisions.md): design decisions and their rationale.

## Deployment (Render)

Live on Render's free plan, defined in [`render.yaml`](render.yaml) (a Render Blueprint):

- Web + QR API (Go): https://qr-stats-go-api.onrender.com
- Stats API (Node): https://qr-stats-node-api.onrender.com

Free services sleep after 15 minutes idle, so the first request can take 30 to 50 seconds. Free services can't use Render's private network, so the Go API calls the Node API through its public URL (`STATS_API_URL`). `JWT_SECRET` is generated by Render and shared through the `qr-stats-shared` env group, where `AUTH_USERNAME` and `AUTH_PASSWORD` are also set.

Render was chosen over Google Cloud Run because GCP billing signup failed (`OR_BACR2_59`) and Render needs no credit card. Both run the same Docker images.

Smoke test against the live deploy:

```sh
set -a; . ./.env; set +a
GO_API_URL=https://qr-stats-go-api.onrender.com node scripts/integration-test.mjs
```
