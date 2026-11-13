# QR → Stats pipeline

Dos APIs REST que trabajan en cadena:

- **go-api** (Go + Fiber v3): recibe una matriz rectangular, calcula su **factorización QR** completa con reflexiones de Householder escritas a mano, envía Q y R a la API Node y devuelve Q, R y las estadísticas.
- **node-api** (Node + Express 5, TypeScript): calcula máximo, mínimo, promedio, suma y si alguna matriz es diagonal, en una sola pasada.

Ambas usan JWT (HS256), corren en Docker y se prueban dentro de contenedores. La API Go sirve un frontend mínimo en `/`.

```
Cliente ──POST /api/v1/qr──▶ go-api :3000 ──POST /api/v1/stats (reenvía Authorization)──▶ node-api :4000
```

## En vivo (Render)

- Web + API QR (Go): https://qr-stats-go-api.onrender.com
- API de estadísticas (Node): https://qr-stats-node-api.onrender.com

## Ejecutar en local

Solo requiere Docker con Compose v2+.

```sh
cp .env.example .env    # edita JWT_SECRET y las credenciales demo
docker compose up --build
```

Frontend y API Go en http://localhost:3000; API Node en http://localhost:4000.

## Tests

```sh
docker compose --profile test run --rm --build go-api-test     # go test ./...
docker compose --profile test run --rm --build node-api-test   # node:test + supertest

# end-to-end contra el stack en ejecución
docker compose up -d --build --wait go-api node-api
docker compose --profile test run --rm --build integration-test
docker compose down
```

## API

Los errores siempre son JSON: `{"error": "..."}`.

| API | Método | Ruta | Auth | Descripción |
|---|---|---|---|---|
| go | GET | `/health` | pública | `{"status":"ok","statsApi":"ok"\|"down"}` |
| go | POST | `/auth/token` | pública | `{"username","password"}` → `{"token"}` (HS256, 1 h); 401 si falla |
| go | POST | `/api/v1/qr` | Bearer | `{"matrix": number[][]}` → `{"q","r","stats"}` |
| node | GET | `/health` | pública | `{"status":"ok"}` |
| node | POST | `/api/v1/stats` | Bearer | `{"matrices": number[][][]}` → `{"max","min","average","sum","anyDiagonal","diagonal"}` |

`/api/v1/qr` responde **400** ante JSON inválido, matriz o filas vacías, filas de distinto largo, valores no finitos o más de 100 filas/columnas; **401** sin token válido; **502** si node-api falla, tarda más de 5 s o rechaza la petición. Q es m×m ortogonal y R es m×n triangular superior; los signos pueden diferir de un cálculo a mano porque `(-Q)(-R) = QR`.

### Autenticación y ejemplo

Se pide un token a `/auth/token` y se envía como `Authorization: Bearer <token>`. Go reenvía ese mismo header a Node, que lo valida con el `JWT_SECRET` compartido.

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

### Variables de entorno

| Variable | Usada por | Descripción |
|---|---|---|
| `JWT_SECRET` | ambas | secreto HS256 compartido (obligatorio) |
| `AUTH_USERNAME`, `AUTH_PASSWORD` | go-api | credenciales demo de `/auth/token` |
| `STATS_API_URL` | go-api | URL base de node-api (compose usa `http://node-api:4000`) |
| `PORT` | ambas | 3000 / 4000 |

## Despliegue (Render)

Definido en [`render.yaml`](render.yaml) (Blueprint de Render, plan gratuito):

- Los servicios gratuitos se duermen tras ~15 min sin tráfico; la primera petición puede tardar 30 a 50 s.
- El plan gratuito no tiene red privada, así que Go llama a Node por su URL pública (`STATS_API_URL`).
- `JWT_SECRET` lo genera Render y se comparte en el grupo `qr-stats-shared`, junto con `AUTH_USERNAME` y `AUTH_PASSWORD`.
- Se eligió Render en lugar de Google Cloud Run porque el alta de facturación en GCP falló (`OR_BACR2_59`) y Render no pide tarjeta. Ambos usan las mismas imágenes Docker.

Smoke test contra el despliegue:

```sh
set -a; . ./.env; set +a
GO_API_URL=https://qr-stats-go-api.onrender.com node scripts/integration-test.mjs
```

## Documentación

- [docs/01-problem.md](docs/01-problem.md): el problema explicado sin jerga, con un ejemplo resuelto.
- [docs/02-decisions.md](docs/02-decisions.md): decisiones de diseño y su justificación.
