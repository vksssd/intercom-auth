FROM  golang:1.20 AS builder 

WORKDIR  /app

COPY . .

