# LinkedIn Automation Setup Script (PowerShell)
# Run this script to set up the project

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "LinkedIn Automation - Setup Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "Checking for Go installation..." -ForegroundColor Yellow
$goVersion = go version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Go is not installed!" -ForegroundColor Red
    Write-Host "Please install Go from: https://golang.org/dl/" -ForegroundColor Yellow
    Write-Host "Minimum version required: 1.21" -ForegroundColor Yellow
    exit 1
}
Write-Host "✓ Go is installed: $goVersion" -ForegroundColor Green
Write-Host ""

# Create required directories
Write-Host "Creating required directories..." -ForegroundColor Yellow
$directories = @("data", "logs")
foreach ($dir in $directories) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir | Out-Null
        Write-Host "✓ Created directory: $dir" -ForegroundColor Green
    } else {
        Write-Host "✓ Directory already exists: $dir" -ForegroundColor Green
    }
}
Write-Host ""

# Create .env file if it doesn't exist
Write-Host "Setting up environment file..." -ForegroundColor Yellow
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "✓ Created .env file from template" -ForegroundColor Green
    Write-Host "⚠️  Please edit .env and add your LinkedIn credentials!" -ForegroundColor Yellow
} else {
    Write-Host "✓ .env file already exists" -ForegroundColor Green
}
Write-Host ""

# Download dependencies
Write-Host "Downloading Go dependencies..." -ForegroundColor Yellow
go mod download
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Dependencies downloaded successfully" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to download dependencies" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Verify dependencies
Write-Host "Verifying dependencies..." -ForegroundColor Yellow
go mod verify
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Dependencies verified" -ForegroundColor Green
} else {
    Write-Host "❌ Dependency verification failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Build the application
Write-Host "Building application..." -ForegroundColor Yellow
go build -o linkedin-automation.exe ./cmd/main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Application built successfully" -ForegroundColor Green
} else {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Summary
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Edit .env file with your LinkedIn credentials" -ForegroundColor White
Write-Host "2. Review and customize config.yaml" -ForegroundColor White
Write-Host "3. Run the application: .\linkedin-automation.exe" -ForegroundColor White
Write-Host ""
Write-Host "⚠️  IMPORTANT DISCLAIMER:" -ForegroundColor Red
Write-Host "This is an educational proof-of-concept only." -ForegroundColor Yellow
Write-Host "Do NOT use for actual LinkedIn automation." -ForegroundColor Yellow
Write-Host "It may violate LinkedIn's Terms of Service." -ForegroundColor Yellow
Write-Host ""
Write-Host "For more information, see README.md" -ForegroundColor Cyan
Write-Host ""
