# Verify Release Files

Use the release checksum manifest before running a downloaded installer or executable.

## Identify The Files

A release contains files similar to:

```text
wslc-tui-v1.2.3-windows-amd64.msi
wslc-tui-v1.2.3-windows-amd64.exe
wslc-tui-v1.2.3-windows-amd64-portable.zip
wslc-tui-v1.2.3-checksums.json
update-policy.json
```

The manifest uses SHA-256 and covers the MSI, bootstrapper, portable ZIP, and policy file. It does not checksum itself.

## Calculate A SHA-256 Hash

In PowerShell, run:

```powershell
Get-FileHash .\wslc-tui-v1.2.3-windows-amd64-portable.zip -Algorithm SHA256
```

Compare the returned `Hash` value with the matching `sha256` value in the checksum manifest. Repeat for each file you plan to use.

## Investigate A Mismatch

Do not run a file whose hash differs from the published manifest. Download it again from the release page. If the mismatch continues, report the release tag, asset name, and calculated hash without sharing private files or credentials.
