.DEFAULT_GOAL := all

include .bingo/Variables.mk

TARGET     := kubelogin
OS         := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH       := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOARM      := $(if $(GOARM),$(GOARM),)
BIN         = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET)
ifeq ($(OS),windows)
  BIN = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET).exe
endif

GIT_TAG    := $(if $(GIT_TAG),$(GIT_TAG),)

LDFLAGS    := -X main.gitTag=$(GIT_TAG)

all: $(TARGET)

help:
	@echo "Available targets:"
	@echo "  all                        - Build the kubelogin binary (default)"
	@echo "  $(TARGET)                  - Build the kubelogin binary"
	@echo "  lint                       - Run linting checks"
	@echo "  test                       - Run unit tests (includes linting)"
	@echo "  integration-test           - Run integration tests (bypasses cache)"
	@echo "  integration-test-with-output - Run integration tests with output saving (bypasses cache)"
	@echo "  clean                      - Remove built binaries"
	@echo "  clean-integration          - Remove integration test outputs"
	@echo "  build-image                - Build Docker image with kubelogin binary"
	@echo ""
	@echo "Docker image build options:"
	@echo "  make build-image                    # Build with 'latest' tag"
	@echo "  GIT_TAG=v1.0.0 make build-image   # Build with specific tag"
	@echo ""
	@echo "Environment variables:"
	@echo "  GOOS         - Target OS (default: $(OS))"
	@echo "  GOARCH       - Target architecture (default: $(ARCH))"
	@echo "  GIT_TAG      - Git tag for version info and Docker tagging"

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

test: lint
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# Integration tests - run with minimal output by default
integration-test: $(TARGET)
	cp $(BIN) $(TARGET)
	cd test/integration && go test -v -count=1 ./...

# Integration tests with output saving for manual verification
integration-test-with-output: $(TARGET)
	cp $(BIN) $(TARGET)
	cd test/integration && KUBELOGIN_SAVE_TEST_OUTPUT=true go test -v -count=1 ./...

# Clean integration test outputs
clean-integration:
	-rm -rf test/integration/convert/_output/*

$(TARGET): clean
	CGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0) go build -o $(BIN) -ldflags "$(LDFLAGS)"

clean:
	-rm -f $(BIN)
	-rm -f $(TARGET)

# Docker image build target
IMAGE_NAME    := ghcr.io/azure/kubelogin
IMAGE_TAG     := $(if $(GIT_TAG),$(GIT_TAG),latest)

build-image: $(TARGET)
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@if [ "$(GIT_TAG)" != "" ]; then \
		docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest; \
	fi
