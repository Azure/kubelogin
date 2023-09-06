# Service Principal

This login mode uses the service principal to login. The credential may be provided via environment variables or flag.
The supported credentials are password and pfx client certificate.

The token will not be cached on the filesystem.

## Usage Examples

### Client secret in environment variable

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l spn

export AAD_SERVICE_PRINCIPAL_CLIENT_ID=<spn client id>
export AAD_SERVICE_PRINCIPAL_CLIENT_SECRET=<spn secret>

kubectl get nodes
```

### Client secret in environment variable

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l spn

export AZURE_CLIENT_ID=<spn client id>
export AZURE_CLIENT_SECRET=<spn secret>

kubectl get nodes
```

### Client secret in command-line flag

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l spn --client-id <spn client id> --client-secret <spn client secret>

kubectl get nodes
```

> ### Warning
> this will leave the secret in the kubeconfig

### Client certificate

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l spn

export AAD_SERVICE_PRINCIPAL_CLIENT_ID=<spn client id>
export AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE=/path/to/cert.pfx
export AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE_PASSWORD=<pfx password>

kubectl get nodes
```

### Client certificate

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l spn

export AZURE_CLIENT_ID=<spn client id>
export AZURE_CLIENT_CERTIFICATE_PATH=/path/to/cert.pfx
export AZURE_CLIENT_CERTIFICATE_PASSWORD=<pfx password>

kubectl get nodes
```

### Proof-of-possession (PoP) token with client secret from environment variables
```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l spn --pop-enabled --pop-claims "u=/ARM/ID/OF/CLUSTER"

export AAD_SERVICE_PRINCIPAL_CLIENT_ID=<spn client id>
export AAD_SERVICE_PRINCIPAL_CLIENT_SECRET=<spn secret>

kubectl get nodes
```

## Restrictions

- on AKS, it will only work with managed AAD
- the service principal can be member of [maximum 200 AAD groups](https://learn.microsoft.com/en-us/azure/active-directory/hybrid/how-to-connect-fed-group-claims) 
