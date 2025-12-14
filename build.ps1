
# Build script for glog TUI applications

Write-Host "Building glog applications..." -ForegroundColor Cyan
Write-Host ""

# Build the main menu
Write-Host "Building glog.exe (main menu)..." -ForegroundColor Yellow
go build -o glog.exe cmd/glog/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to build glog.exe!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ glog.exe built successfully" -ForegroundColor Green

# Build the create command
Write-Host "Building glog-create.exe..." -ForegroundColor Yellow
go build -o glog-create.exe cmd/create/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to build glog-create.exe!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ glog-create.exe built successfully" -ForegroundColor Green

# Build the browse command
Write-Host "Building glog-browse.exe..." -ForegroundColor Yellow
go build -o glog-browse.exe cmd/browse/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to build glog-browse.exe!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ glog-browse.exe built successfully" -ForegroundColor Green

Write-Host ""
Write-Host "✓ All builds successful!" -ForegroundColor Green
Write-Host ""
Write-Host "Run the applications with:" -ForegroundColor Yellow
Write-Host "  .\glog.exe         - Main menu (create or browse)" -ForegroundColor White
Write-Host "  .\glog-create.exe  - Create new document" -ForegroundColor White
Write-Host "  .\glog-browse.exe  - Browse documents" -ForegroundColor White
