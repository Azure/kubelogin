# Dockerfile for kubelogin
# This is a simple Dockerfile that copies a pre-built binary into a minimal scratch image.
# The binary should be built before running docker build using: make kubelogin
# 
# Usage:
#   make build-image                    # Build with latest tag
#   GIT_TAG=v1.0.0 make build-image   # Build with specific tag
#
FROM scratch

# Copy the pre-built binary from local build to /usr/local/bin
COPY bin/linux_amd64/kubelogin /usr/local/bin/kubelogin

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/kubelogin"]