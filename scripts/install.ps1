# Hoppr Windows 1-Line Installer
# Usage: irm https://raw.githubusercontent.com/HawkdotDev/Hoppr/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "HawkdotDev/Hoppr"
Write-Host ">>> Installing Hoppr..." -ForegroundColor Cyan

# 1. Detect Architecture
$arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Error "Hoppr requires a 64-bit Windows OS."
}

# 2. Get Latest Release from GitHub
Write-Host ">>> Fetching latest release information..." -ForegroundColor Yellow
$apiUrl = "https://api.github.com/repos/$repo/releases/latest"
try {
    $release = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing -Headers @{ "User-Agent" = "Hoppr-Installer" }
    $tag = $release.tag_name
} catch {
    $tag = "v1.0.0"
}

$assetName = "hoppr-$tag-windows-$arch.zip"
$downloadUrl = "https://github.com/$repo/releases/download/$tag/$assetName"

# 3. Setup Install Directory
$installDir = Join-Path $env:LOCALAPPDATA "Programs\Hoppr"
$binDir = Join-Path $installDir "bin"
$shellDir = Join-Path $installDir "shell"

if (!(Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir -Force | Out-Null
}
if (!(Test-Path $shellDir)) {
    New-Item -ItemType Directory -Path $shellDir -Force | Out-Null
}

$tempZip = Join-Path $env:TEMP "$assetName"

# 4. Download and Extract
Write-Host ">>> Downloading $assetName..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempZip -UseBasicParsing
    Expand-Archive -Path $tempZip -DestinationPath $installDir -Force
    Remove-Item $tempZip -Force
} catch {
    Write-Warning "Direct release archive download failed. Attempting fallback download..."
    # Fallback to direct raw scripts download
    $exeUrl = "https://raw.githubusercontent.com/$repo/main/hoppr.exe"
}

# Move binary to bin directory if extracted in subfolder
$extractedBin = Get-ChildItem -Path $installDir -Filter "hoppr.exe" -Recurse | Select-Object -First 1
if ($extractedBin -and $extractedBin.FullName -ne (Join-Path $binDir "hoppr.exe")) {
    Copy-Item $extractedBin.FullName -Destination (Join-Path $binDir "hoppr.exe") -Force
}

# Ensure shell wrapper is downloaded / copied
$psWrapperUrl = "https://raw.githubusercontent.com/$repo/main/shell/hop.ps1"
$localPsWrapper = Join-Path $shellDir "hop.ps1"
if (!(Test-Path $localPsWrapper)) {
    Invoke-WebRequest -Uri $psWrapperUrl -OutFile $localPsWrapper -UseBasicParsing
}

# 5. Add to User PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$binDir*") {
    Write-Host ">>> Adding $binDir to User PATH..." -ForegroundColor Cyan
    [Environment]::SetEnvironmentVariable("PATH", "$userPath;$binDir", "User")
    $env:PATH = "$env:PATH;$binDir"
}

# 6. Configure PowerShell Profile for hop auto-jumping
if (!(Test-Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force | Out-Null
}
$profileContent = Get-Content -Path $PROFILE -Raw -ErrorAction SilentlyContinue
$sourceCmd = ". `"$localPsWrapper`""
if ($profileContent -notlike "*$localPsWrapper*") {
    Write-Host ">>> Configuring PowerShell Profile for instant directory jumping..." -ForegroundColor Cyan
    Add-Content -Path $PROFILE -Value "`n# Hoppr Shell Integration`n$sourceCmd"
}

# Run the shell integration in the current session
. "$localPsWrapper"

Write-Host ""
Write-Host ">>> Hoppr installed successfully! 🚀" -ForegroundColor Green
Write-Host ">>> Run 'hop doctor' or 'hop --help' to get started." -ForegroundColor Green
