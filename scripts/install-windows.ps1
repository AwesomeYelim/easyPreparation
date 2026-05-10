# easyPreparation Windows Installer
# Usage: irm https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/install-windows.ps1 | iex

$ErrorActionPreference = "Stop"
$ProgressPreference    = "SilentlyContinue"

$AppName     = "easyPreparation"
$SetupName   = "${AppName}_desktop_windows_amd64_setup.exe"
$DownloadUrl = "https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/$SetupName"
$TempSetup   = Join-Path $env:TEMP $SetupName

function Find-InstalledExe {
    # 1) Registry
    $regPaths = @(
        "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\$AppName",
        "HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\$AppName"
    )
    foreach ($rp in $regPaths) {
        if (Test-Path $rp) {
            $dir = (Get-ItemProperty -Path $rp -ErrorAction SilentlyContinue).InstallLocation
            if ($dir) {
                $candidate = Join-Path $dir "$AppName.exe"
                if (Test-Path $candidate) { return $candidate }
            }
        }
    }
    # 2) Known paths
    $paths = @(
        "$env:LOCALAPPDATA\Programs\$AppName\$AppName.exe",
        "$env:LOCALAPPDATA\$AppName\$AppName.exe",
        "$env:ProgramFiles\$AppName\$AppName.exe",
        "${env:ProgramFiles(x86)}\$AppName\$AppName.exe"
    )
    foreach ($p in $paths) {
        if (Test-Path $p) { return $p }
    }
    # 3) Broad search
    $found = Get-ChildItem -Path $env:LOCALAPPDATA -Filter "$AppName.exe" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($found) { return $found.FullName }
    return $null
}

# --- 1. Kill running instance ---
$running = Get-Process -Name $AppName -ErrorAction SilentlyContinue
if ($running) {
    Write-Host "  Running instance detected, stopping..." -ForegroundColor Yellow
    $running | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
}

# --- 2. Download ---
Write-Host ""
Write-Host "  Downloading $AppName..." -ForegroundColor Cyan
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempSetup -UseBasicParsing
} catch {
    Write-Host "  Download failed: $_" -ForegroundColor Red
    exit 1
}
Unblock-File -Path $TempSetup -ErrorAction SilentlyContinue
Write-Host "  Download complete." -ForegroundColor Green

# --- 3. Silent install ---
Write-Host "  Installing (silent)..." -ForegroundColor Cyan
$proc = Start-Process -FilePath $TempSetup -ArgumentList "/S" -Wait -PassThru
Remove-Item -Path $TempSetup -Force -ErrorAction SilentlyContinue

if ($proc.ExitCode -ne 0) {
    Write-Host "  Install failed (exit code: $($proc.ExitCode))." -ForegroundColor Red
    exit 1
}

# --- 4. Find and launch ---
Start-Sleep -Seconds 2
$exePath = Find-InstalledExe

if ($exePath) {
    Write-Host "  Installed: $exePath" -ForegroundColor Green
    Start-Process -FilePath $exePath
    Write-Host "  Launched!" -ForegroundColor Green
} else {
    Write-Host "  Install completed but exe not found." -ForegroundColor Yellow
    Write-Host "  Check Start Menu for $AppName" -ForegroundColor Gray
}
Write-Host ""
