# Multi-stage Dockerfile for kubelogin
# Stage 1: Copy pre-built binary from local build
FROM scratch

# Copy the binary from local build to /usr/local/bin
COPY bin/linux_amd64/kubelogin /usr/local/bin/kubelogin

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/kubelogin"]