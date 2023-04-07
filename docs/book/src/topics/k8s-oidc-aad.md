# Setup k8s OIDC Provider using Azure AD

`kubelogin` can be used to authenticate to general kubernetes clusters using AAD as an OIDC provider. 

1. Create an AAD Enterprise Application and the corresponding App Registration. Check the `Allow public client flows` checkbox. 
Configure groups to be included in the response. Take a note of the directory (tenant) ID as `$AAD_TENANT_ID` and the application (client) ID as `$AAD_CLIENT_ID`
1. Configure the API server with the following flags:

   * Issuer URL: `--oidc-issuer-url=https://sts.windows.net/$AAD_TENANT_ID/`
   * Client ID: `--oidc-client-id=$AAD_CLIENT_ID`
   * Username claim: `--oidc-username-claim=upn`

   See the [kubernetes docs for optional flags](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server). For EKS clusters [configure this on the Management Console](https://docs.amazonaws.cn/en_us/eks/latest/userguide/authenticate-oidc-identity-provider.html) or via [terraform](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_identity_provider_config).

3. Configure the [Exec plugin](../concepts/exec-plugin.md) with `kubelogin` to use the application from the first step:

   ```sh
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

