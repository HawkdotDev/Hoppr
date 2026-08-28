# Hoppr Windows Uninstaller
# Usage: irm https://raw.githubusercontent.com/HawkdotDev/Hoppr/main/scripts/uninstall.ps1 | iex

$ErrorActionPreference = "SilentlyContinue"

Write-Host ""
Write-Host "=========================================" -ForegroundColor Yellow
Write-Host "       Hoppr Uninstaller for Windows     " -ForegroundColor Yellow
Write-Host "=========================================" -ForegroundColor Yellow
Write-Host ""

$InstallRoot = Join-Path $env:LOCALAPPDATA "Programs\Hoppr"
$BinDir      = Join-Path $InstallRoot "bin"
$ConfigDir   = Join-Path $env:USERPROFILE ".hoppr"

# 1. Remove Installation Folder
if (Test-Path $InstallRoot) {
    Write-Host "[*] Removing application files from $InstallRoot..." -ForegroundColor Cyan
    Remove-Item -Recurse -Force $InstallRoot
    Write-Host "[+] Application files removed." -ForegroundColor Green
}

# 2. Remove from User PATH
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -like "*$BinDir*") {
    Write-Host "[*] Removing from User PATH..." -ForegroundColor Cyan
    $NewPath = ($UserPath -split ';' | Where-Object { $_ -and $_ -ne $BinDir }) -join ';'
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    $env:PATH = ($env:PATH -split ';' | Where-Object { $_ -and $_ -ne $BinDir }) -join ';'
    Write-Host "[+] Removed from User PATH." -ForegroundColor Green
}

# 3. Clean up PowerShell Profile
if (Test-Path $PROFILE) {
    Write-Host "[*] Cleaning up PowerShell Profile..." -ForegroundColor Cyan
    $ProfileContent = Get-Content -Path $PROFILE -Raw
    if ($ProfileContent -match "hop\.ps1") {
        $CleanedLines = (Get-Content -Path $PROFILE) | Where-Object { 
            $_ -notmatch "Hoppr Shell Integration" -and $_ -notmatch "hop\.ps1" 
        }
        Set-Content -Path $PROFILE -Value $CleanedLines
        Write-Host "[+] Shell integration removed from `$PROFILE." -ForegroundColor Green
    }
}

# 4. Prompt to Remove Config / Saved Projects
Write-Host ""
$RemoveConfig = Read-Host "Do you also want to delete your saved projects and config (~/.hoppr)? (y/N)"
if ($RemoveConfig -eq 'y' -or $RemoveConfig -eq 'Y') {
    if (Test-Path $ConfigDir) {
        Remove-Item -Recurse -Force $ConfigDir
        Write-Host "[+] Config and saved projects removed." -ForegroundColor Green
    }
} else {
    Write-Host "[*] Preserved ~/.hoppr configuration." -ForegroundColor DarkGray
}

Write-Host ""
Write-Host "=========================================" -ForegroundColor Green
Write-Host "     Hoppr uninstalled successfully!     " -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
