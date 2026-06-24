# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25
ARG APP_PORT=8080

# ── Stage 1: builder ───────────────────────────────────────────
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Dependencies
COPY go.mod go.sum* ./
COPY vendor/ ./vendor/

# Build
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -mod=vendor -trimpath -ldflags="-w -s" -o /bin/server ./cmd/server

# ── Stage 2: runner ────────────────────────────────────────────
FROM scratch

ARG APP_PORT

WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/server /server
COPY --from=builder /app/sql/migrations /sql/migrations

# Run as a non-root, unprivileged numeric UID/GID
USER 65532:65532

EXPOSE ${APP_PORT}

HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=5 \
    CMD ["/server", "health"]

ENTRYPOINT ["/server"]
