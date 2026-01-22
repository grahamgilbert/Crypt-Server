# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o crypt-server ./cmd/crypt-server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/crypt-server .

# Copy web assets
COPY --from=builder /app/web ./web

# Create non-root user
RUN adduser -D -u 1000 crypt
USER crypt

EXPOSE 8080

ENTRYPOINT ["/app/crypt-server"]
