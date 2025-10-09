# Chained Credential

This login mode uses Azure SDK's `DefaultAzureCredential` which automatically tries multiple credential types in a chain to authenticate with Azure AD. This provides a seamless authentication experience by attempting various authentication methods in a predefined order until one succeeds.

The credential chain tries the following methods in order:

1. **Environment Credential** - Authenticates using environment variables (`AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`)
2. **Workload Identity Credential** - Authenticates using workload identity in Azure Kubernetes Service
3. **Managed Identity Credential** - Authenticates using managed identity assigned to the Azure resource
4. **Azure CLI Credential** - Authenticates using the logged-in user from Azure CLI (`az login`)

This login mode is particularly useful in scenarios where you want to support multiple authentication methods without explicitly specifying which one to use, or when deploying the same application across different environments that may use different authentication mechanisms.

> ### NOTE
>
> The chained credential does not require any additional configuration flags and will automatically detect and use the first available authentication method from the chain.

## Usage Examples

### Basic Usage

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l chained

kubectl get nodes
```

### Using Environment Variables for Service Principal

```sh
export KUBECONFIG=/path/to/kubeconfig
export AZURE_CLIENT_ID=<service-principal-client-id>
export AZURE_CLIENT_SECRET=<service-principal-client-secret>
export AZURE_TENANT_ID=<tenant-id>

kubelogin convert-kubeconfig -l chained

kubectl get nodes
```

### Using with Azure CLI Authentication

```sh
# First login with Azure CLI
az login

export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l chained

kubectl get nodes
```

### Using with Managed Identity in Azure

```sh
# On Azure VMs, Azure Container Instances, Azure App Service, etc.
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l chained

kubectl get nodes
```

### Using with Workload Identity in AKS

```sh
# In AKS pods with workload identity configured
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l chained

kubectl get nodes
```

### Direct Token Retrieval

```bash
kubelogin get-token \
  --login chained \
  --server-id <cluster-server-id>
```

## How It Works

The chained credential automatically detects the authentication method to use based on the environment:

1. **Checks Environment Variables**: If `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, and `AZURE_TENANT_ID` are set, it uses service principal authentication
2. **Checks Workload Identity**: If running in AKS with workload identity configured, it uses the federated token
3. **Checks Managed Identity**: If running on Azure resources with managed identity, it uses the metadata service
4. **Checks Azure CLI**: If Azure CLI is installed and logged in, it uses the cached credentials

The first successful method is used, and subsequent methods in the chain are not attempted.

## Benefits

- **Flexibility**: Works across different environments without code changes
- **Simplicity**: No need to specify authentication method explicitly
- **Fallback**: Automatically tries alternative methods if the primary method fails
- **Best Practices**: Follows Azure SDK's recommended authentication patterns

## Troubleshooting

### Common Issues

- **No authentication method available**: Ensure at least one of the supported authentication methods is properly configured
- **Multiple authentication methods configured**: The credential will use the first available method in the chain order
- **Permissions**: Ensure the authenticated identity has appropriate permissions to access the AKS cluster

### Debugging

To understand which authentication method is being used, check the kubelogin logs:

```bash
kubelogin get-token --login chained --server-id <server-id> -v 5
```

## References

- https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/credential-chains
- https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#DefaultAzureCredential
- https://learn.microsoft.com/en-us/dotnet/azure/sdk/authentication/credential-chains