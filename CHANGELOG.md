# Change Log

## [0.1.0]

### Enhancements

- [library usage] Move modules under `pkg` to `pkg/internal` by @bcho in https://github.com/Azure/kubelogin/pull/376
- [library usage] Update module version usages by @bcho in https://github.com/Azure/kubelogin/pull/377
- [library usage] Refine internal token types by @bcho in https://github.com/Azure/kubelogin/pull/379
- [library usage] Implement library token provider by @bcho in https://github.com/Azure/kubelogin/pull/380
- [library usage] fix: downgrade required go version to 1.20 by @bcho in https://github.com/Azure/kubelogin/pull/386

### Maintenance

- Bump github.com/spf13/cobra from 1.7.0 to 1.8.0 by @dependabot in https://github.com/Azure/kubelogin/pull/359
- Bump golang.org/x/crypto from 0.14.0 to 0.17.0 by @dependabot in https://github.com/Azure/kubelogin/pull/378
- Bump github.com/golang-jwt/jwt/v5 from 5.0.0 to 5.2.0 by @dependabot in https://github.com/Azure/kubelogin/pull/370
- Bump github.com/Azure/azure-sdk-for-go/sdk/azcore from 1.8.0 to 1.9.1 by @dependabot in https://github.com/Azure/kubelogin/pull/372
- Bump go.uber.org/mock from 0.3.0 to 0.4.0 by @dependabot in https://github.com/Azure/kubelogin/pull/385
- Bump github.com/google/uuid from 1.4.0 to 1.5.0 by @dependabot in https://github.com/Azure/kubelogin/pull/383

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.0.34...v0.1.0

## [0.0.34]

### Enhancements

* feat(timeout): Implement customizable timeout for Azure CLI token ret… by @Aricg in https://github.com/Azure/kubelogin/pull/362
* added github token support by @weinong in https://github.com/Azure/kubelogin/pull/366
* added armv7 support by @weinong in https://github.com/Azure/kubelogin/pull/367

### Maintenance

* bump golang to 1.21 by @weinong in https://github.com/Azure/kubelogin/pull/356
* Bump k8s.io/klog/v2 from 2.100.1 to 2.110.1 by @dependabot in https://github.com/Azure/kubelogin/pull/357
* Bump github.com/google/uuid from 1.3.1 to 1.4.0 by @dependabot in https://github.com/Azure/kubelogin/pull/355

## New Contributors
* @Aricg made their first contribution in https://github.com/Azure/kubelogin/pull/362

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.0.33...v0.0.34

## [0.0.33]

### Enhancements

- use the adal library for spn when --legacy is specified by @weinong in https://github.com/Azure/kubelogin/pull/338

### Maintenance

- Bump github.com/google/uuid from 1.3.0 to 1.3.1 by @dependabot in https://github.com/Azure/kubelogin/pull/334
- Add 1P client/server app IDs to docs by @rharpavat in https://github.com/Azure/kubelogin/pull/336
- Update install.md by @torreymicrosoft in https://github.com/Azure/kubelogin/pull/342
- Bump golang.org/x/net from 0.10.0 to 0.17.0 by @dependabot in https://github.com/Azure/kubelogin/pull/347
- Bump github.com/Azure/azure-sdk-for-go/sdk/azcore from 1.6.1 to 1.8.0 by @dependabot in https://github.com/Azure/kubelogin/pull/344
- Bump github.com/Azure/azure-sdk-for-go/sdk/azidentity from 1.3.0 to 1.4.0 by @dependabot in https://github.com/Azure/kubelogin/pull/346
- Bump k8s.io/cli-runtime from 0.27.2 to 0.28.2 by @dependabot in https://github.com/Azure/kubelogin/pull/340
- Bump k8s.io/cli-runtime from 0.28.2 to 0.28.3 by @dependabot in https://github.com/Azure/kubelogin/pull/351
- Bump github.com/google/go-cmp from 0.5.9 to 0.6.0 by @dependabot in https://github.com/Azure/kubelogin/pull/349
- Bump github.com/stretchr/testify from 1.8.2 to 1.8.4 by @dependabot in https://github.com/Azure/kubelogin/pull/348

## New Contributors

- @torreymicrosoft made their first contribution in https://github.com/Azure/kubelogin/pull/342

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.0.32...v0.0.33

## [0.0.32]

### Enhancements

- Add PoP token support to interactive+spn get-token/convert-kubeconfig flows by @rharpavat in https://github.com/Azure/kubelogin/pull/319

### Maintenance

- Fixed typo in top header for convert-kubeconfig documentation by @byk0t in https://github.com/Azure/kubelogin/pull/323
- Bump golang.org/x/crypto from 0.11.0 to 0.12.0 by @dependabot in https://github.com/Azure/kubelogin/pull/315
- Bump k8s.io/apimachinery from 0.27.3 to 0.27.4 by @dependabot in https://github.com/Azure/kubelogin/pull/310

## New Contributors

- @byk0t made their first contribution in https://github.com/Azure/kubelogin/pull/323
- @rharpavat made their first contribution in https://github.com/Azure/kubelogin/pull/319

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.0.31...v0.0.32

## [0.0.31]

### Enhancements

- upgrade klog from v1 to v2 by @peterbom in https://github.com/Azure/kubelogin/pull/306

### Maintenance

- Bump k8s.io/apimachinery from 0.27.2 to 0.27.3 by @dependabot in https://github.com/Azure/kubelogin/pull/297
- Bump golang.org/x/crypto from 0.10.0 to 0.11.0 by @dependabot in https://github.com/Azure/kubelogin/pull/303
- Bump github.com/Azure/azure-sdk-for-go/sdk/azcore from 1.6.0 to 1.6.1 by @dependabot in https://github.com/Azure/kubelogin/pull/292
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
- Update devicecode.md by @madhurgupta03 in https://github.com/Azu
