# Known Issues

* [Maximum 200 groups will be included in the Azure AD JWT](https://docs.microsoft.com/en-us/azure/active-directory/hybrid/how-to-connect-fed-group-claims). 
For more than 200 groups, consider using [Application Roles](https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-add-app-roles-in-azure-ad-apps)
* Groups created in Azure AD can only be included by their ObjectID and not name, as [`sAMAccountName` is only available for groups synchronized from Active Directory](https://docs.microsoft.com/en-us/azure/active-directory/hybrid/how-to-connect-fed-group-claims#group-claims-for-applications-migrating-from-ad-fs-and-other-identity-providers)
* [`kubelogin` may not work with MSI when run in Azure Container Instance](https://github.com/Azure/kubelogin/issues/79)
* On AKS, [service principal](./concepts/login-modes/sp.md) login mode will only work with managed AAD, not legacy AAD.
* [Device code](./concepts/login-modes/devicecode.md) login mode does not work when Conditional Access policy is configured on Azure AD tenant.
Use [web browser interactive](./concepts/login-modes/interactive.md) instead.
