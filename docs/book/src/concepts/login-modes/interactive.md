# Web Browser Interactive

This login mode will automatically open a browser to login the user. 
Once authenticated, the browser will redirect back to a local web server with the credentials. 
This login mode complies with Conditional Access policy.

In this login mode, the access token will be cached at `${HOME}/.kube/cache/kubelogin` directory. This path can be overriden by `--token-cache-dir`.

## Usage Examples

### Bearer token with interactive flow
```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l interactive

kubectl get nodes
```

### Proof-of-possession (PoP) token with interactive flow
```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l interactive --pop-enabled --pop-claims "u=/ARM/ID/OF/CLUSTER"

kubectl get nodes
```

## References

- https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.interactivebrowsercredential?view=azure-python
