# Resource Owner Password Credential (ropc)

> ### Warning: 
> [Microsoft recommends you do not use the ROPC flow](https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth-ropc)

> ### Note: 
> ROPC is not supported in hybrid identity federation scenarios (for example, Azure AD and ADFS used to authenticate on-premises accounts). If users are redirected to an on-premises identity providers, Azure AD is not able to test the username and password against that identity provider. Pass-through authentication is supported with ROPC, however.
> It also does not work when MFA policy is enabled
> Personal accounts that are invited to an Azure AD tenant can't use ROPC

In this login mode, the access token and refresh token will be cached at `${HOME}/.kube/cache/kubelogin` directory. This path can be overriden by `--token-cache-dir`.

## Usage Examples

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l ropc

export AAD_USER_PRINCIPAL_NAME=foo@bar.com
export AAD_USER_PRINCIPAL_PASSWORD=<password>

kubectl get nodes
```

## Reference

https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth-ropc
