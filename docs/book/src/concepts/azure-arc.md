# Using kubelogin with Azure Arc

kubelogin can be used to authenticate with Azure Arc-enabled clusters by requesting a [proof-of-possession (PoP) token](https://learn.microsoft.com/en-us/entra/msal/dotnet/advanced/proof-of-possession-tokens). This can be done by providing both of the following flags together:

1. `--pop-enabled`: indicates that `kubelogin` should request a PoP token instead of a regular bearer token
2. `--pop-claims`: is a comma-separated list of `key=value` claims to include in the PoP token. At minimum, this must include the u-claim as `u=ARM_ID_OF_CLUSTER`, which specifies the host that the requested token should allow access on.

These flags can be provided to either `kubelogin get-token` directly to get a PoP token, or to `kubelogin convert-kubeconfig` for `kubectl` to request the token internally. 

PoP token requests only work with `interactive` and `spn` login modes; these flags will be ignored if provided for other login modes.

## AAD Server App

```
applicationID: 6256c85f-0aad-4d50-b960-e6e9b21efe35
```

This is the application used by the server side. The access token needs to be issued for this app to access a 1P Arc-enabled cluster.

This server app ID is a required parameter for [`web browser interactive`](./login-modes/interactive.md) login mode supporting PoP token authentication.

## AAD Client App

```
applicationID: 3f4439ff-e698-4d6d-84fe-09c9d574f06b
```

This is a 1P client application used by `kubelogin` to perform login on behalf of the user. It should be used for [`web browser interactive`](./login-modes/interactive.md) login mode when using PoP token authentication.
