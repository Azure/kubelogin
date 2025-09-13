package env

const (
	// env vars
	LoginMethod                        = "AAD_LOGIN_METHOD"
	KubeloginROPCUsername              = "AAD_USER_PRINCIPAL_NAME"
	KubeloginROPCPassword              = "AAD_USER_PRINCIPAL_PASSWORD"
	KubeloginClientID                  = "AAD_SERVICE_PRINCIPAL_CLIENT_ID"
	KubeloginClientSecret              = "AAD_SERVICE_PRINCIPAL_CLIENT_SECRET"
	KubeloginClientCertificatePath     = "AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE"
	KubeloginClientCertificatePassword = "AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE_PASSWORD"

	// env vars used by Terraform
	TerraformClientID                  = "ARM_CLIENT_ID"
	TerraformClientSecret              = "ARM_CLIENT_SECRET"
	TerraformClientCertificatePath     = "ARM_CLIENT_CERTIFICATE_PATH"
	TerraformClientCertificatePassword = "ARM_CLIENT_CERTIFICATE_PASSWORD"
	TerraformTenantID                  = "ARM_TENANT_ID"

	// env vars following azure sdk naming convention
	AzureAuthorityHost             = "AZURE_AUTHORITY_HOST"
	AzureClientCertificatePassword = "AZURE_CLIENT_CERTIFICATE_PASSWORD"
	AzureClientCertificatePath     = "AZURE_CLIENT_CERTIFICATE_PATH"
	AzureClientID                  = "AZURE_CLIENT_ID"
	AzureClientSecret              = "AZURE_CLIENT_SECRET"
	AzureFederatedTokenFile        = "AZURE_FEDERATED_TOKEN_FILE"
	AzureTenantID                  = "AZURE_TENANT_ID"
	AzureUsername                  = "AZURE_USERNAME"
	AzurePassword                  = "AZURE_PASSWORD"
)
