# Azure CLI

This login mode uses the already logged-in context performed by Azure CLI to get the [access token](https://docs.microsoft.com/en-us/cli/azure/account?view=azure-cli-latest#az_account_get_access_token). 
The token will be issued in the same Azure AD tenant as in `az login`. 

`kubelogin` will not cache any token since it's already managed by Azure CLI.

> ### NOTE
> This login mode only works with managed AAD in AKS.

## Usage Examples

```sh
az login

export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l azurecli

kubectl get nodes
```


## References

- https://learn.microsoft.com/en-us/cli/azure/
- https://github.com/Azure/azure-cli
