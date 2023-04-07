# Managed Service Identity

This login mode should be used in an environment where 
[Managed Service Identity](https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview) 
is available such as Azure Virtual Machine, Azure Virtual Machine ScaleSet, Cloud Shell, Azure Container Instance, and Azure App Service.

The token will not be cached on the filesystem.

## Usage Examples

### Using default Managed Service Identity

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l msi

kubectl get nodes
```

### Using Managed Service Identity with specific identity

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l msi --client-id <msi-client-id>

kubectl get nodes
```
