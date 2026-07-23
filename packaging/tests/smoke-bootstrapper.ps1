[CmdletBinding()]
param([Parameter(Mandatory = $true)][string]$Bootstrapper, [Parameter(Mandatory = $true)][string]$InstallRoot)
$ErrorActionPreference = 'Stop'
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]$identity
if ($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Run bootstrapper smoke as a standard user.' }
$log = Join-Path $env:TEMP 'wslc-tui-bootstrapper-smoke.log'
$result = Start-Process $Bootstrapper -ArgumentList "/quiet /norestart /log `"$log`"" -Wait -PassThru
if ($result.ExitCode -notin @(0, 3010)) { throw "Bootstrapper failed with $($result.ExitCode)" }
$exe = Join-Path $InstallRoot 'wslc-tui.exe'
if (-not (Test-Path $exe)) { throw 'Bootstrapper did not install wslc-tui.exe' }
if ((& $exe --version) -notmatch 'distribution=installer') { throw 'Bootstrapper installed non-installer payload' }
$logText = Get-Content $log -Raw
if ($logText -notmatch '(?i)\.msi') { throw 'Bootstrapper log does not identify an MSI payload' }
Write-Output 'Bootstrapper smoke passed.'
