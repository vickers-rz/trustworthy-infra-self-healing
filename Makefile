LOCALBIN ?= $(CURDIR)/bin
CONTROLLER_TOOLS_VERSION ?= v0.19.0
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)

.PHONY: test run run-controller fmt vet manifests generate verify-generated install-crd sample

test:
	go test ./...

run:
	go run ./cmd/controlplane

run-controller:
	go run ./cmd/controller

fmt:
	go fmt ./...

vet:
	go vet ./...

manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) rbac:roleName=infraheal-healingpolicy-controller crd paths="./..." output:crd:artifacts:config=config/crd/bases output:rbac:artifacts:config=config/rbac

generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

verify-generated: manifests generate
	git diff --exit-code -- api/v1alpha1/zz_generated.deepcopy.go config/crd/bases config/rbac

install-crd:
	kubectl apply -f config/crd/bases/infraheal.io_healingpolicies.yaml

sample:
	kubectl apply -f config/samples/infraheal_v1alpha1_healingpolicy.yaml

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

$(CONTROLLER_GEN): | $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
	mv $(LOCALBIN)/controller-gen $(CONTROLLER_GEN)
