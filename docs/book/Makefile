TOOLS_BIN_DIR ?= $(PWD)/bin
MDBOOK_VERSION ?= v0.4.28
MDBOOK_INSTALL := $(realpath ../../hack/install-mdbook.sh)

MDBOOK := $(TOOLS_BIN_DIR)/mdbook
$(MDBOOK):
	$(MDBOOK_INSTALL) ${MDBOOK_VERSION} ${TOOLS_BIN_DIR}

DEPS := $(MDBOOK)

.PHONY: build
build: $(DEPS)
	$(MDBOOK) build

.PHONY: serve
serve: $(DEPS)
	$(MDBOOK) serve
