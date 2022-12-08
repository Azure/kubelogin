#!/bin/bash -e

# this script is invoked by the Makefile and is not used in the pipeline as it's done by a github actiono
# so this script is intended for local testing
if ! command -v golangci-lint &> /dev/null
then
    echo "WARNING: golangci-lint is not found. Hence linting is skipped. Visit https://golangci-lint.run/usage/install/#local-installation to install"
else
    golangci-lint run
fi

