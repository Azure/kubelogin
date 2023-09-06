# convert-kubeconfig

This subcommand converts kubeconfig to [Exec plugin](../concepts/exec-plugin.md) using `kubelogin get-token` with specified [login mode](../concepts/login-modes.md).

Note that when `--context` is specified, only the matching kubeconfig context will be converted. Otherwise, every kubeconfig context that uses azure auth or Exec plugin will be converted.

## Usage

```sh
kubelogin convert-kubeconfig -h
convert kubeconfig to use exec auth module

Usage:
  kubelogin convert-kubeconfig [flags]

Flags:
      --authority-host string                Workload Identity authority host. It may be specified in AZURE_AUTHORITY_HOST environment variable
      --azure-config-dir string              Azure CLI config path
      --client-certificate string            AAD client cert in pfx. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE or AZURE_CLIENT_CERTIFICATE_PATH environment variable
      --client-certificate-password string   Password for AAD client cert. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE_PASSWORD or AZURE_CLIENT_CERTIFICATE_PASSWORD environment variable
      --client-id string                     AAD client application ID. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_ID or AZURE_CLIENT_ID environment variable
      --client-secret string                 AAD client application secret. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_SECRET or AZURE_CLIENT_SECRET environment variable
      --context string                       The name of the kubeconfig context to use
  -e, --environment string                   Azure environment name (default "AzurePublicCloud")
      --federated-token-file string          Workload Identity federated token file. It may be specified in AZURE_FEDERATED_TOKEN_FILE environment variable
  -h, --help                                 help for convert-kubeconfig
      --identity-resource-id string          Managed Identity resource id.
      --kubeconfig string                    Path to the kubeconfig file to use for CLI requests.
      --legacy                               set to true to get token with 'spn:' prefix in audience claim
  -l, --login string                         Login method. Supported methods: devicecode, interactive, spn, ropc, msi, azurecli, workloadidentity. It may be specified in AAD_LOGIN_METHOD environment variable (default "devicecode")
      --password string                      password for ropc login flow. It may be specified in AAD_USER_PRINCIPAL_PASSWORD or AZURE_PASSWORD environment variable
      --pop-enabled                          set to true to request a proof-of-possession/PoP token, or false to request a regular bearer token. Only works with interactive and spn login modes. --pop-claims must be provided if --pop-enabled is true
      --pop-claims                           claims to include when requesting a PoP token, formatted as a comma-separated string of key=value pairs. Must include the u-claim, `u=ARM_ID` containing the ARM ID of the cluster (host). --pop-enabled must be set to true if --pop-claims are provided
      --server-id string                     AAD server application ID
  -t, --tenant-id string                     AAD tenant ID. It may be specified in AZURE_TENANT_ID environment variable
      --token-cache-dir string               directory to cache token (default "${HOME}/.kube/cache/kubelogin/")
      --use-azurerm-env-vars                 Use environment variable names of Terraform Azure Provider (ARM_CLIENT_ID, ARM_CLIENT_SECRET, ARM_CLIENT_CERTIFICATE_PATH, ARM_CLIENT_CERTIFICATE_PASSWORD, ARM_TENANT_ID)
      --username string                      user name for ropc login flow. It may be specified in AAD_USER_PRINCIPAL_NAME or AZURE_USERNAME environment variable

Global Flags:
      --logtostderr   log to standard error instead of files (default true)
  -v, --v Level       number for the log level verbosity
```
