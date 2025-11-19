# N7AKG-UDP-Translator

Listen for UDP broadcast from HF apps, reformat and re-broadcast as UDP in N1MM format.

UDP Logger Relay is a Go application that acts as a bridge between various Ham Radio HF applications and N1MM Logger Plus. It listens for UDP broadcast messages from applications like WSJT-X, FLDigi, JS8Call, and others, then reformats these messages into N1MM-compatible XML format and re-broadcasts them via UDP.

## Features

- **Multi-format Support**: Automatically detects and parses messages from:
  - WSJT-X (FT8, FT4, MSK144, etc.)
  - FLDigi (PSK31, RTTY, etc.)
  - JS8Call
  - VarAC (VARA HF/FM digital modes)
  - N1MM Logger Plus (XML contactinfo format)
  - Generic amateur radio logging formats
- **Bi-directional N1MM Support**: Both converts TO N1MM format and accepts FROM N1MM format
- **Configurable**: Flexible configuration via YAML files or command-line options
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Verbose Logging**: Optional detailed logging for troubleshooting

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/akgordon/N7AKG-UDP-Translator/releases).

### Build from Source

1. Install Go 1.21 or later
2. Clone the repository:
   ```bash
   git clone https://github.com/akgordon/N7AKG-UDP-Translator.git
   cd N7AKG-UDP-Translator
   ```
3. Build the application:
   ```bash
   go build -o N7AKG-UDP-Translator .
   ```

## Quick Start

1. **Basic usage with default settings:**
   ```bash
   N7AKG-UDP-Translator
   ```
   This starts the relay listening on `0.0.0.0:2333` and forwarding to `127.0.0.1:12060` (N1MM default port).

2. **With custom ports:**
   ```bash
   N7AKG-UDP-Translator --listen-port 2334 --target-port 12061
   ```

3. **With verbose logging:**
   ```bash
   N7AKG-UDP-Translator --verbose
   ```

## Configuration

### Command Line Options

```
Usage:
  N7AKG-UDP-Translator [flags]

Flags:
  -c, --config string        config file (default is $HOME/.N7AKG-UDP-Translator.yaml)
      --listen-addr string   address to listen for incoming UDP messages (default "0.0.0.0")
      --listen-port int      port to listen for incoming UDP messages (default 2333)
      --source-type string   expected source message type (auto, wsjt-x, fldigi, js8call, varac, n1mm) (default "auto")
      --target-addr string   address to send reformatted UDP messages (default "127.0.0.1")
      --target-port int      port to send reformatted UDP messages (N1MM default) (default 12060)
  -v, --verbose              enable verbose logging
  -h, --help                 help for N7AKG-UDP-Translator
```

### Configuration File

Create a configuration file at `$HOME/.N7AKG-UDP-Translator.yaml`:

```yaml
# UDP Logger Relay Configuration
listen:
  address: "0.0.0.0"
  port: 2333

target:
  address: "127.0.0.1"
  port: 12060    # N1MM Logger Plus default UDP port

verbose: false

formatting:
  auto_detect: true
  source_type: "auto"  # Options: auto, wsjt-x, fldigi, js8call, etc.
  
  n1mm:
    station: "UDP-RELAY"
    operator: "OP"
    contest: "GENERAL"
```

## Usage Examples

### WSJT-X Integration

1. Configure WSJT-X to send UDP messages:
   - Go to **File** → **Settings** → **Reporting**
   - Enable "UDP Server"
   - Set UDP Server to your relay listen address (e.g., `127.0.0.1:2333`)

2. Start UDP Logger Relay:
   ```bash
   N7AKG-UDP-Translator --verbose
   ```

3. Start N1MM Logger Plus and ensure it's listening on port 12060

4. Make QSOs in WSJT-X - they should automatically appear in N1MM Logger Plus

### VarAC Integration

1. Configure VarAC to send UDP messages:
   - Go to **Settings** → **Integration** → **UDP**
   - Enable "Send UDP messages"
   - Set UDP target to your relay listen address (e.g., `127.0.0.1:2333`)
   - Enable "Send QSO data" and "Send on QSO complete"

2. Start UDP Logger Relay:
   ```bash
   N7AKG-UDP-Translator --verbose
   ```

3. Start N1MM Logger Plus and ensure it's listening on port 12060

4. Complete QSOs in VarAC - they should automatically appear in N1MM Logger Plus

### N1MM Logger Plus Integration

The relay can both receive and send messages to N1MM Logger Plus, making it useful for:
- Forwarding N1MM QSOs to other logging software
- Acting as a bridge between multiple N1MM instances
- Relaying digital mode QSOs through N1MM format

1. Configure N1MM Logger Plus UDP broadcasts:
   - Go to **Config** → **Configure Ports, Mode Control, Audio, Other**
   - In the **Broadcast Data** tab, enable "Broadcast contact info"
   - Set broadcast address to `127.0.0.1:2334` (or your relay listen port)
   - Check "When contact is logged"

2. Start UDP Logger Relay with N1MM as source:
   ```bash
   N7AKG-UDP-Translator --source-type n1mm --verbose
   ```

3. Configure destination logging software to receive N1MM-formatted UDP on port 12060

4. Log contacts in N1MM - they will be reformatted and broadcast to other applications

**Note**: When using N1MM as both source and destination, ensure different ports to avoid feedback loops.

### Custom Configuration Example

For a contest setup with specific station information:

```yaml
listen:
  address: "0.0.0.0"
  port: 2333

target:
  address: "192.168.1.100"  # N1MM computer IP
  port: 12060

verbose: true

formatting:
  auto_detect: true
  source_type: "auto"  # Options: auto, wsjt-x, fldigi, js8call, varac, n1mm
  
  n1mm:
    station: "W1AW"
    operator: "K1ABC"
    contest: "ARRL-DX-CW"
```

## Supported Input Formats

### WSJT-X
- ADIF-style UDP messages
- Automatically extracts: callsign, frequency, mode, RST reports, date/time
- Supports all WSJT-X digital modes (FT8, FT4, MSK144, etc.)

### FLDigi
- XML and text-based formats
- PSK31, RTTY, and other digital modes

### JS8Call
- JS8-specific message formats
- Heartbeat and directed messages

### VarAC
- JSON format UDP broadcasts when QSOs are completed
- Supports both VARA HF and VARA FM modes
- Automatically extracts: callsign, frequency, mode, RST reports, timestamp
- Example JSON format: `{"app":"VarAC","call":"W1ABC","freq":"14.105","mode":"VARA HF"}`
- Also supports plain text format: "QSO with W1ABC on 14.105 VARA"

### N1MM Logger Plus
- XML contactinfo format messages
- Bi-directional: accepts N1MM broadcasts and outputs N1MM-compatible XML
- Automatically extracts: callsign, frequency, mode, band, RST reports, exchange, timestamp
- Example XML format: `<contactinfo app="N1MM Logger Plus"><call>W1ABC</call><mode>CW</mode><band>20m</band></contactinfo>`
- Useful for relay chains and multi-station setups

### Generic Format
- Attempts to parse any message containing:
  - Valid amateur radio callsigns
  - Frequency information
  - Mode information
  - Band designations

## N1MM Logger Plus Setup

1. **Enable UDP listening in N1MM:**
   - Go to **Config** → **Configure Ports, Mode Control, Audio, Other**
   - Enable "Accept UDP broadcast information on port"
   - Set port to 12060 (or match your relay target port)

2. **Contest Setup:**
   - Configure your contest in N1MM as usual
   - The relay will send QSO data that N1MM can log automatically

## Troubleshooting

### Common Issues

1. **No messages received:**
   - Check that your HF application is configured to send UDP broadcasts
   - Verify the listen address and port match your HF app settings
   - Use `--verbose` flag to see incoming messages

2. **Messages not reaching N1MM:**
   - Verify N1MM is listening on the target port (default 12060)
   - Check firewall settings
   - Ensure target address is correct

3. **Parsing errors:**
   - Use `--verbose` to see parsing details
   - Check if your HF app sends a supported format
   - Try different `source_type` settings in configuration

### Debug Commands

```bash
# Show version information
N7AKG-UDP-Translator version

# Test with maximum verbosity
N7AKG-UDP-Translator --verbose --listen-port 2333 --target-port 12060

# Use custom config file
N7AKG-UDP-Translator --config /path/to/config.yaml --verbose
```

### Network Testing

Test UDP connectivity:
```bash
# Listen for incoming messages (Linux/macOS)
nc -u -l 2333

# Send test message (Linux/macOS)  
echo "test message" | nc -u 127.0.0.1 2333
```

## Examples

The `examples/` directory contains demonstration programs that show how the relay works with different message formats.

### VarAC Message Format Demo

Test VarAC message detection and parsing:

```bash
# Run the VarAC demo
go run examples/varac_demo.go
```

This demo shows:
- Full JSON format VarAC messages
- Plain text VarAC messages
- Minimal JSON formats
- VARA FM examples
- How messages are detected, parsed, and formatted for N1MM

The demo is useful for:
- Testing your VarAC configuration
- Understanding supported message formats
- Troubleshooting parsing issues
- Learning how to use the formatter package

## Development

### Building

All build artifacts are organized into the `output/` directory.

**Using Make (Linux/macOS):**
```bash
# Build for current platform
make build

# Build for all platforms  
make build-all

# Run tests with coverage
make test-coverage

# Clean build artifacts
make clean

# See all available targets
make help
```

**Using PowerShell script (Windows):**
```powershell
# Build for current platform
.\build.ps1 build

# Build for all platforms
.\build.ps1 build-all

# Run tests with coverage
.\build.ps1 test-coverage

# Clean build artifacts  
.\build.ps1 clean

# See all available targets
.\build.ps1 help
```

**Manual building:**
```bash
# Create output directory
mkdir output

# Build for current platform
go build -ldflags "-X main.version=v1.0.0" -o output/N7AKG-UDP-Translator .

# Build for multiple platforms
GOOS=windows GOARCH=amd64 go build -o output/N7AKG-UDP-Translator-windows.exe .
GOOS=linux GOARCH=amd64 go build -o output/N7AKG-UDP-Translator-linux .
GOOS=darwin GOARCH=amd64 go build -o output/N7AKG-UDP-Translator-macos .
```

**Output Directory Structure:**
```
output/
├── N7AKG-UDP-Translator.exe          # Windows build
├── N7AKG-UDP-Translator-linux        # Linux build  
├── N7AKG-UDP-Translator-macos        # macOS build
├── coverage.out                  # Test coverage data
└── coverage.html                 # Test coverage report
```

### Testing

```bash
# Run tests
go test ./...

# Run with race detection
go test -race ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- **Issues**: Report bugs and request features on [GitHub Issues](https://github.com/akgordon/N7AKG-UDP-Translator/issues)
- **Discussions**: Join the conversation in [GitHub Discussions](https://github.com/akgordon/N7AKG-UDP-Translator/discussions)

## Acknowledgments

- N1MM Logger Plus team for the excellent logging software
- WSJT-X developers for the digital mode innovations
- Ham radio community for testing and feedback
