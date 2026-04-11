# easyPreparation Windows 설치 스크립트
# 사용법: irm https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/install-windows.ps1 | iex

$ErrorActionPreference = "Stop"

$AppName = "easyPreparation"
$ExeName = "${AppName}_desktop_windows_amd64_setup.exe"
$DownloadUrl = "https://github.com/AwesomeYelim/easyPreparation/releases/latest/download/$ExeName"
$TmpDir = Join-Path $env:TEMP "easyprep_install_$(Get-Random)"

try {
    New-Item -ItemType Directory -Force $TmpDir | Out-Null
    $ExePath = Join-Path $TmpDir $ExeName

    Write-Host "다운로드 중..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ExePath -UseBasicParsing

    Write-Host "SmartScreen 차단 해제 중..." -ForegroundColor Cyan
    Unblock-File -Path $ExePath

    Write-Host "설치 프로그램 실행 중..." -ForegroundColor Cyan
    Start-Process -FilePath $ExePath

    Write-Host "설치 완료! 설치 마법사의 안내를 따라주세요." -ForegroundColor Green
}
finally {
    # 설치 프로그램이 실행 중이므로 즉시 삭제하지 않음
    Start-Sleep -Seconds 3
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}
