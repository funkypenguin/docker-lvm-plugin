#!/bin/bash

set -e
set -x

#install
go get github.com/golang/lint/golint

#script
test -z "$(go vet .)"
test -z "$(golint .)"
test -z "$(gofmt -s -l *.go)"
go list ./... | go test -v
