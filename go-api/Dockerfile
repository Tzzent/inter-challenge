# El contexto es la raíz del repo (ver docker-compose.yml) para poder copiar web/.

# build: compila el binario y permite correr los tests.
FROM golang:1.27.4-alpine AS build
WORKDIR /src
COPY go-api/go.mod go-api/go.sum ./
RUN go mod download
COPY go-api/ .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server
CMD ["go", "test", "./..."]

# runtime: binario estático + frontend sobre distroless, sin root (uid 65532).
FROM gcr.io/distroless/static-debian13:nonroot AS runtime
WORKDIR /app
COPY --from=build /out/server /app/server
COPY web /app/web
ENV PORT=3000 WEB_DIR=/app/web
EXPOSE 3000
USER nonroot
ENTRYPOINT ["/app/server"]
