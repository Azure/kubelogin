# remove-cache-dir

This subcommand removes the cached access/refresh token from filesystem. Note that only `devicelogin`, `interactive`, and `ropc` login modes will cache the token.

## Usage

```sh
kubelogin remove-cache-dir -h
Remove all cached authentication record from filesystem

Usage:
  kubelogin remove-cache-dir [flags]

Flags:
      --cache-dir string   directory to cache authentication record (default "/home/weinongw/.kube/cache/kubelogin/")
  -h, --help               help for remove-cache-dir

Global Flags:
      --logtostderr   log to standard error instead of files (default true)
  -v, --v Level       number for the log level verbosity
```
