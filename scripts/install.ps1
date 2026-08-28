# Hoppr — Windows Installer
# Usage: irm https://raw.githubusercontent.com/HawkdotDev/Hoppr/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

# ── Paths & Constants ──────────────────────────────────────────────────────────
$Repo        = "HawkdotDev/Hoppr"
$BinName     = "hop.exe"
$InstallRoot = Join-Path $env:LOCALAPPDATA "Programs\Hoppr"
$BinDir      = Join-Path $InstallRoot "bin"
$ShellDir    = Join-Path $InstallRoot "shell"

function Write-Step { param([string]$Msg) Write-Host "  [*] $Msg" -ForegroundColor Cyan }
function Write-Ok   { param([string]$Msg) Write-Host "  [+] $Msg" -ForegroundColor Green }
function Write-Warn { param([string]$Msg) Write-Host "  [!] $Msg" -ForegroundColor Yellow }
function Write-Fail { param([string]$Msg) Write-Host "  [-] $Msg" -ForegroundColor Red }

function Abort {
    param([string]$Msg)
    Write-Fail $Msg
    if ($script:TempDir -and (Test-Path $script:TempDir)) {
        Remove-Item -Recurse -Force $script:TempDir -ErrorAction SilentlyContinue
    }
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

# ── Header Banner ──────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  +-------------------------------------------------------------+" -ForegroundColor Magenta
Write-Host "  |                    HOPPR WINDOWS INSTALLER                  |" -ForegroundColor Magenta
Write-Host "  |             The fast lane to your favourite projects        |" -ForegroundColor Magenta
Write-Host "  +-------------------------------------------------------------+" -ForegroundColor Magenta
Write-Host ""

# ── 1. Detect Architecture ────────────────────────────────────────────────────
Write-Step "Detecting CPU architecture..."

if (-not [System.Environment]::Is64BitOperatingSystem) {
    Abort "Hoppr requires a 64-bit Windows operating system."
}

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "ARM64" { "arm64" }
    default { "amd64" }
}
Write-Ok "Architecture resolved: windows/$Arch"

# ── 2. Resolve Latest Release ─────────────────────────────────────────────────
Write-Step "Checking latest release from GitHub..."

$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
try {
    $Headers  = @{ "User-Agent" = "Hoppr-Installer/1.1" }
    $Release  = Invoke-RestMethod -Uri $ApiUrl -Headers $Headers -UseBasicParsing
    $Tag      = $Release.tag_name
} catch {
    # Fallback to zero-quota web redirect
    try {
        $WebReq = [System.Net.HttpWebRequest]::Create("https://github.com/$Repo/releases/latest")
        $WebReq.AllowAutoRedirect = $false
        $WebResp = $WebReq.GetResponse()
        $Location = $WebResp.GetResponseHeader("Location")
        $Tag = $Location.Substring($Location.LastIndexOf("/") + 1)
        $WebResp.Close()
    } catch {
        $Tag = "v1.2.0"
    }
}
Write-Ok "Target version: $Tag"

# ── 3. Prepare Temp Directory ─────────────────────────────────────────────────
$script:TempDir = Join-Path $env:TEMP "hoppr-install-$(Get-Random)"
New-Item -ItemType Directory -Path $script:TempDir -Force | Out-Null

$AssetName   = "hoppr-$Tag-windows-$Arch.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$AssetName"
$ChecksumUrl = "https://github.com/$Repo/releases/download/$Tag/checksums.txt"
$ZipPath     = Join-Path $script:TempDir $AssetName

# ── 4. Download Release Archive ───────────────────────────────────────────────
Write-Step "Downloading $AssetName..."

try {
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
} catch {
    Abort "Failed to download $AssetName from GitHub. Please check your network connection."
}

$FileSizeMb = [math]::Round((Get-Item $ZipPath).Length / 1MB, 2)
Write-Ok "Downloaded archive payload: $FileSizeMb MB"

# ── 5. Cryptographic SHA256 Verification ──────────────────────────────────────
Write-Step "Verifying cryptographic SHA256 signature..."

$ChecksumPath = Join-Path $script:TempDir "checksums.txt"
try {
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumPath -UseBasicParsing
} catch {
    Write-Warn "Could not download checksums.txt (skipping hash verification)."
    $ChecksumPath = $null
}

if ($ChecksumPath -and (Test-Path $ChecksumPath)) {
    $ExpectedLine = Get-Content $ChecksumPath | Where-Object { $_ -match [regex]::Escape($AssetName) } | Select-Object -First 1
    if ($ExpectedLine) {
        $ExpectedHash = ($ExpectedLine -split '\s+')[0].Trim().ToUpper()
        $ActualHash   = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash.ToUpper()

        if ($ExpectedHash -ne $ActualHash) {
            Abort "SECURITY ALERT: SHA256 checksum mismatch! Expected: $ExpectedHash Got: $ActualHash"
        }
        Write-Ok "SHA256 verified successfully: $($ActualHash.Substring(0, 16))..."
    } else {
        Write-Warn "Package hash not found in checksum manifest."
    }
}

# ── 6. Extract and Install Binary ─────────────────────────────────────────────
Write-Step "Installing Hoppr to $InstallRoot..."

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
    Abort "Could not find $BinName inside the release archive."
}

$DestBin = Join-Path $BinDir $BinName
Copy-Item -Path $ExtractedBin.FullName -Destination $DestBin -Force

# Copy shell integration scripts
$ExtractedShellDir = Get-ChildItem -Path $ExtractDir -Filter "shell" -Directory -Recurse | Select-Object -First 1
if ($ExtractedShellDir) {
    Get-ChildItem -Path $ExtractedShellDir.FullName -File | ForEach-Object {
        Copy-Item -Path $_.FullName -Destination (Join-Path $ShellDir $_.Name) -Force
    }
} else {
    $PsWrapperUrl = "https://raw.githubusercontent.com/$Repo/main/shell/hop.ps1"
    try {
        Invoke-WebRequest -Uri $PsWrapperUrl -OutFile (Join-Path $ShellDir "hop.ps1") -UseBasicParsing
    } catch {
        Write-Warn "Could not download hop.ps1 shell wrapper directly."
    }
}

Write-Ok "Binary and shell integration installed successfully."

# ── 7. Configure User Environment PATH ────────────────────────────────────────
Write-Step "Configuring User PATH environment variable..."

$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$BinDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$BinDir", "User")
    $env:PATH = "$env:PATH;$BinDir"
    Write-Ok "Added $BinDir to User PATH."
} else {
    Write-Ok "User PATH already configured."
}

# ── 8. Configure PowerShell Profile ───────────────────────────────────────────
Write-Step "Setting up PowerShell profile shell integration..."

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
            Write-Ok "Backed up existing profile to $BackupPath."
        }
        Add-Content -Path $PROFILE -Value "`r`n# Hoppr Shell Integration`r`n$SourceLine"
        Write-Ok "Added shell integration to $PROFILE."
    } else {
        Write-Ok "PowerShell profile already contains Hoppr integration."
    }

    # Activate in active session
    try {
        . $PsWrapper
    } catch {
        Write-Warn "Could not load function into current session. (Restart your terminal)."
    }
}

# ── 9. Cleanup ─────────────────────────────────────────────────────────────────
Remove-Item -Recurse -Force $script:TempDir -ErrorAction SilentlyContinue

# ── 10. Verification Test ─────────────────────────────────────────────────────
Write-Host ""
Write-Host "  +-------------------------------------------------------------+" -ForegroundColor Green
Write-Host "  |               HOPPR WAS INSTALLED SUCCESSFULLY!             |" -ForegroundColor Green
Write-Host "  +-------------------------------------------------------------+" -ForegroundColor Green
Write-Host ""

Write-Host "  Installation Summary:" -ForegroundColor White
Write-Host "    Version:       $Tag" -ForegroundColor Cyan
Write-Host "    Location:      $DestBin" -ForegroundColor Gray
Write-Host "    Integration:   PowerShell Profile Configured" -ForegroundColor Gray
Write-Host ""

Write-Host "  Self-Test Output:" -ForegroundColor White
& $DestBin --version
Write-Host ""

Write-Host "  Quick Start:" -ForegroundColor Yellow
Write-Host "    hop add .          - Save current folder as a project" -ForegroundColor Gray
Write-Host "    hop list           - View all saved projects" -ForegroundColor Gray
Write-Host "    hop <project>      - Jump directly to any project" -ForegroundColor Gray
Write-Host "    hop doctor         - Check environment health" -ForegroundColor Gray
Write-Host "    hop update         - Self-update to future releases" -ForegroundColor Gray
Write-Host ""
Write-Host "  Documentation: https://github.com/$Repo" -ForegroundColor DarkGray
Write-Host ""

# ── Optional Windows Notification Toast ───────────────────────────────────────
try {
    [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
    [Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
    $template = @"
<toast>
    <visual>
        <binding template="ToastGeneric">
            <text>Hoppr Installed Successfully!</text>
            <text>Hoppr $Tag is ready. Open a terminal and run 'hop doctor'.</text>
        </binding>
    </visual>
</toast>
"@
    $xml = New-Object Windows.Data.Xml.Dom.XmlDocument
    $xml.LoadXml($template)
    $toast = New-Object Windows.UI.Notifications.ToastNotification $xml
    [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Hoppr").Show($toast)
} catch {
    # Non-critical if Toast API is not available
}
