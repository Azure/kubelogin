Instruction for recording response using [GO-VCR](https://github.com/dnaeon/go-vcr) for unit test
# Things to know if you want to record new recording
* All the recorded responses have been saved under folder `/testdata`

* Highly recommand using `RecordOnly` if you want completely new recording, otherwise, current recordings have been modified without the sensitive contents
* Here's the variable you need to input for recording 
Modify these variables 
modify authorizer clientID `AZURE_CLIENT_ID="<specify with real value>"`
modify authorizer clientSecret `AAD_SERVICE_PRINCIPAL_CLIENT_SECRET="<specify with real value>" `
modify authorizer clientCert `AZURE_CLIENT_CER="<specify with real value>"`
modify authorizer clientCertPass `AZURE_CLIENT_CERTIFICATE_PASSWORD="<specify with real value>" `
modify authorizer resourceID `AZURE_RESOURCE_ID="<specify with real value>"`
modify authorizer tenantID `AZURE_TENANT_ID="<specify with real value>" `
modify go-vcr record mode `VCR_MODE="<specify vcr mode>" `
you can set to record mode by setting vcr mode to RecordOnly `VCR_MODE="RecordOnly"`
To return to replay mode, simply unset the enviroment variable by `unset VCR_MODE`

Examples: 
# Recording Mode
* Navigate to `pkg/token` folder in terminal
* Setup your enviroment variables

```
export AZURE_CLIENT_ID="<specify with real value>"
export AAD_SERVICE_PRINCIPAL_CLIENT_SECRET="<specify with real value>"
export AZURE_CLIENT_CER="<specify with real value>"
export AZURE_CLIENT_CERTIFICATE_PASSWORD="<specify with real value>"
export AZURE_RESOURCE_ID="<specify with real value>"
export AZURE_TENANT_ID="<specify with real value>"
export VCR_MODE="RecordOnly"
go test
```

# Replay Mode
```
unset VCR_MODE
go test
```
