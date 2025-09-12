SHELL := /bin/bash

.PHONY: run test bench vet fmt hooks-install hooks-uninstall

build:
	go build -o bin/uniswap-estimator cmd/api/main.go

clean:
	rm -rf bin

tidy:
	go mod tidy

golangci-lint:
	golangci-lint run

test:
	go test ./...

bench:
	go test -run=^$$ -bench=. -benchmem ./pkg/uniswapv2

vet:
	go vet ./...

fmt:
	gofmt -s -w .

hooks-install:
	chmod +x .githooks/* || true
	git config core.hooksPath .githooks
	@echo "Git hooks installed (core.hooksPath=.githooks)"

hooks-uninstall:
	git config --unset core.hooksPath || true
	@echo "Git hooks path unset"
