ARGS :=

all: test

test:
	go test

run:
	go run . $(ARGS)

deps:
	go get github.com/rakyll/portmidi
	go get github.com/bradhe/stopwatch

.PHONY: all test run deps