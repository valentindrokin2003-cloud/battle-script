.PHONY: fmt fmt-check lint vet test check

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "not gofmt-formatted:"; gofmt -l .; exit 1)

lint:
	golangci-lint run ./...

vet:
	go vet ./...

test:
	go test ./... -race

check: fmt-check vet lint test
