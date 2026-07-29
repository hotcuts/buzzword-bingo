# Install bingo from the latest (or pinned) GitHub Release.
#
# No Go toolchain or git clone required — downloads the published
# Windows amd64 binary.
#
# Usage:
#   irm https://raw.githubusercontent.com/hotcuts/buzzword-bingo/main/scripts/install.ps1 | iex
#   # or from a checkout:
#   .\scripts\install.ps1
#
# Env overrides:
#   REPO         default: hotcuts/buzzword-bingo
#   INSTALL_DIR  default: $env:LOCALAPPDATA\bingo
#   VERSION      latest | vX.Y.Z  (default: latest)
#   ASSET        default: bingo_windows_amd64.exe

$ErrorActionPreference = "Stop"

$Repo = if ($env:REPO) { $env:REPO } else { "hotcuts/buzzword-bingo" }
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "bingo" }
$Version = if ($env:VERSION) { $env:VERSION } else { "latest" }
$Asset = if ($env:ASSET) { $env:ASSET } else { "bingo_windows_amd64.exe" }
$BinaryName = "bingo.exe"

function Write-Step {
    param([string]$Label, [string]$Message)
    Write-Host ("  · {0,-8} {1}" -f $Label, $Message) -ForegroundColor DarkGray
}

function Write-Ok {
    param([string]$Label, [string]$Message)
    Write-Host ("  ✓ {0,-8} {1}" -f $Label, $Message) -ForegroundColor Green
}

function Write-Err {
    param([string]$Message)
    Write-Host ("  ✗ {0}" -f $Message) -ForegroundColor Red
    exit 1
}

function Test-WindowsAmd64 {
    $arch = $env:PROCESSOR_ARCHITECTURE
    if ($arch -ne "AMD64") {
        Write-Err "bingo on Windows only supports amd64, got: $arch"
    }
}

function Get-DownloadUrl {
    $base = "https://github.com/$Repo/releases"
    if ($Version -eq "latest") {
        return "$base/latest/download/$Asset"
    }
    $tag = $Version
    if (-not $tag.StartsWith("v")) {
        $tag = "v$tag"
    }
    return "$base/download/$tag/$Asset"
}

function Ensure-DirOnPath {
    param([string]$Dir)

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (-not $userPath) { $userPath = "" }

    $parts = $userPath -split ";" | Where-Object { $_ -ne "" }
    $already = $parts | Where-Object { $_.TrimEnd("\") -ieq $Dir.TrimEnd("\") }
    if ($already) {
        Write-Ok "path" "already in user PATH"
    }
    else {
        Write-Step "path" "adding $Dir to user PATH"
        $newPath = if ($userPath.Trim() -eq "") { $Dir } else { "$Dir;$userPath" }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Ok "path" "added to user PATH"
    }

    # Current session
    $sessionParts = $env:Path -split ";" | Where-Object { $_ -ne "" }
    $inSession = $sessionParts | Where-Object { $_.TrimEnd("\") -ieq $Dir.TrimEnd("\") }
    if (-not $inSession) {
        $env:Path = "$Dir;$env:Path"
    }
}

Write-Host ""
Write-Host "  bingo soft install · Windows amd64" -ForegroundColor Cyan
Write-Host ""

Write-Step "detect" "checking platform…"
Test-WindowsAmd64
Write-Ok "detect" "Windows amd64 → $Asset"

$url = Get-DownloadUrl
$dest = Join-Path $InstallDir $BinaryName
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("bingo-install-" + [guid]::NewGuid().ToString() + ".exe")

try {
    Write-Step "fetch" "downloading release…"
    try {
        Invoke-WebRequest -Uri $url -OutFile $tmp -UseBasicParsing
    }
    catch {
        Write-Err "download failed — is there a release with asset $Asset on $Repo?"
    }

    if (-not (Test-Path $tmp) -or (Get-Item $tmp).Length -eq 0) {
        Write-Err "downloaded file is empty"
    }
    Write-Ok "fetch" $Asset

    Write-Step "install" "writing $dest"
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Force -Path $tmp -Destination $dest
    Write-Ok "install" $dest

    Ensure-DirOnPath $InstallDir

    $playable = $false
    $verMsg = "bingo"
    $bingoCmd = Get-Command bingo -ErrorAction SilentlyContinue
    if ($bingoCmd) {
        $playable = $true
        try {
            $verMsg = (& bingo version 2>$null)
            if (-not $verMsg) { $verMsg = "bingo" }
        }
        catch {
            $verMsg = "bingo"
        }
    }

    Write-Host ""
    Write-Host "  Ready · $verMsg" -ForegroundColor Cyan
    if ($playable) {
        Write-Host "  Run     bingo play" -ForegroundColor Cyan
    }
    else {
        Write-Host "  Open a new terminal, then bingo play" -ForegroundColor Cyan
    }
    Write-Host ""
}
finally {
    if (Test-Path $tmp) {
        Remove-Item -Force $tmp -ErrorAction SilentlyContinue
    }
}
