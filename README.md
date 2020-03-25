# kubelogin

This is a [client-go credential (exec) plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) implementing azure authentication. This plugin provides features that are not available in kubectl.

## Features

* convert command to convert kubeconfig with existing azure auth provider to exec credential plugin
* device code login
* non-interactive service principal login
* non-interactive user principal login using [Resource owner login flow](https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth-ropc) 
* AAD token will be cached locally for renewal. By default, it is saved in `~/.kube/cache/kubelogin/azure.json`
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

> On AKS, it will only work with AADv2

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

### Clean up

Whenever you want to remove the cached token, to change login method, or to change tenant, you should remove the cached token

```sh
kubelogin remove-token
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
