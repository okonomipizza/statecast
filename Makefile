.PHONY: build

build:
	go build -o statecast ./cmd/statecast

.PHONY: test

test:
	go test ./...

.PHONY: lint

lint:
	go vet ./...
	@test -z "$$(gofmt -l . | grep -v '^\.go/')"