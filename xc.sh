#!/usr/bin/env zsh
env GOOS=linux GOARCH=amd64 go build -o lights-server cmd/lights-server/lights-server.go
