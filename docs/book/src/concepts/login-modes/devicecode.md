# Device Code

This is the default login mode in `convert-kubeconfig` subcommand. So `-l devicecode` is optional. This login will prompt the device code for user to login on a browser. 

Before `kubelogin` and [Exec plugin](./concepts/exec-plugin.md) were introduced, the azure authentication mode in `kubectl` supports device code flow only. 
It uses an old library that produces the token with `audience` claim that has `spn:` prefix 
which is not compatible with AKS Managed AAD using On-Behalf-Of mode ([Issue86410](https://github.com/kubernetes/kubernetes/issues/86410)).
So when running `convert-kubeconfig` subcommand, `kubelogin` will remove the `spn:` prefix in `audience` claim.
If it's desired to keep the old behavior, add `--legacy`. 

If you are using kubeconfig from AKS Legacy AAD (AADv1) clusters, `kubelogin` will automatically add `--legacy` flag.

## Usage Examples

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig

kubectl get nodes

```

## Using Interactive Mode Instead

Device code login asks the user to open a URL and type a code by hand. Entra ID does not return the
optional `verification_uri_complete` field from [RFC 8628](https://datatracker.ietf.org/doc/html/rfc8628#section-3.3.1),
so the URL and the code cannot be combined into a single link. When the machine running `kubectl`
has a browser, [web browser interactive mode](./interactive.md) avoids the copying altogether.

If the kubeconfig was provisioned for you and already pins `--login=devicecode` in its exec args,
you don't have to edit it. Environment variables are applied after the flags are parsed, so
`AAD_LOGIN_METHOD` takes precedence over the login mode stored in the kubeconfig:

```sh
export AAD_LOGIN_METHOD=interactive

kubectl get nodes
```

Passing `--disable-environment-override` turns this off and keeps the login mode from the kubeconfig.

## Restrictions

- Device code login mode doesn't work when Conditional Access policy is configured on AAD tenant. Use [web browser interactive mode](./interactive.md) instead.


## References

- https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-device-code
