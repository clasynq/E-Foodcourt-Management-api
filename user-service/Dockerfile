# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service ./cmd/server/main.go

# Stage 2: Create a minimal production container
FROM alpine:3.19

# Add certificates for HTTPS/SMTP requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from stage 1
COPY --from=builder /app/user-service .

# Copy env template just in case
COPY --from=builder /app/.env.example ./

# Expose the service port
EXPOSE 8081

# Command to run the application
CMD ["./user-service"]
