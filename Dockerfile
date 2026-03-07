# TikLab Launcher - runs the sandbox CLI with all dependencies inside Docker
# Use: docker run --rm -it --privileged -v /var/run/docker.sock:/var/run/docker.sock -v tiklab-data:/sandbox ghcr.io/aburizq/tiklab create
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o sandbox .

FROM ubuntu:22.04
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    docker.io \
    curl \
    qemu-utils \
    iproute2 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && curl -SL https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-linux-x86_64 -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose

COPY --from=builder /build/sandbox /usr/local/bin/sandbox

ENV SANDBOX_WORKDIR=/sandbox
WORKDIR /sandbox

ENTRYPOINT ["sandbox"]
CMD ["--help"]
