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

## Usage Examples

### Basic Usage in Pipeline

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
      kubelogin convert-kubeconfig \
        --login azurepipelines \
        --tenant-id $(tenant-id) \
        --client-id $(client-id) \
        --server-id $(server-id) \
        --azure-pipelines-service-connection-id $(service-connection-resource-id)
      
      # Now kubectl commands will authenticate using Azure Pipelines credentials
      kubectl get nodes
    addSpnToEnvironment: true  # This enables SYSTEM_ACCESSTOKEN
```

### Direct Token Retrieval

```bash
# In Azure DevOps pipeline (with "Allow scripts to access the OAuth token" enabled)
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
- task: TerraformTaskV3@3
  displayName: 'Terraform Apply'
  inputs:
    provider: 'azurerm'
    command: 'apply'
    workingDirectory: '$(System.DefaultWorkingDirectory)/terraform'
    environmentServiceNameAzureRM: 'my-service-connection'
    commandOptions: |
      -auto-approve
  env:
    # Configure kubeconfig for kubectl provider in Terraform
    KUBECONFIG: $(Agent.TempDirectory)/kubeconfig
- script: |
    # Convert kubeconfig to use azurepipelines authentication
    kubelogin convert-kubeconfig \
      --login azurepipelines \
      --tenant-id $(tenant-id) \
      --client-id $(client-id) \
      --server-id $(server-id) \
      --azure-pipelines-service-connection-id $(service-connection-resource-id) \
      --kubeconfig $(Agent.TempDirectory)/kubeconfig
  displayName: 'Configure kubectl authentication'
  condition: always()
```

## How It Works

1. **Service Connection**: Azure DevOps service connections provide managed identity or service principal authentication to Azure resources
2. **System Access Token**: When "Allow scripts to access the OAuth token" is enabled, Azure Pipelines provides a `SYSTEM_ACCESSTOKEN` environment variable
3. **OIDC Integration**: The `azurepipelines` login method uses Azure SDK's `AzurePipelinesCredential` to exchange the system access token for an Azure AD token
4. **Token Caching**: Authentication tokens are cached to improve performance across multiple kubectl operations

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