.DEFAULT_GOAL := build

VERSION := $(shell cat version.txt)

.PHONY: test fmt vet build
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test -cover ./...

build: test
	GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/sylvain-bataille/fujigo/cmd.CmdVersion=$(VERSION)" -o fujigo .
	GOOS=windows GOARCH=amd64 go build  -ldflags "-X github.com/sylvain-bataille/fujigo/cmd.CmdVersion=$(VERSION)" -o fujigo.exe .

