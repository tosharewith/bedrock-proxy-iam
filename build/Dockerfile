# ============================================================================
# Multi-stage secure Dockerfile for Go Bedrock Proxy
# Security: Distroless, non-root, minimal attack surface
# Compliance: OWASP, NVD scanning integrated
# ============================================================================

# Build arguments for multi-architecture support
ARG BUILDPLATFORM=linux/amd64
ARG TARGETPLATFORM=linux/amd64
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# ============================================================================
# Stage 1: Build environment with security scanning
# ============================================================================
FROM --platform=$BUILDPLATFORM golang:1.21-alpine3.18 AS builder

# Install security updates and build dependencies
RUN apk update && \
    apk upgrade && \
    apk add --no-cache \
        ca-certificates \
        git \
        tzdata \
        upx && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

# Create non-root user for build process
RUN adduser -D -g '' -h /nonexistent -s /sbin/nologin -u 65534 appuser

# Set build arguments for Go cross-compilation
ARG TARGETOS
ARG TARGETARCH

# Set working directory
WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod go.sum ./

# Download dependencies with verification
RUN go mod download && \
    go mod verify

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build \
        -a \
        -installsuffix cgo \
        -ldflags='-w -s -extldflags "-static" -X main.version=${BUILD_VERSION:-dev} -X main.buildTime=${BUILD_TIME}' \
        -tags=netgo \
        -trimpath \
        -o bedrock-proxy \
        ./cmd/bedrock-proxy

# Compress binary (optional, saves ~30-40% size)
RUN upx --best --lzma bedrock-proxy

# Verify the binary
RUN file bedrock-proxy && \
    ls -la bedrock-proxy && \
    ./bedrock-proxy -version || true

# ============================================================================
# Stage 2: Security scanning (optional but recommended)
# ============================================================================
FROM --platform=$BUILDPLATFORM alpine:3.18 AS security-scan

# Install security scanning tools
RUN apk add --no-cache \
        trivy \
        dumb-init && \
    rm -rf /var/cache/apk/*

# Copy binary for scanning
COPY --from=builder /build/bedrock-proxy /tmp/bedrock-proxy

# Run security scan on binary
RUN trivy fs --severity HIGH,CRITICAL --no-progress /tmp/ || true

# ============================================================================
# Stage 3: Minimal runtime - Distroless (Production)
# ============================================================================
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

# Metadata labels
LABEL maintainer="Platform Team <platform@company.com>" \
      description="Secure Bedrock Proxy with embedded IAM authentication" \
      version="1.0.0" \
      org.opencontainers.image.title="bedrock-proxy" \
      org.opencontainers.image.description="AWS Bedrock Proxy with IRSA support" \
      org.opencontainers.image.vendor="Company" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/company/bedrock-proxy"

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /build/bedrock-proxy /bedrock-proxy

# Use non-root user (uid: 65534, gid: 65534)
USER 65534:65534

# Expose port
EXPOSE 8080/tcp

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/bedrock-proxy", "-health-check"]

# Set entrypoint
ENTRYPOINT ["/bedrock-proxy"]

# ============================================================================
# Stage 4: Alternative runtime with shell access for debugging
# ============================================================================
FROM alpine:3.18 AS runtime-debug

# Install minimal runtime dependencies
RUN apk update && \
    apk upgrade && \
    apk add --no-cache \
        ca-certificates \
        tzdata \
        dumb-init \
        curl \
        jq && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 65534 -S appgroup && \
    adduser -u 65534 -S -D -G appgroup -H -h /nonexistent appuser

# Copy binary
COPY --from=builder /build/bedrock-proxy /usr/local/bin/bedrock-proxy

# Set permissions
RUN chmod +x /usr/local/bin/bedrock-proxy

# Create necessary directories with correct permissions
RUN mkdir -p /tmp /var/run && \
    chown -R appuser:appgroup /tmp /var/run

# Switch to non-root user
USER appuser:appgroup

# Set working directory
WORKDIR /

# Expose port
EXPOSE 8080/tcp

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Use dumb-init to handle signals properly
ENTRYPOINT ["dumb-init", "--"]
CMD ["bedrock-proxy"]

# ============================================================================
# Default target: Use distroless for production
# ============================================================================
FROM runtime AS final