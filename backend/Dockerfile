# ===== Stage 1: Build =====
FROM golang:1.25.5-alpine3.23 AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /weekly-dashboard .

# ===== Stage 2: Runtime =====
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

# Set timezone to Asia/Jakarta (WIB)
ENV TZ=Asia/Jakarta

WORKDIR /app

# Copy binary from builder
COPY --from=builder /weekly-dashboard .

# Copy .env as fallback defaults (env vars from docker-compose take priority)
COPY .env .env

EXPOSE 8080

CMD ["./weekly-dashboard"]
