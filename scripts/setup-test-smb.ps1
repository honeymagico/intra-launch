$ErrorActionPreference = 'Stop'
$root = Join-Path $PSScriptRoot '..\testdata\smb'
$sourceExe = Join-Path $env:WINDIR 'System32\notepad.exe'
$targets = @(
    'report-system\ReportSystem.exe',
    'purchase-manager\PurchaseManager.exe',
    'ops-toolkit\bin\OpsToolkit.exe'
)

foreach ($relativePath in $targets) {
    $target = Join-Path $root $relativePath
    New-Item -ItemType Directory -Force -Path (Split-Path $target) | Out-Null
    Copy-Item -LiteralPath $sourceExe -Destination $target -Force
}

Write-Host "Test SMB directory ready: $((Resolve-Path $root).Path)"
Write-Host 'Set smbBasePath in %AppData%\intra-launch\config.json to this directory.'
