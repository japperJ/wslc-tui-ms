[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)][string]$Tag,
  [string]$OutputDirectory = (Join-Path (Get-Location) 'artifacts'),
  [string]$DependencyRoot,
  [string]$OscdimgPath
)
$ErrorActionPreference = 'Stop'
if ($Tag -notmatch '^v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$') { throw "Tag must be a SemVer tag: $Tag" }
$repo = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$bundleName = "wslc-tui-ms-phase1-packaging-smoke-$Tag"
$bundle = Join-Path $OutputDirectory $bundleName
$iso = Join-Path $OutputDirectory "$bundleName.iso"
if (Test-Path $iso) { throw "Refusing to overwrite existing ISO: $iso" }
if (Test-Path $bundle) { throw "Refusing to overwrite existing staging directory: $bundle" }
New-Item -ItemType Directory -Force -Path $bundle, (Join-Path $bundle 'repository'), (Join-Path $bundle 'tools') | Out-Null

function Copy-RepoItem([string]$RelativePath) {
  $source = Join-Path $repo $RelativePath
  if (-not (Test-Path $source)) { throw "Required bundle input is missing: $RelativePath" }
  $destination = Join-Path $bundle ('repository\' + $RelativePath)
  $parent = Split-Path $destination -Parent
  New-Item -ItemType Directory -Force -Path $parent | Out-Null
  Copy-Item $source $destination -Recurse -Force
}

function Write-ImapiIso([string]$SourceDirectory, [string]$IsoPath, [string]$VolumeName) {
  if (-not ('ImapiIsoWriter' -as [type])) {
    Add-Type -TypeDefinition @'
using System;
using System.IO;
using System.Runtime.InteropServices;
using System.Runtime.InteropServices.ComTypes;
public static class ImapiIsoWriter {
  public static long Write(object source, string path) {
    IntPtr unknown = Marshal.GetIUnknownForObject(source);
    IntPtr interfacePointer = IntPtr.Zero;
    Guid iid = typeof(IStream).GUID;
    Marshal.QueryInterface(unknown, iid, out interfacePointer);
    try {
      IStream stream = (IStream)Marshal.GetTypedObjectForIUnknown(interfacePointer, typeof(IStream));
      IntPtr count = Marshal.AllocHGlobal(4);
      byte[] buffer = new byte[1048576];
      try {
        long total = 0;
        using (FileStream output = new FileStream(path, FileMode.CreateNew, FileAccess.Write, FileShare.None)) {
          while (true) {
            Marshal.WriteInt32(count, 0);
            stream.Read(buffer, buffer.Length, count);
            int read = Marshal.ReadInt32(count);
            if (read <= 0) break;
            output.Write(buffer, 0, read);
            total += read;
          }
        }
        return total;
      } finally { Marshal.FreeHGlobal(count); }
    } finally {
      if (interfacePointer != IntPtr.Zero) Marshal.Release(interfacePointer);
      Marshal.Release(unknown);
    }
  }
}
'@
  }
  $filesystemImage = New-Object -ComObject IMAPI2FS.MsftFileSystemImage
  try {
    $filesystemImage.FileSystemsToCreate = 7 # FsiFileSystemISO9660 + Joliet + UDF
    $filesystemImage.VolumeName = $VolumeName
    $filesystemImage.Root.AddTree($SourceDirectory, $false)
    $resultImage = $filesystemImage.CreateResultImage()
    [void][ImapiIsoWriter]::Write($resultImage.ImageStream, $IsoPath)
  } finally {
    if ($filesystemImage) { [Runtime.InteropServices.Marshal]::ReleaseComObject($filesystemImage) | Out-Null }
  }
}

foreach ($item in @('README.md', 'go.mod', 'go.sum', 'update-policy.json', 'main.go', 'internal', 'packaging', 'scripts', 'docs', '.github')) {
  Copy-RepoItem $item
}
Remove-Item (Join-Path $bundle 'repository\packaging\tests\node_modules') -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item (Join-Path $bundle 'repository\packaging\tests\*.log') -Force -ErrorAction SilentlyContinue
Copy-Item (Join-Path $repo 'packaging\iso\README.md') (Join-Path $bundle 'README.md') -Force
Copy-Item (Join-Path $repo 'packaging\iso\dependencies.json') (Join-Path $bundle 'tools\dependencies.json') -Force
if ($DependencyRoot) {
  if (-not (Test-Path $DependencyRoot -PathType Container)) { throw "Dependency root is not a directory: $DependencyRoot" }
  Copy-Item (Join-Path $DependencyRoot '*') (Join-Path $bundle 'tools') -Recurse -Force
}
New-Item -ItemType Directory -Force -Path (Join-Path $bundle 'results') | Out-Null
Copy-Item (Join-Path $repo 'packaging\tests\RUN-TESTS.ps1') (Join-Path $bundle 'RUN-TESTS.ps1') -Force

$manifest = [ordered]@{ schemaVersion = 1; tag = $Tag; files = @() }
foreach ($file in Get-ChildItem $bundle -File -Recurse | Where-Object { $_.Name -ne 'bundle-manifest.json' } | Sort-Object FullName) {
  $relative = $file.FullName.Substring($bundle.Length + 1).Replace('\', '/')
  $manifest.files += [ordered]@{ path = $relative; sizeBytes = $file.Length; sha256 = (Get-FileHash $file -Algorithm SHA256).Hash.ToLowerInvariant() }
}
$manifest | ConvertTo-Json -Depth 6 | Set-Content (Join-Path $bundle 'bundle-manifest.json') -Encoding utf8NoBOM

if (-not $OscdimgPath) {
  $command = Get-Command oscdimg.exe -ErrorAction SilentlyContinue
  if ($command) { $OscdimgPath = $command.Source }
}
if ($OscdimgPath) {
  if (-not (Test-Path $OscdimgPath)) { throw "ISO tool not found: $OscdimgPath" }
  & $OscdimgPath -m -o -u2 -udfver102 $bundle $iso
  if ($LASTEXITCODE -ne 0) { throw 'oscdimg failed.' }
  Write-Output "ISO=$iso"
} else {
  Write-Warning 'oscdimg.exe was not found. Using the built-in Windows IMAPI2FS ISO authoring fallback.'
  Write-ImapiIso $bundle $iso $bundleName
  if (-not (Test-Path $iso) -or (Get-Item $iso).Length -eq 0) { throw "IMAPI2FS created an empty or missing ISO: $iso" }
  Write-Output "ISO=$iso"
}
