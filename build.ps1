# Build script for UDP Logger Relay (Windows PowerShell)
# This is an alternative to the Makefile for Windows users

param(
    [string]$Target = "build",
    [string]$Version = "dev",
    [string]$OutputDir = "output"
)

# Get version info
if ($Version -eq "dev") {
    try {
        $Version = git describe --tags --always --dirty 2>$null
        if (-not $Version) { $Version = "dev" }
    } catch {
        $Version = "dev"
    }
}

try {
    $Commit = git rev-parse --short HEAD 2>$null
    if (-not $Commit) { $Commit = "none" }
} catch {
    $Commit = "none"
}

$Date = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$LdFlags = "-X main.version=$Version -X main.commit=$Commit -X main.date=$Date"

# Ensure output directory exists
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    Write-Host "Created output directory: $OutputDir"
}

function Invoke-Build {
    param([string]$Os, [string]$Arch, [string]$Output)
    
    Write-Host "Building for $Os/$Arch -> $Output"
    $env:GOOS = $Os
    $env:GOARCH = $Arch
    
    go build -ldflags $LdFlags -o "$OutputDir/$Output" .
    
    # Clear environment variables
    $env:GOOS = $null
    $env:GOARCH = $null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Build successful: $OutputDir/$Output"
    } else {
        Write-Host "✗ Build failed for $Os/$Arch"
        exit 1
    }
}

function Show-Help {
    Write-Host "UDP Logger Relay Build Script"
    Write-Host "Usage: .\build.ps1 [-Target <target>] [-Version <version>] [-OutputDir <dir>]"
    Write-Host ""
    Write-Host "Available targets:"
    Write-Host "  build            Build for current platform"
    Write-Host "  build-all        Build for Windows, Linux, and macOS"
    Write-Host "  build-windows    Build for Windows"
    Write-Host "  build-linux      Build for Linux" 
    Write-Host "  build-macos      Build for macOS"
    Write-Host "  test             Run tests"
    Write-Host "  test-coverage    Run tests with coverage"
    Write-Host "  clean            Remove build artifacts"
    Write-Host "  run              Run application with verbose logging"
    Write-Host "  run-example      Run VarAC demo example"
    Write-Host "  help             Show this help"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\build.ps1 build"
    Write-Host "  .\build.ps1 -Target build-all -Version v1.0.0"
    Write-Host "  .\build.ps1 test-coverage"
}

switch ($Target.ToLower()) {
    "build" {
        Invoke-Build "" "" "N7AKG-UDP-Translator.exe"
    }
    
    "build-windows" {
        Invoke-Build "windows" "amd64" "N7AKG-UDP-Translator-windows.exe"
    }
    
    "build-linux" {
        Invoke-Build "linux" "amd64" "N7AKG-UDP-Translator-linux"
    }
    
    "build-macos" {
        Invoke-Build "darwin" "amd64" "N7AKG-UDP-Translator-macos"
    }
    
    "build-all" {
        Invoke-Build "windows" "amd64" "N7AKG-UDP-Translator-windows.exe"
        Invoke-Build "linux" "amd64" "N7AKG-UDP-Translator-linux"
        Invoke-Build "darwin" "amd64" "N7AKG-UDP-Translator-macos"
        Write-Host "✓ All builds completed in $OutputDir/"
    }
    
    "test" {
        Write-Host "Running tests..."
        go test -v ./...
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ All tests passed"
        } else {
            Write-Host "✗ Some tests failed"
            exit 1
        }
    }
    
    "test-coverage" {
        Write-Host "Running tests with coverage..."
        go test -v -coverprofile="$OutputDir/coverage.out" ./...
        if ($LASTEXITCODE -eq 0) {
            go tool cover -html="$OutputDir/coverage.out" -o "$OutputDir/coverage.html"
            Write-Host "✓ Coverage report generated: $OutputDir/coverage.html"
        } else {
            Write-Host "✗ Tests failed"
            exit 1
        }
    }
    
    "clean" {
        if (Test-Path $OutputDir) {
            Remove-Item -Recurse -Force $OutputDir
            Write-Host "✓ Cleaned $OutputDir/"
        } else {
            Write-Host "Nothing to clean - $OutputDir/ doesn't exist"
        }
    }
    
    "run" {
        Write-Host "Running UDP Logger Relay with verbose logging..."
        go run . --verbose
    }
    
    "run-example" {
        Write-Host "Running VarAC demo..."
        go run examples/varac_demo.go
    }
    
    "help" {
        Show-Help
    }
    
    default {
        Write-Host "Unknown target: $Target"
        Write-Host "Use '.\build.ps1 help' to see available targets"
        exit 1
    }
}