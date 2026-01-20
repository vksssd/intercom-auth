# Dockerfile
FROM golang:1.26rc2

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o auth ./cmd/main.go

EXPOSE 8080

CMD ["./auth"]
