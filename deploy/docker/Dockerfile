# ================================================================
# OPTRION — Multi-stage Docker Build
# ================================================================
# Stage 1: Build the Go binary
# Stage 2: Minimal runtime image
# ================================================================

# --- Build Stage ---
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /bin/optrion \
    ./cmd/optrion

# --- Runtime Stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S optrion \
    && adduser -S -G optrion optrion

COPY --from=builder /bin/optrion /usr/local/bin/optrion

USER optrion

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/healthz"]

ENTRYPOINT ["optrion"]
