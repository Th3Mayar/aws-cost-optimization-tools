# PowerShell installation script for coaws (Windows)
# Usage: irm https://raw.githubusercontent.com/Th3Mayar/aws-cost-optimization-tools/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'

# Repository information
$REPO = "Th3Mayar/aws-cost-optimization-tools"
$BINARY_NAME = "coaws.exe"
$INSTALL_DIR = "$env:LOCALAPPDATA\coaws"

Write-Host ""
Write-Host "    ██████╗ ██████╗  █████╗ ██╗    ██╗███████╗" -ForegroundColor Cyan
Write-Host "   ██╔════╝██╔═══██╗██╔══██╗██║    ██║██╔════╝" -ForegroundColor Cyan
Write-Host "   ██║     ██║   ██║███████║██║ █╗ ██║███████╗" -ForegroundColor Cyan
Write-Host "   ██║     ██║   ██║██╔══██║██║███╗██║╚════██║" -ForegroundColor Cyan
Write-Host "   ╚██████╗╚██████╔╝██║  ██║╚███╔███╔╝███████║" -ForegroundColor Cyan
Write-Host "    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝ ╚══╝╚══╝ ╚══════╝" -ForegroundColor Cyan
Write-Host ""
Write-Host "AWS Cost Optimization & Savings Tool" -ForegroundColor Blue
Write-Host ""

# Detect architecture
$ARCH = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

Write-Host "Detected Architecture: $ARCH" -ForegroundColor Green
Write-Host ""

# Get latest release version
Write-Host "Fetching latest release..." -ForegroundColor Blue
try {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest"
    $LATEST_VERSION = $release.tag_name
} catch {
    Write-Host "Failed to fetch latest version. Please check your internet connection." -ForegroundColor Red
    exit 1
}

Write-Host "Latest version: $LATEST_VERSION" -ForegroundColor Green

# Construct download URL
$VERSION_NO_V = $LATEST_VERSION.TrimStart('v')
$ARCHIVE_NAME = "coaws_${VERSION_NO_V}_windows_${ARCH}.zip"
$DOWNLOAD_URL = "https://github.com/$REPO/releases/download/$LATEST_VERSION/$ARCHIVE_NAME"

Write-Host "Downloading $ARCHIVE_NAME..." -ForegroundColor Blue

# Create temporary directory
$TMP_DIR = Join-Path $env:TEMP "coaws-install-$(Get-Random)"
New-Item -ItemType Directory -Path $TMP_DIR | Out-Null
$ARCHIVE_PATH = Join-Path $TMP_DIR $ARCHIVE_NAME

# Download archive
try {
    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $ARCHIVE_PATH
} catch {
    Write-Host "Failed to download $ARCHIVE_NAME" -ForegroundColor Red
    Write-Host "URL: $DOWNLOAD_URL" -ForegroundColor Red
    Remove-Item -Path $TMP_DIR -Recurse -Force
    exit 1
}

Write-Host "Download complete!" -ForegroundColor Green

# Extract archive
Write-Host "Extracting..." -ForegroundColor Blue
Expand-Archive -Path $ARCHIVE_PATH -DestinationPath $TMP_DIR -Force

# Create install directory if it doesn't exist
if (!(Test-Path $INSTALL_DIR)) {
    Write-Host "Creating $INSTALL_DIR..." -ForegroundColor Blue
    New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
}

# Install binary
Write-Host "Installing to $INSTALL_DIR..." -ForegroundColor Blue
$SOURCE_BINARY = Join-Path $TMP_DIR $BINARY_NAME
$DEST_BINARY = Join-Path $INSTALL_DIR $BINARY_NAME

Copy-Item -Path $SOURCE_BINARY -Destination $DEST_BINARY -Force

# Add to PATH if not already there
$USER_PATH = [Environment]::GetEnvironmentVariable("Path", "User")
if ($USER_PATH -notlike "*$INSTALL_DIR*") {
    Write-Host "Adding $INSTALL_DIR to PATH..." -ForegroundColor Blue
    [Environment]::SetEnvironmentVariable(
        "Path",
        "$USER_PATH;$INSTALL_DIR",
        "User"
    )
    $env:Path = "$env:Path;$INSTALL_DIR"
}

# Cleanup
Remove-Item -Path $TMP_DIR -Recurse -Force

Write-Host ""
Write-Host "✓ Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "To get started, run:" -ForegroundColor Cyan
Write-Host "  coaws --help"
Write-Host "  coaws start"
Write-Host ""
Write-Host "Note: You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
Write-Host ""
Write-Host "For more information, visit:" -ForegroundColor Cyan
Write-Host "  https://github.com/$REPO"
Write-Host ""
