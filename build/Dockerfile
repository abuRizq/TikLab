# TikLab Sandbox - Multi-stage build
# Build arg for version tagging (default: latest)
ARG VERSION=latest

# Stage 1: Build behavior engine
FROM golang:1.25-alpine AS engine-builder
WORKDIR /build
COPY go.mod ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 go build -o tiklab-engine ./cmd/tiklab-engine

# Stage 2: Runtime with RouterOS + engine
FROM evilfreelancer/docker-routeros:7

# Re-declare for runtime stage
ARG VERSION=latest

# Copy behavior engine binary
COPY --from=engine-builder /build/tiklab-engine /usr/local/bin/tiklab-engine

# Copy entrypoint script
COPY build/scripts/entrypoint.sh /usr/local/bin/tiklab-entrypoint.sh
RUN sed -i 's/\r$//' /usr/local/bin/tiklab-entrypoint.sh && chmod +x /usr/local/bin/tiklab-entrypoint.sh

# Expose ports: SSH, API, Winbox, Control
EXPOSE 22 8728 8291 9090

ENTRYPOINT ["/usr/local/bin/tiklab-entrypoint.sh"]
