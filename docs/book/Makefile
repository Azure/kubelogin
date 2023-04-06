TOOLS_BIN_DIR ?= $(PWD)/bin
# include tools bin dir in path so that mdbook-toc can be run by mdbook
PATH := ${PATH}:${TOOLS_BIN_DIR}
MDBOOK_VERSION ?= v0.4.27
# this version of mdbook-toc is built against mdbook 0.4.27
MDBOOK_TOC_VERSION ?= 0.11.2
MDBOOK_INSTALL := $(realpath ../../hack/install-mdbook.sh)
MDBOOK_TOC_INSTALL := $(realpath ../../hack/install-mdbook-toc.sh)

MDBOOK := $(TOOLS_BIN_DIR)/mdbook
$(MDBOOK):
	$(MDBOOK_INSTALL) ${MDBOOK_VERSION} ${TOOLS_BIN_DIR}

MDBOOK_TOC := $(TOOLS_BIN_DIR)/mdbook-toc
$(MDBOOK_TOC):
	$(MDBOOK_TOC_INSTALL) ${MDBOOK_TOC_VERSION} ${TOOLS_BIN_DIR}

DEPS := $(MDBOOK) $(MDBOOK_TOC)

.PHONY: build
build: $(DEPS)
	$(MDBOOK) build

.PHONY: serve
serve: $(DEPS)
	$(MDBOOK) serve
