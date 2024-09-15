# Start from the official Golang image for building the Go application
FROM golang:alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Download and cache dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o file-sharing-system main.go

# Start a new stage to create a minimal runtime image
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Install necessary libraries for the application
RUN apk --no-cache add ca-certificates

# Copy the built Go binary from the builder stage
COPY --from=builder /app/file-sharing-system .

# Copy static files if necessary
COPY --from=builder /app/static ./static

# Expose the application port
EXPOSE 9080

# Run the Go binary
CMD ["./file-sharing-system"]
