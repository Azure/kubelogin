# Development

## Prerequisites

### System Dependencies

kubelogin uses secure token storage that requires platform-specific libraries:

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install libsecret-1-0 libsecret-1-dev
```

#### Linux (CentOS/RHEL/Fedora)
```bash
# CentOS/RHEL
sudo yum install libsecret-devel

# Fedora
sudo dnf install libsecret-devel
```

#### macOS
No additional dependencies required (uses Keychain)

#### Windows
No additional dependencies required (uses Windows Credential Manager)

### Go Dependencies
- Go 1.23 or later
- Make

## Building

```bash
make build
```

## Testing

```bash
make test
```

**Note**: Tests require the system dependencies listed above. If you encounter errors related to `libsecret-1.so` or "encrypted storage isn't possible", ensure the libsecret library is installed.
