# ---- Build stage for Go backend ----
FROM golang:1.26-bookworm AS backend-builder

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /netbox_go ./cmd/netbox_go

# ---- Runtime stage ----
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary
COPY --from=backend-builder /netbox_go /app/netbox_go

# Copy configs
COPY configs/ /app/configs/

# Copy docs (swagger)
COPY docs/ /app/docs/

# Web dist will be mounted or copied by docker-compose
# The frontend is built separately and copied into /app/web/dist

EXPOSE 8080 7682 7683

HEALTHCHECK --interval=10s --timeout=5s --retries=3 --start-period=15s \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/netbox_go", "-c", "/app/configs/netbox_go.yml"]