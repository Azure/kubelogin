# Azure Pipelines

This login mode uses Azure Pipelines service connections and the built-in `SYSTEM_ACCESSTOKEN` to authenticate with Azure AD. This is particularly useful when running kubelogin as an exec plugin within Azure DevOps pipelines, such as in Terraform deployments that need to interact with Azure Kubernetes Service clusters.

The authentication leverages Azure Pipelines' managed identity integration through service connections, providing a seamless way to authenticate without additional credential management.

> ### NOTE
>
> This login mode only works within Azure DevOps pipelines and requires proper pipeline configuration.

## Prerequisites

1. **Service Connection**: An Azure Resource Manager service connection configured in your Azure DevOps project
2. **Pipeline Configuration**: The pipeline must have "Allow scripts to access the OAuth token" enabled in the agent job settings
3. **Environment Variables**: The following environment variables must be available (automatically set by Azure Pipelines when OAuth token access is enabled):
   - `SYSTEM_ACCESSTOKEN`: The OAuth token provided by Azure Pipelines
   - `SYSTEM_OIDCREQUESTURI`: The OIDC request URI (automatically set by Azure Pipelines)

## Required Parameters

- `--tenant-id`: Azure AD tenant ID where the service connection is configured
- `--client-id`: Application ID of the client application (typically the AKS cluster's client ID)  
- `--server-id`: Application ID of the server/resource (typically the AKS cluster's server ID)
- `--azure-pipelines-service-connection-id`: The resource ID of the Azure Resource Manager service connection

> **Note**: When using Azure Pipelines tasks with Azure Resource Manager service connections, Azure Pipelines automatically sets the following environment variables which kubelogin will use if the corresponding flags are not provided:
> - `AZURESUBSCRIPTION_TENANT_ID` - Automatically used for `--tenant-id`
> - `AZURESUBSCRIPTION_CLIENT_ID` - Automatically used for `--client-id`
> - `AZURESUBSCRIPTION_SERVICE_CONNECTION_ID` - Automatically used for `--azure-pipelines-service-connection-id`
>
> This means you only need to provide the `--server-id` parameter when these environment variables are available.

## Usage Examples

### Basic Usage in Pipeline (Simplified with Environment Variables)

```yaml
# azure-pipelines.yml
steps:
- task: AzureCLI@2
  displayName: 'Deploy to AKS'
  inputs:
    azureSubscription: 'my-service-connection'
    scriptType: 'bash'
    scriptLocation: 'inlineScript'
    inlineScript: |
      # Configure kubeconfig to use azurepipelines login
      # tenant-id, client-id, and service-connection-id are automatically detected
      kubelogin convert-kubeconfig \
        --login azurepipelines \
        --server-id $(server-id)
      
      # Now kubectl commands will authenticate using Azure Pipelines credentials
      kubectl get nodes
```

### Basic Usage with Explicit Parameters

If you prefer to explicitly provide all parameters:

```yaml
# azure-pipelines.yml
steps:
- task: AzureCLI@2
  displayName: 'Deploy to AKS'
  inputs:
    azureSubscription: 'my-service-connection'
    scriptType: 'bash'
    scriptLocation: 'inlineScript'
    inlineScript: |
      # Configure kubeconfig to use azurepipelines login with explicit parameters
      kubelogin convert-kubeconfig \
        --login azurepipelines \
        --tenant-id $(tenant-id) \
        --client-id $(client-id) \
        --server-id $(server-id) \
        --azure-pipelines-service-connection-id $(service-connection-resource-id)
      
      # Now kubectl commands will authenticate using Azure Pipelines credentials
      kubectl get nodes
```

### Direct Token Retrieval

```bash
# In Azure DevOps pipeline (with "Allow scripts to access the OAuth token" enabled)
# Simplified version - uses environment variables automatically set by Azure Pipelines
kubelogin get-token \
  --login azurepipelines \
  --server-id <cluster-server-id>

# Or with explicit parameters
kubelogin get-token \
  --login azurepipelines \
  --tenant-id <tenant-id> \
  --client-id <client-id> \
  --server-id <cluster-server-id> \
  --azure-pipelines-service-connection-id <service-connection-resource-id>
```

### Terraform Integration

```yaml
# azure-pipelines.yml for Terraform deployments
steps:
- task: AzureCLI@2
  displayName: 'Terraform Apply'
  inputs:
    azureSubscription: 'my-service-connection'
    scriptType: 'bash'
    scriptLocation: 'inlineScript'
    inlineScript: |
      # Setup Terraform
      cd $(System.DefaultWorkingDirectory)/terraform
      terraform init
      
      # Configure kubeconfig for kubectl provider in Terraform
      # tenant-id, client-id, and service-connection-id are automatically detected
      kubelogin convert-kubeconfig \
        --login azurepipelines \
        --server-id $(server-id)
      
      # Run Terraform
      terraform apply -auto-approve
```

## Environment Variable Support

When using Azure Pipelines tasks with Azure Resource Manager service connections, Azure Pipelines automatically sets environment variables for the service connection. Kubelogin automatically detects and uses these variables:

| Environment Variable | Used For | Command-line Flag Equivalent |
|---------------------|----------|------------------------------|
| `AZURESUBSCRIPTION_TENANT_ID` | Tenant ID | `--tenant-id` |
| `AZURESUBSCRIPTION_CLIENT_ID` | Client ID | `--client-id` |
| `AZURESUBSCRIPTION_SERVICE_CONNECTION_ID` | Service Connection ID | `--azure-pipelines-service-connection-id` |

**Precedence**: Command-line flags always take precedence over environment variables. This allows you to override specific values when needed.

**Disabling Environment Variables**: You can use the `--disable-environment-override` flag to ignore all environment variables and require explicit parameters.

## How It Works

1. **Service Connection**: Azure DevOps service connections provide managed identity or service principal authentication to Azure resources
2. **System Access Token**: When "Allow scripts to access the OAuth token" is enabled, Azure Pipelines provides a `SYSTEM_ACCESSTOKEN` environment variable
3. **Environment Variables**: When using Azure Pipelines tasks with Azure Resource Manager service connections, Azure Pipelines automatically sets subscription-specific environment variables
4. **OIDC Integration**: The `azurepipelines` login method uses Azure SDK's `AzurePipelinesCredential` to exchange the system access token for an Azure AD token
5. **Token Caching**: Authentication tokens are cached to improve performance across multiple kubectl operations

## Troubleshooting

### Common Errors

- **"SYSTEM_ACCESSTOKEN environment variable not set"**: Enable "Allow scripts to access the OAuth token" in your pipeline job settings
- **"SYSTEM_OIDCREQUESTURI environment variable not set"**: This should be automatically set by Azure Pipelines; check your Azure DevOps version and configuration
- **"tenant ID is required"**: Provide the `--tenant-id` parameter
- **"--azure-pipelines-service-connection-id is required"**: Provide the service connection resource ID parameter

### Finding Service Connection Resource ID

The service connection resource ID can be found in the Azure DevOps portal:
1. Go to Project Settings → Service connections
2. Select your Azure Resource Manager service connection
3. The resource ID is displayed in the connection details

## References

- https://learn.microsoft.com/en-us/azure/devops/pipelines/process/system-and-variable-groups
- https://learn.microsoft.com/en-us/azure/devops/pipelines/library/service-endpoints
- https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#AzurePipelinesCredential