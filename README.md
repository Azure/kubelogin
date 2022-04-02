# kubelogin

This is a [client-go credential (exec) plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) implementing azure authentication. This plugin provides features that are not available in kubectl. It is supported on kubectl v1.11+

## Features

- `convert-kubeconfig` command to converts kubeconfig with existing azure auth provider format to exec credential plugin format
- [device code login](<#device-code-flow-interactive>)
- [non-interactive service principal login](<#service-principal-login-flow-non-interactive>)
- [non-interactive user principal login](<#user-principal-login-flow-non-interactive>) using [Resource owner login flow](https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth-ropc)
- [non-interactive managed service identity login](<#managed-service-identity-non-interactive>)
- [non-interactive Azure CLI token login (AKS only)](<#azure-cli-token-login-non-interactive>)
- [non-interactive workload identity login](<#azure-workload-federated-identity-non-interactive>)
- AAD token will be cached locally for renewal in device code login and user principal login (ropc) flow. By default, it is saved in `~/.kube/cache/kubelogin/`
- addresses <https://github.com/kubernetes/kubernetes/issues/86410> to remove `spn:` prefix in `audience` claim, if necessary. (based on kubeconfig or commandline argument `--legacy`)
- [Setup for Kubernetes OIDC Provider using Azure AD](<#setup-for-kubernetes-oidc-provider-using-azure-ad>)

## Getting Started

### Setup

Copy the latest [Releases](https://github.com/Azure/kubelogin/releases) to shell's search path.

### Setup (homebrew)

```sh
# install
brew install Azure/kubelogin/kubelogin

# upgrade
brew update
brew upgrade Azure/kubelogin/kubelogin
```

### Run

#### Device code flow (interactive)

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig

kubectl get no
```

If you are using kubeconfig from AKS AADv1 clusters, `convert-kubeconfig` command will automatically add `--legacy` flag so that `audience` claim will have `spn:` prefix.

#### Service principal login flow (non interactive)

> On AKS, it will only work with managed AAD. Service principal can be member of maximum 250 AAD groups.

Create a service principal or use an existing one.

```sh
az ad sp create-for-rbac --skip-assignment --name myAKSAutomationServicePrincipal
```

The output is similar to the following example.

```json
{
  "appId": "<spn client id>",
  "displayName": "myAKSAutomationServicePrincipal",
  "name": "http://myAKSAutomationServicePrincipal",
  "password": "<spn secret>",
  "tenant": "<aad tenant id>"
}
```

Query your service principal AAD Object ID by using the command below.

```sh
az ad sp show --id <spn client id> --query "objectId"
```

To configure the role binding on Azure Kubernetes Service, the user in rolebinding should be the AAD Object ID.

For example,

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: sp-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: <service-principal-object-id>
```

Use Kubelogin to convert your kubeconfig

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l spn

export AAD_SERVICE_PRINCIPAL_CLIENT_ID=<spn client id>
export AAD_SERVICE_PRINCIPAL_CLIENT_SECRET=<spn secret>

kubectl get no
```

#### User Principal login flow (non interactive)

> Note: ROPC is not supported in hybrid identity federation scenarios (for example, Azure AD and ADFS used to authenticate on-premises accounts). If users are full-page redirected to an on-premises identity providers, Azure AD is not able to test the username and password against that identity provider. Pass-through authentication is supported with ROPC, however.
> It also does not work when MFA policy is enabled
> Personal accounts that are invited to an Azure AD tenant can't use ROPC

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l ropc

export AAD_USER_PRINCIPAL_NAME=foo@bar.com
export AAD_USER_PRINCIPAL_PASSWORD=<password>

kubectl get no
```

#### Managed Service Identity (non interactive)

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l msi

kubectl get no
```

To configure the role binding on Azure Kubernetes Service, the user in rolebinding should be the MSI's AAD Object ID.

For example,

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: msi-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: <service-principal-object-id>
```

#### Managed Service Identity with specific identity (non interactive)

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l msi --client-id msi-client-id

kubectl get no
```

#### Azure CLI token login (non interactive)

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l azurecli

kubectl get no
```

Uses an [access token](https://docs.microsoft.com/en-us/cli/azure/account?view=azure-cli-latest#az_account_get_access_token) from [Azure CLI](https://github.com/Azure/azure-cli) to log in. The token will be issued against whatever tenant was logged in at the time `kubelogin convert-kubeconfig -l azurecli` was run. This login option only works with managed AAD in AKS.

#### Azure Workload Federated Identity (non interactive)

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l workloadidentity

kubectl get no
```

Workload identity uses [Azure AD federated identity credentials](https://docs.microsoft.com/en-us/graph/api/resources/federatedidentitycredentials-overview?view=graph-rest-beta) to authenticate to Kubernetes clusters with AAD integration. This works by setting the environment variables:
* `AZURE_CLIENT_ID` is Azure Active Directory application ID that is federated with workload identity
* `AZURE_TENANT_ID` is Azure Active Directory tenant ID
* `AZURE_FEDERATED_TOKEN_FILE` is the file containing signed assertion of workload identity. E.g. Kubernetes projected service account (jwt) token
* `AZURE_AUTHORITY_HOST` is the base URL of an Azure Active Directory authority. E.g. `https://login.microsoftonline.com/`

### Clean up

Whenever you want to remove cached tokens

```sh
kubelogin remove-tokens
```

## Azure Environment

`kubelogin` supports Azure Environments:

- AzurePublicCloud (default value)
- AzureChinaCloud
- AzureUSGovernmentCloud
- AzureStackCloud

You can specify `--environment` for `kubelogin convert-kubeconfig`.

When using `AzureStackCloud` you will need to specify the actual endpoints in a config file, and set the environment variable `AZURE_ENVIRONMENT_FILEPATH` to that file.

The configuration parameters of this file:

```
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

The full configuration is available in the source code at <https://github.com/Azure/go-autorest/blob/master/autorest/azure/environments.go>.

## Exec Plugin Format

Below is what a kubeconfig with exec plugin would look like. By default, the `audience` claim will not have `spn:` prefix. If it's desired to keep the prefix, add `--legacy` to the args.

Note: The AAD server app ID of AKS Managed AAD is always `6dae42f8-4368-4678-94ff-3960e28e3630` in any environments.

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

## Setup for Kubernetes OIDC Provider using Azure AD

Kubelogin can be used to authenticate to general kubernetes clusters using AAD as an OIDC provider. 

1. Create an AAD Enterprise Application and the corresponding App Registration. Check the `Allow public client flows` checkbox. Configure groups to be included in the response. Take a note of the directory (tenant) ID as `$AAD_TENANT_ID` and the application (client) ID as `$AAD_CLIENT_ID`
2. Configure the API server with the following flags:
* Issuer URL: `--oidc-issuer-url=https://sts.windows.net/$AAD_TENANT_ID`
* Client ID: `--oidc-client-id=$AAD_CLIENT_ID`
* Username claim: `--oidc-username-claim=upn`

See the [kubernetes docs for optional flags](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server). For EKS clusters [configure this on the Management Console](https://docs.amazonaws.cn/en_us/eks/latest/userguide/authenticate-oidc-identity-provider.html) or via [terraform](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_identity_provider_config).

3. Configure kubelogin to use the application from the first step:
```
kubectl config set-credentials "azure-user" \
  --exec-api-version=client.authentication.k8s.io/v1beta1 \
  --exec-command=kubelogin \
  --exec-arg=get-token \
  --exec-arg=--environment \
  --exec-arg=AzurePublicCloud \
  --exec-arg=--server-id \
  --exec-arg=$AAD_CLIENT_ID \
  --exec-arg=--client-id \
  --exec-arg=$AAD_CLIENT_ID \
  --exec-arg=--tenant-id \
  --exec-arg=$AAD_TENANT_ID
```
4. Use this credential to connect to the cluster:
```
kubectl config set-context "$CLUSTER_NAME" --cluster="$CLUSTER_NAME" --user=azure-user
kubectl config use-context "$CLUSTER_NAME"
```

### Known limitation

* [Maximum 200 groups will be included in the OIDC JWT](https://docs.microsoft.com/en-us/azure/active-directory/hybrid/how-to-connect-fed-group-claims). For more than 200 groups, consider using [Application Roles](https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-add-app-roles-in-azure-ad-apps)
* Groups created in AAD can only be included by their ObjectID and not name, as [`sAMAccountName` is only available for groups synchronized from Active Directory](https://docs.microsoft.com/en-us/azure/active-directory/hybrid/how-to-connect-fed-group-claims#group-claims-for-applications-migrating-from-ad-fs-and-other-identity-providers)

## Contributing

This project welcomes contributions and suggestions. Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit <https://cla.opensource.microsoft.com>.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
