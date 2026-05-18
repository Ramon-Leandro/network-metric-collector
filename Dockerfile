# Step 1: Build stage
FROM golang:1.23-alpine AS builder

# Install git for private modules if necessary
RUN apk add --no-cache git

WORKDIR /app

# Leverage Docker cache by copying dependencies first
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static linking ensures the binary runs without external C libraries
RUN CGO_ENABLED=0 GOOS=linux go build -o collector ./cmd/collector/main.go

# Step 2: Final stage
FROM scratch

# Import certificates to allow HTTPS calls
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/collector .
COPY --from=builder /app/configs/ ./configs/

# Running as a non-privileged user is a security best practice
ENTRYPOINT ["./collector"]