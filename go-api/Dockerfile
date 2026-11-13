# Build context is the repo root (see docker-compose.yml) so web/ can be copied in.

# build: toolchain + sources; compiles the binary and can run the tests.
FROM golang:1.27.4-alpine AS build
WORKDIR /src
COPY go-api/go.mod go-api/go.sum ./
RUN go mod download
COPY go-api/ .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server
CMD ["go", "test", "./..."]

# runtime: static binary + frontend on distroless, non-root (uid 65532).
FROM gcr.io/distroless/static-debian13:nonroot AS runtime
WORKDIR /app
COPY --from=build /out/server /app/server
COPY web /app/web
ENV PORT=3000 WEB_DIR=/app/web
EXPOSE 3000
USER nonroot
ENTRYPOINT ["/app/server"]
