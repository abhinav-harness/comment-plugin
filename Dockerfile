# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary for the target platform
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -X main.version=${VERSION:-dev}" \
    -o /build/comment-plugin \
    ./cmd/plugin

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/comment-plugin /app/comment-plugin

# Use non-root user
USER appuser

ENTRYPOINT ["/app/comment-plugin"]

