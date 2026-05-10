# easyPreparation Windows Installer
# Usage: irm https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/install-windows.ps1 | iex

$ErrorActionPreference = "Stop"
$ProgressPreference    = "SilentlyContinue"

$AppName     = "easyPreparation"
$SetupName   = "${AppName}_desktop_windows_amd64_setup.exe"
$DownloadUrl = "https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/$SetupName"
$TempSetup   = Join-Path $env:TEMP $SetupName

# --- 1. Check for existing installation ---
$regPath    = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\$AppName"
$regPathM   = "HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\$AppName"
$knownPaths = @(
    "$env:LOCALAPPDATA\$AppName\$AppName.exe",
    "$env:LOCALAPPDATA\Programs\$AppName\$AppName.exe",
    "$env:ProgramFiles\$AppName\$AppName.exe",
    "${env:ProgramFiles(x86)}\$AppName\$AppName.exe"
)

function Find-InstalledExe {
    # Try registry first
    foreach ($rp in @($regPath, $regPathM)) {
        if (Test-Path $rp) {
            $dir = (Get-ItemProperty -Path $rp -ErrorAction SilentlyContinue).InstallLocation
            if ($dir) {
                $candidate = Join-Path $dir "$AppName.exe"
                if (Test-Path $candidate) { return $candidate }
            }
        }
    }
    # Fallback to known paths
    return ($knownPaths | Where-Object { Test-Path $_ } | Select-Object -First 1)
}

$existingExe = Find-InstalledExe

if ($existingExe) {
    Write-Host ""
    Write-Host "  Existing installation detected: $existingExe" -ForegroundColor Yellow
    Write-Host "  Upgrading..." -ForegroundColor Cyan
}

# --- 2. Download installer ---
Write-Host ""
Write-Host "  Downloading $AppName installer..." -ForegroundColor Cyan

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempSetup -UseBasicParsing
} catch {
    Write-Host "  ERROR: Download failed — $_" -ForegroundColor Red
    exit 1
}

Unblock-File -Path $TempSetup -ErrorAction SilentlyContinue

# --- 3. Run NSIS installer (shows GUI) ---
Write-Host "  Starting installer GUI..." -ForegroundColor Cyan
Write-Host "  (Follow the on-screen instructions to complete setup)" -ForegroundColor Gray
Write-Host ""

$proc = Start-Process -FilePath $TempSetup -Wait -PassThru

Remove-Item -Path $TempSetup -Force -ErrorAction SilentlyContinue

if ($proc.ExitCode -ne 0) {
    Write-Host ""
    Write-Host "  Installation was cancelled or failed (exit code: $($proc.ExitCode))." -ForegroundColor Yellow
    exit 1
}

# --- 4. Launch option ---
$exePath = Find-InstalledExe

Write-Host ""
Write-Host "  Installation complete!" -ForegroundColor Green

if ($exePath) {
    Write-Host "  Installed at: $exePath" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Launching $AppName..." -ForegroundColor Cyan
    try {
        Start-Process -FilePath $exePath
        Write-Host "  Launched!" -ForegroundColor Green
    } catch {
        Write-Host "  Auto-launch failed: $_" -ForegroundColor Yellow
        Write-Host "  Please run manually: $exePath" -ForegroundColor Gray
    }
} else {
    Write-Host "  Could not find installed exe." -ForegroundColor Yellow
    Write-Host "  Please launch from Start Menu or find easyPreparation.exe manually." -ForegroundColor Gray
}

Write-Host ""
