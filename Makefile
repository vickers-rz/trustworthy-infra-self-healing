.PHONY: test run run-controller fmt vet install-crd sample

test:
	go test ./...

run:
	go run ./cmd/controlplane

run-controller:
	go run ./cmd/controller

fmt:
	gofmt -w ./api ./cmd ./internal

vet:
	go vet ./...

install-crd:
	kubectl apply -f config/crd/bases/infraheal.io_healingpolicies.yaml

sample:
	kubectl apply -f config/samples/infraheal_v1alpha1_healingpolicy.yaml
