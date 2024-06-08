FROM  golang:1.22.4 AS builder 

WORKDIR  /app

COPY . .

