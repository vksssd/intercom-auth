# Dockerfile
FROM golang:1.24rc1

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o auth ./cmd/main.go

EXPOSE 8080

CMD ["./auth"]
