[CmdletBinding()]
param([Parameter(Mandatory = $true)][string]$Msi, [Parameter(Mandatory = $true)][string]$Bootstrapper, [Parameter(Mandatory = $true)][string]$InstallRoot)
$ErrorActionPreference = 'Stop'
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]$identity
if ($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Run this smoke test as a standard user, not administrator.' }
$log = Join-Path $env:TEMP 'wslc-tui-msi-smoke.log'
$result = Start-Process msiexec.exe -ArgumentList "/i `"$Msi`" /qn /l*v `"$log`"" -Wait -PassThru
if ($result.ExitCode -ne 0) { throw "Per-user MSI install failed with $($result.ExitCode); see $log" }
$exe = Join-Path $InstallRoot 'wslc-tui.exe'
if (-not (Test-Path $exe)) { throw "Installed executable missing: $exe" }
$metadata = & $exe --version
if ($metadata -notmatch 'distribution=installer') { throw "Installer metadata missing: $metadata" }
$key = 'HKCU:\Software\japperJ\wslc-tui-ms'
foreach ($name in 'InstallRoot', 'Version', 'Distribution') { if (-not (Get-ItemProperty $key -Name $name -ErrorAction SilentlyContinue)) { throw "Missing HKCU value $name" } }
if (Test-Path 'HKLM:\Software\japperJ\wslc-tui-ms') { throw 'Unexpected HKLM registration' }
if (Get-Service -Name 'wslc-tui-ms' -ErrorAction SilentlyContinue) { throw 'Unexpected Windows service' }
Write-Output "Per-user MSI smoke passed. Bootstrapper artifact: $Bootstrapper"
