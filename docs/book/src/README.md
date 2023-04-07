# Introduction

`kubelogin` is a [client-go credential (exec) plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) implementing azure authentication. This plugin provides features that are not available in kubectl. It is supported on kubectl v1.11+

## Features

- [interactive device code login](./concepts/login-modes/devicecode.md)
- [interactive web browser login](./concepts/login-modes/interactive.md)
- [non-interactive service principal login](./concepts/login-modes/sp.md)
- [non-interactive user principal login](./concepts/login-modes/ropc.md) using [Resource owner login flow](https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth-ropc)
- [non-interactive managed service identity login](./concepts/login-modes/msi.md)
- [non-interactive Azure CLI token login (AKS only)](./concepts/login-modes/azurecli.md)
- [non-interactive workload identity login](./concepts/login-modes/workloadidentity.md)
