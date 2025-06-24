# Build stage
FROM golang:1.23-alpine AS builder

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
RUN make build

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