# UDP-Logger-Relay

Listen for UDP broadcast from HF apps, reformat and re-broadcast as UDP in N1MM format.

UDP Logger Relay is a Go application that acts as a bridge between various Ham Radio HF applications and N1MM Logger Plus. It listens for UDP broadcast messages from applications like WSJT-X, FLDigi, JS8Call, and others, then reformats these messages into N1MM-compatible XML format and re-broadcasts them via UDP.

## Features

- **Multi-format Support**: Automatically detects and parses messages from:
  - WSJT-X (FT8, FT4, MSK144, etc.)
  - FLDigi (PSK31, RTTY, etc.)
  - JS8Call
  - VarAC (VARA HF/FM digital modes)
  - Generic amateur radio logging formats
- **N1MM Integration**: Converts QSO data to N1MM Logger Plus XML format
- **Configurable**: Flexible configuration via YAML files or command-line options
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Verbose Logging**: Optional detailed logging for troubleshooting

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/akgordon/UDP-Logger-Relay/releases).

### Build from Source

1. Install Go 1.21 or later
2. Clone the repository:
   ```bash
   git clone https://github.com/akgordon/UDP-Logger-Relay.git
   cd UDP-Logger-Relay
   ```
3. Build the application:
   ```bash
   go build -o udp-logger-relay .
   ```

## Quick Start

1. **Basic usage with default settings:**
   ```bash
   udp-logger-relay
   ```
   This starts the relay listening on `0.0.0.0:2333` and forwarding to `127.0.0.1:12060` (N1MM default port).

2. **With custom ports:**
   ```bash
   udp-logger-relay --listen-port 2334 --target-port 12061
   ```

3. **With verbose logging:**
   ```bash
   udp-logger-relay --verbose
   ```

## Configuration

### Command Line Options

```
Usage:
  udp-logger-relay [flags]

Flags:
  -c, --config string        config file (default is $HOME/.udp-logger-relay.yaml)
      --listen-addr string   address to listen for incoming UDP messages (default "0.0.0.0")
      --listen-port int      port to listen for incoming UDP messages (default 2333)
      --target-addr string   address to send reformatted UDP messages (default "127.0.0.1")
      --target-port int      port to send reformatted UDP messages (N1MM default) (default 12060)
  -v, --verbose              enable verbose logging
  -h, --help                 help for udp-logger-relay
```

### Configuration File

Create a configuration file at `$HOME/.udp-logger-relay.yaml`:

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
   udp-logger-relay --verbose
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
   udp-logger-relay --verbose
   ```

3. Start N1MM Logger Plus and ensure it's listening on port 12060

4. Complete QSOs in VarAC - they should automatically appear in N1MM Logger Plus

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
  source_type: "auto"  # Options: auto, wsjt-x, fldigi, js8call, varac
  
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
udp-logger-relay version

# Test with maximum verbosity
udp-logger-relay --verbose --listen-port 2333 --target-port 12060

# Use custom config file
udp-logger-relay --config /path/to/config.yaml --verbose
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

```bash
# Build for current platform
go build -o udp-logger-relay .

# Build for multiple platforms
GOOS=windows GOARCH=amd64 go build -o udp-logger-relay-windows.exe .
GOOS=linux GOARCH=amd64 go build -o udp-logger-relay-linux .
GOOS=darwin GOARCH=amd64 go build -o udp-logger-relay-macos .
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

- **Issues**: Report bugs and request features on [GitHub Issues](https://github.com/akgordon/UDP-Logger-Relay/issues)
- **Discussions**: Join the conversation in [GitHub Discussions](https://github.com/akgordon/UDP-Logger-Relay/discussions)

## Acknowledgments

- N1MM Logger Plus team for the excellent logging software
- WSJT-X developers for the digital mode innovations
- Ham radio community for testing and feedback
