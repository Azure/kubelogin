# Login Modes

Most of the interaction with `kubelogin` is around `convert-kubeconfig` subcommand which uses the input kubeconfig specified in `--kubeconfig` or `KUBECONFIG` environment variable to convert to the final kubeconfig in [exec format](./concepts/exec-plugin.md) based on specified login mode.
In this section, the login modes will be explained in details.
