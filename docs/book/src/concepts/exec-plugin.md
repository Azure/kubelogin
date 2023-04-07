# Exec Plugin

[Exec plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) 
is one of Kubernetes authentication strategies which allows `kubectl` to execute an external command to receive user credentials to send to api-server.
Since Kubernetes 1.26, [the default azure auth plugin is removed from `client-go` and `kubectl`](https://github.com/kubernetes/kubernetes/blob/ad18954259eae3db51bac2274ed4ca7304b923c4/CHANGELOG/CHANGELOG-1.26.md).

To interact with an Azure AD enabled Kubernetes cluster, Exec plugin using `kubelogin` will be required.

A kubeconfig using exec plugin will look somewhat like:

```yaml
kind: Config
preferences: {}
users:
  - name: user-name
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        command: kubelogin
        args:
          - get-token
          - --environment
          - AzurePublicCloud
          - --server-id
          - <AAD server app ID>
          - --client-id
          - <AAD client app ID>
          - --tenant-id
          - <AAD tenant ID>
```

When using `kubelogin` in Exec plugin, the kubeconfig tells `kubectl` to execute `kubelogin get-token` subcommand to perform various Azure AD [login modes](./login-modes.md) to get the access token.
