[CmdletBinding()]
param(
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
if ($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Run bootstrapper smoke as a standard user.' }
$log = Join-Path $env:TEMP 'wslc-tui-bootstrapper-smoke.log'
$result = Start-Process $Bootstrapper -ArgumentList "/quiet /norestart /log `"$log`"" -Wait -PassThru
if ($result.ExitCode -notin @(0, 3010)) { throw "Bootstrapper failed with $($result.ExitCode)" }
$exe = Join-Path $InstallRoot 'wslc-tui.exe'
if (-not (Test-Path $exe)) { throw 'Bootstrapper did not install wslc-tui.exe' }
if ((& $exe --version) -notmatch 'distribution=installer') { throw 'Bootstrapper installed non-installer payload' }
$logText = Get-Content $log -Raw
$msiNames = @([regex]::Matches($logText, '(?i)[^\s"'']+\.msi') | ForEach-Object { $_.Value.ToLowerInvariant() } | Sort-Object -Unique)
if ($msiNames.Count -ne 1) { throw "Bootstrapper did not identify exactly one MSI payload: $($msiNames -join ', ')" }
Write-Output 'Bootstrapper smoke passed.'
