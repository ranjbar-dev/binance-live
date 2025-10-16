# PowerShell script to generate protobuf Go code
# This script downloads protoc if not available and generates Go code

$protocVersion = "25.1"
$protocUrl = "https://github.com/protocolbuffers/protobuf/releases/download/v$protocVersion/protoc-$protocVersion-win64.zip"
$protocDir = "$env:TEMP\protoc"
$protocExe = "$protocDir\bin\protoc.exe"

# Check if protoc is already available
if (-not (Get-Command protoc -ErrorAction SilentlyContinue)) {
    Write-Host "protoc not found. Downloading and installing..."
    
    # Create temp directory
    if (-not (Test-Path $protocDir)) {
        New-Item -ItemType Directory -Path $protocDir -Force | Out-Null
    }
    
    # Download protoc if not exists
    $protocZip = "$protocDir\protoc.zip"
    if (-not (Test-Path $protocExe)) {
        Write-Host "Downloading protoc from $protocUrl"
        Invoke-WebRequest -Uri $protocUrl -OutFile $protocZip
        
        # Extract
        Write-Host "Extracting protoc..."
        Expand-Archive -Path $protocZip -DestinationPath $protocDir -Force
    }
}

# Use local protoc if available, otherwise try system protoc
$protocCmd = if (Test-Path $protocExe) { $protocExe } else { "protoc" }

# Generate Go code
Write-Host "Generating protobuf Go code..."
& $protocCmd --go_out=. --go_opt=paths=source_relative proto/binance.proto

if ($LASTEXITCODE -eq 0) {
    Write-Host "Protobuf Go code generated successfully!"
} else {
    Write-Error "Failed to generate protobuf Go code"
    exit 1
}
