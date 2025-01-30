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

GIT_TAG := $(shell git describe --tags --exact-match --abbrev=0 2>/dev/null || echo "")

LDFLAGS    := -X main.gitTag=$(GIT_TAG)

all: $(TARGET)

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

test: lint
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

version:
	@echo VERSION: $(VERSION)

$(TARGET): clean
	CGO_ENABLED=0 go build -o $(BIN) -ldflags "$(LDFLAGS)"

clean:
	-rm -f $(BIN)
