#!/bin/bash
go mod tidy
go mod vendor
go run cmd/main.go
