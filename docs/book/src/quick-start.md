# Quick Start

After `kubelogin` is installed, do the following on Azure AD enabled AKS clusters

## Using Azure CLI login mode

```sh
az login

# by default, this command merges the kubeconfig into ${HOME}/.kube/config
az aks get-credentials -g ${RESOURCE_GROUP_NAME} -n ${AKS_NAME}


# kubelogin by default will use the kubeconfig from ${KUBECONFIG}. Specify --kubeconfig to override
# this converts to use azurecli login mode
kubelogin convert-kubeconfig -l azurecli

# voila!
kubectl get nodes
```
