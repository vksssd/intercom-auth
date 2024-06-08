FROM  golang:1.21 AS builder 

WORKDIR  /app

COPY . .

