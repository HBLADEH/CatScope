$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot

function Invoke-CheckedCommand {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$FilePath,
        [string[]]$Arguments = @(),
        [string]$WorkingDirectory = $Root
    )

    Write-Host ""
    Write-Host "==> $Name"
    Push-Location $WorkingDirectory
    try {
        & $FilePath @Arguments
        $exitCode = $LASTEXITCODE
    }
    finally {
        Pop-Location
    }

    if ($exitCode -ne 0) {
        throw "$Name failed with exit code $exitCode."
    }
}

Invoke-CheckedCommand -Name "Go tests" -FilePath "go" -Arguments @("test", "./...")
Invoke-CheckedCommand -Name "Frontend build" -FilePath "npm" -Arguments @("run", "build") -WorkingDirectory (Join-Path $Root "frontend")
Invoke-CheckedCommand -Name "Wails Windows build" -FilePath "wails" -Arguments @("build")

$artifactDirs = @(
    (Join-Path $Root "frontend\dist"),
    (Join-Path $Root "build\bin"),
    (Join-Path $Root "dist")
)

Write-Host ""
Write-Host "Build artifacts:"
foreach ($dir in $artifactDirs) {
    if (Test-Path $dir) {
        Write-Host "- $dir"
        Get-ChildItem -Path $dir -File -Recurse | ForEach-Object {
            $size = "{0:N2} MB" -f ($_.Length / 1MB)
            Write-Host "  $($_.FullName) ($size)"
        }
    }
}

$exeFiles = Get-ChildItem -Path (Join-Path $Root "build\bin") -Filter "*.exe" -File -ErrorAction SilentlyContinue
if ($exeFiles) {
    Write-Host ""
    Write-Host "Executable sizes:"
    foreach ($exe in $exeFiles) {
        $size = "{0:N2} MB" -f ($exe.Length / 1MB)
        Write-Host "- $($exe.FullName): $size"
    }
}
else {
    Write-Host ""
    Write-Host "No executable was found under build\bin."
}

Write-Host ""
Write-Host "Windows build completed successfully."
