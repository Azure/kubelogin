# Using kubelogin with AKS

AKS uses a pair of first party Azure AD applications. These application IDs are the same in all environments.

## Azure Kubernetes Service AAD Server

applicationID: 6dae42f8-4368-4678-94ff-3960e28e3630

This is the application used by the server side. The access token accessing AKS clusters need to be issued for this app.
In most of `kubelogin` [login modes](./login-modes.md), `--server-id` is required parameter in `kubelogin get-token`.

## Azure Kubernetes Service AAD Client

applicationID: 80faf920-1908-4b52-b5ef-a8e7bedfc67a

This is a public client application used by `kubelogin` to perform login on behalf of the user. 
It's used in [device code](./login-modes/devicecode.md), [web browser interactive](./login-modes/interactive.md), and [ropc](./login-modes/ropc.md) login modes.
