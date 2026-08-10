.PHONY: test run fmt vet

test:
	go test ./...

run:
	go run ./cmd/controlplane

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...
