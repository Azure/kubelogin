# Change Log

## [0.0.32]

### Enhancements

* Add PoP token support to interactive+spn get-token/convert-kubeconfig flows by @rharpavat in https://github.com/Azure/kubelogin/pull/319

### Maintenance

* Fixed typo in top header for convert-kubeconfig documentation by @byk0t in https://github.com/Azure/kubelogin/pull/323
* Bump golang.org/x/crypto from 0.11.0 to 0.12.0 by @dependabot in https://github.com/Azure/kubelogin/pull/315
* Bump k8s.io/apimachinery from 0.27.3 to 0.27.4 by @dependabot in https://github.com/Azure/kubelogin/pull/310

## New Contributors

* @byk0t made their first contribution in https://github.com/Azure/kubelogin/pull/323
* @rharpavat made their first contribution in https://github.com/Azure/kubelogin/pull/319

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.0.31...v0.0.32

## [0.0.31]

### Enhancements

- upgrade klog from v1 to v2 by @peterbom in https://github.com/Azure/kubelogin/pull/306

### Maintenance

- Bump k8s.io/apimachinery from 0.27.2 to 0.27.3 by @dependabot in https://github.com/Azure/kubelogin/pull/297 
- Bump golang.org/x/crypto from 0.10.0 to 0.11.0 by @dependabot in https://github.com/Azure/kubelogin/pull/303
- Bump github.com/Azure/azure-sdk-for-go/sdk/azcore from 1.6.0 to 1.6.1  by @dependabot in https://github.com/Azure/kubelogin/pull/292
- Bump golang.org/x/crypto from 0.9.0 to 0.10.0 by @dependabot in https://github.com/Azure/kubelogin/pull/294 

### Doc Update

- docs: Use asdf-plugins index instead of hard coded repo https://github.com/Azure/kubelogin/pull/298
- Add chocolatey installation instructions https://github.com/Azure/kubelogin/pull/299

### New Contributors

- @peterbom made their first contribution in https://github.com/Azure/kubelogin/pull/306
- @sechmann made their first contribution in https://github.com/Azure/kubelogin/pull/298
- @moredatapls made their first contribution in https://github.com/Azure/kubelogin/pull/299

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.0.30...v0.0.31

## [0.0.30]

### Enhancements

- added verbose logging in convert-kubeconfig by @weinong in https://github.com/Azure/kubelogin/pull/272
- Adding installHint field to kubeconfigs that have been converted to the exec format by @cirvine-MSFT in https://github.com/Azure/kubelogin/pull/282

### Maintenance

- Bump github.com/Azure/azure-sdk-for-go/sdk/azcore from 1.1.1 to 1.5.0 by @dependabot in https://github.com/Azure/kubelogin/pull/249
- Bump github.com/AzureAD/microsoft-authentication-library-for-go from 0.9.0 to 1.0.0 by @dependabot in https://github.com/Azure/kubelogin/pull/259
- Bump k8s.io/cli-runtime from 0.26.3 to 0.27.1 by @dependabot in https://github.com/Azure/kubelogin/pull/262
- Bump github.com/Azure/go-autorest/autorest from 0.11.28 to 0.11.29 by @dependabot in https://github.com/Azure/kubelogin/pull/273
- add unit tests for `manualtoken_test.go` by @khareyash05 in https://github.com/Azure/kubelogin/pull/268
- Bump github.com/Azure/azure-sdk-for-go/sdk/azcore from 1.5.0 to 1.6.0 by @dependabot in https://github.com/Azure/kubelogin/pull/274
- Bump golang.org/x/crypto from 0.8.0 to 0.9.0 by @dependabot in https://github.com/Azure/kubelogin/pull/277
- Bump github.com/Azure/azure-sdk-for-go/sdk/azidentity from 1.2.2 to 1.3.0 by @dependabot in https://github.com/Azure/kubelogin/pull/278
- Bump k8s.io/apimachinery from 0.27.1 to 0.27.2 by @dependabot in https://github.com/Azure/kubelogin/pull/283
- Bump k8s.io/cli-runtime from 0.27.1 to 0.27.2 by @dependabot in https://github.com/Azure/kubelogin/pull/285
- Azidentity migration for service principal token by @ekoehn in https://github.com/Azure/kubelogin/pull/287
- update go to address CVE by @weinong in https://github.com/Azure/kubelogin/pull/290

### Doc Update

- update doc for v0.0.29 by @weinong in https://github.com/Azure/kubelogin/pull/270

### New Contributors

- @khareyash05 made their first contribution in https://github.com/Azure/kubelogin/pull/268
- @ekoehn made their first contribution in https://github.com/Azure/kubelogin/pull/287

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.0.29...v0.0.30

## [0.0.29]

### Enhancements

- add --context support in convert subcommand by @weinong in https://github.com/Azure/kubelogin/pull/260
- return error when specified context is not found by @weinong in https://github.com/Azure/kubelogin/pull/261
- add --azure-config-dir in convert-kubeconfig subcommand by @weinong in https://github.com/Azure/kubelogin/pull/263

### Maintenance

- Enable Code Cov for this repo. by @Tatsinnit in https://github.com/Azure/kubelogin/pull/229
- Bump golang.org/x/crypto from 0.6.0 to 0.7.0 by @dependabot in https://github.com/Azure/kubelogin/pull/230
- Bump k8s.io/client-go from 0.26.2 to 0.26.3 by @dependabot in https://github.com/Azure/kubelogin/pull/234
- Feature/addtests by @Tatsinnit in https://github.com/Azure/kubelogin/pull/238
- Bump k8s.io/cli-runtime from 0.26.2 to 0.26.3 by @dependabot in https://github.com/Azure/kubelogin/pull/237
- Bump github.com/spf13/cobra from 1.6.1 to 1.7.0 by @dependabot in https://github.com/Azure/kubelogin/pull/245
- Bump golang.org/x/crypto from 0.7.0 to 0.8.0 by @dependabot in https://github.com/Azure/kubelogin/pull/250
- Add codecov badge to this repo. by @Tatsinnit in https://github.com/Azure/kubelogin/pull/252
- Bump k8s.io/apimachinery from 0.26.3 to 0.27.1 by @dependabot in https://github.com/Azure/kubelogin/pull/257
- Bump k8s.io/client-go from 0.26.3 to 0.27.1 by @dependabot in https://github.com/Azure/kubelogin/pull/258
- Fix merge conflicts and breaking changes in PR 221 by @cirvine-MSFT in https://github.com/Azure/kubelogin/pull/264
- Fix merge conflicts in PR 232 updating adal from 0.9.22 to 0.9.23 by @cirvine-MSFT in https://github.com/Azure/kubelogin/pull/265

### Doc Update

- refactor windows install doc by @weinong in https://github.com/Azure/kubelogin/pull/233
- adding github pages by @weinong in https://github.com/Azure/kubelogin/pull/241
- added inline toc by @weinong in https://github.com/Azure/kubelogin/pull/244
- Document scoop installation option by @goostleek in https://github.com/Azure/kubelogin/pull/242
- revamp the website by @weinong in https://github.com/Azure/kubelogin/pull/246
- update readme and docs by @weinong in https://github.com/Azure/kubelogin/pull/247
- ignore docs and readme on some workflows by @weinong in https://github.com/Azure/kubelogin/pull/248
- Add reference to a context. by @Tatsinnit in https://github.com/Azure/kubelogin/pull/253
- How to install kubelogin with asdf tool manager by @daveneeley in https://github.com/Azure/kubelogin/pull/256
- Update devicecode.md by @madhurgupta03 in https://github.com/Azure/kubelogin/pull/266

## New Contributors

- @goostleek made their first contribution in https://github.com/Azure/kubelogin/pull/242
- @daveneeley made their first contribution in https://github.com/Azure/kubelogin/pull/256
- @cirvine-MSFT made their first contribution in https://github.com/Azure/kubelogin/pull/264
- @madhurgupta03 made their first contribution in https://github.com/Azure/kubelogin/pull/266

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.0.28...v0.0.29

## [0.0.28]

### What's Changed

- Create dependabot.yml by @bcho in https://github.com/Azure/kubelogin/pull/201
- fix: set package ecosystem by @bcho in https://github.com/Azure/kubelogin/pull/203
- document the default device code login doesn't work with conditional … by @weinong in https://github.com/Azure/kubelogin/pull/202
- ci: remove snapstore publish step from push action by @bcho in https://github.com/Azure/kubelogin/pull/210
- Bump golang.org/x/text from 0.3.7 to 0.3.8 by @dependabot in https://github.com/Azure/kubelogin/pull/209
- Bump k8s.io/cli-runtime from 0.24.2 to 0.26.1 by @dependabot in https://github.com/Azure/kubelogin/pull/208
- Bump github.com/Azure/go-autorest/autorest/adal from 0.9.21 to 0.9.22 by @dependabot in https://github.com/Azure/kubelogin/pull/204
- Bump github.com/spf13/cobra from 1.6.0 to 1.6.1 by @dependabot in https://github.com/Azure/kubelogin/pull/213
- Bump github.com/Azure/go-autorest/autorest from 0.11.27 to 0.11.28 by @dependabot in https://github.com/Azure/kubelogin/pull/212
- Bump golang.org/x/net from 0.3.1-0.20221206200815-1e63c2f08a10 to 0.7.0 by @dependabot in https://github.com/Azure/kubelogin/pull/214
- Bump golang.org/x/crypto from 0.0.0-20220722155217-630584e8d5aa to 0.6.0 by @dependabot in https://github.com/Azure/kubelogin/pull/211
- Bump k8s.io/apimachinery from 0.26.1 to 0.26.2 by @dependabot in https://github.com/Azure/kubelogin/pull/217
- Bump k8s.io/cli-runtime from 0.26.1 to 0.26.2 by @dependabot in https://github.com/Azure/kubelogin/pull/218

### New Contributors

- @bcho made their first contribution in https://github.com/Azure/kubelogin/pull/201
- @dependabot made their first contribution in https://github.com/Azure/kubelogin/pull/209

## [0.0.27]

### Whats Changed

- Added Binaries for windows/arm64 by @ssrahul96 in https://github.com/Azure/kubelogin/pull/195
- update go-restful mod. by @Tatsinnit in https://github.com/Azure/kubelogin/pull/198

## [0.0.26]

### Whats Changed

- Add support of env var convention used by azure sdk by @weinong in https://github.com/Azure/kubelogin/pull/174
- update release archives to omit unnecessary file by @weinong in https://github.com/Azure/kubelogin/pull/176

### Bug fixes

- fix/release tagging by Tatsinnit in https://github.com/Azure/kubelogin/pull/189

### Doc Update

- update doc with interactive login index by @weinong in https://github.com/Azure/kubelogin/pull/175
- Go-report and cli flare addition. by @Tatsinnit in https://github.com/Azure/kubelogin/pull/178
- Add go reference for this repo. by @Tatsinnit in https://github.com/Azure/kubelogin/pull/181
- Enable CodeQL Analysis. by @Tatsinnit in https://github.com/Azure/kubelogin/pull/179
- Possible enhancement - Add changelog for this repo for automating release tags. ❤️☕️💡 by @Tatsinnit in https://github.com/Azure/kubelogin/pull/177

### Experimental Features

- build: add support for Ubuntu snap package by @Exodus in https://github.com/Azure/kubelogin/pull/182
- update workflow
