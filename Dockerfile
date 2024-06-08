# Use the official Golang image as the builder
FROM golang:1.22.4 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the Working Directory inside the container
COPY . .

# Fetch dependencies
RUN go mod tidy

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux  build -o auth ./cmd/main.go

# Use an Alpine image for the final build targeting arm64 architecture
FROM arm64v8/alpine:latest

# Install necessary CA certificates and compatibility libraries
RUN apk --no-cache add ca-certificates libc6-compat

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/auth .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./auth"]
