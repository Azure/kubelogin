# Azure CLI

This login mode uses the already logged-in context performed by Azure CLI to get the [access token](https://docs.microsoft.com/en-us/cli/azure/account?view=azure-cli-latest#az_account_get_access_token).
The token will be issued in the same Azure AD tenant as in `az login`.

`kubelogin` will not cache any token since it's already managed by Azure CLI.

> ### NOTE
>
> This login mode only works with managed AAD in AKS.

## Usage Examples

```sh
az login

export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l azurecli

kubectl get nodes
```

When Azure CLI's config directory is outside the `${HOME}` directory, `--azure-config-dir` should be specified in `convert-kubeconfig` subcommand. It will generate the kubeconfig with environment variable configured. The same thing can also be achieved by setting environment variable `AZURE_CONFIG_DIR` to this directory while running `kubectl` command.

## References

- https://learn.microsoft.com/en-us/cli/azure/
- https://github.com/Azure/azure-cli
