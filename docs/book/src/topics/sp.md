# Using Service Principal

This section documents the end to end flow to use `kubelogin` to access AKS cluster with a service principal.

## 1. Create a service principal or use an existing one.

```sh
az ad sp create-for-rbac --skip-assignment --name myAKSAutomationServicePrincipal
```

The output is similar to the following example.

```json
{
  "appId": "<spn client id>",
  "displayName": "myAKSAutomationServicePrincipal",
  "name": "http://myAKSAutomationServicePrincipal",
  "password": "<spn secret>",
  "tenant": "<aad tenant id>"
}
```

## 2. Query your service principal AAD Object ID by using the command below.

```sh
az ad sp show --id <spn client id> --query "id"
```

## 3. To configure the role binding on Azure Kubernetes Service, the user in rolebinding should be the SP's Object ID.

For example,

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: sp-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: <service-principal-object-id>
```

## 4. Use `kubelogin` to convert the kubeconfig

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l spn

export AAD_SERVICE_PRINCIPAL_CLIENT_ID=<spn client id>
export AAD_SERVICE_PRINCIPAL_CLIENT_SECRET=<spn secret>

kubectl get nodes
```

or write your spn secret permanently into the kubeconfig (not preferred!):

```sh
export KUBECONFIG=/path/to/kubeconfig

kubelogin convert-kubeconfig -l spn --client-id <spn client id> --client-secret <spn secret>

kubectl get nodes
```
