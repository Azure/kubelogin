# Using in different environments

`kubelogin` supports Azure Environments:

- AzurePublicCloud (default value)
- AzureChinaCloud
- AzureUSGovernmentCloud
- AzureStackCloud

You can specify `--environment` in `kubelogin convert-kubeconfig`.

When using `AzureStackCloud` you will need to specify the actual endpoints in a config file, and set the environment variable `AZURE_ENVIRONMENT_FILEPATH` to that file.

The configuration parameters of this file:

```json
{
  "name": "AzureStackCloud",
  "managementPortalURL": "...",
  "publishSettingsURL": "...",
  "serviceManagementEndpoint": "...",
  "resourceManagerEndpoint": "...",
  "activeDirectoryEndpoint": "...",
  "galleryEndpoint": "...",
  "keyVaultEndpoint": "...",
  "graphEndpoint": "...",
  "serviceBusEndpoint": "...",
  "batchManagementEndpoint": "...",
  "storageEndpointSuffix": "...",
  "sqlDatabaseDNSSuffix": "...",
  "trafficManagerDNSSuffix": "...",
  "keyVaultDNSSuffix": "...",
  "serviceBusEndpointSuffix": "...",
  "serviceManagementVMDNSSuffix": "...",
  "resourceManagerVMDNSSuffix": "...",
  "containerRegistryDNSSuffix": "...",
  "cosmosDBDNSSuffix": "...",
  "tokenAudience": "...",
  "resourceIdentifiers": {
    "graph": "...",
    "keyVault": "...",
    "datalake": "...",
    "batch": "...",
    "operationalInsights": "..."
  }
}
```

The full configuration is available in the source code at <https://github.com/Azure/go-autorest/blob/main/autorest/azure/environments.go>.
