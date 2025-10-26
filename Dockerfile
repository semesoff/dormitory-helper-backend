# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/dormitory-helper ./cmd/app

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/dormitory-helper .
COPY --from=builder /app/migrations ./migrations

# Expose ports
# 50051 - gRPC
# 8081 - HTTP Gateway
EXPOSE 50051 8081

# Run the application
CMD ["./dormitory-helper"]
