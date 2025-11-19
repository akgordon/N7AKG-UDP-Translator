# UDP Logger Relay Testing Script
# This script helps test the UDP Logger Relay with the WSJT-X simulator

Write-Host "UDP Logger Relay - Testing Setup" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

# Check if the relay executable exists
if (-not (Test-Path ".\N7AKG-UDP-Translator.exe")) {
    Write-Host "Building UDP Logger Relay..." -ForegroundColor Yellow
    go build -o N7AKG-UDP-Translator.exe .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to build relay. Please check for errors." -ForegroundColor Red
        exit 1
    }
}

# Check if the simulator exists
if (-not (Test-Path ".\tools\wsjtx_simulator.exe")) {
    Write-Host "Building WSJT-X Simulator..." -ForegroundColor Yellow
    go build -o tools\wsjtx_simulator.exe tools\wsjtx_simulator.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to build simulator. Please check for errors." -ForegroundColor Red
        exit 1
    }
}

Write-Host "All binaries are ready!" -ForegroundColor Green
Write-Host ""
Write-Host "To test the UDP Logger Relay:" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. In Terminal 1, start the relay in verbose mode:" -ForegroundColor White
Write-Host "   .\N7AKG-UDP-Translator.exe --verbose" -ForegroundColor Gray
Write-Host ""
Write-Host "2. In Terminal 2, run the simulator:" -ForegroundColor White
Write-Host "   .\tools\wsjtx_simulator.exe" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Watch the relay terminal for incoming messages" -ForegroundColor White
Write-Host ""
Write-Host "Additional simulator options:" -ForegroundColor Cyan
Write-Host "  -interval 3        # Send messages every 3 seconds" -ForegroundColor Gray
Write-Host "  -callsign W1ABC   # Use your callsign" -ForegroundColor Gray
Write-Host "  -mode FT4         # Simulate FT4 instead of FT8" -ForegroundColor Gray
Write-Host "  -port 2334        # Send to a different port" -ForegroundColor Gray
Write-Host ""

# Offer to start the relay
$response = Read-Host "Would you like to start the relay now? (y/n)"
if ($response -eq "y" -or $response -eq "Y") {
    Write-Host ""
    Write-Host "Starting UDP Logger Relay in verbose mode..." -ForegroundColor Green
    Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
    Write-Host ""
    .\N7AKG-UDP-Translator.exe --verbose
}
