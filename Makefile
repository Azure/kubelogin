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
	@echo "  all          - Build the kubelogin binary (default)"
	@echo "  $(TARGET)    - Build the kubelogin binary"
	@echo "  lint         - Run linting checks"
	@echo "  test         - Run tests (includes linting)"
	@echo "  clean        - Remove built binaries"
	@echo "  build-image  - Build Docker image with kubelogin binary"
	@echo "  changelog    - Generate a CHANGELOG.md entry for a new release"
	@echo ""
	@echo "Docker image build options:"
	@echo "  make build-image                    # Build with 'latest' tag"
	@echo "  GIT_TAG=v1.0.0 make build-image   # Build with specific tag"
	@echo ""
	@echo "Changelog generation options:"
	@echo "  VERSION=0.2.15 make changelog                    # auto-detect previous tag"
	@echo "  VERSION=0.2.15 SINCE_TAG=v0.2.14 make changelog # explicit previous tag"
	@echo ""
	@echo "Environment variables:"
	@echo "  GOOS         - Target OS (default: $(OS))"
	@echo "  GOARCH       - Target architecture (default: $(ARCH))"
	@echo "  GIT_TAG      - Git tag for version info and Docker tagging"
	@echo "  VERSION      - New version number for changelog generation"
	@echo "  SINCE_TAG    - Previous tag to compare from for changelog"
	@echo "  GITHUB_TOKEN - GitHub token for changelog generation (API access)"

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

test: lint
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

$(TARGET): clean
	CGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0) go build -o $(BIN) -ldflags "$(LDFLAGS)"

clean:
	-rm -f $(BIN)

changelog:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: VERSION=0.2.15 make changelog"; \
		exit 1; \
	fi
	GITHUB_TOKEN=$(GITHUB_TOKEN) go run hack/changelog-generator/main.go \
		--version="$(VERSION)" \
		$(if $(SINCE_TAG),--since-tag="$(SINCE_TAG)",) \
		--repo="Azure/kubelogin"

# Docker image build target
IMAGE_NAME    := ghcr.io/azure/kubelogin
IMAGE_TAG     := $(if $(GIT_TAG),$(GIT_TAG),latest)

build-image: $(TARGET)
	docker build --build-arg VERSION=$(IMAGE_TAG) -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@if [ "$(GIT_TAG)" != "" ]; then \
		docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest; \
	fi
