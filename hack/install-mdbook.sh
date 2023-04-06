#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

VERSION=${1}
OUTPUT_PATH=${2}

# Ensure the output folder exists
mkdir -p "${OUTPUT_PATH}"

RELEASE_NAME=""
case "$OSTYPE" in
  darwin*) RELEASE_NAME="x86_64-apple-darwin.tar.gz"  ;;
  linux*)  RELEASE_NAME="x86_64-unknown-linux-gnu.tar.gz" ;;
  *)       echo "No mdBook release available for: $OSTYPE" && exit 1;;
esac

# Download and extract the mdBook release
curl -L "https://github.com/rust-lang/mdBook/releases/download/${VERSION}/mdbook-${VERSION}-${RELEASE_NAME}" | tar -xvz -C "${OUTPUT_PATH}"
