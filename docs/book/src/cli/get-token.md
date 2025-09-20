# get-token

This subcommand uses specified [login mode](../concepts/login-modes.md) to authenticate with Azure AD and return the access token to standard out.

## Usage

```sh
kubelogin get-token -h
get AAD token

Usage:
  kubelogin get-token [flags]

Flags:
      --authority-host string                          Workload Identity authority host. It may be specified in AZURE_AUTHORITY_HOST environment variable
      --azure-pipelines-service-connection-id string   Service connection (resource) ID used by azurepipelines login method
      --cache-dir string                               directory to cache authentication record (default "/home/weinongw/.kube/cache/kubelogin/")
      --client-certificate string            AAD client cert in pfx. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE or AZURE_CLIENT_CERTIFICATE_PATH environment variable
      --client-certificate-password string   Password for AAD client cert. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE_PASSWORD or AZURE_CLIENT_CERTIFICATE_PASSWORD environment variable
      --client-id string                     AAD client application ID. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_ID or AZURE_CLIENT_ID environment variable
      --client-secret string                 AAD client application secret. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_SECRET or AZURE_CLIENT_SECRET environment variable
      --disable-environment-override         Enable or disable the use of env-variables. Default false
      --disable-instance-discovery           set to true to disable instance discovery in environments with their own simple Identity Provider (not AAD) that do not have instance metadata discovery endpoint. Default false
  -e, --environment string                   Azure environment name (default "AzurePublicCloud")
      --federated-token-file string          Workload Identity federated token file. It may be specified in AZURE_FEDERATED_TOKEN_FILE environment variable
  -h, --help                                 help for get-token
      --identity-resource-id string          Managed Identity resource id.
      --legacy                               set to true to get token with 'spn:' prefix in audience claim
  -l, --login string                         Login method. Supported methods: devicecode, interactive, spn, ropc, msi, azurecli, azd, workloadidentity, azurepipelines, chained. It may be specified in AAD_LOGIN_METHOD environment variable (default "devicecode")
      --login-hint string                    The login hint to pre-fill the username in the interactive login flow.
      --password string                      password for ropc login flow. It may be specified in AAD_USER_PRINCIPAL_PASSWORD or AZURE_PASSWORD environment variable
      --pop-claims key=val,key2=val2         contains a comma-separated list of claims to attach to the pop token in the format key=val,key2=val2. At minimum, specify the ARM ID of the cluster as `u=ARM_ID`
      --pop-enabled                          set to true to use a PoP token for authentication or false to use a regular bearer token
      --redirect-url string                  The URL Microsoft Entra ID will redirect to with the access token. This is only used for interactive login. This is an optional parameter.
      --server-id string                     AAD server application ID
  -t, --tenant-id string                     AAD tenant ID. It may be specified in AZURE_TENANT_ID environment variable
      --timeout duration                     Timeout duration for Azure CLI token requests. It may be specified in AZURE_CLI_TIMEOUT environment variable (default 30s)
      --use-azurerm-env-vars                 Use environment variable names of Terraform Azure Provider (ARM_CLIENT_ID, ARM_CLIENT_SECRET, ARM_CLIENT_CERTIFICATE_PATH, ARM_CLIENT_CERTIFICATE_PASSWORD, ARM_TENANT_ID)
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

### Azure Pipelines

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
          - --client-id
          - <AAD client app ID>
          - --tenant-id
          - <AAD tenant ID>
          - --login
          - azurepipelines
          - --azure-pipelines-service-connection-id
          - <service connection resource ID>
        command: kubelogin
        env: null
```

### Chained Credential

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
          - --client-id
          - <AAD client app ID>
          - --tenant-id
          - <AAD tenant ID>
          - --login
          - chained
        command: kubelogin
        env: null
```
