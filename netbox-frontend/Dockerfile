# ---- Build stage for Vue.js frontend ----
FROM node:24.18.0-bookworm-slim AS frontend-builder

WORKDIR /build

# Cache dependencies
COPY package.json package-lock.json ./
RUN npm ci

# Copy source and build
COPY . .
RUN npm run build

# ---- Runtime stage (just the dist output) ----
FROM alpine:3.20

RUN mkdir -p /web/dist
COPY --from=frontend-builder /build/dist /web/dist

# This image just produces the built static files.
# The docker-compose mounts this into the backend container.
CMD ["echo", "Frontend built. Use as a volume source."]
