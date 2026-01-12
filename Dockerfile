# Dockerfile for kubelogin
# This Dockerfile copies a pre-built binary into a minimal scratch image.
# The binary should be built before running docker build using: make kubelogin
# 
# Usage:
#   make build-image                    # Build with latest tag
#   GIT_TAG=v1.0.0 make build-image   # Build with specific tag
#
FROM scratch

# Build arguments for multi-architecture support
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=""

# OpenContainers Image Spec labels
LABEL org.opencontainers.image.source="https://github.com/Azure/kubelogin"
LABEL org.opencontainers.image.description="Kubernetes credential plugin for Azure authentication"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.version="${VERSION}"

# Copy the pre-built binary from local build to /usr/local/bin
COPY bin/${TARGETOS}_${TARGETARCH}/kubelogin /usr/local/bin/kubelogin

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/kubelogin"]
