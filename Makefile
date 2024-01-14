#!make

all: build

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	rm -rf ./bin/*

.PHONY: build
build: clean fmt
	mkdir -p ./bin/
	go build -o ./bin/cf ./cf.go

.PHONY: run
run: clean fmt vet
	go run ./...

.PHONY: test
test: vet
	go test -v  -failfast ./...
