# Login Modes

Most of the interaction with `kubelogin` is around `convert-kubeconfig` subcommand 
which uses the input kubeconfig specified in `--kubeconfig` or `KUBECONFIG` environment variable 
to convert to the final kubeconfig in [exec format](./concepts/exec-plugin.md) based on specified login mode.

In this section, the login modes will be explained in details.

## How Login Works

The login modes that `kubelogin` implements are [AAD OAuth 2.0 token grant flows](https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-auth-code-flow).
Throughout `kubelogin` subcommands, you will see below common flags. In general, these flags are already setup when you get the kubeconfig from AKS.

- `--tenant-id`: [Azure AD tenant ID](https://learn.microsoft.com/en-us/azure/active-directory/fundamentals/active-directory-how-to-find-tenant)
- `--client-id`: the application ID of the [public client application](https://learn.microsoft.com/en-us/azure/active-directory/develop/msal-client-applications).
This client app is only used in [device code](./login-modes/devicecode.md), [web browser interactive](./login-modes/interactive.md), and [ropc](./login-modes/ropc.md) login modes.
- `--server-id`: the application ID of the [web app, or resource server](https://learn.microsoft.com/en-us/azure/active-directory/fundamentals/auth-oauth2). 
The token should be issued to this resource.

## References

* https://learn.microsoft.com/en-us/azure/active-directory/fundamentals/auth-oauth2
* https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-auth-code-flow
* https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-client-creds-grant-flow
* https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth-ropc
* https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-protocols-oidc
