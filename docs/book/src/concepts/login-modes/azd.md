# Azure Developer CLI (azd)

This login mode uses the already logged-in context performed by Azure Developer CLI to get the access token.
The token will be issued in the same Azure AD tenant as in `azd auth login`.

`kubelogin` will not cache any token since it's already managed by Azure Developer CLI.

> ### NOTE
>
> This login mode only works with managed AAD in AKS.

## Usage Examples

```sh
azd auth login

export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l azd

kubectl get nodes
```

## References

- https://learn.microsoft.com/azure/developer/azure-developer-cli/overview
- https://github.com/azure/azure-dev
