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

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

test: lint
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

$(TARGET): clean
	CGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0) go build -o $(BIN) -ldflags "$(LDFLAGS)"

clean:
	-rm -f $(BIN)

# Docker image build target
IMAGE_NAME    := ghcr.io/azure/kubelogin
IMAGE_TAG     := $(if $(GIT_TAG),$(GIT_TAG),latest)

build-image: $(TARGET)
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@if [ "$(GIT_TAG)" != "" ]; then \
		docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest; \
	fi
