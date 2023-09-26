# Installation

## Download from Release

Copy the latest [Releases](https://github.com/Azure/kubelogin/releases) to shell's search path.

## Homebrew

```sh
# install
brew install Azure/kubelogin/kubelogin

# upgrade
brew update
brew upgrade Azure/kubelogin/kubelogin
```

## Linux

### Using [asdf](https://asdf-vm.com/)

_asdf and the asdf-kubelogin plugin are not maintained by Microsoft._

```sh
# install
asdf plugin add kubelogin
asdf install kubelogin latest
asdf global kubelogin latest

# upgrade
asdf update
asdf plugin update kubelogin
asdf install kubelogin latest
asdf global kubelogin latest
```
### Using azure cli
There is another option to install Kubectl and Kubectl login. Documentation on this is here: https://learn.microsoft.com/en-us/cli/azure/aks?view=azure-cli-latest#az-aks-install-cli()

```
# install (May require using the command ‘sudo’)
az aks install-cli
```

## Windows

### Using winget

From Powershell:

```powershell
winget install --id=Kubernetes.kubectl  -e
winget install --id=Microsoft.Azure.Kubelogin  -e
```

### Using scoop

This package is not maintained by Microsoft.

From Powershell:

```powershell
scoop install kubectl azure-kubelogin
```

### Using chocolatey

This package is not maintained by Microsoft.

From Powershell:

```powershell
choco install kubernetes-cli azure-kubelogin
```

### Using azure cli

From Powershell:

```powershell
az aks install-cli
$targetDir="$env:USERPROFILE\.azure-kubelogin"
$oldPath = [System.Environment]::GetEnvironmentVariable("Path","User")
$oldPathArray=($oldPath) -split ";"
if(-Not($oldPathArray -Contains "$targetDir")) {
    write-host "Permanently adding $targetDir to User Path"
    $newPath = "$oldPath;$targetDir" -replace ";+", ";"
    [System.Environment]::SetEnvironmentVariable("Path",$newPath,"User")
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","User"),[System.Environment]::GetEnvironmentVariable("Path","Machine") -join ";"
}
```
