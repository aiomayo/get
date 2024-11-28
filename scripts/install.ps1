$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$REPO_OWNER = "AIO-Develope"
$REPO_NAME = "get"
$CLI_NAME = "get"

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] " -NoNewline -ForegroundColor Blue
    Write-Host $Message
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] " -NoNewline -ForegroundColor Green
    Write-Host $Message
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] " -NoNewline -ForegroundColor Red
    Write-Host $Message -ForegroundColor Red
}

function Test-Dependencies {
    $requiredModules = @('Microsoft.PowerShell.Management', 'Microsoft.PowerShell.Utility')
    foreach ($module in $requiredModules) {
        if (!(Get-Module -ListAvailable -Name $module)) {
            Write-Error "Required PowerShell module '$module' is not available."
            exit 1
        }
    }
}

function Get-SystemArch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "x86" { return "386" }
        "ARM64" { return "arm64" }
        default {
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

function Update-UserPath {
    param([string]$InstallDir)
    if (!(Test-Path $InstallDir)) {
        New-Item -Path $InstallDir -ItemType Directory | Out-Null
    }
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -split ';' -notcontains $InstallDir) {
        $newPath = "$InstallDir;$currentPath"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Info "Updated user PATH to include $InstallDir"
        $env:Path = "$InstallDir;$env:Path"
    }
}

function Get-LatestRelease {
    $url = "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
    try {
        Invoke-RestMethod -Uri $url -Method Get
    }
    catch {
        Write-Error "Failed to fetch latest release: $_"
        exit 1
    }
}

function Find-DownloadUrl {
    param(
        [string]$Arch,
        [object]$Release
    )
    $windowsBinary = $Release.assets |
        Where-Object {
            $_.name -match "windows" -and
            $_.name -match $Arch -and
            $_.name -match "\.exe$"
        } |
        Select-Object -First 1
    if ($windowsBinary) {
        return $windowsBinary.browser_download_url
    }
    else {
        Write-Error "No compatible Windows binary found for architecture: $Arch"
        exit 1
    }
}

function Install-Cli {
    Test-Dependencies
    $arch = Get-SystemArch
    Write-Info "Detected Architecture: $arch"
    $installDir = "$env:USERPROFILE\.local\bin"
    $release = Get-LatestRelease
    $downloadUrl = Find-DownloadUrl -Arch $arch -Release $release
    $tmpPath = Join-Path $env:TEMP "$CLI_NAME.exe"
    Write-Info "Downloading $CLI_NAME for Windows $arch..."
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tmpPath
    }
    catch {
        Write-Error "Failed to download binary: $_"
        exit 1
    }
    if (!(Test-Path $installDir)) {
        New-Item -Path $installDir -ItemType Directory | Out-Null
    }
    $finalPath = Join-Path $installDir "$CLI_NAME.exe"
    try {
        Move-Item -Path $tmpPath -Destination $finalPath -Force
    }
    catch {
        Write-Error "Failed to move binary: $_"
        exit 1
    }
    Update-UserPath -InstallDir $installDir
    try {
        $version = & "$finalPath" --version
        if ($?) {
            Write-Success "$CLI_NAME installed successfully!"
            Write-Info "Installed at: $finalPath"
            Write-Host "Version: $version"
        }
        else {
            Write-Error "Installation verification failed."
            exit 1
        }
    }
    catch {
        Write-Error "Failed to verify installation: $_"
        exit 1
    }
}

Install-Cli
