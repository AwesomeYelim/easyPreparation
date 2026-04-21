# easyPreparation Windows 설치 스크립트
# 사용법: irm https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/install-windows.ps1 | iex

$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$AppName    = "easyPreparation"
$ExeName    = "${AppName}_desktop_windows_amd64.exe"
$DownloadUrl = "https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/$ExeName"
$InstallDir = Join-Path $env:LOCALAPPDATA $AppName
$ExePath    = Join-Path $InstallDir "${AppName}.exe"

# 설치 디렉토리 생성
New-Item -ItemType Directory -Force $InstallDir | Out-Null

Write-Host "다운로드 중..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $DownloadUrl -OutFile $ExePath -UseBasicParsing

Write-Host "SmartScreen 차단 해제 중..." -ForegroundColor Cyan
Unblock-File -Path $ExePath

# 바탕화면 바로가기 생성
$ShortcutPath = Join-Path ([Environment]::GetFolderPath("Desktop")) "${AppName}.lnk"
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut($ShortcutPath)
$Shortcut.TargetPath = $ExePath
$Shortcut.WorkingDirectory = $InstallDir
$Shortcut.Description = "easyPreparation"
$Shortcut.Save()

Write-Host "설치 완료!" -ForegroundColor Green
Write-Host "설치 경로: $ExePath" -ForegroundColor Gray
Write-Host "바탕화면 바로가기가 생성되었습니다." -ForegroundColor Gray
