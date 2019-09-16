ARGS :=

all: test

test:
	go test

run:
	go run . $(ARGS)

deps:
	go get gitlab.com/gomidi/midi/...
	go get gitlab.com/gomidi/portmididrv

.PHONY: all test run deps