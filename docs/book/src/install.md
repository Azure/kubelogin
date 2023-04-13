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

```sh
# install
asdf plugin add kubelogin https://github.com/sechmann/asdf-kubelogin
asdf install kubelogin latest
asdf global kubelogin latest

# upgrade
asdf update
asdf plugin update kubelogin
asdf install kubelogin latest
asdf global kubelogin latest
```

## Windows

### Using winget

From Powershell:
```powershell
winget install --id=Kubernetes.kubectl  -e
winget install --id=Microsoft.Azure.Kubelogin  -e
```

### Using scoop

From Powershell:
```powershell
scoop install kubectl azure-kubelogin
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

