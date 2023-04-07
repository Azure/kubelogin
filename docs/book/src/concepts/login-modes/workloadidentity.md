# Workload Identity

This login mode uses [Azure AD federated identity credentials](https://docs.microsoft.com/en-us/graph/api/resources/federatedidentitycredentials-overview?view=graph-rest-beta) to authenticate to Kubernetes clusters with Azure AD integration. This works by setting the environment variables:
* `AZURE_CLIENT_ID` is Azure Active Directory application ID that is federated with workload identity
* `AZURE_TENANT_ID` is Azure Active Directory tenant ID
* `AZURE_FEDERATED_TOKEN_FILE` is the file containing signed assertion of workload identity. E.g. Kubernetes projected service account (jwt) token
* `AZURE_AUTHORITY_HOST` is the base URL of an Azure Active Directory authority. E.g. `https://login.microsoftonline.com/`

With workload identity, it's possible to access Kubernetes clusters from CI/CD system such as Github, ArgoCD, etc. without storing Service Principal credentials in those external systems. To learn more, [here](https://github.com/weinong/azure-federated-identity-samples) is a sample to setup OIDC federation from Github.

In this login mode, token will not be cached on the filesystem.

## Usage Examples

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l workloadidentity

kubectl get nodes
```
