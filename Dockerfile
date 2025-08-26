# Development environment Dockerfile
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine AS dev

# Disable VCS stamping to avoid git issues in container
ENV GOFLAGS="-buildvcs=false"

# Install development tools
RUN apk add --no-cache \
    git \
    curl \
    bash \
    make \
    gcc \
    musl-dev \
    binutils-gold

# Copy Makefile first to leverage layer caching
COPY Makefile ./

# Install golangci-lint
RUN make install-tools

# Set working directory
WORKDIR /workspace

# Copy go mod files for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Default command
CMD ["bash"]

# Build stage
FROM golang:${GO_VERSION}-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o build/i18ngen .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh i18ngen

WORKDIR /home/i18ngen

# Copy binary from builder stage
COPY --from=builder /app/build/i18ngen /usr/local/bin/i18ngen

# Switch to non-root user
USER i18ngen

# Set entrypoint
ENTRYPOINT ["i18ngen"]
CMD ["--help"]
