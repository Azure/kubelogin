# Command Line Tool

`kubelogin` command-line tool has following subcommands:

```sh
kubelogin -h
login to azure active directory and populate kubeconfig with AAD tokens

Usage:
  kubelogin [flags]
  kubelogin [command]

Available Commands:
  completion         Generate the autocompletion script for the specified shell
  convert-kubeconfig convert kubeconfig to use exec auth module
  get-token          get AAD token
  help               Help about any command
  remove-tokens      Remove all cached tokens from filesystem

Flags:
  -h, --help          help for kubelogin
      --logtostderr   log to standard error instead of files (default true)
  -v, --v Level       number for the log level verbosity
      --version       version for kubelogin

Use "kubelogin [command] --help" for more information about a command.

```

Following sections provide in-depth information on these subcommands:

* [`kubelogin convert-kubeconfig`](./cli/convert-kubeconfig.md) - converts the kubeconfig to different login mode
* [`kubelogin get-token`](./cli/get-token.md) - gets the Azure AD token based on configured login mode. This subcommand is typically used in kubeconfig via [exec plugin](./concepts/exec-plugin.md) and is invoked by kubectl or any command-line tool, such as helm, implementing exec plugin.
* [`kubelogin remove-tokens`](./cli/remove-tokens.md) - removes the cached token on the filesystem
