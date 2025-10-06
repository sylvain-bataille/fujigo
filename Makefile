.DEFAULT_GOAL := build

.PHONY: test fmt vet build
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test -cover ./...

build: test
	GOOS=linux GOARCH=amd64 go build -o fujigo .
	GOOS=windows GOARCH=amd64 go build -o fujigo.exe .

