$ErrorActionPreference = 'Stop'

$Repo = if ($env:JLINK_CLI_REPO) { $env:JLINK_CLI_REPO } else { 'guajun/jlink-cli' }
$InstallDir = if ($env:JLINK_CLI_INSTALL_DIR) { $env:JLINK_CLI_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\jlink-cli\bin' }

if (-not [Environment]::Is64BitOperatingSystem) {
    throw 'jlink-cli release binaries require a 64-bit Windows OS.'
}

$Arch = if ([Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq [Runtime.InteropServices.Architecture]::Arm64) { 'arm64' } else { 'amd64' }
$ReleaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
$Headers = @{ 'User-Agent' = 'jlink-cli-install' }
$Release = Invoke-RestMethod -Headers $Headers -Uri $ReleaseUrl
$Asset = $Release.assets | Where-Object { $_.name -match "(?i)windows[_-]$Arch\.zip$" } | Select-Object -First 1

if (-not $Asset) {
    throw "No Windows $Arch release asset found for $Repo release $($Release.tag_name)."
}

$TempDir = Join-Path ([IO.Path]::GetTempPath()) "jlink-cli-$([Guid]::NewGuid())"
$Archive = Join-Path $TempDir $Asset.name
New-Item -ItemType Directory -Path $TempDir, $InstallDir -Force | Out-Null

try {
    Invoke-WebRequest -Headers $Headers -Uri $Asset.browser_download_url -OutFile $Archive
    Expand-Archive -Path $Archive -DestinationPath $TempDir -Force
    $Binary = Get-ChildItem -Path $TempDir -Filter 'jlink-cli.exe' -Recurse | Select-Object -First 1
    if (-not $Binary) {
        throw 'Downloaded archive did not contain jlink-cli.exe.'
    }

    Copy-Item -Path $Binary.FullName -Destination (Join-Path $InstallDir 'jlink-cli.exe') -Force
    Write-Host "Installed jlink-cli $($Release.tag_name) to $InstallDir"

    $UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $PathParts = @($UserPath -split ';' | Where-Object { $_ })
    if ($PathParts -notcontains $InstallDir) {
        [Environment]::SetEnvironmentVariable('Path', ($PathParts + $InstallDir -join ';'), 'User')
        Write-Host 'Added install directory to your user PATH. Open a new terminal to use jlink-cli from PATH.'
    }
} finally {
    Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
}
