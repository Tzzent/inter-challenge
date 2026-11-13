# Feature: qr-stats-pipeline

## Objective
Build the Interseguro coding challenge as two services:
- a Go + Fiber API that computes the QR factorization of a rectangular matrix;
- a Node + Express (TypeScript) API that computes statistics over Q and R.

Both are containerized with Docker, protected with JWT, and backed by tests.

## Problem / Why
The PDF in `assets/` defines the challenge. Plan: `~/.claude/plans/memoized-orbiting-journal.md`.

## Scope and constraints
- Everything runs in Docker, tests included. Go is not installed on the host.
- QR uses hand-written Householder reflections and returns the full QR: Q is m×m, R is m×n. gonum is used only as a test oracle.
- Diagonal check uses epsilon `1e-9`.
- Input limits: matrix at most 100×100, non-empty, rectangular, finite numbers.
- `docs/` is written in Spanish, because it is the user's learning material. Code, README and comments are in English.
- Delivery strategy: `ask-on-risk`. Commits happen only after user confirmation, as stated in the plan.

## Tasks
- [x] T0 Bootstrap: git, `.gitignore`, `docs/01-problem.md`, `docs/02-decisions.md`
- [x] T1 Walking skeleton: `/health` on both APIs, Dockerfiles (build + runtime stages), `docker-compose.yml` with a `test` profile
- [x] T2 Node stats (TDD): `computeStats`, `isDiagonal`, validation, `POST /api/v1/stats`
- [x] T3 Go QR (TDD): Householder, validation, `POST /api/v1/qr`
- [x] T4 Integration Go→Node: `statsclient` with timeout, stats included in the QR response, 502 when Node fails, compose integration test
- [x] T5 JWT: `POST /auth/token`, middleware in both APIs, 401 tests
- [x] T6 Frontend: `web/index.html` served by Fiber
- [ ] T7 Render deploy via `render.yaml` Blueprint (free plan). Cloud Run dropped: GCP billing signup failed with OR_BACR2_59. Pending: push to GitHub, then create the Blueprint on Render.
- [x] T8 README

## Route
- T0: inline (parent).
- T1–T6, T8: delegated to one writer. Trigger: 2 or more non-trivial files.

## Acceptance criteria
- `docker compose up --build` serves both APIs.
- `POST /api/v1/qr` with `[[3,0],[4,5]]` returns Q, R and stats.
- Both test profiles exit 0.

## Progress / evidence
- 2026-11-12: Git initialized on branch `feat/qr-stats-pipeline`.

- T0: wrote `.gitignore`, `.env.example`, `docs/01-problem.md` and `docs/02-decisions.md` (Spanish). Structural readback only.
- T1: `docker compose up -d --build --wait go-api node-api` → both services `(healthy)`; `curl localhost:3000/health` → `{"statsApi":"ok","status":"ok"}`.
- T2 RED: `npm test` (node:26.13.0-alpine) with a stub `stats.ts` → 7 tests fail with `Error: not implemented`. GREEN: after implementation → `tests 7, pass 7, fail 0`.
- T3 RED: `go test ./internal/qr/` (golang:1.27.4-alpine) with a stub → FAIL (`R: shape 1x1, want 3x3`, ...). GREEN: Householder implementation → `ok go-api/internal/qr`, including known example and gonum oracle.
- T4: `docker compose --profile test run --rm --build integration-test` → `integration test passed`, exit 0. With node-api stopped: `/health` → 200 `{"statsApi":"down"}`, `/api/v1/qr` → 502 `{"error":"stats service unavailable"}`.
- T5: Go auth tests (no token, garbage, wrong key, expired, HS512, alg=none → 401; valid → 200) and Node app tests (no token, bad token, alg=none, HS512 → 401; valid → 200) pass.
- T6: `curl -o /dev/null -w '%{http_code}' localhost:3000/` → 200 (frontend served by Fiber static).
- T8: README written in English. Structural readback only.
- Final: `docker compose --profile test run --rm --build go-api-test` → all 4 packages `ok`, exit 0. `... node-api-test` → `tests 19, pass 19, fail 0`, exit 0.

## Versions
Go image `golang:1.27.4-alpine` / runtime `gcr.io/distroless/static-debian13:nonroot`; Fiber v3.7.0; golang-jwt v5.3.1; gonum v0.17.0 (tests only); `node:26.13.0-alpine`; express 5.3.0; typescript 7.0.2 (works, no fallback needed); jsonwebtoken 9.0.3; supertest 7.3.2; tsx 4.23.15.

## Next step
User review, then work-unit commits on `feat/qr-stats-pipeline`. T7 (Render) needs the repo on GitHub.
