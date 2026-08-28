# Hoppr — Professional Windows Installer
# Usage: irm https://raw.githubusercontent.com/HawkdotDev/Hoppr/main/scripts/install.ps1 | iex
#
# Security: Downloads checksums.txt from the release and verifies the SHA256
#           hash of the archive before extraction. Aborts on mismatch.

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

# ── Constants ──────────────────────────────────────────────────────────────────
$Repo        = "HawkdotDev/Hoppr"
$BinName     = "hop.exe"
$InstallRoot = Join-Path $env:LOCALAPPDATA "Programs\Hoppr"
$BinDir      = Join-Path $InstallRoot "bin"
$ShellDir    = Join-Path $InstallRoot "shell"

# ── Helpers ────────────────────────────────────────────────────────────────────
function Write-Step  { param([string]$Msg) Write-Host "  ● $Msg" -ForegroundColor Cyan }
function Write-Ok    { param([string]$Msg) Write-Host "  ✔ $Msg" -ForegroundColor Green }
function Write-Warn  { param([string]$Msg) Write-Host "  ⚠ $Msg" -ForegroundColor Yellow }
function Write-Fail  { param([string]$Msg) Write-Host "  ✖ $Msg" -ForegroundColor Red }

function Abort {
    param([string]$Msg)
    Write-Fail $Msg
    # Cleanup any temp files
    if ($script:TempDir -and (Test-Path $script:TempDir)) {
        Remove-Item -Recurse -Force $script:TempDir -ErrorAction SilentlyContinue
    }
    exit 1
}

# ── Banner ─────────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  ╔═══════════════════════════════════════╗" -ForegroundColor Magenta
Write-Host "  ║       Hoppr Installer for Windows      ║" -ForegroundColor Magenta
Write-Host "  ║   The fast lane to your projects  🚀   ║" -ForegroundColor Magenta
Write-Host "  ╚═══════════════════════════════════════╝" -ForegroundColor Magenta
Write-Host ""

# ── 1. Detect Architecture ────────────────────────────────────────────────────
Write-Step "Detecting system architecture..."

if (-not [System.Environment]::Is64BitOperatingSystem) {
    Abort "Hoppr requires a 64-bit Windows installation."
}

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "ARM64" { "arm64" }
    default { "amd64" }
}
Write-Ok "Architecture: windows/$Arch"

# ── 2. Resolve Latest Release ─────────────────────────────────────────────────
Write-Step "Fetching latest release from GitHub..."

$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
try {
    $Headers  = @{ "User-Agent" = "Hoppr-Installer/1.0" }
    $Release  = Invoke-RestMethod -Uri $ApiUrl -Headers $Headers -UseBasicParsing
    $Tag      = $Release.tag_name
} catch {
    Write-Warn "GitHub API unreachable. Falling back to v1.0.0."
    $Tag = "v1.0.0"
}
Write-Ok "Release: $Tag"

# ── 3. Prepare Temp Directory ─────────────────────────────────────────────────
$script:TempDir = Join-Path $env:TEMP "hoppr-install-$(Get-Random)"
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

$AssetName   = "hoppr-$Tag-windows-$Arch.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$AssetName"
$ChecksumUrl = "https://github.com/$Repo/releases/download/$Tag/checksums.txt"
$ZipPath     = Join-Path $TempDir $AssetName

# ── 4. Download Archive ───────────────────────────────────────────────────────
Write-Step "Downloading $AssetName..."

try {
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
} catch {
    Abort "Failed to download release archive. Check your internet connection or visit https://github.com/$Repo/releases"
}
Write-Ok "Downloaded $(([math]::Round((Get-Item $ZipPath).Length / 1MB, 2))) MB"

# ── 5. Verify SHA256 Checksum ─────────────────────────────────────────────────
Write-Step "Verifying SHA256 checksum..."

$ChecksumPath = Join-Path $TempDir "checksums.txt"
try {
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumPath -UseBasicParsing
} catch {
    Write-Warn "Could not download checksums.txt — skipping verification."
    $ChecksumPath = $null
}

if ($ChecksumPath -and (Test-Path $ChecksumPath)) {
    $ExpectedLine = Get-Content $ChecksumPath | Where-Object { $_ -match [regex]::Escape($AssetName) } | Select-Object -First 1
    if ($ExpectedLine) {
        $ExpectedHash = ($ExpectedLine -split '\s+')[0].Trim().ToUpper()
        $ActualHash   = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash.ToUpper()

        if ($ExpectedHash -ne $ActualHash) {
            Abort "SHA256 MISMATCH! Expected: $ExpectedHash  Got: $ActualHash — The download may be corrupted or tampered with."
        }
        Write-Ok "SHA256 verified: $($ActualHash.Substring(0, 16))..."
    } else {
        Write-Warn "No checksum entry found for $AssetName — skipping verification."
    }
} else {
    Write-Warn "Checksum verification skipped."
}

# ── 6. Extract and Install ────────────────────────────────────────────────────
Write-Step "Installing to $InstallRoot..."

# Create install directories
foreach ($Dir in @($BinDir, $ShellDir)) {
    if (-not (Test-Path $Dir)) {
        New-Item -ItemType Directory -Path $Dir -Force | Out-Null
    }
}

# Extract archive
$ExtractDir = Join-Path $TempDir "extracted"
Expand-Archive -Path $ZipPath -DestinationPath $ExtractDir -Force

# Locate and copy binary
$ExtractedBin = Get-ChildItem -Path $ExtractDir -Filter $BinName -Recurse | Select-Object -First 1
if (-not $ExtractedBin) {
    Abort "Could not find $BinName in the release archive. The release may be corrupt."
}
Copy-Item -Path $ExtractedBin.FullName -Destination (Join-Path $BinDir $BinName) -Force

# Copy shell integration scripts from archive
$ExtractedShellDir = Get-ChildItem -Path $ExtractDir -Filter "shell" -Directory -Recurse | Select-Object -First 1
if ($ExtractedShellDir) {
    Get-ChildItem -Path $ExtractedShellDir.FullName -File | ForEach-Object {
        Copy-Item -Path $_.FullName -Destination (Join-Path $ShellDir $_.Name) -Force
    }
} else {
    # Fallback: download shell wrapper directly
    Write-Warn "Shell scripts not found in archive. Downloading hop.ps1 directly..."
    $PsWrapperUrl = "https://raw.githubusercontent.com/$Repo/main/shell/hop.ps1"
    try {
        Invoke-WebRequest -Uri $PsWrapperUrl -OutFile (Join-Path $ShellDir "hop.ps1") -UseBasicParsing
    } catch {
        Write-Warn "Could not download hop.ps1 — shell integration will need manual setup."
    }
}

Write-Ok "Binary installed to $BinDir"

# ── 7. Configure User PATH ────────────────────────────────────────────────────
Write-Step "Configuring system PATH..."

$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$BinDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$BinDir", "User")
    $env:PATH = "$env:PATH;$BinDir"
    Write-Ok "Added $BinDir to User PATH"
} else {
    Write-Ok "PATH already configured"
}

# ── 8. Configure PowerShell Profile ───────────────────────────────────────────
Write-Step "Setting up PowerShell shell integration..."

$PsWrapper = Join-Path $ShellDir "hop.ps1"
if (Test-Path $PsWrapper) {
    if (-not (Test-Path $PROFILE)) {
        New-Item -ItemType File -Path $PROFILE -Force | Out-Null
    }

    $ProfileContent = Get-Content -Path $PROFILE -Raw -ErrorAction SilentlyContinue
    $SourceLine = ". `"$PsWrapper`""

    if (-not $ProfileContent -or $ProfileContent -notlike "*hop.ps1*") {
        # Backup existing profile before modifying
        if ($ProfileContent) {
            $BackupPath = "$PROFILE.hoppr-backup"
            Copy-Item -Path $PROFILE -Destination $BackupPath -Force
            Write-Ok "Profile backed up to $BackupPath"
        }
        Add-Content -Path $PROFILE -Value "`n# Hoppr Shell Integration (added by installer)`n$SourceLine"
        Write-Ok "Shell integration added to `$PROFILE"
    } else {
        Write-Ok "Shell integration already configured"
    }

    # Activate in current session
    try { . $PsWrapper } catch { Write-Warn "Could not activate in current session. Restart your terminal." }
} else {
    Write-Warn "hop.ps1 not found — shell integration skipped."
}

# ── 9. Cleanup ─────────────────────────────────────────────────────────────────
Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue

# ── Done ───────────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  ╔═══════════════════════════════════════╗" -ForegroundColor Green
Write-Host "  ║     Hoppr installed successfully! 🎉   ║" -ForegroundColor Green
Write-Host "  ╚═══════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""
Write-Host "  Get started:" -ForegroundColor White
Write-Host "    hop doctor          — verify installation health" -ForegroundColor Gray
Write-Host "    hop add .           — save current folder as a project" -ForegroundColor Gray
Write-Host "    hop list            — view all saved projects" -ForegroundColor Gray
Write-Host "    hop <project>       — jump to any saved project" -ForegroundColor Gray
Write-Host ""
Write-Host "  Docs: https://github.com/$Repo" -ForegroundColor DarkGray
Write-Host ""
