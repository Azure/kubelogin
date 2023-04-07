# remove-tokens

This subcommand removes the cached access/refresh token from filesystem. Note that only `devicelogin`, `interactive`, and `ropc` login modes will cache the token.

## Usage

```sh
kubelogin remove-tokens -h
Remove all cached tokens from filesystem

Usage:
  kubelogin remove-tokens [flags]

Flags:
  -h, --help                     help for remove-tokens
      --token-cache-dir string   directory to cache token (default "${HOME}/.kube/cache/kubelogin/")

Global Flags:
      --logtostderr   log to standard error instead of files (default true)
  -v, --v Level       number for the log level verbosity
```
