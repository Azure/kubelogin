# kubelogin

This is a [client-go credential (exec) plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) implementing azure authentication. This plugin provides features that are not available in kubectl. It is supported on kubectl v1.11+

## Features

* `convert-kubeconfig` command to converts kubeconfig with existing azure auth provider format to exec credential plugin format
* device code login
* non-interactive service principal login
* non-interactive user principal login using [Resource owner login flow](https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth-ropc) 
* non-interactive managed service identity login
* AAD token will be cached locally for renewal in device code login and user principal login (ropc) flow. By default, it is saved in `~/.kube/cache/kubelogin/`
* addresses https://github.com/kubernetes/kubernetes/issues/86410 to remove `spn:` prefix in `audience` claim, if necessary. (based on kubeconfig or commandline argument `--legacy`)

## Getting Started

### Setup

Copy the latest [Releases](https://github.com/Azure/kubelogin/releases) to shell's search path.

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

### Clean up

Whenever you want to remove cached tokens

```sh
kubelogin remove-tokens
```

## Exec Plugin Format

Below is what a kubeconfig with exec plugin would look like. By default, the `audience` claim will not have `spn:` prefix. If it's desired to keep the prefix, add `--legacy` to the args.

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
      - <msi-client-id>
      - --login
      - msi
```

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
