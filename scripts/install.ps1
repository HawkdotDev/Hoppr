# Hoppr Windows Installer
# Usage: irm https://raw.githubusercontent.com/HawkdotDev/Hoppr/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

# Constants
$Repo        = "HawkdotDev/Hoppr"
$BinName     = "hop.exe"
$InstallRoot = Join-Path $env:LOCALAPPDATA "Programs\Hoppr"
$BinDir      = Join-Path $InstallRoot "bin"
$ShellDir    = Join-Path $InstallRoot "shell"

function Write-Step { param([string]$Msg) Write-Host "[*] $Msg" -ForegroundColor Cyan }
function Write-Ok   { param([string]$Msg) Write-Host "[+] $Msg" -ForegroundColor Green }
function Write-Warn { param([string]$Msg) Write-Host "[!] $Msg" -ForegroundColor Yellow }
function Write-Fail { param([string]$Msg) Write-Host "[-] $Msg" -ForegroundColor Red }

function Abort {
    param([string]$Msg)
    Write-Fail $Msg
    if ($script:TempDir -and (Test-Path $script:TempDir)) {
        Remove-Item -Recurse -Force $script:TempDir -ErrorAction SilentlyContinue
    }
    exit 1
}

Write-Host ""
Write-Host "=========================================" -ForegroundColor Magenta
Write-Host "       Hoppr Installer for Windows       " -ForegroundColor Magenta
Write-Host "     The fast lane to your projects      " -ForegroundColor Magenta
Write-Host "=========================================" -ForegroundColor Magenta
Write-Host ""

# 1. Detect Architecture
Write-Step "Detecting system architecture..."

if (-not [System.Environment]::Is64BitOperatingSystem) {
    Abort "Hoppr requires a 64-bit Windows installation."
}

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "ARM64" { "arm64" }
    default { "amd64" }
}
Write-Ok "Architecture: windows/$Arch"

# 2. Resolve Latest Release
Write-Step "Fetching latest release from GitHub..."

$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
try {
    $Headers  = @{ "User-Agent" = "Hoppr-Installer/1.0" }
    $Release  = Invoke-RestMethod -Uri $ApiUrl -Headers $Headers -UseBasicParsing
    $Tag      = $Release.tag_name
} catch {
    Write-Warn "GitHub API unreachable. Falling back to v1.1.0."
    $Tag = "v1.1.0"
}
Write-Ok "Release: $Tag"

# 3. Prepare Temp Directory
$script:TempDir = Join-Path $env:TEMP "hoppr-install-$(Get-Random)"
New-Item -ItemType Directory -Path $script:TempDir -Force | Out-Null

$AssetName   = "hoppr-$Tag-windows-$Arch.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$AssetName"
$ChecksumUrl = "https://github.com/$Repo/releases/download/$Tag/checksums.txt"
$ZipPath     = Join-Path $script:TempDir $AssetName

# 4. Download Archive
Write-Step "Downloading $AssetName..."

try {
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
} catch {
    Abort "Failed to download release archive ($downloadUrl). Check your internet connection or visit https://github.com/$Repo/releases"
}

$FileSizeMb = [math]::Round((Get-Item $ZipPath).Length / 1MB, 2)
Write-Ok "Downloaded $FileSizeMb MB"

# 5. Verify SHA256 Checksum
Write-Step "Verifying SHA256 checksum..."

$ChecksumPath = Join-Path $script:TempDir "checksums.txt"
try {
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumPath -UseBasicParsing
} catch {
    Write-Warn "Could not download checksums.txt. Skipping verification."
    $ChecksumPath = $null
}

if ($ChecksumPath -and (Test-Path $ChecksumPath)) {
    $ExpectedLine = Get-Content $ChecksumPath | Where-Object { $_ -match [regex]::Escape($AssetName) } | Select-Object -First 1
    if ($ExpectedLine) {
        $ExpectedHash = ($ExpectedLine -split '\s+')[0].Trim().ToUpper()
        $ActualHash   = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash.ToUpper()

        if ($ExpectedHash -ne $ActualHash) {
            Abort "SHA256 MISMATCH! Expected: $ExpectedHash Got: $ActualHash"
        }
        Write-Ok "SHA256 verified successfully."
    } else {
        Write-Warn "No checksum entry found for $AssetName. Skipping verification."
    }
} else {
    Write-Warn "Checksum verification skipped."
}

# 6. Extract and Install
Write-Step "Installing to $InstallRoot..."

if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
}
if (-not (Test-Path $ShellDir)) {
    New-Item -ItemType Directory -Path $ShellDir -Force | Out-Null
}

$ExtractDir = Join-Path $script:TempDir "extracted"
Expand-Archive -Path $ZipPath -DestinationPath $ExtractDir -Force

$ExtractedBin = Get-ChildItem -Path $ExtractDir -Filter $BinName -Recurse | Select-Object -First 1
if (-not $ExtractedBin) {
    Abort "Could not find $BinName in the release archive."
}
Copy-Item -Path $ExtractedBin.FullName -Destination (Join-Path $BinDir $BinName) -Force

$ExtractedShellDir = Get-ChildItem -Path $ExtractDir -Filter "shell" -Directory -Recurse | Select-Object -First 1
if ($ExtractedShellDir) {
    Get-ChildItem -Path $ExtractedShellDir.FullName -File | ForEach-Object {
        Copy-Item -Path $_.FullName -Destination (Join-Path $ShellDir $_.Name) -Force
    }
} else {
    Write-Warn "Shell scripts not found in archive. Downloading hop.ps1 directly..."
    $PsWrapperUrl = "https://raw.githubusercontent.com/$Repo/main/shell/hop.ps1"
    try {
        Invoke-WebRequest -Uri $PsWrapperUrl -OutFile (Join-Path $ShellDir "hop.ps1") -UseBasicParsing
    } catch {
        Write-Warn "Could not download hop.ps1 directly."
    }
}

Write-Ok "Binary installed to $BinDir"

# 7. Configure User PATH
Write-Step "Configuring system PATH..."

$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$BinDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$BinDir", "User")
    $env:PATH = "$env:PATH;$BinDir"
    Write-Ok "Added $BinDir to User PATH"
} else {
    Write-Ok "PATH already configured"
}

# 8. Configure PowerShell Profile
Write-Step "Setting up PowerShell shell integration..."

$PsWrapper = Join-Path $ShellDir "hop.ps1"
if (Test-Path $PsWrapper) {
    $profileDir = Split-Path -Parent $PROFILE
    if (-not (Test-Path $profileDir)) {
        New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
    }
    if (-not (Test-Path $PROFILE)) {
        New-Item -ItemType File -Path $PROFILE -Force | Out-Null
    }

    $ProfileContent = Get-Content -Path $PROFILE -Raw -ErrorAction SilentlyContinue
    $SourceLine = ". `"$PsWrapper`""

    if (-not $ProfileContent -or $ProfileContent -notlike "*hop.ps1*") {
        if ($ProfileContent) {
            $BackupPath = "$PROFILE.hoppr-backup"
            Copy-Item -Path $PROFILE -Destination $BackupPath -Force
            Write-Ok "Profile backed up to $BackupPath"
        }
        Add-Content -Path $PROFILE -Value "`r`n# Hoppr Shell Integration`r`n$SourceLine"
        Write-Ok "Shell integration added to profile"
    } else {
        Write-Ok "Shell integration already configured"
    }

    try {
        . $PsWrapper
    } catch {
        Write-Warn "Could not activate in current session. Restart your terminal."
    }
} else {
    Write-Warn "hop.ps1 not found. Shell integration skipped."
}

# 9. Cleanup
Remove-Item -Recurse -Force $script:TempDir -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "=========================================" -ForegroundColor Green
Write-Host "     Hoppr installed successfully!       " -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Get started:" -ForegroundColor White
Write-Host "  hop doctor          - verify installation health" -ForegroundColor Gray
Write-Host "  hop add .           - save current folder as a project" -ForegroundColor Gray
Write-Host "  hop list            - view all saved projects" -ForegroundColor Gray
Write-Host "  hop <project>       - jump to any saved project" -ForegroundColor Gray
Write-Host ""
Write-Host "Docs: https://github.com/$Repo" -ForegroundColor DarkGray
Write-Host ""
