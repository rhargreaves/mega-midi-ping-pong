all: test

test:
	go test

run:
	go run .

.PHONY: all test run