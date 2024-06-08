FROM  golang:1.18 AS builder 

WORKDIR  /app

COPY . .

