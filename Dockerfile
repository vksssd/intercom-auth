# Use the official Golang image as the builder
FROM golang:1.22.4 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the necessary files for building the binary
COPY . .

# Fetch dependencies
RUN go mod tidy

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o auth ./cmd/main.go

# Use a minimal base image for the final build
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/auth .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./auth"]
