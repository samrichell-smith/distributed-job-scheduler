# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY ../../go.mod ../../go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY ../.. .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api.go

# Final stage
FROM alpine:3.18

WORKDIR /app

# Install certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Copy the binary from builder
COPY --from=builder /app/api .

# Copy any additional required files
COPY .env.example .env

# Expose the API port
EXPOSE 8080

# Run the application
CMD ["./api"]