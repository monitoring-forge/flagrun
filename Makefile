VERSION=0.0.9

.PHONY: check lint

check: *.go
	go test -v ./...
	go test -race ./...

lint: *.go
	golangci-lint run ./...