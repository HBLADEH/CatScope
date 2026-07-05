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
Invoke-CheckedCommand -Name "Wails doctor" -FilePath "wails" -Arguments @("doctor")
Invoke-CheckedCommand -Name "Git whitespace check" -FilePath "git" -Arguments @("diff", "--check")

Write-Host ""
Write-Host "Pre-release checks completed successfully."
