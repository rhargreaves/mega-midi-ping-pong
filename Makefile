ARGS :=

all: test

test:
	go test ./...

run:
	go run . $(ARGS)

deps:
	go mod download
	go mod tidy

.PHONY: all test run deps