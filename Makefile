GOLANGCI_LINT = bin/golangci-lint

.PHONY: build
build:
	go build -o bin/owners ./cmd

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint: | $(GOLANGCI_LINT)
	$| run

.PHONY: docker-build
docker-build:
	docker build -t owners:latest .

$(GOLANGCI_LINT):
	GOBIN=$(shell pwd)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
