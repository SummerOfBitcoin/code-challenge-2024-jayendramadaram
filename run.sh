#!/bin/bash


if ! command -v go &> /dev/null
then
    echo "couldn't find Go, Please install Go"
    exit 1
else
    # go mod tidy
    go run cmd/main.go > info.log
fi
