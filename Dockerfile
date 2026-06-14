# Multi-stage build producing the three binaries (server, engine, seed).
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -o /out/engine ./cmd/engine \
 && CGO_ENABLED=0 go build -o /out/seed   ./cmd/seed

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/ /app/
# The engine loads strategy YAMLs at runtime (news is embedded; strategies are not).
COPY --from=build /src/internal/engine/strategies /app/internal/engine/strategies
# Default command is the API server; compose overrides it for the engine/seed.
CMD ["/app/server"]
