# Web Browser Interactive

This login mode will automatically open a browser to login the user. 
Once authenticated, the browser will redirect back to a local web server with access token.
The redirect URL can be set via `--redirect-url`.
This login mode complies with Conditional Access policy.

## Usage Examples

### Bearer token with interactive flow

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l interactive

kubectl get nodes
```

### Specifying Redirect URL

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l interactive --redirect-url http://localhost:8080

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
