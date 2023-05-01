# Device Code

This is the default login mode in `convert-kubeconfig` subcommand. So `-l devicecode` is optional. This login will prompt the device code for user to login on a browser. 

Before `kubelogin` and [Exec plugin](./concepts/exec-plugin.md) were introduced, the azure authentication mode in `kubectl` supports device code flow only. 
It uses an old library that produces the token with `audience` claim that has `spn:` prefix 
which is not compatible with AKS Managed AAD using On-Behalf-Of mode ([Issue86410](https://github.com/kubernetes/kubernetes/issues/86410)).
So when running `convert-kubeconfig` subcommand, `kubelogin` will remove the `spn:` prefix in `audience` claim.
If it's desired to keep the old behavior, add `--legacy`. 

If you are using kubeconfig from AKS Legacy AAD (AADv1) clusters, `kubelogin` will automatically add `--legacy` flag.

In this login mode, the access token and refresh token will be cached at `${HOME}/.kube/cache/kubelogin` directory. This path can be overriden by `--token-cache-dir`.

## Usage Examples

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig

kubectl get nodes

# clean up cached token
kubelogin remove-tokens
```

## Restrictions

- Device code login mode doesn't work when Conditional Access policy is configured on AAD tenant. Use [web browser interactive mode](./interactive.md) instead.


## References

- https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-device-code
