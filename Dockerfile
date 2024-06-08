FROM  golang:1.16 AS builder 

WORKDIR  /app

COPY . .

