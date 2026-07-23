[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)][string]$Msi,
  [Parameter(Mandatory = $true)][string]$Bootstrapper,
  [Parameter(Mandatory = $true)][string]$InstallRoot,
  [switch]$PreflightOnly,
  [switch]$AllowDeferred
)
$ErrorActionPreference = 'Stop'
$missing = [System.Collections.Generic.List[string]]::new()
if (-not $env:windir) { $missing.Add('Windows x64 test host; run on a Windows 10/11 VM.') }
if (-not [Environment]::Is64BitOperatingSystem) { $missing.Add('Windows x64 packaging tools and operating system.') }
$wix = Get-Command wix -ErrorAction SilentlyContinue
if (-not $wix) { $missing.Add('WiX Toolset v4.0.5; install with dotnet tool install --global wix --version 4.0.5.') }
elseif ((& wix --version 2>$null) -notmatch '4\.0\.5') { $missing.Add('pinned WiX Toolset v4.0.5; the active wix command has a different version.') }
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]$identity
if ($env:windir -and $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { $missing.Add('fresh standard-user account with a medium-integrity token; do not run as administrator.') }
$uac = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System' -Name EnableLUA -ErrorAction SilentlyContinue
if ($env:windir -and $uac.EnableLUA -ne 1) { $missing.Add('UAC enabled (HKLM EnableLUA=1).') }
if ($missing.Count -gt 0) {
  $message = "SMOKE_RESULT=blocked; missing prerequisite(s): $($missing -join ' | ')"
  if ($AllowDeferred) { Write-Warning "$message Set up the prerequisite and rerun without -AllowDeferred."; exit 0 }
  throw $message
}
if ($PreflightOnly) { Write-Output 'SMOKE_RESULT=ready'; exit 0 }
if ($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Run this smoke test as a standard user, not administrator.' }
$log = Join-Path $env:TEMP 'wslc-tui-msi-smoke.log'
$result = Start-Process msiexec.exe -ArgumentList "/i `"$Msi`" /qn /l*v `"$log`"" -Wait -PassThru
if ($result.ExitCode -ne 0) { throw "Per-user MSI install failed with $($result.ExitCode); see $log" }
$exe = Join-Path $InstallRoot 'wslc-tui.exe'
if (-not (Test-Path $exe)) { throw "Installed executable missing: $exe" }
$metadata = & $exe --version
if ($metadata -notmatch 'distribution=installer') { throw "Installer metadata missing: $metadata" }
$key = 'HKCU:\Software\japperJ\wslc-tui-ms'
$values = Get-ItemProperty $key -ErrorAction Stop
if ($values.InstallRoot.TrimEnd('\') -ne $InstallRoot.TrimEnd('\')) { throw "HKCU InstallRoot mismatch: $($values.InstallRoot)" }
$metadataVersion = [regex]::Match($metadata, 'wslc-tui ([^ ]+)').Groups[1].Value
if ($values.Version -ne $metadataVersion) { throw "HKCU Version mismatch: $($values.Version)" }
if ($values.Distribution -ne 'installer') { throw "HKCU Distribution mismatch: $($values.Distribution)" }
if (Test-Path 'HKLM:\Software\japperJ\wslc-tui-ms') { throw 'Unexpected HKLM registration' }
if (Get-Service -Name 'wslc-tui-ms' -ErrorAction SilentlyContinue) { throw 'Unexpected Windows service' }
if (Get-ScheduledTask -TaskName 'wslc-tui-ms' -ErrorAction SilentlyContinue) { throw 'Unexpected scheduled task' }
if ($InstallRoot -like "$env:ProgramFiles*") { throw 'Installed under Program Files instead of LocalAppData' }
Write-Output "Per-user MSI smoke passed. Bootstrapper artifact: $Bootstrapper"
