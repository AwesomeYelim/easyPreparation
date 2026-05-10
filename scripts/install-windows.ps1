# easyPreparation Windows Installer
# Usage: irm https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/install-windows.ps1 | iex

$ErrorActionPreference = "Stop"
$ProgressPreference    = "SilentlyContinue"

$AppName     = "easyPreparation"
$SetupName   = "${AppName}_desktop_windows_amd64_setup.exe"
$DownloadUrl = "https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/$SetupName"
$TempSetup   = Join-Path $env:TEMP $SetupName

# --- 1. Check for existing installation ---
$knownPaths = @(
    "$env:LOCALAPPDATA\$AppName\$AppName.exe",
    "$env:LOCALAPPDATA\Programs\$AppName\$AppName.exe",
    "$env:ProgramFiles\$AppName\$AppName.exe",
    "${env:ProgramFiles(x86)}\$AppName\$AppName.exe"
)

function Find-InstalledExe {
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
    return ($knownPaths | Where-Object { Test-Path $_ } | Select-Object -First 1)
}

# --- 2. Kill running instance ---
$running = Get-Process -Name $AppName -ErrorAction SilentlyContinue
if ($running) {
    Write-Host "  Running instance detected, stopping..." -ForegroundColor Yellow
    $running | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
}

$existingExe = Find-InstalledExe
if ($existingExe) {
    Write-Host "  Existing installation: $existingExe" -ForegroundColor Yellow
    Write-Host "  Upgrading..." -ForegroundColor Cyan
} else {
    Write-Host "  Fresh install" -ForegroundColor Cyan
}

# --- 3. Download ---
Write-Host "  Downloading $AppName..." -ForegroundColor Cyan
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempSetup -UseBasicParsing
} catch {
    Write-Host "  ERROR: Download failed - $_" -ForegroundColor Red
    exit 1
}
Unblock-File -Path $TempSetup -ErrorAction SilentlyContinue

# --- 4. Silent install ---
Write-Host "  Installing..." -ForegroundColor Cyan
$proc = Start-Process -FilePath $TempSetup -ArgumentList "/S" -Wait -PassThru
Remove-Item -Path $TempSetup -Force -ErrorAction SilentlyContinue

if ($proc.ExitCode -ne 0) {
    Write-Host "  Installation failed (exit code: $($proc.ExitCode))." -ForegroundColor Red
    exit 1
}

# --- 5. Launch ---
Start-Sleep -Seconds 1
$exePath = Find-InstalledExe

if ($exePath) {
    Write-Host ""
    Write-Host "  Installed: $exePath" -ForegroundColor Green
    Start-Process -FilePath $exePath
    Write-Host "  Launched!" -ForegroundColor Green
} else {
    Write-Host "  Install completed but exe not found." -ForegroundColor Yellow
    Write-Host "  Check Start Menu for $AppName" -ForegroundColor Gray
}
Write-Host ""
