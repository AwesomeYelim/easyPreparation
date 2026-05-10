# easyPreparation Windows Installer
# Usage: irm https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/install-windows.ps1 | iex

$ErrorActionPreference = "Stop"
$ProgressPreference    = "SilentlyContinue"

$AppName     = "easyPreparation"
$SetupName   = "${AppName}_desktop_windows_amd64_setup.exe"
$DownloadUrl = "https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/$SetupName"
$TempSetup   = Join-Path $env:TEMP $SetupName

# NSIS 설치 경로 후보 (wails.json companyName: "AwesomeYelim")
$CompanyName = "AwesomeYelim"

function Find-InstalledExe {
    # 1) Known paths (우선순위순)
    $paths = @(
        "$env:ProgramFiles\$CompanyName\$AppName\$AppName.exe",
        "${env:ProgramFiles(x86)}\$CompanyName\$AppName\$AppName.exe",
        "$env:LOCALAPPDATA\Programs\$AppName\$AppName.exe",
        "$env:ProgramFiles\$AppName\$AppName.exe",
        "$env:LOCALAPPDATA\$AppName\$AppName.exe"
    )
    foreach ($p in $paths) {
        if (Test-Path $p) { return $p }
    }
    # 2) Registry
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
    # 3) Broad search
    foreach ($root in @("$env:ProgramFiles", "$env:LOCALAPPDATA")) {
        $found = Get-ChildItem -Path $root -Filter "$AppName.exe" -Recurse -Depth 3 -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($found) { return $found.FullName }
    }
    return $null
}

# --- 1. Kill running instance ---
$running = Get-Process -Name $AppName -ErrorAction SilentlyContinue
if ($running) {
    Write-Host "  Stopping running instance..." -ForegroundColor Yellow
    $running | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
}

# --- 2. Clean up old installation (Programs\ 아닌 곳) ---
$oldPath = "$env:LOCALAPPDATA\$AppName"
if (Test-Path "$oldPath\$AppName.exe") {
    Write-Host "  Removing old installation: $oldPath" -ForegroundColor Yellow
    Remove-Item -Path $oldPath -Recurse -Force -ErrorAction SilentlyContinue
}

# --- 3. Download ---
Write-Host "  Downloading $AppName..." -ForegroundColor Cyan
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempSetup -UseBasicParsing
} catch {
    Write-Host "  Download failed: $_" -ForegroundColor Red
    exit 1
}
Unblock-File -Path $TempSetup -ErrorAction SilentlyContinue

# --- 4. Silent install ---
Write-Host "  Installing..." -ForegroundColor Cyan
$proc = Start-Process -FilePath $TempSetup -ArgumentList "/S" -Wait -PassThru
Remove-Item -Path $TempSetup -Force -ErrorAction SilentlyContinue

if ($proc.ExitCode -ne 0) {
    Write-Host "  Install failed (exit code: $($proc.ExitCode))." -ForegroundColor Red
    exit 1
}

# --- 5. Find and launch ---
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
