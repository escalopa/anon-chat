# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o anon-chat-app main.go

# Stage 2: Create a smaller image
FROM alpine:3.20.3

# Set the working directory in the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/anon-chat-app .

# Set the entrypoint command to run the application
ENTRYPOINT ["./anon-chat-app"]
