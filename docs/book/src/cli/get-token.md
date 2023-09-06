# get-token

This subcommand uses specified [login mode](../concepts/login-modes.md) to authenticate with Azure AD and return the access token to standard out.

## Usage

```sh
kubelogin get-token -h
get AAD token

Usage:
  kubelogin get-token [flags]

Flags:
      --authority-host string                Workload Identity authority host. It may be specified in AZURE_AUTHORITY_HOST environment variable
      --client-certificate string            AAD client cert in pfx. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE or AZURE_CLIENT_CER
TIFICATE_PATH environment variable
      --client-certificate-password string   Password for AAD client cert. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE_PASSWORD or A
ZURE_CLIENT_CERTIFICATE_PASSWORD environment variable
      --client-id string                     AAD client application ID. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_ID or AZURE_CLIENT_ID environment variable
      --client-secret string                 AAD client application secret. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_SECRET or AZURE_CLIENT_S
ECRET environment variable
  -e, --environment string                   Azure environment name (default "AzurePublicCloud")
      --federated-token-file string          Workload Identity federated token file. It may be specified in AZURE_FEDERATED_TOKEN_FILE environment variable
  -h, --help                                 help for get-token
      --identity-resource-id string          Managed Identity resource id.
      --legacy                               set to true to get token with 'spn:' prefix in audience claim
  -l, --login string                         Login method. Supported methods: devicecode, interactive, spn, ropc, msi, azurecli, workloadidentity. It may be specified in A
AD_LOGIN_METHOD environment variable (default "devicecode")
      --password string                      password for ropc login flow. It may be specified in AAD_USER_PRINCIPAL_PASSWORD or AZURE_PASSWORD environment variable
      --pop-enabled                          set to true to request a proof-of-possession/PoP token, or false to request a regular bearer token. Only works with interactive and spn login modes. --pop-claims must be provided if --pop-enabled is true
      --pop-claims                           claims to include when requesting a PoP token, formatted as a comma-separated string of key=value pairs. Must include the u-claim, `u=ARM_ID` containing the ARM ID of the cluster (host). --pop-enabled must be set to true if --pop-claims are provided
      --server-id string                     AAD server application ID
  -t, --tenant-id string                     AAD tenant ID. It may be specified in AZURE_TENANT_ID environment variable
      --token-cache-dir string               directory to cache token (default "${HOME}/.kube/cache/kubelogin/")
      --use-azurerm-env-vars                 Use environment variable names of Terraform Azure Provider (ARM_CLIENT_ID, ARM_CLIENT_SECRET, ARM_CLIENT_CERTIFICATE_PATH, ARM
_CLIENT_CERTIFICATE_PASSWORD, ARM_TENANT_ID)
      --username string                      user name for ropc login flow. It may be specified in AAD_USER_PRINCIPAL_NAME or AZURE_USERNAME environment variable

Global Flags:
      --logtostderr   log to standard error instead of files (default true)
  -v, --v Level       number for the log level verbosity
```

## Exec Plugin Examples

> cluster info including cluster CA and FQDN are omitted in below examples

### Device Code Flow (default)

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
### web browser Flow (default)

```yaml
kind: Config
preferences: {}
users:
  - name: user-name
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args:
        - get-token
        - --login
        - interactive
        - --server-id
        - <AAD server app ID>
        - --client-id
        - <AAD client app ID>
        - --tenant-id
        - <AAD tenant ID>
        - --environment
        - AzurePublicCloud
        command: kubelogin
```

### Spn login with secret

```yaml
kind: Config
preferences: {}
users:
  - name: demouser
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args:
          - get-token
          - --environment
          - AzurePublicCloud
          - --server-id
          - <AAD server app ID>
          - --client-id
          - <AAD client app ID>
          - --client-secret
          - <client_secret>
          - --tenant-id
          - <AAD tenant ID>
          - --login
          - spn
        command: kubelogin
        env: null
```

### Spn login with pfx certificate

```yaml
kind: Config
preferences: {}
users:
  - name: demouser
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args:
          - get-token
          - --environment
          - AzurePublicCloud
          - --server-id
          - <AAD server app ID>
          - --client-id
          - <AAD client app ID>
          - --client-certificate
          - <client_certificate_path>
          - --tenant-id
          - <AAD tenant ID>
          - --login
          - spn
        command: kubelogin
        env: null
```

### Managed Service Identity

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
          - --server-id
          - <AAD server app ID>
          - --login
          - msi
```

### Managed Service Identity with specific client ID

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
          - --server-id
          - <AAD server app ID>
          - --client-id
          - <MSI app ID>
          - --login
          - msi
```

### Azure CLI token login

```yaml
kind: Config
preferences: {}
users:
  - name: demouser
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args:
          - get-token
          - --server-id
          - <AAD server app ID>
          - --login
          - azurecli
        command: kubelogin
        env: null
```

### Workload Identity

```yaml
kind: Config
preferences: {}
users:
  - name: demouser
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args:
          - get-token
          - --server-id
          - <AAD server app ID>
          - --login
          - workloadidentity
        command: kubelogin
        env: null
```
