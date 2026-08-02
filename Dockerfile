# =========================================================================
# Unified multi-stage Dockerfile for NetBox Go (frontend + backend)
# Build:  docker build -t netbox-go .
# Run:    docker-compose up
# =========================================================================

# ---- Stage 1: Build Vue.js frontend ----
FROM node:24.18.0-bookworm-slim AS frontend-builder

WORKDIR /frontend
COPY netbox-frontend/package.json netbox-frontend/package-lock.json ./
RUN npm ci
COPY netbox-frontend/ .
RUN npm run build

# ---- Stage 2: Build Go backend ----
FROM golang:1.26.0-bookworm AS backend-builder

WORKDIR /backend
COPY netbox-backend/go.mod netbox-backend/go.sum ./
RUN go mod download

# Copy backend source (the docs symlink will be broken, fix it below)
COPY netbox-backend/ .

# Replace the broken docs symlink with actual docs content from repo root
RUN rm -rf docs
COPY docs/ ./docs/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /netbox_go ./cmd/netbox_go \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /netbox_go_admin ./cmd/netbox_go_admin

# ---- Stage 3: Final runtime image ----
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /netbox_go /app/netbox_go
COPY --from=backend-builder /netbox_go_admin /app/netbox_go_admin

# Copy configs (paths are relative to repo root build context)
COPY netbox-backend/configs/ /app/configs/

# Copy docs (swagger) — these are needed for the API docs endpoint
COPY docs/ /app/docs/

# Copy frontend dist into the location the Go server expects
COPY --from=frontend-builder /frontend/dist /app/web/dist

EXPOSE 8080 7682

HEALTHCHECK --interval=10s --timeout=5s --retries=3 --start-period=15s \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/netbox_go", "-c", "/app/configs/netbox_go.docker.yml"]
