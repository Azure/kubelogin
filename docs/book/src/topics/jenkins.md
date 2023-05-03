# Using kubelogin in Jenkins

In Jenkins, since workspaces are most likely run under `jenkins` user, different login modes may have different configuration requirements to allow multiple builds to run concurrently. When it is not configured properly, there may be clashing in cache or login context that results in `You must be logged in to the server (Unauthorized)` error message.

## Using Azure CLI Login mode

When Azure CLI is installed in Jenkins environment, Azure CLI's config directory likely resides in Jenkins workspace directory. To use the Azure CLI, environment variable `AZURE_CONFIG_DIR` should be specified.

Using kubelogin `convert-kubeconfig` subcommand with `--azure-config-dir`, the generated kubeconfig will configure the environment variable for `get-token` subcommand to find the corresponding Azure config directory. For example,

```sh
stage('Download kubeconfig and convert') {
    steps {
        sh 'az aks get-credentials -g ${RESOURCE_GROUP} -n ${CLUSTER_NAME}'
        sh 'kubelogin convert-kubeconfig -l azurecli --azure-config-dir ${WORKSPACE}/.azure'
    }
}

stage('Run kubectl') {
    steps {
        sh 'kubectl get nodes'
    }
}
```

## Using Device Code, Web Browser, and ROPC Login Modes

Since `kubelogin` by default caches tokens at `${HOME}/.kube/cache/kubelogin` in [device code](../concepts/login-modes/devicecode.md),
[web browser interactive](../concepts/login-modes/interactive.md), and [ropc](../concepts/login-modes/ropc.md) [login modes](../concepts/login-modes.md),
`kubelogin covert-kubeconfig --token-cache-dir` should be specified to a directory under Jenkins workspace such as `${WORKSPACE}/.kube/cache/kubelogin`.
