TARGET     := kubelogin
OS         := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH       := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
BIN         = bin/$(OS)_$(ARCH)/$(TARGET)
ifeq ($(OS),windows)
  BIN = bin/$(OS)_$(ARCH)/$(TARGET).exe
endif

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_HASH   := $(shell git rev-parse --verify HEAD)
GIT_TAG    := $(shell git describe --tags --exact-match --abbrev=0 2>/dev/null || echo "")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

ifdef GIT_TAG
	VERSION := $(GIT_TAG)/$(GIT_HASH)
else
	VERSION := $(GIT_BRANCH)/$(GIT_HASH)
endif

LDFLAGS    := -X main.version=$(VERSION) \
    -X main.goVersion=$(shell go version | cut -d " " -f 3) \
	-X main.buildTime=$(BUILD_TIME)

all: $(TARGET)

version:
	@echo VERSION: $(VERSION)

$(TARGET): clean
	go build -o $(BIN) -ldflags "$(LDFLAGS)"

clean:
	-rm -f $(BIN)
